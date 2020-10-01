package integration

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"
)

func TestGETContentLengthResponses(t *testing.T) {
	MethodHasContentLengthAndBody(t, "GET", "http://localhost:8080/about", "Identity")
	MethodHasContentLengthAndBody(t, "GET", "http://localhost:8080/about", "identity")
	MethodHasContentLengthAndBody(t, "GET", "http://localhost:8080/about", "gzip")
	MethodHasContentLengthAndBody(t, "GET", "http://localhost:8080/about", "Gzip")
	MethodHasContentLengthAndBody(t, "GET", "http://localhost:8080/mse6/get", "gzip")
	MethodHasContentLengthAndBody(t, "GET", "http://localhost:8080/mse6/get", "identity")
}

func TestOPTIONSContentLengthResponses(t *testing.T) {
	MethodHasZeroContentLengthAndNoBody(t, "OPTIONS", "http://localhost:8080/mse6/options", "identity")
	MethodHasZeroContentLengthAndNoBody(t, "OPTIONS", "http://localhost:8080/mse6/options", "gzip")
	//golang removes content-length from http 204 response.
	MethodHasNoContentLengthHeaderAndNoBody(t, "OPTIONS", "http://localhost:8080/mse6/options?code=204", "identity")
	MethodHasNoContentLengthHeaderAndNoBody(t, "OPTIONS", "http://localhost:8080/mse6/options?code=204", "gzip")
}

func TestHEADContentLengthResponses(t *testing.T) {
	//upstream server does not send content-length during HEAD
	MethodHasZeroContentLengthAndNoBody(t, "HEAD", "http://localhost:8080/mse6/getorhead", "identity")
	MethodHasZeroContentLengthAndNoBody(t, "HEAD", "http://localhost:8080/mse6/getorhead", "gzip")

	//upstream server does send content-length of would-be resource as per RFC7231: https://tools.ietf.org/html/rfc7231#page-25
	MethodHasContentLengthAndNoBody(t, "HEAD", "http://localhost:8080/mse6/getorhead?cl=y", "identity")
	MethodHasContentLengthAndNoBody(t, "HEAD", "http://localhost:8080/mse6/getorhead?cl=y", "gzip")

	//upstream server serves actual resource with content-length and full body
	MethodHasContentLengthAndBody(t, "GET", "http://localhost:8080/mse6/getorhead", "identity")
	MethodHasContentLengthAndBody(t, "GET", "http://localhost:8080/mse6/getorhead", "gzip")
}

func TestPUTContentLengthResponses(t *testing.T) {
	//upstream server serves actual resource with content-length and full body
	MethodHasContentLengthAndBody(t, "PUT", "http://localhost:8080/mse6/put", "identity")
	MethodHasContentLengthAndBody(t, "PUT", "http://localhost:8080/mse6/put", "gzip")
}

func TestPATCHContentLengthResponses(t *testing.T) {
	//upstream server serves actual resource with content-length and full body
	MethodHasContentLengthAndBody(t, "PATCH", "http://localhost:8080/mse6/patch", "identity")
	MethodHasContentLengthAndBody(t, "PATCH", "http://localhost:8080/mse6/patch", "gzip")
}

func TestDELETEContentLengthResponses(t *testing.T) {
	//upstream server serves actual resource with content-length and full body
	MethodHasContentLengthAndBody(t, "DELETE", "http://localhost:8080/mse6/delete", "identity")
	MethodHasContentLengthAndBody(t, "DELETE", "http://localhost:8080/mse6/delete", "gzip")
}

func TestTRACEContentLengthResponses(t *testing.T) {
	//upstream server serves actual resource with content-length and full body
	MethodHasZeroContentLengthAndNoBody(t, "TRACE", "http://localhost:8080/mse6/trace", "identity")
	MethodHasZeroContentLengthAndNoBody(t, "TRACE", "http://localhost:8080/mse6/trace", "gzip")
}

func MethodHasContentLengthAndBody(t *testing.T, method string, url string, acceptEncoding string) {
	client := &http.Client{}

	var buf *bytes.Buffer
	var req *http.Request
	if method == "PUT" || method == "POST" || method == "PATCH" || method == "DELETE" {
		jsonData := map[string]string{"scandi": "grind", "convex": "grind", "concave": "grind"}
		jsonValue, _ := json.Marshal(jsonData)
		buf = bytes.NewBuffer(jsonValue)
		req, _ = http.NewRequest(method, url, buf)
	} else {
		req, _ = http.NewRequest(method, url, nil)
	}

	req.Header.Add("Accept-Encoding", acceptEncoding)
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("error connecting to server for method %s, url %s, cause: %s", method, url, err)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	got := int(resp.ContentLength)
	notwant := -1
	if got == notwant {
		t.Errorf("illegal response for method %s, url %s received Content-Length got %d", method, url, got)
	}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	if got != len(bodyBytes) {
		t.Errorf("illegal response for method %s, url %s content-length %d does not match body size %d", method, url, got, len(bodyBytes))
	}
}

func MethodHasContentLengthAndNoBody(t *testing.T, method string, url string, acceptEncoding string) {
	client := &http.Client{}
	req, _ := http.NewRequest(method, url, nil)
	req.Header.Add("Accept-Encoding", acceptEncoding)
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("error connecting to server for method %s, url %s, cause: %s", method, url, err)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	got2, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
	if got2 < 1 {
		t.Errorf("illegal response for method %s url %s want Content-Length >0 but got %d", method, url, got2)
	}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	if len(bodyBytes) > 0 {
		t.Errorf("illegal response for method %s, url %s should not have body, got %d bytes", method, url, len(bodyBytes))
	}
}

func MethodHasZeroContentLengthAndNoBody(t *testing.T, method, url string, acceptEncoding string) {
	client := &http.Client{}
	req, _ := http.NewRequest(method, url, nil)
	req.Header.Add("Accept-Encoding", acceptEncoding)
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("error connecting to server for method %s, url %s, cause: %s", method, url, err)
	}

	if resp != nil && resp.Body != nil {
		bod, _ := ioutil.ReadAll(resp.Body)
		if len(bod) > 0 {
			t.Errorf("illegal response contains body for url %s, method %s, got: %v", url, method, bod)
		}
		defer resp.Body.Close()
	}

	got2 := resp.Header.Get("Content-Length")
	want2 := "0"
	if got2 != want2 {
		t.Errorf("illegal response for method %s Content-Length, want %s got %s", method, want2, got2)
	}
}

func MethodHasNoContentLengthHeaderAndNoBody(t *testing.T, method, url string, acceptEncoding string) {
	client := &http.Client{}
	req, _ := http.NewRequest(method, url, nil)
	req.Header.Add("Accept-Encoding", acceptEncoding)
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("error connecting to server for method %s, url %s, cause: %s", method, url, err)
	}

	if resp != nil && resp.Body != nil {
		bod, _ := ioutil.ReadAll(resp.Body)
		if len(bod) > 0 {
			t.Errorf("illegal response contains body for method %s, got: %v", method, bod)
		}
		defer resp.Body.Close()
	}

	got2 := resp.Header.Get("Content-Length")
	want2 := ""
	if got2 != want2 {
		t.Errorf("illegal response for method %s Content-Length, want %s got %s", method, want2, got2)
	}
}
