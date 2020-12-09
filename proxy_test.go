package j8a

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
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

func TestExtractKid(t *testing.T) {
	tok := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6ImsxIn0.eyJpc3MiOiJqb2UiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwianRpIjoiYjE1ZWM5YzctYjZiMi00MGE1LTg3ZGEtN2ExMDVhMWY2ZTk0IiwiaWF0IjoxNjA2MjUwNTE4fQ.RNjqTVFkFTzgnkW0rJvW1yZbYFSr48g6gOKXPF34tEtebT6P5LbCh4JLKSmtIwEJ2PF6Tu6az2VIa9KiRTqWwqwQT5qJmI6Nyy9hMNY5PdmBV8HDTofAkGnvvlSG2iF0d5bVkJ223VN-mYRoWCR9S5D4kfjM3ZFhYQgfMi_k-kiU9KfPLxeUqcSjFx9jVYJj0diT--3GRejJU8VYpox40TwYf_KmS0IKmCu62SCXLXmiqNarAJ1R6zc8iNab5r05mqv1zJZcwRebj3Er0WdFbpIhwYR9lFYHjuxizJHo19-NW30g5NS6wLuk6QS8plK6_-kCgvYCzjLg_8ZFOyJLzg"
	want := "k1"
	got := extractKid(tok)
	if got != want {
		t.Errorf("unable to extract kid header from token, got %v, want %v", got, want)
	}
}

func TestExtractKidInvalid(t *testing.T) {
	tok := "eyJ0eXAiOiJKV1xQiLCJhbGciOiJSUzI1NiIsImtpZCI6ImsxIn0.eyJpc3MiOiJqb2UiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwianRpIjoiYjE1ZWM5YzctYjZiMi00MGE1LTg3ZGEtN2ExMDVhMWY2ZTk0IiwiaWF0IjoxNjA2MjUwNTE4fQ.RNjqTVFkFTzgnkW0rJvW1yZbYFSr48g6gOKXPF34tEtebT6P5LbCh4JLKSmtIwEJ2PF6Tu6az2VIa9KiRTqWwqwQT5qJmI6Nyy9hMNY5PdmBV8HDTofAkGnvvlSG2iF0d5bVkJ223VN-mYRoWCR9S5D4kfjM3ZFhYQgfMi_k-kiU9KfPLxeUqcSjFx9jVYJj0diT--3GRejJU8VYpox40TwYf_KmS0IKmCu62SCXLXmiqNarAJ1R6zc8iNab5r05mqv1zJZcwRebj3Er0WdFbpIhwYR9lFYHjuxizJHo19-NW30g5NS6wLuk6QS8plK6_-kCgvYCzjLg_8ZFOyJLzg"
	want := ""
	got := extractKid(tok)
	if got != want {
		t.Errorf("want empty kid header from token, got %v, want %v", got, want)
	}
}

func TestExtractKidInvalidHeader(t *testing.T) {
	tok := "PF34tEtebT6P5LbCh4JLKSmtIwEJ2PF6Tu6az2VIa9KiRTqWwqwQT5qJmI6Nyy9hMNY5PdmBV8HDTofAkGnvvlSG2iF0d5bVkJ223VN-mYRoWCR9S5D4kfjM3ZFhYQgfMi_k-kiU9KfPLxeUqcSjFx9jVYJj0diT--3GRejJU8VYpox40TwYf_KmS0IKmCu62SCXLXmiqNarAJ1R6zc8iNab5r05mqv1zJZcwRebj3Er0WdFbpIhwYR9lFYHjuxizJHo19-NW30g5NS6wLuk6QS8plK6_-kCgvYCzjLg_8ZFOyJLzg"
	want := ""
	got := extractKid(tok)
	if got != want {
		t.Errorf("want empty kid header from token, got %v, want %v", got, want)
	}
}

func TestExtractKidNoHeader(t *testing.T) {
	tok := ""
	want := ""
	got := extractKid(tok)
	if got != want {
		t.Errorf("want empty kid header from token, got %v, want %v", got, want)
	}
}

func TestExtractBadKidNoString(t *testing.T) {
	tok := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6MX0.eyJpc3MiOiJqb2UiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwianRpIjoiZTQ3ZTQyMDEtYTU5Zi00NTgzLTg0ZGEtODJhMmFhZjIyOTliIiwiaWF0IjoxNjA2NTExODYxfQ.Bu9qKyrctz8VToGaI8DdczBcaA_NEcDWwoRf7j-W68hoX-z8LkVwl9Ono4JziypQZJA8DJs6FinbSO54IiEszHKIh7J1TAiQxSpNL7YtjZDKConHaxREqDsXxEAW9edgaSFMth6Tclw8nOIYiCTrq678hBFHnTUYni4WCLVCZ1UYliw1sjoOKrUmk6teCna_sHBuXiht4fyZuKiT6X4ONU3HM0OBGLppKmTLmMadfOKmIy0QrJfTcH2C2UUehTJxR0l4qudIFTd5BU1YToDqNmZI9wAtXDf3iDPANn67NOqCdRhepmX4ztYkpcduOVu7X6mJBZXlujh_ld30Dpr7FQ"
	want := ""
	got := extractKid(tok)
	if got != want {
		t.Errorf("want empty kid header from token, got %v, want %v", got, want)
	}
}

func TestJwt_IatFail(t *testing.T) {
	now := time.Now()
	iat := now.Add(time.Second * 180)
	skew := 120
	payload := dummyHs256TokenFactory(t, jwt.IssuedAtKey, iat)

	err2 := verifyDateClaims(string(payload), skew)
	if err2 == nil {
		t.Error("got nil err but token should not have satisfied iat")
	} else {
		t.Logf("normal. token not validated, iat %d, skewSecs %d, now %d, delta %d, cause: %v", iat.Unix(), skew, now.Unix(), now.Unix()-iat.Unix(), err2)
	}
}

func TestJwt_IatFailSkew(t *testing.T) {
	now := time.Now()
	iat := now.Add(time.Second * 60)
	skew := 30
	payload := dummyHs256TokenFactory(t, jwt.IssuedAtKey, iat)

	err2 := verifyDateClaims(string(payload), skew)
	if err2 == nil {
		t.Error("got nil err but token should not have satisfied iat")
	} else {
		t.Logf("normal. token not validated, iat %d, skewSecs %d, now %d, delta %d, cause: %v", iat.Unix(), skew, now.Unix(), now.Unix()-iat.Unix(), err2)
	}
}

func TestJwt_IatPassWithinSkew(t *testing.T) {
	now := time.Now()
	iat := now.Add(time.Second * 60)
	skew := 120
	payload := dummyHs256TokenFactory(t, jwt.IssuedAtKey, iat)

	err2 := verifyDateClaims(string(payload), skew)
	if err2 != nil {
		t.Error("iat should have satisfied")
	} else {
		t.Logf("normal. iat satisfied time %d, skewSecs %d, now %d, delta %d", iat.Unix(), skew, now.Unix(), now.Unix()-iat.Unix())
	}
}

func dummyHs256TokenFactory(t *testing.T, key string, value time.Time) []byte {
	var err error
	var payload []byte

	tok := jwt.New()
	tok.Set(key, value)
	tok.Set("foo", "bar")
	payload, err = jwt.Sign(tok, jwa.HS256, []byte("secret"))
	t.Logf("token %s", payload)
	if err != nil {
		t.Errorf("cannot sign token, cause: %v", err)
	}
	return payload
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
			Path:      path,
			PathRegex: pr,
			Transform: transform,
			Resource:  "mse7",
			Policy:    "",
		},
	}
	return proxy
}
