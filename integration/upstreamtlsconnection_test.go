package integration

import (
	"net/http"
	"testing"
)

func TestUpstreamInsecureVerifyOn(t *testing.T) {
	//tlsInsecureVerify=true for this server, should return 200 from upstream
	resp, err := http.Get("http://localhost:8080/badssl/get")
	if err != nil {
		t.Errorf("upstream insecure tls connection skip verify, cause: %v", err)
	}

	want := 200
	if resp.StatusCode != want {
		t.Errorf("upstream insecure tls connection skip verify, want %d, got %v", want, resp.StatusCode)
	}
}

func TestUpstreamInsecureVerifyOff(t *testing.T) {
	//tlsInsecureVerify=false for this server, should return 50x from upstream
	resp, err := http.Get("http://localhost:8081/badssl/get")
	if err != nil {
		t.Errorf("upstream insecure tls connection skip verify, cause: %v", err)
	}

	want := 502
	if resp.StatusCode != want {
		t.Errorf("upstream insecure tls connection skip verify, want %d, got %v", want, resp.StatusCode)
	}
}
