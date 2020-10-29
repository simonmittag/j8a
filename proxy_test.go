package j8a

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"
)

func TestAbortAllUpstreamAttempts(t *testing.T) {
	Runner = mockRuntime()

	want := true
	got := false

	mockAtmpt := func() Atmpt {
		return Atmpt{
			URL:            nil,
			Label:          "",
			Count:          1,
			StatusCode:     0,
			isGzip:         false,
			resp:           nil,
			respBody:       nil,
			CompleteHeader: nil,
			CompleteBody:   nil,
			Aborted:        nil,
			AbortedFlag:    false,
			CancelFunc: func() {
				fmt.Println("cancelfunc called")
				got = true
			},
			startDate: time.Now(),
		}
	}

	atmpt := mockAtmpt()

	proxy := Proxy{
		XRequestID:    "",
		XRequestDebug: false,
		Up: Up{
			Atmpts: []Atmpt{mockAtmpt()},
			Atmpt:  &atmpt,
		},
		Dwn: Down{
			Req:         nil,
			Resp:        Resp{},
			Method:      "",
			Path:        "",
			URI:         "",
			UserAgent:   "",
			Body:        nil,
			ReqTooLarge: false,
			Aborted:     nil,
			AbortedFlag: false,
			startDate:   time.Now(),
		},
	}

	proxy.abortAllUpstreamAttempts()

	if want != got {
		t.Errorf("cancel func on proxy upstream attempt not triggered")
	}
}

func TestParseTlsVersionV12(t *testing.T) {
	req, _ := http.NewRequest("GET", "/hello", nil)
	req.TLS = &tls.ConnectionState{
		Version: tls.VersionTLS12,
	}
	if "TLS1.2" != parseTlsVersion(req) {
		t.Errorf("wrong TLS version")
	}
}

func TestParseTlsVersionV13(t *testing.T) {
	req, _ := http.NewRequest("GET", "/hello", nil)
	req.TLS = &tls.ConnectionState{
		Version: tls.VersionTLS13,
	}
	if "TLS1.3" != parseTlsVersion(req) {
		t.Errorf("wrong TLS version")
	}
}

func TestParseRequestBody(t *testing.T) {
	Runner = mockRuntime()
	Runner.Connection.Downstream.MaxBodyBytes = 65535

	body := []byte(`{"key":"value"}`)

	req, _ := http.NewRequest("PUT", "/hello", bytes.NewReader(body))
	req.ContentLength = int64(len(body))

	proxy := Proxy{}
	proxy.Dwn.startDate = time.Now()
	proxy.parseRequestBody(req)

	got := proxy.Dwn.ReqTooLarge
	want := false
	if got != want {
		t.Errorf("request entity should not be too large, sent %d, max %d", req.ContentLength, Runner.Connection.Downstream.MaxBodyBytes)
	}
}

func TestParseRequestBodyTooLarge(t *testing.T) {
	Runner = mockRuntime()
	Runner.Connection.Downstream.MaxBodyBytes = 65535
	req, _ := http.NewRequest("PUT", "/hello", nil)
	req.ContentLength = 65536

	proxy := Proxy{}
	proxy.Dwn.startDate = time.Now()
	proxy.parseRequestBody(req)

	got := proxy.Dwn.ReqTooLarge
	want := true
	if got != want {
		t.Errorf("request entity should be too large, sent %d, max %d", req.ContentLength, Runner.Connection.Downstream.MaxBodyBytes)
	}
}

func TestSuccessParseUpstreamContentLength(t *testing.T) {
	upBody := []byte("body")
	proxy := mockProxy(upBody, fmt.Sprint(len(upBody)), "", "", "")
	proxy.setContentLengthHeader()

	got := proxy.Dwn.Resp.Writer.Header().Get("Content-Length")
	want := fmt.Sprint(len(upBody))
	if got != want {
		t.Errorf("content-length was not properly set from upstream, got %s, want %s", got, want)
	}
}

func TestFailParseUpstreamContentLength(t *testing.T) {
	upBody := []byte("body")
	proxy := mockProxy(upBody, "NAN", "", "", "")
	proxy.setContentLengthHeader()

	got := proxy.Dwn.Resp.Writer.Header().Get("Content-Length")
	want := "0"
	if got != want {
		t.Errorf("content-length was not properly set from upstream, got %s, want %s", got, want)
	}
}

func TestPathTransformation(t *testing.T) {
	pathTransformation(t, "/mse6", "/mse7/v2/api", "/mse6/mse6/get/me/treats", "/mse7/v2/api/mse6/get/me/treats")
	pathTransformation(t, "/mse6", "/mse7/v2/api", "/mse6/get/me/treats", "/mse7/v2/api/get/me/treats")
	pathTransformation(t, "/mse6", "/mse7", "/mse6/get/me/treats", "/mse7/get/me/treats")
	pathTransformation(t, "/mse6", "/mse7", "/mse6/", "/mse7/")
	pathTransformation(t, "/mse6", "/mse6long", "/mse6?p=v", "/mse6long?p=v")
	pathTransformation(t, "/mse6", "/", "/mse6/get/me/treats", "/get/me/treats")
	pathTransformation(t, "/mse6", "/", "/mse6/", "/")
	pathTransformation(t, "/mse6", "", "/mse6/get/me/treats", "/mse6/get/me/treats")
	pathTransformation(t, "/mse6", "", "/mse6/", "/mse6/")
}

func pathTransformation(t *testing.T, routePath string, transform string, requestUri string, want string) {
	p := mockProxy(make([]byte, 1), "", routePath, transform, requestUri)
	got := p.resolveUpstreamURI()
	want = "http://upstreamhost:8080" + want
	if got != want {
		t.Errorf("path transformation error, got %s, want %s", got, want)
	}
}

func mockProxy(upBody []byte, cl string, path string, transform string, requestUri string) Proxy {
	pr, _ := regexp.Compile(path)
	tr, _ := regexp.Compile(transform)
	proxy := Proxy{
		XRequestID: "12345",
		Up: Up{
			Atmpt: &Atmpt{
				URL: &URL{
					Scheme: "http",
					Host:   "upstreamhost",
					Port:   8080,
				},
				resp: &http.Response{
					Body: ioutil.NopCloser(bytes.NewReader(upBody)),
					Header: map[string][]string{
						"Content-Length": []string{cl},
					},
				},
			},
		},
		Dwn: Down{
			URI:    requestUri,
			Method: "HEAD",
			Resp: Resp{
				Writer:        httptest.NewRecorder(),
				Body:          &upBody,
				ContentLength: 0,
			},
			startDate: time.Time{},
		},
		Route: &Route{
			Path:           path,
			PathRegex:      pr,
			Transform:      transform,
			TransformRegex: tr,
			Resource:       "mse7",
			Policy:         "",
		},
	}
	return proxy
}
