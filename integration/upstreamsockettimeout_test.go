package integration

import (
	"net/http"
	"testing"
	"time"
)

func TestServerUpstreamConnectionTimeoutFailsWithBadDNS(t *testing.T) {
	start := time.Now()
	resp, err := http.Get("http://localhost:8080/notdns")
	elapsed := time.Since(start)

	//check integration test config file for correct values if test is failing
	maxAttempts:=4
	socketTimeout:=3
	want := maxAttempts*socketTimeout
	if elapsed > time.Duration(want) * time.Second {
		t.Errorf("upstream connection timeout exceed, want: %vs, got: %vs", want, elapsed)
	}

	if err != nil {
		t.Errorf("error connecting to upstream, cause: %v", err)
	}

	if resp.StatusCode != 502 {
		t.Errorf("upstream connection failure, want 502, got %v", resp.StatusCode)
	}
}