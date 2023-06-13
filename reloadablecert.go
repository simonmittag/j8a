package j8a

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/rs/zerolog/log"
	"sync"
)

type ReloadableCert struct {
	Cert *tls.Certificate
	mu   sync.Mutex
	Init bool
	//required to use runtime internally without global pointer for testing.
	runtime *Runtime
}

func (r *ReloadableCert) GetCertificateFunc(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	return r.Cert, nil
}

func (r *ReloadableCert) triggerInit() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Init = true

	var cert tls.Certificate
	var err error

	c := []byte(r.runtime.Connection.Downstream.Tls.Cert)
	k := []byte(r.runtime.Connection.Downstream.Tls.Key)

	cert, err = tls.X509KeyPair(c, k)
	if err == nil {
		r.Cert = &cert
		r.Cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	}
	if err == nil {
		log.Info().Msgf("TLS certificate #%v initialized", formatSerial(cert.Leaf.SerialNumber))
	}
	r.Init = false
	return err
}
