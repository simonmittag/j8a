package j8a

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
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
		XRequestID:   "",
		XRequestInfo: false,
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
	if "1.2" != parseTlsVersion(req) {
		t.Errorf("wrong TLS version")
	}
}

func TestParseTlsVersionV13(t *testing.T) {
	req, _ := http.NewRequest("GET", "/hello", nil)
	req.TLS = &tls.ConnectionState{
		Version: tls.VersionTLS13,
	}
	if "1.3" != parseTlsVersion(req) {
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
	proxy := mockProxy(upBody, fmt.Sprint(len(upBody)), "", "", "", "", "")
	proxy.setContentLengthHeader()

	got := proxy.Dwn.Resp.Writer.Header().Get("Content-Length")
	want := fmt.Sprint(len(upBody))
	if got != want {
		t.Errorf("content-length was not properly set from upstream, got %s, want %s", got, want)
	}
}

func TestFailParseUpstreamContentLength(t *testing.T) {
	upBody := []byte("body")
	proxy := mockProxy(upBody, "NAN", "", "", "", "", "")
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

	err2 := verifyDateClaims(string(payload), skew, log.Trace())
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

	err2 := verifyDateClaims(string(payload), skew, log.Trace())
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

	err2 := verifyDateClaims(string(payload), skew, log.Trace())
	if err2 != nil {
		t.Error("iat should have satisfied")
	} else {
		t.Logf("normal. iat satisfied time %d, skewSecs %d, now %d, delta %d", iat.Unix(), skew, now.Unix(), now.Unix()-iat.Unix())
	}
}

func TestJwt_ExpFail(t *testing.T) {
	now := time.Now()
	exp := now.Add(-time.Second * 180)
	skew := 120
	payload := dummyHs256TokenFactory(t, jwt.ExpirationKey, exp)

	err2 := verifyDateClaims(string(payload), skew, log.Trace())
	if err2 == nil {
		t.Error("got nil err but token should not have satisfied exp")
	} else {
		t.Logf("normal. token not validated, exp %d, skewSecs %d, now %d, delta %d, cause: %v", exp.Unix(), skew, now.Unix(), now.Unix()-exp.Unix(), err2)
	}
}

func TestJwt_ExpFailSkew(t *testing.T) {
	now := time.Now()
	exp := now.Add(-time.Second * 60)
	skew := 30
	payload := dummyHs256TokenFactory(t, jwt.ExpirationKey, exp)

	err2 := verifyDateClaims(string(payload), skew, log.Trace())
	if err2 == nil {
		t.Error("got nil err but token should not have satisfied exp")
	} else {
		t.Logf("normal. token not validated, exp %d, skewSecs %d, now %d, delta %d, cause: %v", exp.Unix(), skew, now.Unix(), now.Unix()-exp.Unix(), err2)
	}
}

func TestJwt_ExpPassWithinSkew(t *testing.T) {
	now := time.Now()
	exp := now.Add(-time.Second * 60)
	skew := 120
	payload := dummyHs256TokenFactory(t, jwt.ExpirationKey, exp)

	err2 := verifyDateClaims(string(payload), skew, log.Trace())
	if err2 != nil {
		t.Error("exp should have satisfied")
	} else {
		t.Logf("normal. exp satisfied time %d, skewSecs %d, now %d, delta %d", exp.Unix(), skew, now.Unix(), now.Unix()-exp.Unix())
	}
}

func TestJwt_NbfFail(t *testing.T) {
	now := time.Now()
	nbf := now.Add(time.Second * 180)
	skew := 120
	payload := dummyHs256TokenFactory(t, jwt.NotBeforeKey, nbf)

	err2 := verifyDateClaims(string(payload), skew, log.Trace())
	if err2 == nil {
		t.Error("got nil err but token should not have satisfied nbf")
	} else {
		t.Logf("normal. token not validated, nbf %d, skewSecs %d, now %d, delta %d, cause: %v", nbf.Unix(), skew, now.Unix(), now.Unix()-nbf.Unix(), err2)
	}
}

func TestJwt_NbfFailSkew(t *testing.T) {
	now := time.Now()
	nbf := now.Add(time.Second * 60)
	skew := 30
	payload := dummyHs256TokenFactory(t, jwt.NotBeforeKey, nbf)

	err2 := verifyDateClaims(string(payload), skew, log.Trace())
	if err2 == nil {
		t.Error("got nil err but token should not have satisfied nbf")
	} else {
		t.Logf("normal. token not validated, nbf %d, skewSecs %d, now %d, delta %d, cause: %v", nbf.Unix(), skew, now.Unix(), now.Unix()-nbf.Unix(), err2)
	}
}

func TestJwt_NbfPassWithinSkew(t *testing.T) {
	now := time.Now()
	nbf := now.Add(time.Second * 60)
	skew := 120
	payload := dummyHs256TokenFactory(t, jwt.NotBeforeKey, nbf)

	err2 := verifyDateClaims(string(payload), skew, log.Trace())
	if err2 != nil {
		t.Error("nbf should have satisfied")
	} else {
		t.Logf("normal. nbf satisfied time %d, skewSecs %d, now %d, delta %d", nbf.Unix(), skew, now.Unix(), now.Unix()-nbf.Unix())
	}
}

func TestJwt_NbfPassNoSkew(t *testing.T) {
	now := time.Now()
	nbf := now.Add(time.Second * -3)
	skew := 0
	payload := dummyHs256TokenFactory(t, jwt.NotBeforeKey, nbf)

	err2 := verifyDateClaims(string(payload), skew, log.Trace())
	if err2 != nil {
		t.Error("nbf should have satisfied")
	} else {
		t.Logf("normal. nbf satisfied time %d, skewSecs %d, now %d, delta %d", nbf.Unix(), skew, now.Unix(), now.Unix()-nbf.Unix())
	}
}

func TestJwt_NbfFailNoSkew(t *testing.T) {
	now := time.Now()
	nbf := now.Add(time.Second * 3)
	skew := 0
	payload := dummyHs256TokenFactory(t, jwt.NotBeforeKey, nbf)

	err2 := verifyDateClaims(string(payload), skew, log.Trace())
	if err2 == nil {
		t.Error("got nil err but token should not have satisfied nbf")
	} else {
		t.Logf("normal token not validated, nbf %d, skewSecs %d, now %d, delta %d", nbf.Unix(), skew, now.Unix(), now.Unix()-nbf.Unix())
	}
}

func TestUpstreamNobody(t *testing.T) {
	Runner = &Runtime{
		Config: Config{
			Connection: Connection{
				Upstream: Upstream{
					MaxAttempts: 4,
				},
			},
		},
		ReloadableCert: NewReloadableCert(),
	}

	proxy := mockProxy([]byte(""), "0", "/path", "/path", "/get", "", "")
	proxy.encodeUpstreamResponseBody()

	if proxy.Dwn.Resp.Body == nil {
		t.Errorf("downstream body should have been initialized")
	}
}

func TestValidateJwtNoClaims(t *testing.T) {
	Runner = mockJwtRuntime("jwt",
		"RS256",
		"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAuvtFgDnIcdB/jqSLICns\nz7FXU/uiFSdJGVpGc5Dy+xm8wZwgiy6lJdL9/TtYjnmJefkPVyYdazabvGvOcns7\n3rshkt0g6Ackqa72yiUEsv1kzCvBObPYNXgr1dNda8/F/ZiO3V9BtcTgQs9Y6rdO\nWJq7zNpees8pfuhEamk3sQp8AmKImFNfuZceNeglMHLLt0NcmSQp4VmhDCladFa1\nEdLirtFM9BtEIOlX20SRcN1LjeRsos8JywpQRxe6M3bnGFXcDQHqrsvwkkzu+vBt\nnPFa2e+jkBSDWkf6ZwvdJnEEUiJkHYTgJuXD1sbGeUkQL1Jb5NaQHhQ1mt3xn1z0\ntwIDAQAB\n-----END PUBLIC KEY-----\n",
		"")

	proxy := mockProxy([]byte(""),
		"0",
		"/path",
		"/path",
		"/get",
		"eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6Ims0MiJ9.eyJpc3MiOiJqb2UiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwianRpIjoiNGFkOTcxMDMtZmQ4ZC00Zjc5LWIwNmEtZTRmOTE1NzkxYjUyIiwiaWF0IjoxNjA3ODU1MTU3fQ.irL3sYTzkM4yFGKBfTzoAAe5H7mGaHECXFy-lkOfVoaaPwuL29b-je_ROoeR8uqw_441QE2P-Ky5582tG2dcu7s3EC3FNuPN_CaZPmbhzV8YIKdzRY7GiTj9sij1_2uRB61b5Qns7H3AJjMuZeCcaGA9t3gSJlVZwkpy9qU46JpX13SPqdSSR9Sg2kZhNFmrRDM5ZGN2VMuzvK34dW72NUkHVaBJUmIRASAfKQnA39xGMskTjP8ZZSzGdYiu0MMhBCA9DZmiS9YBw2Sj6J6Vo7_6SyKAoQyd44JACWbM28jZpfSWDPe-nkMu5ccxNa3A4ocFibnGXaKItWER9MTfeA",
		"jwt")

	ok := proxy.validateJwt()

	if !ok {
		t.Errorf("jwt token did not validate")
	}
}

func TestValidateJwtWithClaims(t *testing.T) {
	Runner = mockJwtRuntime("jwty",
		"RS256",
		"-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n",
		`.sub | select(.=="subscriber")`,
		`.sub | select(.=="admin")`)

	proxy := mockProxy([]byte(""),
		"0",
		"/path",
		"/path",
		"/get",
		"eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6ImszIn0.eyJpc3MiOiJpc3N1ZXIiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwic3ViIjoiYWRtaW4iLCJub25jZSI6InNlc3Npb25JRCIsImp0aSI6IjFkYmM3ZjBmLWY1NjQtNGYyNy04MWViLWVhOTJkYTU5NzgzMSIsImlhdCI6MTYwODY2ODY4N30.W3iv8DDmwkMyt5M5KmKgJ1-Tns1LBm7ov-AdzDulDkp7NhrAqrtg6SnLr3KlBCqJQFh84PYEZ2uuOTapEkXkL2xiRIdEenumWMl_65qazpvdbkWWnZ52FP4tHH-3pWWcb0VEc1HSAJuvFN7pDO1Y9lIYeMPGAJY_4nRDHj-60MTNhd4MP6tf29wyBjvaHRlq1a6dCvPhNj6LESqTrGq1DnrvsdZf2FXHPDsv_DvbLOeh_l4-A1hKbrae7OTFYJijbfLwYNs3B12dUxHJ_bSyLmV84kAPZk-IBVUhusx2kbLEVEKT1upblv9ltgnmnsgbwSv3ClYr_1VPOTvpZDSMxhf2zTHIo1W7R0ZdF5f7aFSNmKW59ya5gUHgK9XjEKryyQuUXU2FCJXDARKGie-4VvHZiJo0Nv2De6PGutB_cXjPRs9lyFVui6XtakMaDKVUrE1BwjyXRlf0cGSARTv3wC9x2VmW1ZuoHm9mYCUV3dZiQ0M0gLjZZcLWF4Jq8MtLl-d0hjq5VoBqmmOBnga6JFROFom8Y0ak-5tRXbpJ67GBgyNTXuJ3iBOUXs0Od3t9ZjUfPQElii1q19pac9vtHsfMp9Otur6tKukHvPC-6kLKM4z0OpzvgaMQm7YhlV882GEFaSviW3hYMtyiwT9Ib3FPPsyGySTQWl-4QLk-b3o",
		"jwty")

	ok := proxy.validateJwt()

	if !ok {
		t.Errorf("jwt token did not validate")
	}
}

func BenchmarkValidateJwtNoClaims(b *testing.B) {
	os.Setenv("LOGLEVEL", "DEBUG")
	initLogger()

	Runner = mockJwtRuntime("jwt",
		"RS256",
		"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAuvtFgDnIcdB/jqSLICns\nz7FXU/uiFSdJGVpGc5Dy+xm8wZwgiy6lJdL9/TtYjnmJefkPVyYdazabvGvOcns7\n3rshkt0g6Ackqa72yiUEsv1kzCvBObPYNXgr1dNda8/F/ZiO3V9BtcTgQs9Y6rdO\nWJq7zNpees8pfuhEamk3sQp8AmKImFNfuZceNeglMHLLt0NcmSQp4VmhDCladFa1\nEdLirtFM9BtEIOlX20SRcN1LjeRsos8JywpQRxe6M3bnGFXcDQHqrsvwkkzu+vBt\nnPFa2e+jkBSDWkf6ZwvdJnEEUiJkHYTgJuXD1sbGeUkQL1Jb5NaQHhQ1mt3xn1z0\ntwIDAQAB\n-----END PUBLIC KEY-----\n",
		"")

	proxy := mockProxy([]byte(""),
		"0",
		"/path",
		"/path",
		"/get",
		"eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6Ims0MiJ9.eyJpc3MiOiJqb2UiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwianRpIjoiNGFkOTcxMDMtZmQ4ZC00Zjc5LWIwNmEtZTRmOTE1NzkxYjUyIiwiaWF0IjoxNjA3ODU1MTU3fQ.irL3sYTzkM4yFGKBfTzoAAe5H7mGaHECXFy-lkOfVoaaPwuL29b-je_ROoeR8uqw_441QE2P-Ky5582tG2dcu7s3EC3FNuPN_CaZPmbhzV8YIKdzRY7GiTj9sij1_2uRB61b5Qns7H3AJjMuZeCcaGA9t3gSJlVZwkpy9qU46JpX13SPqdSSR9Sg2kZhNFmrRDM5ZGN2VMuzvK34dW72NUkHVaBJUmIRASAfKQnA39xGMskTjP8ZZSzGdYiu0MMhBCA9DZmiS9YBw2Sj6J6Vo7_6SyKAoQyd44JACWbM28jZpfSWDPe-nkMu5ccxNa3A4ocFibnGXaKItWER9MTfeA",
		"jwt")

	for i := 0; i < b.N; i++ {
		ok := proxy.validateJwt()
		if !ok {
			b.Errorf("jwt token did not validate")
		}
	}
}

func BenchmarkValidateJWT0Claim(b *testing.B) {
	doBenchValidateJwtWithClaims(b, ``)
}

func BenchmarkValidateJWT1Claim(b *testing.B) {
	doBenchValidateJwtWithClaims(b, `.sub | select(.=="admin")`)
}

func BenchmarkValidateJWT2Claims(b *testing.B) {
	doBenchValidateJwtWithClaims(b, `.sub | select(.=="subscriber")`,
		`.sub | select(.=="admin")`)
}

func BenchmarkValidateJWT3Claims(b *testing.B) {
	doBenchValidateJwtWithClaims(b, `.sub | select(.=="subscriber")`,
		`.sub | select(.=="developer")`,
		`.sub | select(.=="admin")`)
}

func BenchmarkValidateJWT10Claims(b *testing.B) {
	doBenchValidateJwtWithClaims(b, `.sub | select(.=="subscriber")`,
		`.sub | select(.=="developer")`,
		`.sub | select(.=="developer1")`,
		`.sub | select(.=="developer2")`,
		`.sub | select(.=="developer3")`,
		`.sub | select(.=="developer4")`,
		`.sub | select(.=="developer5")`,
		`.sub | select(.=="developer6")`,
		`.sub | select(.=="developer7")`,
		`.sub | select(.=="admin")`)
}

//.9ms this is too slow
func BenchmarkValidateJWT25Claims(b *testing.B) {
	doBenchValidateJwtWithClaims(b, `.sub | select(.=="subscriber")`,
		`.sub | select(.=="developer")`,
		`.sub | select(.=="developer1")`,
		`.sub | select(.=="developer2")`,
		`.sub | select(.=="developer3")`,
		`.sub | select(.=="developer4")`,
		`.sub | select(.=="developer5")`,
		`.sub | select(.=="developer6")`,
		`.sub | select(.=="developer7")`,
		`.sub | select(.=="developer8")`,
		`.sub | select(.=="developer9")`,
		`.sub | select(.=="developer10")`,
		`.sub | select(.=="developer11")`,
		`.sub | select(.=="developer12")`,
		`.sub | select(.=="developer13")`,
		`.sub | select(.=="developer14")`,
		`.sub | select(.=="developer15")`,
		`.sub | select(.=="developer16")`,
		`.sub | select(.=="developer17")`,
		`.sub | select(.=="developer18")`,
		`.sub | select(.=="developer19")`,
		`.sub | select(.=="developer20")`,
		`.sub | select(.=="developer21")`,
		`.sub | select(.=="developer22")`,
		`.sub | select(.=="developer23")`,
		`.sub | select(.=="developer24")`,
		`.sub | select(.=="developer25")`,
		`.sub | select(.=="admin")`)
}

func doBenchValidateJwtWithClaims(b *testing.B, claims ...string) {
	os.Setenv("LOGLEVEL", "DEBUG")
	initLogger()

	Runner = mockJwtRuntime("jwty",
		"RS256",
		"-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n",
		claims...,
	)

	token := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6ImszIn0.eyJpc3MiOiJpc3N1ZXIiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwic3ViIjoiYWRtaW4iLCJub25jZSI6InNlc3Npb25JRCIsImp0aSI6IjFkYmM3ZjBmLWY1NjQtNGYyNy04MWViLWVhOTJkYTU5NzgzMSIsImlhdCI6MTYwODY2ODY4N30.W3iv8DDmwkMyt5M5KmKgJ1-Tns1LBm7ov-AdzDulDkp7NhrAqrtg6SnLr3KlBCqJQFh84PYEZ2uuOTapEkXkL2xiRIdEenumWMl_65qazpvdbkWWnZ52FP4tHH-3pWWcb0VEc1HSAJuvFN7pDO1Y9lIYeMPGAJY_4nRDHj-60MTNhd4MP6tf29wyBjvaHRlq1a6dCvPhNj6LESqTrGq1DnrvsdZf2FXHPDsv_DvbLOeh_l4-A1hKbrae7OTFYJijbfLwYNs3B12dUxHJ_bSyLmV84kAPZk-IBVUhusx2kbLEVEKT1upblv9ltgnmnsgbwSv3ClYr_1VPOTvpZDSMxhf2zTHIo1W7R0ZdF5f7aFSNmKW59ya5gUHgK9XjEKryyQuUXU2FCJXDARKGie-4VvHZiJo0Nv2De6PGutB_cXjPRs9lyFVui6XtakMaDKVUrE1BwjyXRlf0cGSARTv3wC9x2VmW1ZuoHm9mYCUV3dZiQ0M0gLjZZcLWF4Jq8MtLl-d0hjq5VoBqmmOBnga6JFROFom8Y0ak-5tRXbpJ67GBgyNTXuJ3iBOUXs0Od3t9ZjUfPQElii1q19pac9vtHsfMp9Otur6tKukHvPC-6kLKM4z0OpzvgaMQm7YhlV882GEFaSviW3hYMtyiwT9Ib3FPPsyGySTQWl-4QLk-b3o"

	proxy := mockProxy([]byte(""),
		"0",
		"/path",
		"/path",
		"/get",
		token,
		"jwty")

	parsed, err2 := jwt.Parse([]byte(token))
	if err2 != nil {
		b.Errorf("token not parsed %v", err2)
	}

	for i := 0; i < b.N; i++ {
		err := proxy.verifyMandatoryJwtClaims(parsed, log.Trace())
		if err != nil {
			b.Errorf("jwt token did not validate")
		}
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
	p := mockProxy(make([]byte, 1), "", routePath, transform, requestUri, "", "")
	got := p.resolveUpstreamURI()
	want = "http://upstreamhost:8080" + want
	if got != want {
		t.Errorf("path transformation error, got %s, want %s", got, want)
	}
}

func mockJwtRuntime(jwtName string, alg string, key string, claims ...string) *Runtime {
	jwtConfig := NewJwt(jwtName,
		alg,
		key,
		"",
		"121",
		claims...)

	jwaAlg := *new(jwa.SignatureAlgorithm)
	jwaAlg.Accept(jwtConfig.Alg)

	jwtConfig.parseKey(jwaAlg)
	jwtConfig.Validate()

	return &Runtime{
		Config: Config{
			Jwt: map[string]*Jwt{
				jwtName: jwtConfig,
			},
		},
		ReloadableCert: NewReloadableCert(),
	}
}

func mockProxy(upBody []byte, cl string, path string, transform string, requestUri string, bearer string, jwtName string) Proxy {
	pr, _ := regexp.Compile(path)
	req, _ := http.NewRequest("GET", "/blah", nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", bearer))

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
			Req:    req,
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
			Jwt:       jwtName,
		},
	}
	return proxy
}
