package method

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"
)

func TestHttpMethods(t *testing.T) {
	tests := []struct {
		method       string
		url          string
		responseCode int
		cl           int
	}{
		{
			method:       "CONNECT",
			url:          "/mse6/connect",
			responseCode: 200,
			cl:           0,
		},
		//we never proxy the body of a CONNECT request, see: https://httpwg.org/specs/rfc9110.html#CONNECT
		{
			method:       "CONNECT",
			url:          "/mse6/connect?body=true",
			responseCode: 200,
			cl:           0,
		},
		{
			method:       "GET",
			url:          "/mse6/get",
			responseCode: 200,
			cl:           38,
		},
		{
			method:       "HEAD",
			url:          "/mse6/getorhead",
			responseCode: 200,
			cl:           0,
		},
		//does not send a body but may send content length header which needs proxying
		{
			method:       "HEAD",
			url:          "/mse6/getorhead?cl=true",
			responseCode: 200,
			cl:           44,
		},
		{
			method:       "POST",
			url:          "/mse6/post",
			responseCode: 201,
			cl:           39,
		},
		{
			method:       "PUT",
			url:          "/mse6/put",
			responseCode: 200,
			cl:           38,
		},
		{
			method:       "OPTIONS",
			url:          "/mse6/options?code=200",
			responseCode: 200,
			cl:           0,
		},
		//if OPTIONS returns content we proxy it. The spec is not clear on this.
		{
			method:       "OPTIONS",
			url:          "/mse6/options?code=200&body=true",
			responseCode: 200,
			cl:           42,
		},
		{
			method:       "DELETE",
			url:          "/mse6/delete",
			responseCode: 204,
			cl:           0,
		},
		{
			method:       "PATCH",
			url:          "/mse6/patch",
			responseCode: 200,
			cl:           40,
		},
		//TRACE content and headers are proxied.
		{
			method:       "TRACE",
			url:          "/mse6/trace",
			responseCode: 200,
			cl:           40,
		},
	}

	for _, test := range tests {
		t.Run("TestWith"+test.method, func(t *testing.T) {
			methodWithResponseCodeBody(t, test.method, test.url, test.responseCode, test.cl)
		})
	}
}

func methodWithResponseCodeBody(t *testing.T, method string, url string, wantResponseCode int, wantCl int) {
	client := &http.Client{}

	url = "http://localhost:8080" + url

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

	req.Header.Add("Accept-Encoding", "identity")
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("error connecting to server for method %s, url %s, cause: %s", method, url, err)
	}
	if resp != nil && resp.Body != nil {
		//don't try to read the body response for head request
		if method != "HEAD" {
			b, _ := ioutil.ReadAll(resp.Body)
			if len(b) != wantCl {
				t.Errorf("incorrect response body bytes, want %v, got %v", wantCl, len(b))
			}
		}
		defer resp.Body.Close()
	}

	gotCl := resp.Header.Get("Content-Length")
	gotCli, _ := strconv.Atoi(gotCl)
	if wantCl != gotCli {
		t.Errorf("incorrect content length response header, want %v, got %v", wantCl, gotCli)
	}

	gotCode := resp.StatusCode
	if wantResponseCode != gotCode {
		t.Errorf("illegal response for method %s, url %s received %d want %d", method, url, gotCode, wantResponseCode)
	}
}
