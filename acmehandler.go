package j8a

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/simonmittag/lego/v4/certcrypto"
	"github.com/simonmittag/lego/v4/certificate"
	"github.com/simonmittag/lego/v4/challenge/http01"
	"github.com/simonmittag/lego/v4/lego"
	"github.com/simonmittag/lego/v4/registration"
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

	log.Info().Msgf("ACME handler for domain %s activated, ready to serve challenge response for token %s.", domain, token)
	return nil
}

func (a *AcmeHandler) CleanUp(domain, token, keyAuth string) error {
	delete(a.Active, token)
	delete(a.Domains, token)
	delete(a.KeyAuths, token)

	log.Info().Msgf("ACME handler for domain %s deactivated.", domain)
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
			log.Warn().Msgf("unsuccessful response to remote ACME challenge, URI %s, cause: %v", r.RequestURI, r1)
		}
	}()

	tokens := strings.Split(r.RequestURI, "/")
	token := tokens[len(tokens)-1]

	a := Runner.AcmeHandler
	path := http01.ChallengePath(token)
	w.WriteHeader(200)
	w.Write([]byte(a.KeyAuths[token]))
	log.Info().Msgf(acmeEvent, path, a.KeyAuths[token], a.Domains[token])
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

const acmeKeyFile = "tls.pk"
const acmeCertFile = "tls.cert"
const acmeHashFile = "confighash"

func (runtime *Runtime) loadAcmeCertAndKeyFromCache(provider string) error {
	var e error

	if !runtime.cacheDirIsActive() {
		return errors.New("cache directory not active, cannot load TLS from cache")
	}

	keyFile := filepath.FromSlash(Runner.cacheDir + "/" + provider + "/" + acmeKeyFile)
	key, e1 := ioutil.ReadFile(keyFile)
	if e1 != nil {
		return e1
	}

	certFile := filepath.FromSlash(Runner.cacheDir + "/" + provider + "/" + acmeCertFile)
	cert, e2 := ioutil.ReadFile(certFile)
	if e2 != nil {
		return e2
	}

	hashFile := filepath.FromSlash(Runner.cacheDir + "/" + provider + "/" + acmeHashFile)
	hash, e3 := ioutil.ReadFile(hashFile)
	if e3 != nil {
		return e3
	}

	clearCache := func(msg string) {
		os.Remove(keyFile)
		os.Remove(certFile)
		os.Remove(hashFile)
		log.Info().Msg(msg)
	}

	if c := bytes.Compare(hash, runtime.acmeProviderHashAsBytes()); c != 0 {
		msg := fmt.Sprintf("active TLS configuration does not match cached cert and key, clearing cache for ACME provider %s", provider)
		clearCache(msg)
		return errors.New(msg)
	}

	if _, e3 := checkFullCertChainFromBytes(cert, key); e3 == nil {
		runtime.Connection.Downstream.Tls.Key = string(key)
		runtime.Connection.Downstream.Tls.Cert = string(cert)
		log.Info().Msgf("TLS cert and key for ACME provider %s loaded from cache", provider)
	} else {
		//if delete doesn't work ignore this it may already be gone (partially).
		msg := fmt.Sprintf("unable to load data, clearing TLS cache for ACME provider %s", provider)
		clearCache(msg)
		return errors.New(msg)
	}

	return e
}

const acmeRwx fs.FileMode = 0700

func (runtime *Runtime) cacheAcmeCertAndKey(provider string) error {
	var e error

	if !runtime.cacheDirIsActive() {
		return errors.New("cache directory not active, cannot cache keys")
	} else {
		//it doesn't matter if this fails because dir already exists
		os.Mkdir(runtime.cacheDir+"/"+provider, acmeRwx)
	}

	e1 := ioutil.WriteFile(
		filepath.FromSlash(Runner.cacheDir+"/"+provider+"/"+acmeKeyFile),
		[]byte(runtime.Connection.Downstream.Tls.Key),
		acmeRwx)
	if e1 != nil {
		return e1
	}

	e2 := ioutil.WriteFile(
		filepath.FromSlash(Runner.cacheDir+"/"+provider+"/"+acmeCertFile),
		[]byte(runtime.Connection.Downstream.Tls.Cert),
		acmeRwx)
	if e2 != nil {
		return e2
	}

	e3 := ioutil.WriteFile(
		filepath.FromSlash(Runner.cacheDir+"/"+provider+"/"+acmeHashFile),
		runtime.acmeProviderHashAsBytes(),
		acmeRwx)
	if e3 != nil {
		return e3
	}

	log.Info().Msgf("stored TLS cert and key for ACME provider %s in cache", provider)
	return e
}

func asSha256(o interface{}) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", o)))

	return fmt.Sprintf("%x", h.Sum(nil))
}

func (runtime *Runtime) acmeProviderHashAsBytes() []byte {
	return []byte(asSha256(runtime.Connection.Downstream.Tls.Acme))
}

func (runtime *Runtime) fetchAcmeCertAndKey(url string) error {
	var e error

	defer func() {
		if r := recover(); r != nil {
			msg := fmt.Sprintf("TLS ACME certificate not fetched, cause: %s", r)
			log.Warn().Msg(msg)
			e = errors.New(msg)
		}
	}()

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

	//we always register because it's safer than to cache credentials
	myUser.Registration, e = client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: runtime.Connection.Downstream.Tls.Acme.AcceptTOS})
	if e != nil {
		return e
	}

	request := certificate.ObtainRequest{
		Domains: runtime.Connection.Downstream.Tls.Acme.Domains,
		Bundle:  true,
	}

	var c *certificate.Resource
	c, e = client.Certificate.Obtain(request)
	if e != nil {
		log.Warn().Msgf("ACME certificate from %s unsuccessful, cause %v", url, e)
		return e
	}

	runtime.Connection.Downstream.Tls.Cert = string(c.Certificate)
	runtime.Connection.Downstream.Tls.Key = string(c.PrivateKey)

	log.Info().Msgf("ACME certificate successfully fetched from %s", url)

	return e
}
