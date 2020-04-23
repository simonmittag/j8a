package jabba

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

//this Mock object wraps the httpClient and prevents actuall outgoing HTTP requests.
type HttpMock struct{}

func (m *HttpMock) Do(req *http.Request) (*http.Response, error) {
	return canned200Ok()
}

func (m *HttpMock) Get(uri string) (*http.Response, error) {
	return canned200Ok()
}

func canned200Ok() (*http.Response, error) {
	json := `{"key":"value"}`
	return &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(json))),
	}, nil
}

//this testHandler binds the mock HTTP server to proxyHandler.
type TestHttpHandler struct{}

func (t TestHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	proxyHandler(w,r)
}

func TestProxyHandler(t *testing.T) {
	// Create Runtime with hardcoded config.
	Runner = getRuntime()

	// create a Mock http client that returns hardcoded 200 ok
	httpClient = &HttpMock{}

	//create a Mock server that is bound to a real proxyhandler
	h := &TestHttpHandler{}
	server := httptest.NewServer(h)
	defer server.Close()

	// Test all of Jabba's goodness, use Runner config to match a route, find a resource, map an upstream server, use
	// proxy object to serve upstream response.
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	//and here we want to receive the 200 ok from the Mock upstream to prove that proxyHandler did it's job based
	//on the config.
	want := 200
	if resp.StatusCode != want {
		t.Fatalf("uh oh, received incorrect status code from proxyhandler, want %v, got %v", want, resp.StatusCode)
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
					MaxAttempts:          1,
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
						Name:   "upstream",
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