package integration

import (
	"net/http"
	"testing"
	"time"
)

//check integration test config file for correct values if test is failing
func connectionTimeoutTestPeriod() int {
	maxAttempts := 4
	socketTimeout := 3
	grace := 1
	want := maxAttempts*socketTimeout + grace
	return want
}

func TestServerUpstreamSocketTimeoutWithBadLocal(t *testing.T) {
	start := time.Now()
	resp, err := http.Get("http://localhost:8080/badlocal")
	elapsed := time.Since(start)

	want := connectionTimeoutTestPeriod()
	if elapsed > time.Duration(want)*time.Second {
		t.Errorf("upstream connection timeout exceed, want: %vs, got: %vs", want, elapsed)
	}

	if err != nil {
		t.Errorf("error connecting to upstream, cause: %v", err)
	}

	if resp.StatusCode != 502 {
		t.Errorf("upstream connection failure, want 502, got %v", resp.StatusCode)
	}
}

func TestServerUpstreamSocketTimeoutWithBadRemote(t *testing.T) {
	start := time.Now()
	resp, err := http.Get("http://localhost:8080/badremote")
	elapsed := time.Since(start)

	want := connectionTimeoutTestPeriod()
	if elapsed > time.Duration(want)*time.Second {
		t.Errorf("upstream connection timeout exceed, want: %vs, got: %vs", want, elapsed)
	}

	if err != nil {
		t.Errorf("error connecting to upstream, cause: %v", err)
	}

	if resp.StatusCode != 504 {
		t.Errorf("upstream connection failure, want 504, got %v", resp.StatusCode)
	}
}

func TestServerUpstreamSocketTimeoutWithBadIp(t *testing.T) {
	start := time.Now()
	resp, err := http.Get("http://localhost:8080/badip")
	elapsed := time.Since(start)

	want := connectionTimeoutTestPeriod()
	if elapsed > time.Duration(want)*time.Second {
		t.Errorf("upstream connection timeout exceed, want: %vs, got: %vs", want, elapsed)
	}

	if err != nil {
		t.Errorf("error connecting to upstream, cause: %v", err)
	}

	if resp.StatusCode != 504 {
		t.Errorf("upstream connection failure, want 504, got %v", resp.StatusCode)
	}
}

func TestServerUpstreamSocketTimeoutWithBadDNS(t *testing.T) {
	start := time.Now()
	resp, err := http.Get("http://localhost:8080/baddns")
	elapsed := time.Since(start)

	//check integration test config file for correct values if test is failing
	want := connectionTimeoutTestPeriod()
	if elapsed > time.Duration(want)*time.Second {
		t.Errorf("upstream connection timeout exceed, want: %vs, got: %vs", want, elapsed)
	}

	if err != nil {
		t.Errorf("error connecting to upstream, cause: %v", err)
	}

	if resp.StatusCode != 502 {
		t.Errorf("upstream connection failure, want 502, got %v", resp.StatusCode)
	}
}
