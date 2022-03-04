package headers

import (
	"github.com/simonmittag/j8a"
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

	//must be case lowercase HTTP/2
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

func TestSupplyXRequestInfoUpstreamGzip(t *testing.T) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://localhost:8080/mse6/gzip", nil)
	req.Header.Add("Accept-Encoding", "identity")
	req.Header.Add("X-Request-Info", "true")
	req.Header.Add("X-Request-Id", "test1")
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("error connecting to server, cause: %s", err)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	//must be case lowercase HTTP/2
	got := resp.Header.Get("x-request-id")
	want := "test1"
	if got != want {
		t.Errorf("server did not return supplied X-Request-Id, want %s, got %s", want, got)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	utf8 := string(*j8a.Gunzip(body))
	if !strings.Contains(utf8, "gzip endpoint") {
		t.Errorf("should have sent decoded gzip respose with serverside request info but got this instead: %s", body)
	}
}

func TestSupplyXRequestInfoUpstreamTinyGzip(t *testing.T) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://localhost:8080/mse6/tinygzip", nil)
	req.Header.Add("Accept-Encoding", "identity")
	req.Header.Add("X-Request-Info", "true")
	req.Header.Add("X-Request-Id", "test2")
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("error connecting to server, cause: %s", err)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	//must be case lowercase HTTP/2
	got := resp.Header.Get("x-request-id")
	want := "test2"
	if got != want {
		t.Errorf("server did not return supplied X-Request-Id, want %s, got %s", want, got)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	utf8 := string(body)
	if !strings.Contains(utf8, "{}") {
		t.Errorf("should have sent decoded gzip respose with tiny serverside request info but got this instead: %s", body)
	}
}

func TestSupplyXRequestInfoUpstreamIdentity(t *testing.T) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://localhost:8080/mse6/get", nil)
	req.Header.Add("Accept-Encoding", "identity")
	req.Header.Add("X-Request-Info", "true")
	req.Header.Add("X-Request-Id", "test1")
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("error connecting to server, cause: %s", err)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	//must be case lowercase HTTP/2
	got := resp.Header.Get("x-request-id")
	want := "test1"
	if got != want {
		t.Errorf("server did not return supplied X-Request-Id, want %s, got %s", want, got)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	utf8 := string(body)
	if !strings.Contains(utf8, "get endpoint") {
		t.Errorf("should have sent decoded gzip respose with serverside request info but got this instead: %s", body)
	}
}

func TestSupplyXRequestInfoUpstreamTinyIdentity(t *testing.T) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://localhost:8080/mse6/tiny", nil)
	req.Header.Add("Accept-Encoding", "identity")
	req.Header.Add("X-Request-Info", "true")
	req.Header.Add("X-Request-Id", "test3")
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("error connecting to server, cause: %s", err)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	//must be case lowercase HTTP/2
	got := resp.Header.Get("x-request-id")
	want := "test3"
	if got != want {
		t.Errorf("server did not return supplied X-Request-Id, want %s, got %s", want, got)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	utf8 := string(body)
	if !strings.Contains(utf8, "{}") {
		t.Errorf("should have sent decoded gzip respose with tiny serverside request info but got this instead: %s", body)
	}
}
