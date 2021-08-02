package j8a

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/simonmittag/lego/v4/certcrypto"
	"github.com/simonmittag/lego/v4/certificate"
	"github.com/simonmittag/lego/v4/challenge/http01"
	"github.com/simonmittag/lego/v4/lego"
	"github.com/simonmittag/lego/v4/registration"
	"net/http"
	"strings"
)

const acmeChallenge = "/.well-known/acme-challenge/"

var acmeProviders = map[string]string{
	"letsencrypt":  "https://acme-v02.api.letsencrypt.org/directory",
	"let'sencrypt": "https://acme-v02.api.letsencrypt.org/directory",
	//"letsencrypt":  "https://acme-staging-v02.api.letsencrypt.org/directory",
	//"let'sencrypt": "https://acme-staging-v02.api.letsencrypt.org/directory",
}

type AcmeHandler struct {
	Active   map[string]bool
	Domains  map[string]string
	KeyAuths map[string][]byte
}

func NewAcmeHandler() *AcmeHandler {
	return &AcmeHandler{
		Active:   make(map[string]bool),
		Domains:  make(map[string]string),
		KeyAuths: make(map[string][]byte),
	}
}

func (a *AcmeHandler) Present(domain, token, keyAuth string) error {
	a.Active[token] = true
	a.Domains[token] = domain
	a.KeyAuths[token] = []byte(keyAuth)

	log.Debug().Msgf("ACME handler for domain %s activated, ready to serve challenge response for token %s.", domain, token)
	return nil
}

func (a *AcmeHandler) CleanUp(domain, token, keyAuth string) error {
	delete(a.Active, token)
	delete(a.Domains, token)
	delete(a.KeyAuths, token)

	log.Debug().Msgf("ACME handler for domain %s deactivated.", domain)
	return nil
}

func (a *AcmeHandler) isActive() bool {
	var c = false
	for _, v := range a.Active {
		c = c || v
	}
	return c
}

const acmeEvent = "responded to remote ACME challenge path %s, with %s for domain %s"

func acmeHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r1 := recover(); r1 != nil {
			log.Debug().Msgf("unsuccessful response to remote ACME challenge, URI %s, cause: %v", r.RequestURI, r1)
		}
	}()

	tokens := strings.Split(r.RequestURI, "/")
	token := tokens[len(tokens)-1]

	a := Runner.AcmeHandler
	path := http01.ChallengePath(token)
	w.WriteHeader(200)
	w.Write([]byte(a.KeyAuths[token]))
	log.Debug().Msgf(acmeEvent, path, a.KeyAuths[token], a.Domains[token])
}

type AcmeUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *AcmeUser) GetEmail() string {
	return u.Email
}
func (u AcmeUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *AcmeUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func (runtime *Runtime) fetchAcmeCertAndKey(url string) error {
	defer func() {
		if r := recover(); r != nil {
			msg := fmt.Sprintf("TLS ACME certificate not fetched, cause: %s", r)
			log.Debug().Msg(msg)
		}
	}()

	var e error
	var pk *ecdsa.PrivateKey

	pk, e = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if e != nil {
		return e
	}

	myUser := AcmeUser{
		Email: runtime.Connection.Downstream.Tls.Acme.Email,
		key:   pk,
	}

	config := lego.NewConfig(&myUser)
	config.Certificate.KeyType = certcrypto.RSA2048
	var client *lego.Client
	client, e = lego.NewClient(config)
	if e != nil {
		return e
	}

	e = client.Challenge.SetHTTP01Provider(runtime.AcmeHandler)
	if e != nil {
		return e
	}

	//we always register because there is no way to save state.
	myUser.Registration, e = client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})

	request := certificate.ObtainRequest{
		Domains: Runner.Connection.Downstream.Tls.Acme.Domains,
		Bundle:  true,
	}

	var c *certificate.Resource
	c, e = client.Certificate.Obtain(request)
	if e != nil {
		log.Debug().Msgf("ACME certificate from %s unsuccessful, cause %v", url, e)
		return e
	}

	runtime.Connection.Downstream.Tls.Cert = string(c.Certificate)
	runtime.Connection.Downstream.Tls.Key = string(c.PrivateKey)
	log.Debug().Msgf("ACME certificate successfully fetched from %s", url)

	return e
}
