package j8a

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/hako/durafmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"math"
	"math/big"
	"strings"
	"time"
)

type PDuration time.Duration

func (p PDuration) AsString() string {
	return durafmt.Parse(time.Duration(p)).LimitFirstN(2).String()
}

func (p PDuration) AsDuration() time.Duration {
	return time.Duration(p)
}

func (p PDuration) AsDays() int {
	return int(p.AsDuration().Hours() / 24)
}

type TlsLink struct {
	cert              *x509.Certificate
	issued            time.Time
	remainingValidity PDuration
	totalValidity     PDuration
	browserValidity   PDuration
	earliestExpiry    bool
	isCA              bool
}

func (t TlsLink) browserExpiry() PDuration {
	return PDuration(time.Hour * 24 * 398)
}

func (t TlsLink) printRemainingValidity() string {
	rv := t.remainingValidity.AsString()
	if t.earliestExpiry {
		rv = rv + ", which is the earliest in your chain"
	}
	return rv
}

func tlsHealthCheck(conf *tls.Config, daemon bool) {
	defer func() {
		if r := recover(); r != nil {
			log.Trace().Msgf("TLS cert logProcStats not analysed, root cause: %s", r)
		}
	}()

	//safety first
	if conf != nil && len(conf.Certificates) > 0 {
	Daemon:
		for {
			//Andeka is checking our certificate chains forever.
			andeka, _ := checkCertChain(conf.Certificates[0])
			logCertStats(andeka)
			if daemon {
				time.Sleep(time.Hour * 24)
			} else {
				break Daemon
			}
		}
	}
}

func checkCertChain(chain tls.Certificate) ([]TlsLink, error) {
	var tlsLinks []TlsLink
	var err error
	cert, e1 := x509.ParseCertificate(chain.Certificate[0])
	if e1 != nil {
		err = e1
	}
	if cert.DNSNames == nil || len(cert.DNSNames) == 0 {
		err = errors.New("no DNS name specified")
	}
	verified, e2 := cert.Verify(verifyOptions(splitCertPools(chain)))
	if e2 != nil {
		err = e2
	}
	if err == nil {
		tlsLinks = parseTlsLinks(verified[0])
	}
	return tlsLinks, err
}

func verifyOptions(inter *x509.CertPool, root *x509.CertPool) x509.VerifyOptions {
	opts := x509.VerifyOptions{}
	if inter != nil && len(inter.Subjects()) > 0 {
		opts.Intermediates = inter
	}
	if root != nil && len(root.Subjects()) > 0 {
		opts.Roots = root
	}
	return opts
}

func formatSerial(serial *big.Int) string {
	hex := fmt.Sprintf("%X", serial)
	if len(hex) == 1 || len(hex) == 31 {
		hex = "0" + hex
	}
	if len(hex) > 2 {
		frm := strings.Builder{}
		for i := 0; i < len(hex); i += 2 {
			frm.WriteString(hex[i : i+2])
			if i != len(hex)-2 {
				frm.WriteString("-")
			}
		}
		hex = frm.String()
	}
	return hex
}

func logCertStats(tlsLinks []TlsLink) {
	month := PDuration(time.Hour * 24 * 30)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Snapshot of your cert chain size %d explained. ", len(tlsLinks)))
	for i, link := range tlsLinks {
		if !link.isCA {
			sb.WriteString(fmt.Sprintf("[%d/%d] TLS cert #%s for DNS names %s, issued on %s, signed by [%s], expires in %s. ",
				i+1,
				len(tlsLinks),
				formatSerial(link.cert.SerialNumber),
				link.cert.DNSNames,
				link.issued.Format("2006-01-02"),
				link.cert.Issuer.CommonName,
				link.printRemainingValidity(),
			))
			if link.totalValidity > link.browserExpiry() {
				sb.WriteString(fmt.Sprintf("Total validity period of %d days is above legal browser max %d. ",
					int(link.totalValidity.AsDays()),
					int(link.browserExpiry().AsDays())))
			}
			if link.browserValidity > 0 {
				sb.WriteString(fmt.Sprintf("You may experience disruption in %s. ",
					link.browserValidity.AsString()))
			} else {
				sb.WriteString(fmt.Sprintf("Validity grace period expired %s ago, update now. ",
					link.browserValidity.AsString()))
			}
		} else {
			caType := "Intermediate"
			if isRoot(link.cert) {
				caType = "Root"
			}
			sb.WriteString(fmt.Sprintf("[%d/%d] %s CA #%s Common name [%s], signed by [%s], expires in %s. ",
				i+1,
				len(tlsLinks),
				caType,
				formatSerial(link.cert.SerialNumber),
				link.cert.Subject.CommonName,
				link.cert.Issuer.CommonName,
				link.remainingValidity.AsString(),
			))
		}
	}

	for _, t := range tlsLinks {
		if t.earliestExpiry {
			var ev *zerolog.Event
			//if the certificate expires in less than 30 days we send this as a log.Warn event instead.
			if t.remainingValidity < month {
				ev = log.Warn()
			} else {
				ev = log.Debug()
			}
			ev.Msg(sb.String())
		}
	}
}

func parseTlsLinks(chain []*x509.Certificate) []TlsLink {
	earliestExpiry := PDuration(math.MaxInt64)
	browserExpiry := TlsLink{}.browserExpiry().AsDuration()
	var tlsLinks []TlsLink
	si := 0
	for i, cert := range chain {
		link := TlsLink{
			cert:              cert,
			issued:            cert.NotBefore,
			remainingValidity: PDuration(time.Until(cert.NotAfter)),
			totalValidity:     PDuration(cert.NotAfter.Sub(cert.NotBefore)),
			browserValidity:   PDuration(time.Until(cert.NotBefore.Add(browserExpiry))),
			earliestExpiry:    false,
			isCA:              cert.IsCA,
		}
		tlsLinks = append(tlsLinks, link)
		if link.remainingValidity < earliestExpiry {
			si = i
			earliestExpiry = link.remainingValidity
		}
	}
	tlsLinks[si].earliestExpiry = true
	return tlsLinks
}

func splitCertPools(chain tls.Certificate) (*x509.CertPool, *x509.CertPool) {
	root := x509.NewCertPool()
	inter := x509.NewCertPool()
	for _, c := range chain.Certificate {
		c1, _ := x509.ParseCertificate(c)
		//for CA's we treat you as intermediate unless you signed yourself
		if c1.IsCA {
			//as above, you're intermediate in the last position unless you signed yourself, that makes you a root cert.
			if isRoot(c1) {
				root.AddCert(c1)
			} else {
				inter.AddCert(c1)
			}
		}
	}
	return inter, root
}

func isRoot(c *x509.Certificate) bool {
	//TODO: this seems to work but should we really check signature instead?
	return c.IsCA && c.Issuer.CommonName == c.Subject.CommonName
}
