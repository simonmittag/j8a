package jabba

import (
	"net/http"
	"testing"
)

func TestServerBootStrap(t *testing.T) {
	resp, _ := http.Get("http://localhost:8080/about")
	if resp.StatusCode != 200 {
		t.Errorf("server does not return ok status response after starting, want 200, got %v", resp.StatusCode)
	}
}

func TestServerMakesSuccessfulUpstreamConnection(t *testing.T) {
	resp, err := http.Get("http://localhost:8080/mse6/get")

	if err != nil {
		t.Errorf("error connecting to upstream, cause: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("server does not return ok from working upstream, want 200, got %v", resp.StatusCode)
	}
}

func TestServerUpstreamReadTimeoutFailsWithSlowHeader(t *testing.T) {
	Runner.Connection.Upstream.ReadTimeoutSeconds = 1
	httpClient = scaffoldHTTPClient(*Runner)
	resp, err := http.Get("http://localhost:8080/mse6/slowheader")

	if err != nil {
		t.Errorf("error connecting to upstream, cause: %v", err)
	}

	if resp.StatusCode != 502 {
		t.Errorf("slow header writes from server > upstream timeout should not return ok and fail after max attempts, want 502, got %v", resp.StatusCode)
	}
}

func TestServerUpstreamReadTimeoutFailsWithSlowBody(t *testing.T) {
	Runner.Connection.Upstream.ReadTimeoutSeconds = 1
	httpClient = scaffoldHTTPClient(*Runner)

	resp, err := http.Get("http://localhost:8080/mse6/slowbody")
	if err != nil {
		t.Errorf("error connecting to upstream, cause: %v", err)
	}

	if resp.StatusCode != 502 {
		t.Errorf("slow body writes from server > upstream timeout should not return ok and fail after max attempts, want 502, got %v", resp.StatusCode)
	}
}

func TestServerUpstreamReadTimeoutPassesWithSlowHeader(t *testing.T) {
	//set read timeout to 4 seconds, upstream responds with ~3
	Runner.Connection.Upstream.ReadTimeoutSeconds = 4
	httpClient = scaffoldHTTPClient(*Runner)

	resp, err := http.Get("http://localhost:8080/mse6/slowheader")
	if err != nil {
		t.Errorf("error connecting to upstream, cause: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("slow header writes from server < upstream timeout should return ok, want 200, got %v", resp.StatusCode)
	}
}

func TestServerUpstreamReadTimeoutPassesWithSlowBody(t *testing.T) {
	//set read timeout to 10 seconds upstream responds in ~6
	Runner.Connection.Upstream.ReadTimeoutSeconds = 10
	httpClient = scaffoldHTTPClient(*Runner)

	resp, err := http.Get("http://localhost:8080/mse6/slowbody")
	if err != nil {
		t.Errorf("error connecting to upstream, cause: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("slow body writes from server < upstream timeout should return ok, want 200, got %v", resp.StatusCode)
	}
}

func setupJabbaWithMse6() {
	Boot.Add(1)
	go BootStrap()
	Boot.Wait()
}
