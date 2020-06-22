package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestServerMakesSuccessfulUpstreamConnection(t *testing.T) {
	resp, err := http.Get("http://localhost:8080/mse6/get")

	if err != nil {
		t.Errorf("error connecting to upstream, cause: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("server does not return ok from working upstream, want 200, got %v", resp.StatusCode)
	}
}

func TestServer1UpstreamReadTimeoutFailsWithSlowHeader4S(t *testing.T) {
	performJabbaTest(t, "/slowheader", 4, 12, 502, 8080)
}

func TestServer1UpstreamReadTimeoutFailsWithSlowBody4S(t *testing.T) {
	performJabbaTest(t, "/slowbody", 4, 12, 502, 8080)
}

func TestServer1UpstreamReadTimeoutPassesWithSlowHeader2S(t *testing.T) {
	performJabbaTest(t, "/slowheader", 2, 2, 200, 8080)
}

func TestServer1UpstreamReadTimeoutPassesWithSlowBody2S(t *testing.T) {
	performJabbaTest(t, "/slowbody", 2, 2, 200, 8080)
}

func performJabbaTest(t *testing.T, testMethod string, wantUpstreamWaitSeconds int, wantTotalWaitSeconds int, wantStatusCode int, serverPort int) {
	start := time.Now()
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/mse6%s?wait=%d", serverPort, testMethod, wantUpstreamWaitSeconds))
	gotTotalWait := time.Since(start)
	gotStatusCode := resp.StatusCode

	if err != nil {
		t.Errorf("error connecting to upstream for port %d, testMethod %s, cause: %v", serverPort, testMethod, err)
	}

	if !okayTimeDrift(gotTotalWait, wantTotalWaitSeconds) {
		t.Errorf("bad time drift for port %d, testMethod %s, want seconds %d, got %f", serverPort, testMethod, wantTotalWaitSeconds, gotTotalWait.Seconds())
	}

	if gotStatusCode != wantStatusCode {
		t.Errorf("bad status code for port %d, testMethod %s, want statusCode %d, got %d", serverPort, testMethod, wantStatusCode, gotStatusCode)
	}
}

func okayTimeDrift(elapsed time.Duration, waitSeconds int) bool {
	fmin := 1.0
	fmax := 1.1
	elapsedSeconds := elapsed.Seconds()

	if elapsedSeconds > fmax*float64(waitSeconds) {
		return false
	}
	if elapsedSeconds < fmin*float64(waitSeconds) {
		return false
	}
	return true
}
