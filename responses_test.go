package jabba

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

//this testHandler binds the mock HTTP server to proxyHandler.
type AboutHttpHandler struct{}

func (t AboutHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	aboutHandler(w, r)
}

func TestAboutHandlerContentEncodingIdentity(t *testing.T) {
	Runner = mockRuntime()

	server := httptest.NewServer(&AboutHttpHandler{})
	defer server.Close()

	c := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set("Accept-Encoding", "identity")
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	gotBody, _ := ioutil.ReadAll(resp.Body)
	if c := bytes.Compare(gotBody[0:2], gzipMagicBytes); c == 0 {
		t.Errorf("body should not have gzip response magic bytes but does: %v", gotBody[0:2])
	}

	want := "identity"
	got := resp.Header["Content-Encoding"][0]
	if got != want {
		t.Errorf("response does have correct Content-Encoding header, want %v, got %v", want, got)
	}
}

func TestAboutHandlerContentEncodingGzip(t *testing.T) {
	Runner = mockRuntime()

	server := httptest.NewServer(&AboutHttpHandler{})
	defer server.Close()

	c := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set("Accept-Encoding", "gzip")
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	gotBody, _ := ioutil.ReadAll(resp.Body)
	if c := bytes.Compare(gotBody[0:2], gzipMagicBytes); c != 0 {
		t.Errorf("body should not have gzip response magic bytes but does: %v", gotBody[0:2])
	}

	want := "gzip"
	got := resp.Header["Content-Encoding"][0]
	if got != want {
		t.Errorf("response does have correct Content-Encoding header, want %v, got %v", want, got)
	}
}
