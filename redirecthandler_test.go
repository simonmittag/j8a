package j8a

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

//this testHandler binds the mock HTTP server to proxyHandler.
type RedirectHandler struct{}

func (h RedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	redirectHandler(w, r)
}

func TestRedirectHandler(t *testing.T) {
	Runner = mockRuntime()
	Runner.Connection.Downstream.Tls.Port = 443

	h := &RedirectHandler{}
	server := httptest.NewServer(h)
	defer server.Close()

	c := &http.Client{}
	c.Get(server.URL)

	//the only thing we're testing is no nil pointers
}
