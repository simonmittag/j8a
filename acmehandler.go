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
)

var acmeProviders = map[string]string{
	"letsencrypt":  "https://acme-v02.api.letsencrypt.org/directory",
	"let'sencrypt": "https://acme-v02.api.letsencrypt.org/directory",
	//"letsencrypt":  "https://acme-staging-v02.api.letsencrypt.org/directory",
	//"let'sencrypt": "https://acme-staging-v02.api.letsencrypt.org/directory",
}

type AcmeHandler struct {
	Active  bool
	Domain  string
	Token   string
	KeyAuth []byte
}

func (a *AcmeHandler) Present(domain, token, keyAuth string) error {
	a.Active = true
	a.Domain = domain
	a.Token = token
	a.KeyAuth = []byte(keyAuth)

	log.Debug().Msg("ACME handler activated")
	return nil
}

const notConfigured = "not configured"

func (a *AcmeHandler) CleanUp(domain, token, keyAuth string) error {
	a.Active = false
	a.Domain = notConfigured
	a.Token = notConfigured
	a.KeyAuth = []byte(notConfigured)

	log.Debug().Msg("ACME handler deactivated")
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
		Domains: []string{Runner.Connection.Downstream.Tls.Acme.Domain},
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
