package integration

import (
	"net/http"
	"testing"
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

func TestServerUpstreamReadTimeoutFailsWithSlowHeader(t *testing.T) {
	resp, err := http.Get("http://localhost:8080/mse6/slowheader?wait=4")

	if err != nil {
		t.Errorf("error connecting to upstream, cause: %v", err)
	}

	if resp.StatusCode != 502 {
		t.Errorf("slow header writes from server > upstream timeout should not return ok and fail after max attempts, want 502, got %v", resp.StatusCode)
	}
}

func TestServerUpstreamReadTimeoutFailsWithSlowBody(t *testing.T) {
	resp, err := http.Get("http://localhost:8080/mse6/slowbody?wait=4")
	if err != nil {
		t.Errorf("error connecting to upstream, cause: %v", err)
	}

	if resp.StatusCode != 502 {
		t.Errorf("slow body writes from server > upstream timeout should not return ok and fail after max attempts, want 502, got %v", resp.StatusCode)
	}
}

func TestServerUpstreamReadTimeoutPassesWithSlowHeader(t *testing.T) {
	resp, err := http.Get("http://localhost:8080/mse6/slowheader?wait=2")
	if err != nil {
		t.Errorf("error connecting to upstream, cause: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("slow header writes from server < upstream timeout should return ok, want 200, got %v", resp.StatusCode)
	}
}

func TestServerUpstreamReadTimeoutPassesWithSlowBody(t *testing.T) {
	resp, err := http.Get("http://localhost:8080/mse6/slowbody?wait=2")
	if err != nil {
		t.Errorf("error connecting to upstream, cause: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("slow body writes from server < upstream timeout should return ok, want 200, got %v", resp.StatusCode)
	}
}