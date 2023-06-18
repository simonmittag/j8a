package j8a

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

var (
	mockDoFunc  func(*http.Request) (*http.Response, error)
	mockGetFunc func(uri string) (*http.Response, error)
)

// this Mock object wraps the httpClient and prevents actual outgoing HTTP requests.
type MockHttp struct{}

func (m *MockHttp) Do(req *http.Request) (*http.Response, error) {
	return mockDoFunc(req)
}

func (m *MockHttp) Get(uri string) (*http.Response, error) {
	return mockGetFunc(uri)
}

// this testHandler binds the mock HTTP server to proxyHandler.
type ProxyHttpHandler struct{}

func (t ProxyHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	proxyHandler(w, r, handleHTTP)
}

func TestIllegalRequestMethod(t *testing.T) {
	Runner = mockRuntime()
	httpClient = &MockHttp{}
	mockDoFunc = func(req *http.Request) (*http.Response, error) {
		json := `{"key":"value"}`
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(json))),
		}, nil
	}

	h := &ProxyHttpHandler{}
	server := httptest.NewServer(h)
	defer server.Close()

	c := &http.Client{}
	req, _ := http.NewRequest("BAD", server.URL, nil)
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 400 {
		t.Errorf("server did not return 400 error Code on illegal request method")
	}
}

// mocks a 200 upstream response and tests the proxy handler returns clean.
func TestUpstreamSuccess(t *testing.T) {
	Runner = mockRuntime()
	httpClient = &MockHttp{}
	mockDoFunc = func(req *http.Request) (*http.Response, error) {
		json := `{"key":"value"}`
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(json))),
		}, nil
	}

	h := &ProxyHttpHandler{}
	server := httptest.NewServer(h)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	want := 200
	if resp.StatusCode != want {
		t.Fatalf("uh oh, received incorrect status Code from success proxyhandler, want %v, got %v", want, resp.StatusCode)
	}
}

// mocks upstream gzip response that is passed through as gzip by j8a
func TestUpstreamGzipEncodingPassThrough(t *testing.T) {
	Runner = mockRuntime()
	httpClient = &MockHttp{}
	mockDoFunc = func(req *http.Request) (*http.Response, error) {
		json := `{"key":"value"}`
		return &http.Response{
			StatusCode: 200,
			Header: map[string][]string{
				contentEncoding: []string{"gzip"},
			},
			Body: ioutil.NopCloser(bytes.NewReader(*Gzip([]byte(json)))),
		}, nil
	}

	server := httptest.NewServer(&ProxyHttpHandler{})
	defer server.Close()

	c := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set(acceptEncoding, "gzip")
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	gotBody, _ := ioutil.ReadAll(resp.Body)
	if c := bytes.Compare(gotBody[0:2], gzipMagicBytes); c != 0 {
		t.Errorf("uh, oh, body did not have gzip response magic bytes, want %v, got %v", gzipMagicBytes, gotBody[0:2])
	}

	want := "gzip"
	got := resp.Header[contentEncoding][0]
	if got != want {
		t.Errorf("uh oh, did not receive correct Content-Encoding header, want %v, got %v", want, got)
	}
}

func TestUpstreamGzipEncodingPassThroughWhenBrotliRequestedDownstream(t *testing.T) {
	Runner = mockRuntime()
	httpClient = &MockHttp{}
	mockDoFunc = func(req *http.Request) (*http.Response, error) {
		json := `{"key":"value"}`
		return &http.Response{
			StatusCode: 200,
			Header: map[string][]string{
				contentEncoding: []string{"gzip"},
			},
			Body: ioutil.NopCloser(bytes.NewReader(*Gzip([]byte(json)))),
		}, nil
	}

	server := httptest.NewServer(&ProxyHttpHandler{})
	defer server.Close()

	c := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set(acceptEncoding, "gzip, br")
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	gotBody, _ := ioutil.ReadAll(resp.Body)
	if c := bytes.Compare(gotBody[0:2], gzipMagicBytes); c != 0 {
		t.Errorf("uh, oh, body did not have gzip response magic bytes, want %v, got %v", gzipMagicBytes, gotBody[0:2])
	}

	want := "gzip"
	got := resp.Header[contentEncoding][0]
	if got != want {
		t.Errorf("uh oh, did not receive correct Content-Encoding header, want %v, got %v", want, got)
	}
}

func TestUpstreamServerHeaderNotCopied(t *testing.T) {
	Runner = mockRuntime()
	httpClient = &MockHttp{}
	mockDoFunc = func(req *http.Request) (*http.Response, error) {
		json := `{"key":"value"}`
		return &http.Response{
			StatusCode: 200,
			Header: map[string][]string{
				"Server": []string{"r5d4"},
			},
			Body: ioutil.NopCloser(bytes.NewReader([]byte(json))),
		}, nil
	}

	server := httptest.NewServer(&ProxyHttpHandler{})
	defer server.Close()

	c := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set(acceptEncoding, "identity")
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	want := "j8a " + Version + " " + ID
	got := resp.Header["Server"][0]
	if got != want {
		t.Errorf("j8a did not send it's own Server header, want: %v, got: %v", want, got)
	}
}

// mocks upstream identity response that is passed-through as-is
func TestUpstreamIdentityEncodingPassThrough(t *testing.T) {
	Runner = mockRuntime()
	httpClient = &MockHttp{}
	mockDoFunc = func(req *http.Request) (*http.Response, error) {
		json := `{"key":"value"}`
		return &http.Response{
			StatusCode: 200,
			Header: map[string][]string{
				contentEncoding: []string{"identity"},
			},
			Body: ioutil.NopCloser(bytes.NewReader([]byte(json))),
		}, nil
	}

	server := httptest.NewServer(&ProxyHttpHandler{})
	defer server.Close()

	c := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set(acceptEncoding, "identity")
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	gotBody, _ := ioutil.ReadAll(resp.Body)
	if c := bytes.Compare(gotBody[0:2], gzipMagicBytes); c == 0 {
		t.Errorf("body should not have gzip response magic bytes: %v", gotBody[0:2])
	}

	want := "identity"
	got := resp.Header[contentEncoding][0]
	if got != want {
		t.Errorf("uh oh, did not receive correct Content-Encoding header, want %v, got %v", want, got)
	}
}

// mocks upstream response with custom content encoding that is passed-through as-is
func TestContentNegotiationFailsWithBadAcceptEncoding(t *testing.T) {
	Runner = mockRuntime()
	httpClient = &MockHttp{}
	mockDoFunc = func(req *http.Request) (*http.Response, error) {
		json := `{"key":"value"}`
		return &http.Response{
			StatusCode: 200,
			Header: map[string][]string{
				contentEncoding: []string{"identity"},
			},
			Body: ioutil.NopCloser(bytes.NewReader([]byte(json))),
		}, nil
	}

	server := httptest.NewServer(&ProxyHttpHandler{})
	defer server.Close()

	c := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set(acceptEncoding, "uh-oh")
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	gotBody, _ := ioutil.ReadAll(resp.Body)
	if c := bytes.Compare(gotBody[0:2], gzipMagicBytes); c == 0 {
		t.Errorf("body should not have gzip response magic bytes: %v", gotBody[0:2])
	}

	want := 406
	got := resp.StatusCode
	if got != want {
		t.Errorf("uh oh, did not receive bad content encoding status code, want %v, got %v", want, got)
	}
}

// mocks upstream response with custom encoding that is passed through to client that accepts identity
func TestUpstreamCustomEncodingPassThroughWithIdentityAcceptEncodingAndVary(t *testing.T) {
	Runner = mockRuntime()
	httpClient = &MockHttp{}
	mockDoFunc = func(req *http.Request) (*http.Response, error) {
		json := `{"key":"value"}`
		return &http.Response{
			StatusCode: 200,
			Header: map[string][]string{
				contentEncoding: []string{"custom"},
			},
			Body: ioutil.NopCloser(bytes.NewReader([]byte(json))),
		}, nil
	}

	server := httptest.NewServer(&ProxyHttpHandler{})
	defer server.Close()

	c := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set(acceptEncoding, "identity")
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	gotBody, _ := ioutil.ReadAll(resp.Body)
	if c := bytes.Compare(gotBody[0:2], gzipMagicBytes); c == 0 {
		t.Errorf("body should not have gzip response magic bytes: %v", gotBody[0:2])
	}

	want := "custom"
	got := resp.Header[contentEncoding][0]
	if got != want {
		t.Errorf("uh oh, did not receive correct Content-Encoding header, want %v, got %v", want, got)
	}

	vary := resp.Header.Get("Vary")
	if acceptEncoding != vary {
		t.Errorf("should have sent a vary header for accept encoding when changing content encoding")
	}
}

// upstream content encoding header is missing. we assume identity, then gzip for client.
func TestUpstreamMissingContentEncodingIdentityThenGzipReEncoding(t *testing.T) {
	Runner = mockRuntime()
	httpClient = &MockHttp{}
	mockDoFunc = func(req *http.Request) (*http.Response, error) {
		json := `{"key":"value"}`
		return &http.Response{
			StatusCode: 200,
			Header:     map[string][]string{},
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(json))),
		}, nil
	}

	server := httptest.NewServer(&ProxyHttpHandler{})
	defer server.Close()

	c := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set(acceptEncoding, "gzip")
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	gotBody, _ := ioutil.ReadAll(resp.Body)
	if c := bytes.Compare(gotBody[0:2], gzipMagicBytes); c != 0 {
		t.Errorf("uh, oh, body did not have gzip response magic bytes, want %v, got %v", gzipMagicBytes, gotBody[0:2])
	}

	want := "gzip"
	got := resp.Header[contentEncoding][0]
	if got != want {
		t.Errorf("uh oh, did not receive correct Content-Encoding header, want %v, got %v", want, got)
	}
}

// mocks upstream identity that is re-encoded as gzip by j8a
func TestUpstreamGzipReEncoding(t *testing.T) {
	Runner = mockRuntime()
	httpClient = &MockHttp{}
	mockDoFunc = func(req *http.Request) (*http.Response, error) {
		json := `{"key":"value"}`
		return &http.Response{
			StatusCode: 200,
			Header: map[string][]string{
				contentEncoding: []string{"identity"},
			},
			Body: ioutil.NopCloser(bytes.NewReader([]byte(json))),
		}, nil
	}

	server := httptest.NewServer(&ProxyHttpHandler{})
	defer server.Close()

	c := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set(acceptEncoding, "gzip")
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	gotBody, _ := ioutil.ReadAll(resp.Body)
	if c := bytes.Compare(gotBody[0:2], gzipMagicBytes); c != 0 {
		t.Errorf("uh, oh, body did not have gzip response magic bytes, want %v, got %v", gzipMagicBytes, gotBody[0:2])
	}

	want := "gzip"
	got := resp.Header[contentEncoding][0]
	if got != want {
		t.Errorf("uh oh, did not receive correct Content-Encoding header, want %v, got %v", want, got)
	}
}

// mocks upstream identity that is re-encoded as brotli by j8a
func TestUpstreamBrotliReEncoding(t *testing.T) {
	Runner = mockRuntime()
	httpClient = &MockHttp{}

	mockDoFunc = func(req *http.Request) (*http.Response, error) {
		json := `{"key":"value"}`
		return &http.Response{
			StatusCode: 200,
			Header: map[string][]string{
				contentEncoding: []string{"identity"},
			},
			Body: ioutil.NopCloser(bytes.NewReader([]byte(json))),
		}, nil
	}

	server := httptest.NewServer(&ProxyHttpHandler{})
	defer server.Close()

	c := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set(acceptEncoding, "br")
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	gotBody, _ := ioutil.ReadAll(resp.Body)
	if c := bytes.Compare(gotBody[0:2], gzipMagicBytes); c == 0 {
		t.Errorf("uh, oh, body should be br but was gzip %v", gotBody[0:2])
	}

	rawBody := string(*BrotliDecode(gotBody))
	want := `{"key":"value"}`
	if rawBody != want {
		t.Errorf("uh oh, encoded body does not match original")
	}

	want2 := "br"
	got2 := resp.Header[contentEncoding][0]
	if got2 != want2 {
		t.Errorf("uh oh, did not receive correct Content-Encoding header, want %v, got %v", want2, got2)
	}
}

// vary header for incompatible content encoding during negotiation
func TestUpstreamIncompatibleContentEncodingSendVaryHeader(t *testing.T) {
	Runner = mockRuntime()
	httpClient = &MockHttp{}
	mockDoFunc = func(req *http.Request) (*http.Response, error) {
		json := `{"key":"value"}`
		return &http.Response{
			StatusCode: 200,
			Header: map[string][]string{
				contentEncoding: []string{"gzip"},
			},
			Body: ioutil.NopCloser(bytes.NewReader(*Gzip([]byte(json)))),
		}, nil
	}

	server := httptest.NewServer(&ProxyHttpHandler{})
	defer server.Close()

	c := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set(acceptEncoding, "identity")
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	gotBody, _ := ioutil.ReadAll(resp.Body)
	if c := bytes.Compare(gotBody[0:2], gzipMagicBytes); c == 0 {
		t.Logf("normal. upstream sent gzip despite our pleas not to, got %v", gotBody[0:2])
	}

	want := "gzip"
	got := resp.Header.Get(contentEncoding)
	if got != want {
		t.Errorf("upstream should have sent gzip, got %v", got)
	}

	vary := resp.Header.Get("Vary")
	if acceptEncoding != vary {
		t.Errorf("should have sent a vary header for accept encoding when changing content encoding")
	}
}

// tests upstream headers are rewritten
func TestUpstreamHeadersAreRewrittenInOrder(t *testing.T) {
	Runner = mockRuntime()
	httpClient = &MockHttp{}
	mockDoFunc = func(req *http.Request) (*http.Response, error) {
		json := `{"key":"value"}`
		return &http.Response{
			StatusCode: 200,
			Header: map[string][]string{
				"X": []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
			},
			Body: ioutil.NopCloser(bytes.NewReader([]byte(json))),
		}, nil
	}

	h := &ProxyHttpHandler{}
	server := httptest.NewServer(h)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
	got := resp.Header["X"]
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Cookie header was not rewritten, want %v, got %v", want, got)
	}
}

// tests upstream POST requests are not retried.
func TestUpstreamPOSTNonRetry(t *testing.T) {
	Runner = mockRuntime()
	httpClient = &MockHttp{}
	mockDoFunc = func(req *http.Request) (*http.Response, error) {
		json := `{"404":"not found"}`
		return &http.Response{
			StatusCode: 404,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(json))),
		}, nil
	}

	h := &ProxyHttpHandler{}
	server := httptest.NewServer(h)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	want := 404
	if resp.StatusCode != want {
		t.Fatalf("uh oh, received incorrect status Code from non retrying failing proxyhandler, want %v, got %v", want, resp.StatusCode)
	}
}

// TestUpstreamRetryWithProxyHandler mocks a 500 upstream response for a GET request, which is repeatable.
// it will repeat upstream attempts until MaxAttempts is reached, then return a 502 gateway error
func TestUpstreamGETRetry(t *testing.T) {
	Runner = mockRuntime()
	httpClient = &MockHttp{}
	mockDoFunc = func(req *http.Request) (*http.Response, error) {
		json := `{"500":"server error"}`
		return &http.Response{
			StatusCode: 500,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(json))),
		}, nil
	}

	h := &ProxyHttpHandler{}
	server := httptest.NewServer(h)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	want := 502
	if resp.StatusCode != want {
		t.Fatalf("uh oh, received incorrect status Code from retrying failing proxyhandler, want %v, got %v", want, resp.StatusCode)
	}
}

func TestProxyHeaderRewrite(t *testing.T) {
	cl := "conTenT-LEngtH"
	if shouldProxyHeader(cl) {
		t.Errorf("should not proxy %s", cl)
	}
}

func TestJsonifyUpstreamHeadersWithEmptyUp(t *testing.T) {
	res := jsonifyUpstreamHeaders(&Proxy{
		Up: Up{},
	})

	if string(res) != "{}" {
		t.Errorf("should have returned json object")
	}
}

func TestJsonifyUpstreamHeadersWithEmptyProxy(t *testing.T) {
	res := jsonifyUpstreamHeaders(&Proxy{})

	if string(res) != "{}" {
		t.Errorf("should have returned json object")
	}
}

func TestJsonifyUpstreamHeadersWithEmptyAtmpt(t *testing.T) {
	res := jsonifyUpstreamHeaders(&Proxy{
		Up: Up{
			Atmpt: &Atmpt{},
		},
	})

	if string(res) != "{}" {
		t.Errorf("should have returned json object")
	}
}

func TestJsonifyUpstreamHeadersWithNilHeaders(t *testing.T) {
	res := jsonifyUpstreamHeaders(&Proxy{
		Up: Up{
			Atmpt: &Atmpt{
				resp: &http.Response{
					Header: nil,
				},
			},
		},
	})

	if string(res) != "{}" {
		t.Errorf("should have returned json object")
	}
}

func TestJsonifyUpstreamHeadersWithHeader(t *testing.T) {
	h := http.Header{}
	h.Add("K", "v")
	res := jsonifyUpstreamHeaders(&Proxy{
		Up: Up{
			Atmpt: &Atmpt{
				resp: &http.Response{
					Header: h,
				},
			},
		},
	})

	want := `{"K":["v"]}`
	got := string(res)
	if got != want {
		t.Errorf("should have returned json encoded headers, but got: %v", got)
	}
}

func TestParseUpstreamResponseWithNilResponse(t *testing.T) {
	//needed for print function
	Runner = mockRuntime()
	Runner.Connection.Upstream.MaxAttempts = 1

	p := mockProxy(make([]byte, 1), "", "/", "", "/", "", "")
	_, e := parseUpstreamResponse(nil, &p)
	if e == nil {
		t.Errorf("no response error for nil http request")
	}
}

func mockRunner() {
	Runner = mockRuntime()
}

func mockRuntime() *Runtime {
	r := &Runtime{
		Config: Config{
			Connection: Connection{
				Upstream: Upstream{
					ReadTimeoutSeconds:   120,
					SocketTimeoutSeconds: 3,
					IdleTimeoutSeconds:   120,
					MaxAttempts:          3,
					PoolSize:             2,
				},
				Downstream: Downstream{
					IdleTimeoutSeconds:      3,
					ReadTimeoutSeconds:      120,
					RoundTripTimeoutSeconds: 120,
					Http: Http{
						Port:        65534,
						Redirecttls: false,
					},
					Tls: Tls{
						Acme: Acme{
							Domains:         []string{"localhost"},
							GracePeriodDays: 30,
						}},
				},
			},
			Policies: map[string]Policy{
				"simple": {
					LabelWeight{
						Weight: 1,
						Label:  "simple",
					},
				},
			},
			Resources: map[string][]ResourceMapping{
				"default": {
					ResourceMapping{
						Name: "upstream",
						Labels: []string{
							"simple",
						},
						URL: URL{
							Scheme: "http",
							Host:   "localhost",
							Port:   "8083",
						},
					},
				},
				"blahResource": {
					ResourceMapping{
						Name:   "blahResource",
						Labels: nil,
						URL: URL{
							Scheme: "http",
							Host:   "localhost",
							Port:   "8084",
						},
					},
				},
				"ipv4Resource": {
					ResourceMapping{
						Name:   "ipv4Resource",
						Labels: nil,
						URL: URL{
							Scheme: "http",
							Host:   "127.0.0.1",
							Port:   "8084",
						},
					},
				},
				"ipv6Resource": {
					ResourceMapping{
						Name:   "ipv6Resource",
						Labels: nil,
						URL: URL{
							Scheme: "http",
							Host:   "[::1]",
							Port:   "8084",
						},
					},
				},
				"ipv6Resource2": {
					ResourceMapping{
						Name:   "ipv6Resource",
						Labels: nil,
						URL: URL{
							Scheme: "http",
							Host:   "::1",
							Port:   "8085",
						},
					},
				},
			},
			Routes: []Route{
				{
					Path:     "/",
					Resource: "default",
					Policy:   "simple",
				},
				{
					Path:     "/blah",
					Resource: "blahResource",
				},
			},
		},
		Start:             time.Now(),
		AcmeHandler:       NewAcmeHandler(),
		ConnectionWatcher: ConnectionWatcher{dwnOpenConns: 0},
	}

	//simple compiled regexes for prefix matching only
	r.Routes[0].compilePath()
	r.Routes[1].compilePath()

	//we need this to add the reloadable cert.
	r.initReloadableCert()

	//now we can work with you
	return r
}
