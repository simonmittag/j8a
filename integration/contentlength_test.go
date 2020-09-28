package integration

import (
	"io/ioutil"
	"net/http"
	"testing"
)

func TestServerContentLengthResponses(t *testing.T) {
	ServerGETInsertsContentLength(t, "http://localhost:8080/about", "Identity")
	ServerGETInsertsContentLength(t, "http://localhost:8080/about", "identity")
	ServerGETInsertsContentLength(t, "http://localhost:8080/about", "gzip")
	ServerGETInsertsContentLength(t, "http://localhost:8080/about", "Gzip")

	ServerGETInsertsContentLength(t, "http://localhost:8080/mse6/get", "gzip")
	ServerGETInsertsContentLength(t, "http://localhost:8080/mse6/get", "identity")

	ServerOPTIONSNoInsertContentLength(t, "http://localhost:8080/mse6/options", "identity")
	//ServerOPTIONSNoInsertContentLength(t, "http://localhost:8080/mse6/options", "gzip")
}

func ServerGETInsertsContentLength(t *testing.T, url string, acceptEncoding string) {
	method := "GET"
	client := &http.Client{}
	req, _ := http.NewRequest(method, url, nil)
	req.Header.Add("Accept-Encoding", acceptEncoding)
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("error connecting to server, cause: %s", err)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	got := int(resp.ContentLength)
	notwant := -1
	if got == notwant {
		t.Errorf("illegal response Content-Length got %d", got)
	}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	if got != len(bodyBytes) {
		t.Errorf("content-length %d does not match body size %d", got, len(bodyBytes))
	}
}

func ServerOPTIONSNoInsertContentLength(t *testing.T, url string, acceptEncoding string) {
	method := "OPTIONS"
	client := &http.Client{}
	req, _ := http.NewRequest(method, url, nil)
	req.Header.Add("Accept-Encoding", acceptEncoding)
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("error connecting to server, cause: %s", err)
	}

	if resp != nil && resp.Body != nil {
		bod, _ := ioutil.ReadAll(resp.Body)
		if len(bod) > 0 {
			t.Errorf("illegal response contains body for OPTIONS, got: %v", bod)
		}
		defer resp.Body.Close()
	}

	//test golang parsed
	got := int(resp.ContentLength)
	want := 0
	if got != want {
		t.Errorf("illegal response Content-Length, want %d got %d", want, got)
	}

	//not that we don't trust you golang, but test the actual http header sent
	got2 := resp.Header.Get("Content-Length")
	want2 := "0"
	if got2 != want2 {
		t.Errorf("illegal response Content-Length, want %s got %s", want2, got2)
	}
}
