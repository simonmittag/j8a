package jabba

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
type HttpMock struct{}

func (m *HttpMock) Do(req *http.Request) (*http.Response, error) {
	return mockDoFunc(req)
}

func (m *HttpMock) Get(uri string) (*http.Response, error) {
	return mockGetFunc(uri)
}

//this testHandler binds the mock HTTP server to proxyHandler.
type TestHttpHandler struct{}

func (t TestHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	proxyHandler(w, r)
}

// TestUpstreamSuccessWithProxyHandler mocks a 200 upstream response and tests the proxy handler returns clean.
func TestUpstreamSuccessWithProxyHandler(t *testing.T) {
	Runner = getRuntime()
	httpClient = &HttpMock{}
	mockDoFunc = func(req *http.Request) (*http.Response, error) {
		json := `{"key":"value"}`
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(json))),
		}, nil
	}

	h := &TestHttpHandler{}
	server := httptest.NewServer(h)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	want := 200
	if resp.StatusCode != want {
		t.Fatalf("uh oh, received incorrect status code from success proxyhandler, want %v, got %v", want, resp.StatusCode)
	}
}

func TestUpstreamPOSTNonRetryWithProxyHandler(t *testing.T) {
	Runner = getRuntime()
	httpClient = &HttpMock{}
	mockDoFunc = func(req *http.Request) (*http.Response, error) {
		json := `{"404":"not found"}`
		return &http.Response{
			StatusCode: 404,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(json))),
		}, nil
	}

	h := &TestHttpHandler{}
	server := httptest.NewServer(h)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	want := 404
	if resp.StatusCode != want {
		t.Fatalf("uh oh, received incorrect status code from non retrying failing proxyhandler, want %v, got %v", want, resp.StatusCode)
	}
}

// TestUpstreamRetryWithProxyHandler mocks a 500 upstream response for a GET request, which is repeatable.
// it will repeat upstream attempts until MaxAttempts is reached, then return a 502 gateway error
func TestUpstreamGETRetryWithProxyHandler(t *testing.T) {
	Runner = getRuntime()
	httpClient = &HttpMock{}
	mockDoFunc = func(req *http.Request) (*http.Response, error) {
		json := `{"500":"server error"}`
		return &http.Response{
			StatusCode: 500,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(json))),
		}, nil
	}

	h := &TestHttpHandler{}
	server := httptest.NewServer(h)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	want := 502
	if resp.StatusCode != want {
		t.Fatalf("uh oh, received incorrect status code from retrying failing proxyhandler, want %v, got %v", want, resp.StatusCode)
	}
}

func getRuntime() *Runtime {
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
			},
			Routes: []Route{
				{
					Path:     "/",
					Resource: "default",
					Policy:   "simple",
				},
			},
		},
	}
}
