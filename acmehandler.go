package j8a

import (
	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/rs/zerolog/log"
	"net/http"
)

type AcmeHandler struct {
	Active bool
	Domain string
	Token string
	KeyAuth []byte
}

func (a AcmeHandler) Present(domain, token, keyAuth string) error {
	a.Active = true
	a.Domain = domain
	a.Token = token
	a.KeyAuth = []byte(keyAuth)

	return nil
}
const notConfigured = "not configured"
func (a AcmeHandler) CleanUp(domain, token, keyAuth string) error {
	a.Active  = false
	a.Domain = notConfigured
	a.Token = notConfigured
	a.KeyAuth = []byte(notConfigured)

	return nil
}

const acmeEvent = "responded to remote ACME challenge for %s, with %s"
func acmeHandler(w http.ResponseWriter, r *http.Request) {
	a := Runner.AcmeHandler
	path := http01.ChallengePath(a.Token)
	w.WriteHeader(200)
	w.Write([]byte(a.KeyAuth))
	log.Debug().Msgf(acmeEvent, path, a.KeyAuth)
}
