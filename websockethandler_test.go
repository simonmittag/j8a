package j8a

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

//this testHandler binds the mock HTTP server to proxyHandler.
type WebsocketHandler struct{}

func (h WebsocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	websocketHandler(w, r)
}

func TestWebSocketHandler(t *testing.T) {
	Runner = mockRuntime()

	h := &WebsocketHandler{}
	server := httptest.NewServer(h)
	defer server.Close()

	c := &http.Client{}
	resp, err := c.Get(server.URL)

	if resp==nil {
		t.Error("no HTTP resonse")
	} else if resp.StatusCode != 502 {
		t.Errorf("wanted 502 for bad gateway but got: %v", err)
	} else if err!=nil {
		t.Errorf("got HTTP error: %v", err)
	}
}

func TestUpgradeWebsocket(t *testing.T) {
	Runner = mockRuntime()

	p := mockProxyWS()
	upgradeWebsocket(&p)

	//test only nil pointer
}

func TestLogExitStatusNetOpError(t *testing.T) {
	mpws := mockProxyWS()
	x := WebsocketStatus{
		DwnOpCode: 0,
		DwnExit: &net.OpError{
			Op:     "",
			Net:    "",
			Source: nil,
			Addr:   nil,
			Err:    io.EOF,
		},
		UpOpCode:  0,
		UpExit:    nil,
	}
	mpws.logWebsocketConnectionExitStatus(x)
}

func mockProxyWS() Proxy {
	return Proxy{
		XRequestID:    "1",
		XRequestDebug: false,
		Up: Up{
			Atmpt: &Atmpt{
				URL: &URL{
					Scheme: "ws",
					Host:   "localhost",
					Port:   80,
				},
			},
		},
		Dwn: Down{
			Req: &http.Request{
				RemoteAddr: "10.1.1.1",
			},
			Resp: Resp{
				Writer: &httptest.ResponseRecorder{},
			},
			Method: "GET",
			Path:   "/path",
		},
		Route: &Route{
			Path:      "/path",
			PathRegex: nil,
			Resource:  "res",
		},
	}
}

