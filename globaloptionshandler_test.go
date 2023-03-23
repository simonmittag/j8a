package j8a

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// this testHandler binds the mock HTTP server to proxyHandler.
type GlobalOptionsHandler struct{}

func (t GlobalOptionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	globalOptionsHandler(w, r)
}

func TestGlobalOptionsHandlerReturnsWithAcceptHeaders(t *testing.T) {
	eq := func(a, b []string) bool {
		if len(a) != len(b) {
			return false
		}
		for i, v := range a {
			if v != b[i] {
				return false
			}
		}
		return true
	}

	Runner = mockRuntime()

	server := httptest.NewServer(&GlobalOptionsHandler{})
	defer server.Close()

	c := &http.Client{}
	req, _ := http.NewRequest("OPTIONS", server.URL, nil)
	req.Header.Set(acceptEncoding, "identity")
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	got := resp.Header["Allow"]
	want := httpLegalMethods
	if !eq(want, got) {
		t.Errorf("expected server global allowed HTTP methods, want %v, got %v", want, got)
	}

	want2 := "identity"
	got2 := resp.Header[contentEncoding][0]
	if got2 != want2 {
		t.Errorf("response does have correct Content-Encoding header, want %v, got %v", want2, got2)
	}
}
