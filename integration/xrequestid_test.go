package integration

import (
	"net/http"
	"testing"
)

func TestSupplyXRequestID(t *testing.T) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://localhost:8080/mse6/get", nil)
	req.Header.Add("X-Request-Id", "test1")
	resp, err := client.Do(req)
	if err!=nil {
		t.Errorf("error connecting to server, cause: %s", err)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	//must be case insensitive
	got := resp.Header.Get("x-request-id")
	want := "test1"
	if got != want {
		t.Errorf("server did not return supplied X-Request-Id, want %s, got %s", want, got)
	}
}

func TestDoNotSupplyXRequestID(t *testing.T) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://localhost:8080/mse6/get", nil)
	resp, err := client.Do(req)
	if err!=nil {
		t.Errorf("error connecting to server, cause: %s", err)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	//must be case insensitive
	got := resp.Header.Get("x-request-id")
	if len(got)==0 {
		t.Errorf("server did not generate X-Request-Id, want any, got none")
	}
}
