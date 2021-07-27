package j8a

import "net/http"

type AcmeHandler struct {
}

func (a AcmeHandler) Present(domain, token, keyAuth string) error {
	Runner.AcmeChallengeActive = true
	return nil
}

func (a AcmeHandler) CleanUp(domain, token, keyAuth string) error {
	Runner.AcmeChallengeActive = false
	return nil
}

func acmeHandler(response http.ResponseWriter, request *http.Request) {

}
