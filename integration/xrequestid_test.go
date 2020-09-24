package integration

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestSupplyXRequestID(t *testing.T) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://localhost:8080/mse6/echoheader", nil)
	req.Header.Add("X-Request-Id", "test1")
	resp, err := client.Do(req)
	if err != nil {
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

	body, _ := ioutil.ReadAll(resp.Body)
	utf8 := string(body)
	if !strings.Contains(utf8, "test1") {
		t.Errorf("should have sent X-Request-ID header upstream, but sent only this %s", body)
	}
}

func TestDoNotSupplyXRequestID(t *testing.T) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://localhost:8080/mse6/echoheader", nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("error connecting to server, cause: %s", err)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	//must be case insensitive
	got := resp.Header.Get("x-request-id")
	if len(got) == 0 {
		t.Errorf("server did not generate X-Request-Id, want any, got none")
	}

	body, _ := ioutil.ReadAll(resp.Body)
	utf8 := string(body)
	if !strings.Contains(utf8, "X-Request-Id") {
		t.Errorf("should have sent X-Request-ID header upstream, but sent only this %s", body)
	}
}
