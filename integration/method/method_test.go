package method

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestHttpMethods(t *testing.T) {
	tests := []struct {
		method       string
		url          string
		responseCode int
	}{
		{
			method:       "CONNECT",
			url:          "/mse6/connect",
			responseCode: 200,
		},
		{
			method:       "GET",
			url:          "/mse6/get",
			responseCode: 200,
		},
		{
			method:       "HEAD",
			url:          "/mse6/getorhead?cl=y",
			responseCode: 200,
		},
		{
			method:       "POST",
			url:          "/mse6/post",
			responseCode: 201,
		},
		{
			method:       "PUT",
			url:          "/mse6/put",
			responseCode: 200,
		},
		{
			method:       "OPTIONS",
			url:          "/mse6/options?code=200",
			responseCode: 200,
		},
		{
			method:       "DELETE",
			url:          "/mse6/delete",
			responseCode: 204,
		},
		{
			method:       "PATCH",
			url:          "/mse6/patch",
			responseCode: 200,
		},
		{
			method:       "TRACE",
			url:          "/mse6/trace",
			responseCode: 200,
		},
	}

	for _, test := range tests {
		t.Run("TestWith"+test.method, func(t *testing.T) {
			methodWithResponseCodeBody(t, test.method, test.url, test.responseCode)
		})
	}
}

func methodWithResponseCodeBody(t *testing.T, method string, url string, responseCode int) {
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
		defer resp.Body.Close()
	}

	gotCode := resp.StatusCode
	if responseCode != gotCode {
		t.Errorf("illegal response for method %s, url %s received %d want %d", method, url, gotCode, responseCode)
	}
}
