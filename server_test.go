package jabba

import (
	"net/http"
	"testing"
)

func TestServerBootStrap(t *testing.T) {
	setupJabba()
	resp, _ := http.Get("http://localhost:8080/about")
	if resp.StatusCode != 200 {
		t.Errorf("server does not return ok status response after starting, want 200, got %v", resp.StatusCode)
	}
}

func setupJabba() {
	Boot.Add(1)
	go BootStrap()
	Boot.Wait()
}
