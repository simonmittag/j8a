package jabba

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/hako/durafmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"math"
	"strings"
	"time"
)

type TlsLink struct {
	cert           *x509.Certificate
	expiryDuration time.Duration
	expiryString   string
	shortestExpiry bool
	isCA           bool
}

func logCertificateStats(chain tls.Certificate) {
	monthDuration := time.Hour * 24 * 30
	majorBrowserExpiryDuration := time.Hour * 24 * 398
	root, inter := splitCertPools(chain)

	var tlsLinks []TlsLink

	var shortest time.Duration = math.MaxInt64
	si := 0
	for i, c := range chain.Certificate {
		cert, _ := x509.ParseCertificate(c)
		link := TlsLink{
			cert:           cert,
			expiryDuration: time.Until(cert.NotAfter),
			expiryString:   durafmt.Parse(time.Until(cert.NotAfter)).LimitFirstN(3).String(),
			shortestExpiry: false,
			isCA:           cert.IsCA,
		}
		tlsLinks = append(tlsLinks, link)
		if link.expiryDuration < shortest {
			si = i
			shortest = link.expiryDuration
		}
	}
	tlsLinks[si].shortestExpiry = true
	tlsLinks[si].expiryString += " which is the earliest expiry period in your chain"

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Your certificate chain size %d explained. ", len(chain.Certificate)))
	for i, link := range tlsLinks {
		if !link.isCA {
			link.cert.Verify(x509.VerifyOptions{
				Intermediates: inter,
				Roots:         root})
			sb.WriteString(fmt.Sprintf("TLS certificate (%d/%d) for DNS names %s, Common name [%s], signed by issuer [%s], expires in %s. ",
				i+1,
				len(chain.Certificate),
				link.cert.DNSNames,
				link.cert.Subject.CommonName,
				link.cert.Issuer.CommonName,
				link.expiryString,
			))
			if link.expiryDuration > majorBrowserExpiryDuration {
				sb.WriteString(fmt.Sprintf("Note this is above 398 day threshold accepted by major browsers. "))
			}
		} else {
			caType := "Intermediate"
			if link.cert.Issuer.CommonName == link.cert.Subject.CommonName {
				caType = "Root"
			}
			sb.WriteString(fmt.Sprintf("%s CA (%d/%d) for Common name [%s], signed by issuer [%s], expires in %s. ",
				caType,
				i+1,
				len(chain.Certificate),
				link.cert.Subject.CommonName,
				link.cert.Issuer.CommonName,
				link.expiryString,
			))
		}
	}

	var ev *zerolog.Event
	//if the certificate expires in less than 30 days we send this as a log.Warn event instead.
	if shortest < monthDuration {
		ev = log.Warn()
	} else {
		ev = log.Debug()
	}

	ev.Msg(sb.String())
}

func splitCertPools(chain tls.Certificate) (*x509.CertPool, *x509.CertPool) {
	root := x509.NewCertPool()
	inter := x509.NewCertPool()
	for i, c := range chain.Certificate {
		if i > 0 && i < len(chain.Certificate)-1 {
			c1, _ := x509.ParseCertificate(c)
			inter.AddCert(c1)
		}
		if i == len(chain.Certificate)-1 {
			c1, _ := x509.ParseCertificate(c)
			root.AddCert(c1)
		}
	}
	return root, inter
}
