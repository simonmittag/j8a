package j8a

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
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

const Days30 = time.Duration(time.Hour * 24 * 30)
const Days398 = PDuration(time.Hour * 24 * 398)

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

func (t TlsLink) expiresTooCloseForComfort() bool {
	return time.Duration(t.remainingValidity) <= Days30
}

func (t TlsLink) expiryLongerThanLegalBrowserMaximum() bool {
	return t.browserValidity < t.remainingValidity
}

func (t TlsLink) legalBrowserValidityPeriodPassed() bool {
	return t.browserValidity < 0
}

func (t TlsLink) printRemainingValidity() string {
	rv := t.remainingValidity.AsString()
	if t.earliestExpiry {
		rv = rv + ", which is the earliest in your chain"
	}
	return rv
}

func (r *Runtime) tlsHealthCheck(daemon bool) {
	defer func() {
		if r := recover(); r != nil {
			log.Trace().Msgf("TLS cert not analysed, cause: %s", r)
		}
	}()

	//safety first
	if r.ReloadableCert.Cert != nil {
	Daemon:
		for {
			//Andeka is checking our certificate chains forever.
			andeka, _ := checkFullCertChain(r.ReloadableCert.Cert)
			logCertStats(andeka)
			if andeka[0].expiresTooCloseForComfort() {
				r.renewAcmeCertAndKey()
			}

			if daemon {
				time.Sleep(time.Hour * 24)
			} else {
				break Daemon
			}
		}
	}
}

const acmeRetry24h = "unable to renew ACME certificate from provider %s, cause: %s, will retry in 24h"

func (r *Runtime) renewAcmeCertAndKey() error {
	p := r.Connection.Downstream.Tls.Acme.Provider
	log.Debug().Msgf("triggering renewal of ACME certificate from provider %s ", p)

	e1 := r.fetchAcmeCertAndKey(acmeProviders[p])
	if e1 == nil {
		c := []byte(r.Connection.Downstream.Tls.Cert)
		k := []byte(r.Connection.Downstream.Tls.Key)

		if newCerts, e2 := checkFullCertChainFromBytes(c, k); e2 != nil {
			log.Warn().Msgf(acmeRetry24h, p, e2)
			return e2
		} else {
			//if no issues, cache the cert and key. we don't assert whether this works it only matters when loading.
			r.cacheAcmeCertAndKey(acmeProviders[p])

			//now trigger a re-init of TLS cert for the cert we just downloaded.
			e3 := r.ReloadableCert.triggerInit()
			if e3 == nil {
				logCertStats(newCerts)
				log.Debug().Msgf("successful renewal of ACME certificate from provider %s complete", p)
			} else {
				log.Warn().Msgf(acmeRetry24h, p, e3)
				return e3
			}
		}
	}
	return nil
}

func checkFullCertChainFromBytes(cert []byte, key []byte) ([]TlsLink, error) {
	var chain tls.Certificate

	var e1 error
	chain, e1 = tls.X509KeyPair(cert, key)
	if e1 != nil {
		return nil, e1
	}
	return checkFullCertChain(&chain)
}

func checkFullCertChain(chain *tls.Certificate) ([]TlsLink, error) {
	if len(chain.Certificate) == 0 {
		return nil, errors.New("no certificate data found")
	}

	var e2 error
	chain.Leaf, e2 = x509.ParseCertificate(chain.Certificate[0])
	if e2 != nil {
		return nil, e2
	}

	if chain.Leaf.DNSNames == nil || len(chain.Leaf.DNSNames) == 0 {
		return nil, errors.New("no DNS name specified")
	}

	inter, root, e3 := splitCertPools(chain)
	if e3 != nil {
		return nil, e3
	}

	verified, e4 := chain.Leaf.Verify(verifyOptions(inter, root))
	if e4 != nil {
		return nil, e4
	}

	return parseTlsLinks(verified[0]), nil
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
	serial = serial.Abs(serial)
	hex := fmt.Sprintf("%X", serial)
	if len(hex)%2 != 0 {
		hex = "0" + hex
	}
	if len(hex) > 2 {
		frm := strings.Builder{}
		for i := 0; i < len(hex); i += 2 {
			var j = 0
			if i+2 <= len(hex) {
				j = i + 2
			} else {
				j = len(hex)
			}
			w := hex[i:j]
			frm.WriteString(w)
			if i < len(hex)-2 {
				frm.WriteString(":")
			}
		}
		hex = frm.String()
	}
	return hex
}

func sha1Fingerprint(cert *x509.Certificate) string {
	sha1 := sha1.Sum(cert.Raw)
	return "#" + JoinHashString(sha1[:])
}

func sha256Fingerprint(cert *x509.Certificate) string {
	sha256 := sha256.Sum256(cert.Raw)
	return "#" + JoinHashString(sha256[:])
}

func md5Fingerprint(cert *x509.Certificate) string {
	md5 := md5.Sum(cert.Raw)
	return "#" + JoinHashString(md5[:])
}

func JoinHashString(hash []byte) string {
	return strings.Join(ChunkString(strings.ToUpper(hex.EncodeToString(hash[:])), 2), ":")
}

func ChunkString(s string, chunkSize int) []string {
	var chunks []string
	runes := []rune(s)

	if len(runes) == 0 {
		return []string{s}
	}

	for i := 0; i < len(runes); i += chunkSize {
		nn := i + chunkSize
		if nn > len(runes) {
			nn = len(runes)
		}
		chunks = append(chunks, string(runes[i:nn]))
	}
	return chunks
}

func logCertStats(tlsLinks []TlsLink) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Snapshot of your cert chain size %d explained. ", len(tlsLinks)))
	for i, link := range tlsLinks {
		//we only want to know the below for certs, not CAs
		if !link.isCA {
			sb.WriteString(fmt.Sprintf("[%d/%d] TLS cert serial #%s, sha1 fingerprint %s for DNS names %s, valid from %s, signed by [%s], expires in %s. ",
				i+1,
				len(tlsLinks),
				formatSerial(link.cert.SerialNumber),
				sha1Fingerprint(link.cert),
				link.cert.DNSNames,
				link.issued.Format("2006-01-02"),
				link.cert.Issuer.CommonName,
				link.printRemainingValidity(),
			))
			//is this cert valid longer than 398d? tell the admin
			if link.expiryLongerThanLegalBrowserMaximum() {
				sb.WriteString(fmt.Sprintf("Total validity period of %d days is above legal browser period of %d days. ",
					int(link.totalValidity.AsDays()),
					int(Days398.AsDays())))
			}
			//has browser validity passed?
			if link.legalBrowserValidityPeriodPassed() {
				sb.WriteString(fmt.Sprintf("Legal browser period already expired %s ago, update this certificate now. ",
					link.browserValidity.AsString()))
				//if it hasn't warn the user if it's a long-lived cert.
			} else if link.expiryLongerThanLegalBrowserMaximum() {
				sb.WriteString(fmt.Sprintf("Despite valid certificate, You may experience disruption in %s. ",
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
			//if the certificate shows signs of problems but is valid, log.warn instead
			if t.expiresTooCloseForComfort() || t.legalBrowserValidityPeriodPassed() {
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
	var tlsLinks []TlsLink
	si := 0
	for i, cert := range chain {
		link := TlsLink{
			cert:              cert,
			issued:            cert.NotBefore,
			remainingValidity: PDuration(time.Until(cert.NotAfter)),
			totalValidity:     PDuration(cert.NotAfter.Sub(cert.NotBefore)),
			browserValidity:   PDuration(time.Until(cert.NotBefore.Add(Days398.AsDuration()))),
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

func splitCertPools(chain *tls.Certificate) (*x509.CertPool, *x509.CertPool, error) {
	var err error

	root := x509.NewCertPool()
	inter := x509.NewCertPool()
	for _, c := range chain.Certificate {
		c1, caerr := x509.ParseCertificate(c)
		if caerr != nil {
			err = caerr
		}
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
	return inter, root, err
}

func isRoot(c *x509.Certificate) bool {
	return c.IsCA && c.Issuer.CommonName == c.Subject.CommonName
}
