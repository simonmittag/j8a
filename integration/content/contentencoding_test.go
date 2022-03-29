package content

import (
	"bytes"
	"compress/flate"
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
	resp := DownstreamContentEncodingIntegrity("identity", true, "identity", false, "/", t)
	raw := string(resp)
	if !strings.Contains(raw, "404") {
		t.Errorf("identity response should contain 404 in body")
	}
}

func TestBadEncodingOn404Sends406Instead(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("bad", true, "406", "/", t)
}

func TestIdentityCOMMABadEncodingOn404SendsIdentity(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("identity, bad", true, "identity", "/", t)
}

func TestIdentityCOMMAGzipEncodingOn404SendsIdentity(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("identity, gzip", true, "identity", "/", t)
}

func TestGzipEncodingOn404SendsEncodedGzip(t *testing.T) {
	resp := DownstreamContentEncodingIntegrity("gzip", true, "gzip", false, "/", t)
	raw := string(*j8a.Gunzip(resp))
	if !strings.Contains(raw, "404") {
		t.Errorf("gzip response should contain 404 in body")
	}
}

func TestBrotliEncodingOn404SendsEncodedBrotli(t *testing.T) {
	resp := DownstreamContentEncodingIntegrity("br", true, "br", false, "/", t)
	raw := string(*j8a.BrotliDecode(resp))
	if !strings.Contains(raw, "404") {
		t.Errorf("brotli response should contain 404 in body")
	}
}

func TestDeflateAcceptEncodingOn404Sends406(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("deflate", true, "406", "/", t)
	DownstreamAcceptEncodingContentEncodingHTTP11("x-deflate", true, "406", "/", t)
}

func TestCompressAcceptEncodingOn404Sends406(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("compress", true, "406", "/", t)
	DownstreamAcceptEncodingContentEncodingHTTP11("x-compress", true, "406", "/", t)
}

func TestNoAcceptEncodingOnProxyHandlerSendsUpstreamIdentityHeader(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("", false, "identity", "/mse6/get", t)
}

func TestEmptyAcceptEncodingOnProxyHandlerSendsUpstreamIdentityHeader(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("", true, "identity", "/mse6/get", t)
}

func TestNoAcceptEncodingOnProxyHandlerSendsJ8aIdentityHeader(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("", false, "identity", "/mse6/nocontentenc", t)
}

func TestEmptyAcceptEncodingOnProxyHandlerSendsJ8aIdentityHeader(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("", true, "identity", "/mse6/nocontentenc", t)
}

func TestStarAcceptEncodingOnProxyHandlerSendsEncodedGzip(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("*", true, "gzip", "/mse6/get", t)
}

func TestIdentityAcceptEncodingOnProxyHandlerSendsUpstreamIdentityHeader(t *testing.T) {
	resp := DownstreamContentEncodingIntegrity("identity", true, "identity", false, "/mse6/get", t)
	raw := string(resp)
	if !strings.Contains(raw, "mse6") {
		t.Errorf("identity response should contain mse6 response")
	}
}

func TestIdentityAcceptEncodingOnProxyHandlerSendsJ8aIdentityHeader(t *testing.T) {
	resp := DownstreamContentEncodingIntegrity("identity", true, "identity", false, "/mse6/nocontentenc", t)
	raw := string(resp)
	if !strings.Contains(raw, "mse6") {
		t.Errorf("identity response should contain mse6 response")
	}
}

func TestStarAcceptEncodingOnProxyHandlerEncodesUpstreamIdentityAsGzip(t *testing.T) {
	resp := DownstreamContentEncodingIntegrity("*", true, "gzip", false, "/mse6/get", t)
	raw := string(*j8a.Gunzip(resp))
	if !strings.Contains(raw, "get") {
		t.Errorf("unable to find get response after gunzip")
	}
}

func TestStarAcceptEncodingOnProxyHandlerEncodesUpstreamNoContentEncodingAsGzip(t *testing.T) {
	resp := DownstreamContentEncodingIntegrity("*", true, "gzip", false, "/mse6/nocontentenc", t)
	raw := string(*j8a.Gunzip(resp))
	if !strings.Contains(raw, "nocontentenc") {
		t.Errorf("unable to find nocontenenc response after gunzip")
	}
}

func TestStarAcceptEncodingOnProxyHandlerSendsUpstreamGzipAsGzip(t *testing.T) {
	resp := DownstreamContentEncodingIntegrity("*", true, "gzip", false, "/mse6/gzip", t)
	raw := string(*j8a.Gunzip(resp))
	if !strings.Contains(raw, "mse6") {
		t.Errorf("unable to find mse6 response after gunzip")
	}
}

func TestStarAcceptEncodingOnProxyHandlerSendsUpstreamBrotliAsBrotli(t *testing.T) {
	resp := DownstreamContentEncodingIntegrity("*", true, "br", false, "/mse6/brotli", t)
	raw := string(*j8a.BrotliDecode(resp))
	if !strings.Contains(raw, "brotli") {
		t.Errorf("unable to find mse6 response after brotli decode")
	}
}

func TestStarAcceptEncodingOnProxyHandlerSendsUpstreamUnknownAsUnknown(t *testing.T) {
	resp := DownstreamContentEncodingIntegrity("*", true, "unknown", false, "/mse6/unknowncontentenc", t)
	raw := string(resp)
	if !strings.Contains(raw, "unknown") {
		t.Errorf("unable to find mse6 response after with unknown content encoding")
	}
}

func TestStarAcceptEncodingOnProxyHandlerSendsUpstreamDeflateAsDeflate(t *testing.T) {
	resp := DownstreamContentEncodingIntegrity("*", true, "deflate", false, "/mse6/deflate", t)
	raw := string(resp)
	if !strings.Contains(raw, "mse6") {
		t.Errorf("unable to find mse6 response")
	}
}

func TestIdentityEncodingOnProxyHandlerUpstreamBrotliPassthrough(t *testing.T) {
	resp := DownstreamContentEncodingIntegrity("identity", true, "br", true, "/mse6/brotli", t)
	raw := string(*j8a.BrotliDecode(resp))
	want := "{\"mse6\":\"Hello from the brotli endpoint\"}"
	if raw != want {
		t.Errorf("upstream brotli should decode and contain mse6 response passed through")
	}
}

func TestIdentityEncodingOnProxyHandlerUpstreamDeflatePassthroughWithVary(t *testing.T) {
	resp := DownstreamContentEncodingIntegrity("identity", true, "deflate", true, "/mse6/deflate", t)
	raw, _ := ioutil.ReadAll(flate.NewReader(bytes.NewBuffer(resp)))
	want := "{\"mse6\":\"Hello from the deflate endpoint\"}"
	if string(raw) != want {
		t.Errorf("upstream deflate shouold decode and contain mse6 response passed through")
	}
}

func TestIdentityEncodingOnProxyHandlerUpstreamGzipPassthroughWithVary(t *testing.T) {
	resp := DownstreamContentEncodingIntegrity("identity", true, "gzip", true, "/mse6/gzip", t)
	raw := string(*j8a.Gunzip(resp))
	want := "{\"mse6\":\"Hello from the gzip endpoint\"}"
	if raw != want {
		t.Errorf("upstream gzip should decode and contain mse6 response passed through")
	}
}

func TestIdentityEncodingOnProxyHandlerUpstreamBrotliPassthroughWithVary(t *testing.T) {
	resp := DownstreamContentEncodingIntegrity("identity", true, "br", true, "/mse6/brotli", t)
	raw := string(*j8a.BrotliDecode(resp))
	want := "{\"mse6\":\"Hello from the brotli endpoint\"}"
	if raw != want {
		t.Errorf("upstream brotli should decode and contain mse6 response passed through")
	}
}

func TestGzipEncodingOnProxyHandlerUpstreamDeflatePassthroughWithVary(t *testing.T) {
	resp := DownstreamContentEncodingIntegrity("gzip", true, "deflate", true, "/mse6/deflate", t)
	raw, _ := ioutil.ReadAll(flate.NewReader(bytes.NewBuffer(resp)))
	want := "{\"mse6\":\"Hello from the deflate endpoint\"}"
	if string(raw) != want {
		t.Errorf("upstream deflate should decode and contain mse6 response passed through")
	}
}

func TestBrotliEncodingOnProxyHandlerUpstreamDeflatePassthroughWithVary(t *testing.T) {
	resp := DownstreamContentEncodingIntegrity("br", true, "deflate", true, "/mse6/deflate", t)
	raw, _ := ioutil.ReadAll(flate.NewReader(bytes.NewBuffer(resp)))
	want := "{\"mse6\":\"Hello from the deflate endpoint\"}"
	if string(raw) != want {
		t.Errorf("upstream deflate should decode and contain mse6 response passed through")
	}
}

func TestIdentityCOMMAGzipEncodingOnProxyHandlerSendsPreferredGzip(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("identity, gzip", true, "gzip", "/mse6/get", t)
}

func TestIdentityCOMMABrotliEncodingOnProxyHandlerSendsPreferredBrotli(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("identity, br", true, "br", "/mse6/get", t)
}

func TestIdentityCOMMABadEncodingOnProxyHandlerIgnoresBadAndSendsPreferredGzip(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("badd, identity, gzip, bad", true, "gzip", "/mse6/get", t)
}

func TestBadEncodingOnProxyHandlerSends406(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("badd", true, "406", "/mse6/get", t)
}

func TestGzipEncodingOnProxyHandlerSendsJ8aEncodedGzip(t *testing.T) {
	resp := DownstreamContentEncodingIntegrity("gzip", true, "gzip", false, "/mse6/get", t)
	raw := string(*j8a.Gunzip(resp))
	if !strings.Contains(raw, "mse6") {
		t.Errorf("gzip response should contain mse6 response")
	}
}

func TestBrotliEncodingOnProxyHandlerSendsJ8aEncodedBrotli(t *testing.T) {
	resp := DownstreamContentEncodingIntegrity("br", true, "br", false, "/mse6/get", t)
	raw := string(*j8a.BrotliDecode(resp))
	if !strings.Contains(raw, "mse6") {
		t.Errorf("brotli response should contain mse6 response")
	}
}

func TestDeflateEncodingOnProxyHandlerSends406(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("deflate", true, "406", "/mse6/get", t)
	DownstreamAcceptEncodingContentEncodingHTTP11("x-deflate", true, "406", "/mse6/get", t)
}

func TestCompressEncodingOnProxyHandlerSends406(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("compress", true, "406", "/mse6/get", t)
	DownstreamAcceptEncodingContentEncodingHTTP11("x-compress", true, "406", "/mse6/get", t)
}

func TestNoAcceptEncodingOnAboutHandlerSendsIdentity(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("", false, "identity", "/about", t)
}

func TestEmptyAcceptEncodingOnAboutHandlerSendsIdentity(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("", true, "identity", "/about", t)
}

func TestStarAcceptEncodingOnAboutHandlerSendsIdentity(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("*", true, "identity", "/about", t)
}

func TestIdentityCOMMAGzipEncodingOnAboutHandlerSendsPreferredIdentity(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("identity, gzip", true, "identity", "/about", t)
}

func TestIdentityCOMMABadEncodingOnAboutHandlerSendsIdentity(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("deflate, identity", true, "identity", "/about", t)
}

func TestBadEncodingOnAboutHandlerSends406(t *testing.T) {
	DownstreamAcceptEncodingContentEncodingHTTP11("baddddd", true, "406", "/about", t)
}

func TestIdentityEncodingOnAboutHandlerSendsIdentity(t *testing.T) {
	resp := DownstreamContentEncodingIntegrity("identity", true, "identity", false, "/about", t)
	raw := string(resp)
	if !strings.Contains(raw, "j8a") {
		t.Errorf("identity aboutresponse should contain j8a in body")
	}
}

func TestGzipEncodingOnAboutHandlerSendsEncodedGzip(t *testing.T) {
	resp := DownstreamContentEncodingIntegrity("gzip", true, "gzip", false, "/about", t)
	raw := string(*j8a.Gunzip(resp))
	if !strings.Contains(raw, "j8a") {
		t.Errorf("gzip aboutresponse should contain j8a in body")
	}
}

func TestBrotliEncodingOnAboutHandlerSendsEncodedBrotli(t *testing.T) {
	resp := DownstreamContentEncodingIntegrity("br", true, "br", false, "/about", t)
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

func DownstreamContentEncodingIntegrity(ae string, sendAEHeader bool, ce string, wantVaryAE bool, slug string, t *testing.T) []byte {
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

	gotVary := resp.Header.Get("Vary")
	if wantVaryAE == true {
		if gotVary != "Accept-Encoding" {
			t.Errorf("want Vary: Accept-Encoding, but got %s instead", gotVary)
		}
	} else {
		if len(gotVary) > 0 {
			t.Errorf("no vary header should be sent, but got %s", gotVary)
		}
	}

	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	body, _ := ioutil.ReadAll(resp.Body)
	return body
}
