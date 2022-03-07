package content

import (
	"fmt"
	"github.com/simonmittag/j8a"
	"github.com/simonmittag/j8a/integration"
	"io/ioutil"
	"net"
	"net/http"
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
	resp := DownstreamContentEncodingIntegrity("identity", true, "identity", "/", t)
	raw := string(resp)
	if !strings.Contains(raw, "404") {
		t.Errorf("identity response should contain 404 in body")
	}
}

func TestBadEncodingOn404Sends406Instead(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("bad", true, "406", "/", t)
}

func TestIdentityCOMMABadEncodingOn404(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("identity, bad", true, "identity", "/", t)
}

func TestIdentityCOMMAGzipEncodingOn404(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("identity, gzip", true, "identity", "/", t)
}

func TestGzipEncodingOn404(t *testing.T) {
	resp := DownstreamContentEncodingIntegrity("gzip", true, "gzip", "/", t)
	raw := string(*j8a.Gunzip(resp))
	if !strings.Contains(raw, "404") {
		t.Errorf("gzip response should contain 404 in body")
	}
}

func TestBrotliEncodingOn404(t *testing.T) {
	resp := DownstreamContentEncodingIntegrity("br", true, "br", "/", t)
	raw := string(*j8a.BrotliDecode(resp))
	if !strings.Contains(raw, "404") {
		t.Errorf("brotli response should contain 404 in body")
	}
}

func TestDeflateEncodingOn404(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("deflate", true, "406", "/", t)
	DownstreamAcceptEncodingContentEncodingHTTP11("x-deflate", true, "406", "/", t)
}

func TestCompressEncodingOn404(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("compress", true, "406", "/", t)
	DownstreamAcceptEncodingContentEncodingHTTP11("x-compress", true, "406", "/", t)
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
	resp := DownstreamContentEncodingIntegrity("identity", true, "identity", "/mse6/get", t)
	raw := string(resp)
	if !strings.Contains(raw, "mse6") {
		t.Errorf("identity response should contain mse6 response")
	}
}

func TestIdentityCOMMAGzipEncodingOnProxyHandler(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("identity, gzip", true, "gzip", "/mse6/get", t)
}

func TestIdentityCOMMABadEncodingOnProxyHandler(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("badd, identity, gzip, bad", true, "gzip", "/mse6/get", t)
}

func TestBadEncodingOnProxyHandlerSends406(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("badd", true, "406", "/mse6/get", t)
}

func TestGzipEncodingOnProxyHandler(t *testing.T) {
	resp := DownstreamContentEncodingIntegrity("gzip", true, "gzip", "/mse6/get", t)
	raw := string(*j8a.Gunzip(resp))
	if !strings.Contains(raw, "mse6") {
		t.Errorf("gzip response should contain mse6 response")
	}
}

func TestBrotliEncodingOnProxyHandler(t *testing.T) {
	resp := DownstreamContentEncodingIntegrity("br", true, "br", "/mse6/get", t)
	raw := string(*j8a.BrotliDecode(resp))
	if !strings.Contains(raw, "mse6") {
		t.Errorf("brotli response should contain mse6 response")
	}
}

func TestDeflateEncodingOnProxyHandler(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("deflate", true, "406", "/mse6/get", t)
	DownstreamAcceptEncodingContentEncodingHTTP11("x-deflate", true, "406", "/mse6/get", t)
}

func TestCompressEncodingOnProxyHandler(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("compress", true, "406", "/mse6/get", t)
	DownstreamAcceptEncodingContentEncodingHTTP11("x-compress", true, "406", "/mse6/get", t)
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

func TestIdentityCOMMAGzipEncodingOnAboutHandler(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("identity, gzip", true, "identity", "/about", t)
}

func TestIdentityCOMMABadEncodingOnAboutHandler(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("deflate, identity", true, "identity", "/about", t)
}

func TestBadEncodingOnAboutHandlerSends406(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("baddddd", true, "406", "/about", t)
}

func TestIdentityEncodingOnAboutHandler(t *testing.T) {
	resp := DownstreamContentEncodingIntegrity("identity", true, "identity", "/about", t)
	raw := string(resp)
	if !strings.Contains(raw, "j8a") {
		t.Errorf("identity aboutresponse should contain j8a in body")
	}
}

func TestGzipEncodingOnAboutHandler(t *testing.T) {
	resp := DownstreamContentEncodingIntegrity("gzip", true, "gzip", "/about", t)
	raw := string(*j8a.Gunzip(resp))
	if !strings.Contains(raw, "j8a") {
		t.Errorf("gzip aboutresponse should contain j8a in body")
	}
}

func TestBrotliEncodingOnAboutHandler(t *testing.T) {
	resp := DownstreamContentEncodingIntegrity("br", true, "br", "/about", t)
	raw := string(*j8a.BrotliDecode(resp))
	if !strings.Contains(raw, "j8a") {
		t.Errorf("brotli aboutresponse should contain j8a in body")
	}
}

func TestDeflateEncodingOnAboutHandler(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("deflate", true, "406", "/about", t)
	DownstreamAcceptEncodingContentEncodingHTTP11("x-deflate", true, "406", "/about", t)
}

func TestCompressEncodingOnAboutHandler(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("compress", true, "406", "/about", t)
	DownstreamAcceptEncodingContentEncodingHTTP11("x-compress", true, "406", "/about", t)
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

func DownstreamContentEncodingIntegrity(ae string, sendAEHeader bool, ce string, slug string, t *testing.T) []byte {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://localhost:8080"+slug, nil)
	if sendAEHeader {
		req.Header.Add(j8a.AcceptEncodingS, ae)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("error connecting to server, cause: %s", err)
	}

	gotce := resp.Header.Get("Content-Encoding")
	if gotce != ce {
		t.Errorf("want content encoding %s, but got %s instead", ce, gotce)
	}

	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	body, _ := ioutil.ReadAll(resp.Body)
	return body
}
