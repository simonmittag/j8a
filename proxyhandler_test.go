package j8a

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	mockDoFunc  func(*http.Request) (*http.Response, error)
	mockGetFunc func(uri string) (*http.Response, error)
)

//this Mock object wraps the httpClient and prevents actual outgoing HTTP requests.
type MockHttp struct{}

func (m *MockHttp) Do(req *http.Request) (*http.Response, error) {
	return mockDoFunc(req)
}

func (m *MockHttp) Get(uri string) (*http.Response, error) {
	return mockGetFunc(uri)
}

//this testHandler binds the mock HTTP server to proxyHandler.
type ProxyHttpHandler struct{}

func (t ProxyHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	proxyHandler(w, r)
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
				"Content-Encoding": []string{"gzip"},
			},
			Body: ioutil.NopCloser(bytes.NewReader(Gzip([]byte(json)))),
		}, nil
	}

	server := httptest.NewServer(&ProxyHttpHandler{})
	defer server.Close()

	c := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set("Accept-Encoding", "gzip")
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	gotBody, _ := ioutil.ReadAll(resp.Body)
	if c := bytes.Compare(gotBody[0:2], gzipMagicBytes); c != 0 {
		t.Errorf("uh, oh, body did not have gzip response magic bytes, want %v, got %v", gzipMagicBytes, gotBody[0:2])
	}

	want := "gzip"
	got := resp.Header["Content-Encoding"][0]
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
	req.Header.Set("Accept-Encoding", "identity")
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
				"Content-Encoding": []string{"identity"},
			},
			Body: ioutil.NopCloser(bytes.NewReader([]byte(json))),
		}, nil
	}

	server := httptest.NewServer(&ProxyHttpHandler{})
	defer server.Close()

	c := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set("Accept-Encoding", "identity")
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	gotBody, _ := ioutil.ReadAll(resp.Body)
	if c := bytes.Compare(gotBody[0:2], gzipMagicBytes); c == 0 {
		t.Errorf("body should not have gzip response magic bytes: %v", gotBody[0:2])
	}

	want := "identity"
	got := resp.Header["Content-Encoding"][0]
	if got != want {
		t.Errorf("uh oh, did not receive correct Content-Encoding header, want %v, got %v", want, got)
	}
}

// mocks upstream response with custom content encoding that is passed-through as-is
func TestUpstreamCustomEncodingPassThroughWithBadAcceptEncoding(t *testing.T) {
	Runner = mockRuntime()
	httpClient = &MockHttp{}
	mockDoFunc = func(req *http.Request) (*http.Response, error) {
		json := `{"key":"value"}`
		return &http.Response{
			StatusCode: 200,
			Header: map[string][]string{
				"Content-Encoding": []string{"custom"},
			},
			Body: ioutil.NopCloser(bytes.NewReader([]byte(json))),
		}, nil
	}

	server := httptest.NewServer(&ProxyHttpHandler{})
	defer server.Close()

	c := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set("Accept-Encoding", "uh-oh")
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	gotBody, _ := ioutil.ReadAll(resp.Body)
	if c := bytes.Compare(gotBody[0:2], gzipMagicBytes); c == 0 {
		t.Errorf("body should not have gzip response magic bytes: %v", gotBody[0:2])
	}

	want := "custom"
	got := resp.Header["Content-Encoding"][0]
	if got != want {
		t.Errorf("uh oh, did not receive correct Content-Encoding header, want %v, got %v", want, got)
	}
}

// mocks upstream response with custom encoding that is passed through to client that accepts identity
func TestUpstreamCustomEncodingPassThroughWithIdentityAcceptEncoding(t *testing.T) {
	Runner = mockRuntime()
	httpClient = &MockHttp{}
	mockDoFunc = func(req *http.Request) (*http.Response, error) {
		json := `{"key":"value"}`
		return &http.Response{
			StatusCode: 200,
			Header: map[string][]string{
				"Content-Encoding": []string{"custom"},
			},
			Body: ioutil.NopCloser(bytes.NewReader([]byte(json))),
		}, nil
	}

	server := httptest.NewServer(&ProxyHttpHandler{})
	defer server.Close()

	c := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set("Accept-Encoding", "identity")
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	gotBody, _ := ioutil.ReadAll(resp.Body)
	if c := bytes.Compare(gotBody[0:2], gzipMagicBytes); c == 0 {
		t.Errorf("body should not have gzip response magic bytes: %v", gotBody[0:2])
	}

	want := "custom"
	got := resp.Header["Content-Encoding"][0]
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
				"Content-Encoding": []string{"identity"},
			},
			Body: ioutil.NopCloser(bytes.NewReader([]byte(json))),
		}, nil
	}

	server := httptest.NewServer(&ProxyHttpHandler{})
	defer server.Close()

	c := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set("Accept-Encoding", "gzip")
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	gotBody, _ := ioutil.ReadAll(resp.Body)
	if c := bytes.Compare(gotBody[0:2], gzipMagicBytes); c != 0 {
		t.Errorf("uh, oh, body did not have gzip response magic bytes, want %v, got %v", gzipMagicBytes, gotBody[0:2])
	}

	want := "gzip"
	got := resp.Header["Content-Encoding"][0]
	if got != want {
		t.Errorf("uh oh, did not receive correct Content-Encoding header, want %v, got %v", want, got)
	}
}

// mocks upstream gzip that is re-decoded as identity by j8a
func TestUpstreamGzipReDecoding(t *testing.T) {
	Runner = mockRuntime()
	httpClient = &MockHttp{}
	mockDoFunc = func(req *http.Request) (*http.Response, error) {
		json := `{"key":"value"}`
		return &http.Response{
			StatusCode: 200,
			Header: map[string][]string{
				"Content-Encoding": []string{"gzip"},
			},
			Body: ioutil.NopCloser(bytes.NewReader(Gzip([]byte(json)))),
		}, nil
	}

	server := httptest.NewServer(&ProxyHttpHandler{})
	defer server.Close()

	c := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set("Accept-Encoding", "identity")
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	gotBody, _ := ioutil.ReadAll(resp.Body)
	if c := bytes.Compare(gotBody[0:2], gzipMagicBytes); c == 0 {
		t.Errorf("body should not have gzip response magic bytes, got %v", gotBody[0:2])
	}

	want := "identity"
	got := resp.Header["Content-Encoding"][0]
	if got != want {
		t.Errorf("uh oh, did not receive correct Content-Encoding header, want %v, got %v", want, got)
	}
}

// tests upstream headers are rewritten
func TestUpstreamHeadersAreRewritten(t *testing.T) {
	Runner = mockRuntime()
	httpClient = &MockHttp{}
	mockDoFunc = func(req *http.Request) (*http.Response, error) {
		json := `{"key":"value"}`
		return &http.Response{
			StatusCode: 200,
			Header: map[string][]string{
				"Cookie": []string{"$Version=1; Skin=new"},
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

	want := "$Version=1; Skin=new"
	got := resp.Header["Cookie"][0]
	if got != want {
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

func mockRuntime() *Runtime {
	return &Runtime{
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
					Mode:                    "HTTP",
					Port:                    65534,
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
							Port:   8083,
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
							Port:   8084,
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
	}
}
