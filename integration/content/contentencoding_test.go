package content

import (
	"fmt"
	"github.com/simonmittag/j8a/integration"
	"net"
	"strings"
	"testing"
)

func TestNoAcceptEncodingOn404(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("", false, "identity", "/", t)
}

func TestEmptyAcceptEncodingOn404(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("", true, "identity", "/", t)
}

func TestStarAcceptEncodingOn404(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("*", true, "identity", "/", t)
}

func TestIdentityEncodingOn404(t *testing.T) {
	DownstreamAcceptEncodingHTTP11("identity", "/", t)
}

func TestGzipEncodingOn404(t *testing.T) {
	DownstreamAcceptEncodingHTTP11("gzip", "/", t)
}

func TestBrotliEncodingOn404(t *testing.T) {
	DownstreamAcceptEncodingHTTP11("br", "/", t)
}

func TestDeflateEncodingOn404(t *testing.T) {
	DownstreamAcceptEncodingHTTP11("deflate", "/", t)
}

func TestNoAcceptEncodingOnProxyHandler(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("", false, "gzip", "/mse6/get", t)
}

func TestEmptyAcceptEncodingOnProxyHandler(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("", true, "gzip", "/mse6/get", t)
}

func TestStarAcceptEncodingOnProxyHandler(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("*", true, "gzip", "/mse6/get", t)
}

func TestIdentityEncodingOnProxyHandler(t *testing.T) {
	DownstreamAcceptEncodingHTTP11("identity", "/mse6/get", t)
}

func TestGzipEncodingOnProxyHandler(t *testing.T) {
	DownstreamAcceptEncodingHTTP11("gzip", "/mse6/get", t)
}

func TestBrotliEncodingOnProxyHandler(t *testing.T) {
	DownstreamAcceptEncodingHTTP11("br", "/mse6/get", t)
}

func TestDeflateEncodingOnProxyHandler(t *testing.T) {
	DownstreamAcceptEncodingHTTP11("deflate", "/mse6/get", t)
}

func TestNoAcceptEncodingOnAboutHandler(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("", false, "identity", "/about", t)
}

func TestEmptyAcceptEncodingOnAboutHandler(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("", true, "identity", "/about", t)
}

func TestStarAcceptEncodingOnAboutHandler(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("*", true, "identity", "/about", t)
}

func TestIdentityEncodingOnAboutHandler(t *testing.T) {
	DownstreamAcceptEncodingHTTP11("identity", "/about", t)
}

func TestGzipEncodingOnAboutHandler(t *testing.T) {
	DownstreamAcceptEncodingHTTP11("gzip", "/about", t)
}

func TestBrotliEncodingOnAboutHandler(t *testing.T) {
	DownstreamAcceptEncodingHTTP11("br", "/about", t)
}

func TestDeflateEncodingOnAboutHandler(t *testing.T) {
	DownstreamAcceptEncodingHTTP11("deflate", "/about", t)
}

func DownstreamAcceptEncodingHTTP11(enc string, slug string, t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11(enc, true, enc, slug, t)
}

func DownstreamAcceptEncodingContentEncodingHTTP11(ae string, sendAEHeader bool, ce string, slug string, t *testing.T) {
	//if this test fails, check the j8a configuration for connection.downstream.ReadTimeoutSeconds

	//step 1 make a connection
	c, err := net.Dial("tcp", ":8080")
	if err != nil {
		t.Errorf("test failure. unable to connect to j8a server for integration test, cause: %v", err)
	}

	//step 2 we send headers and terminate http message so j8a sends request upstream.
	integration.CheckWrite(t, c, fmt.Sprintf("GET %s HTTP/1.1\r\n", slug))
	integration.CheckWrite(t, c, "Host: localhost:8080\r\n")
	if sendAEHeader {
		integration.CheckWrite(t, c, fmt.Sprintf("Accept-Encoding: %v\r\n", ae))
	}
	integration.CheckWrite(t, c, "\r\n")

	//step 3 we read a response into buffer which returns headers
	buf := make([]byte, 128)
	b, err2 := c.Read(buf)
	t.Logf("normal. j8a responded with %v bytes", b)
	if err2 != nil || !strings.Contains(string(buf), ce) {
		t.Errorf("test failure. accept encoding %s want response content encoding %s but got %s", ae, ce, string(buf))
	} else {
		t.Logf("normal. received response %s", string(buf))
	}
}
