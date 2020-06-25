package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestServer1UpstreamReadTimeoutFireWithSlowHeader31S(t *testing.T) {
	performJabbaTest(t,
		"/slowheader",
		31,
		12,
		504,
		8080)
}

func TestServer1UpstreamReadTimeoutFireWithSlowBody31S(t *testing.T) {
	performJabbaTest(t,
		"/slowbody",
		31,
		12,
		504,
		8080)
}

func TestServer1UpstreamReadTimeoutFireWithSlowHeader25S(t *testing.T) {
	performJabbaTest(t,
		"/slowheader",
		25,
		12,
		504,
		8080)
}

func TestServer1UpstreamReadTimeoutFireWithSlowBody25S(t *testing.T) {
	performJabbaTest(t,
		"/slowbody",
		25,
		12,
		504,
		8080)
}

func TestServer1UpstreamReadTimeoutFireWithSlowHeader4S(t *testing.T) {
	performJabbaTest(t,
		"/slowheader",
		4,
		12,
		504,
		8080)
}

func TestServer1UpstreamReadTimeoutFireWithSlowBody4S(t *testing.T) {
	performJabbaTest(t,
		"/slowbody",
		4,
		12,
		504,
		8080)
}

func TestServer1UpstreamReadTimeoutNotFireWithSlowHeader2S(t *testing.T) {
	performJabbaTest(t,
		"/slowheader",
		2,
		2,
		200,
		8080)
}

func TestServer1UpstreamReadTimeoutNotFireWithSlowBody2S(t *testing.T) {
	performJabbaTest(t,
		"/slowbody",
		2,
		2,
		200,
		8080)
}

func TestServer2DownstreamRoundTripTimeoutFireWithSlowHeader31S(t *testing.T) {
	performJabbaTest(t,
		"/slowheader",
		31,
		20,
		503,
		8081)
}

func TestServer2DownstreamRoundTripTimeoutFireWithSlowBody31S(t *testing.T) {
	performJabbaTest(t,
		"/slowbody",
		31,
		20,
		503,
		8081)
}

func TestServer2DownstreamRoundTripTimeoutFireWithSlowHeader25S(t *testing.T) {
	performJabbaTest(t,
		"/slowheader",
		25,
		20,
		503,
		8081)
}

func TestServer2DownstreamRoundTripTimeoutFireWithSlowBody25S(t *testing.T) {
	performJabbaTest(t,
		"/slowbody",
		25,
		20,
		503,
		8081)
}

func TestServer2DownstreamRoundTripTimeoutNotFireWithSlowHeader4S(t *testing.T) {
	performJabbaTest(t,
		"/slowheader",
		4,
		4,
		200,
		8081)
}

func TestServer2DownstreamRoundTripTimeoutNotFireWithSlowBody4S(t *testing.T) {
	performJabbaTest(t,
		"/slowbody",
		4,
		4,
		200,
		8081)
}

func TestServer2DownstreamRoundTripTimeoutNotFireWithSlowHeader2S(t *testing.T) {
	performJabbaTest(t,
		"/slowheader",
		2,
		2,
		200,
		8081)
}

func TestServer2DownstreamRoundTripTimeoutNotFireWithSlowBody2S(t *testing.T) {
	performJabbaTest(t,
		"/slowbody",
		2,
		2,
		200,
		8081)
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
