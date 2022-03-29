package content

import (
	"bytes"
	"github.com/klauspost/compress/flate"
	"github.com/simonmittag/j8a"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestContentEncodingPermutationsOnProxyHandlerUpstreamGzip(t *testing.T) {
	tests := map[string]struct {
		reqUrlSlug                      string
		reqAcceptEncodingHeader         string
		reqSendAcceptEncodingHeader     bool
		wantResStatusCode               int
		wantResContentEncodingHeader    string
		wantResVaryAcceptEncodingHeader bool
		wantResBodyContent              string
	}{
		"noAcceptEncodingSendsEncoded": {"/mse6/gzip",
			"",
			false,
			200,
			"gzip",
			true,
			"gzip",
		},
		"emptyAcceptEncodingSendsEncoded": {"/mse6/gzip",
			"",
			true,
			200,
			"gzip",
			true,
			"gzip",
		},
		"starAcceptEncodingSendsEncoded": {"/mse6/gzip",
			"*",
			true,
			200,
			"gzip",
			false,
			"gzip",
		},
		"starCommaGzipAcceptEncodingSendsEncoded": {"/mse6/gzip",
			"*,gzip",
			true,
			200,
			"gzip",
			false,
			"gzip",
		},
		"identityCommaGzipAcceptEncodingSendsEncoded": {"/mse6/gzip",
			"identity,gzip",
			true,
			200,
			"gzip",
			false,
			"gzip",
		},
		"gzipCommaIdentityAcceptEncodingSendsEncoded": {"/mse6/gzip",
			"gzip,identity",
			true,
			200,
			"gzip",
			false,
			"gzip",
		},
		"gzipCommaStarAcceptEncodingSendsEncoded": {"/mse6/gzip",
			"gzip,*",
			true,
			200,
			"gzip",
			false,
			"gzip",
		},
		"gzipAcceptEncodingSendsEncoded": {"/mse6/gzip",
			"gzip",
			true,
			200,
			"gzip",
			false,
			"gzip",
		},
		"gzipCommaUnknownAcceptEncodingSendsEncoded": {"/mse6/gzip",
			"gzip, unknown, moreunknown",
			true,
			200,
			"gzip",
			false,
			"gzip",
		},
		"gzipCommaBrotliAcceptEncodingSendsGzip": {"/mse6/gzip",
			"gzip, br",
			true,
			200,
			"gzip",
			false,
			"gzip",
		},
		"brotliAcceptEncodingSendsGzipWithVary": {"/mse6/gzip",
			"br",
			true,
			200,
			"gzip",
			true,
			"gzip",
		},
		"brotliCommaGzipAcceptEncodingSendsGzip": {"/mse6/gzip",
			"br,gzip",
			true,
			200,
			"gzip",
			false,
			"gzip",
		},
		"brotliCommaIdentityAcceptEncodingSendsGzipWithVary": {"/mse6/gzip",
			"br,identity",
			true,
			200,
			"gzip",
			true,
			"gzip",
		},
		"brotliCommaStarAcceptEncodingSendsGzipNoVary": {"/mse6/gzip",
			"br,*",
			true,
			200,
			"gzip",
			false,
			"gzip",
		},
		"deflateAcceptEncodingSends406ResponseCode": {"/mse6/gzip",
			"deflate",
			true,
			406,
			"identity",
			false,
			"406",
		},
		"unknownAcceptEncodingSends406ResponseCode": {"/mse6/gzip",
			"unknown",
			true,
			406,
			"identity",
			false,
			"406",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			DownstreamContentEncodingFullIntegrity(tc.reqUrlSlug, tc.reqAcceptEncodingHeader, tc.reqSendAcceptEncodingHeader,
				tc.wantResContentEncodingHeader, tc.wantResVaryAcceptEncodingHeader, tc.wantResBodyContent, tc.wantResStatusCode, t)
		})
	}
}

func TestContentEncodingPermutationsOnProxyHandlerUpstreamUnencoded(t *testing.T) {
	tests := map[string]struct {
		reqUrlSlug                      string
		reqAcceptEncodingHeader         string
		reqSendAcceptEncodingHeader     bool
		wantResStatusCode               int
		wantResContentEncodingHeader    string
		wantResVaryAcceptEncodingHeader bool
		wantResBodyContent              string
	}{
		"noAcceptEncodingSendsIdentity": {"/mse6/get",
			"",
			false,
			200,
			"identity",
			false,
			"get",
		},
		"emptyAcceptEncodingSendsIdentity": {"/mse6/get",
			"",
			true,
			200,
			"identity",
			false,
			"get",
		},
		"starAcceptEncodingSendsIdentity": {"/mse6/get",
			"*",
			true,
			200,
			"gzip",
			false,
			"get",
		},
		"starCommaGzipAcceptEncodingSendsIdentity": {"/mse6/get",
			"*,gzip",
			true,
			200,
			"gzip",
			false,
			"get",
		},
		"identityCommaGzipAcceptEncodingSendsIdentity": {"/mse6/get",
			"identity,gzip",
			true,
			200,
			"gzip",
			false,
			"get",
		},
		"gzipCommaIdentityAcceptEncodingSendsIdentity": {"/mse6/get",
			"gzip,identity",
			true,
			200,
			"gzip",
			false,
			"get",
		},
		"gzipCommaStarAcceptEncodingSendsIdentity": {"/mse6/get",
			"gzip,*",
			true,
			200,
			"gzip",
			false,
			"get",
		},
		"gzipAcceptEncodingSendsGzip": {"/mse6/get",
			"gzip",
			true,
			200,
			"gzip",
			false,
			"get",
		},
		"gzipCommaUnknownAcceptEncodingSendsGzip": {"/mse6/get",
			"gzip, unknown, moreunknown",
			true,
			200,
			"gzip",
			false,
			"get",
		},
		"gzipCommaBrotliAcceptEncodingSendsGzip": {"/mse6/get",
			"gzip, br",
			true,
			200,
			"gzip",
			false,
			"get",
		},
		"brotliAcceptEncodingSendsBrotli": {"/mse6/get",
			"br",
			true,
			200,
			"br",
			false,
			"get",
		},
		"brotliCommaGzipAcceptEncodingSendsGzip": {"/mse6/get",
			"br,gzip",
			true,
			200,
			"gzip",
			false,
			"get",
		},
		"brotliCommaIdentityAcceptEncodingSendsBrotli": {"/mse6/get",
			"br,identity",
			true,
			200,
			"br",
			false,
			"get",
		},
		"brotliCommaStarAcceptEncodingSendsGzip": {"/mse6/get",
			"br,*",
			true,
			200,
			"gzip",
			false,
			"get",
		},
		"deflateAcceptEncodingSends406ResponseCode": {"/mse6/get",
			"deflate",
			true,
			406,
			"identity",
			false,
			"406",
		},
		"unknownAcceptEncodingSends406ResponseCode": {"/mse6/get",
			"unknown",
			true,
			406,
			"identity",
			false,
			"406",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			DownstreamContentEncodingFullIntegrity(tc.reqUrlSlug, tc.reqAcceptEncodingHeader, tc.reqSendAcceptEncodingHeader,
				tc.wantResContentEncodingHeader, tc.wantResVaryAcceptEncodingHeader, tc.wantResBodyContent, tc.wantResStatusCode, t)
		})
	}
}

//no upstream attempt made.
func TestContentEncodingPermutationsOnAboutHandler(t *testing.T) {
	tests := map[string]struct {
		reqUrlSlug                      string
		reqAcceptEncodingHeader         string
		reqSendAcceptEncodingHeader     bool
		wantResStatusCode               int
		wantResContentEncodingHeader    string
		wantResVaryAcceptEncodingHeader bool
		wantResBodyContent              string
	}{
		"noAcceptEncodingSendsIdentity": {"/about",
			"",
			false,
			200,
			"identity",
			false,
			"ServerID",
		},
		"emptyAcceptEncodingSendsIdentity": {"/about",
			"",
			true,
			200,
			"identity",
			false,
			"ServerID",
		},
		"starAcceptEncodingSendsIdentity": {"/about",
			"*",
			true,
			200,
			"identity",
			false,
			"ServerID",
		},
		"starCommaGzipAcceptEncodingSendsIdentity": {"/about",
			"*,gzip",
			true,
			200,
			"identity",
			false,
			"ServerID",
		},
		"identityCommaGzipAcceptEncodingSendsIdentity": {"/about",
			"identity,gzip",
			true,
			200,
			"identity",
			false,
			"ServerID",
		},
		"gzipCommaIdentityAcceptEncodingSendsIdentity": {"/about",
			"gzip,identity",
			true,
			200,
			"identity",
			false,
			"ServerID",
		},
		"gzipCommaStarAcceptEncodingSendsIdentity": {"/about",
			"gzip,*",
			true,
			200,
			"identity",
			false,
			"ServerID",
		},
		"gzipAcceptEncodingSendsGzip": {"/about",
			"gzip",
			true,
			200,
			"gzip",
			false,
			"ServerID",
		},
		"gzipCommaUnknownAcceptEncodingSendsGzip": {"/about",
			"gzip, unknown",
			true,
			200,
			"gzip",
			false,
			"ServerID",
		},
		"gzipCommaBrotliAcceptEncodingSendsGzip": {"/about",
			"gzip, br",
			true,
			200,
			"gzip",
			false,
			"ServerID",
		},
		"brotliAcceptEncodingSendsBrotli": {"/about",
			"br",
			true,
			200,
			"br",
			false,
			"ServerID",
		},
		"brotliCommaGzipAcceptEncodingSendsGzip": {"/about",
			"br,gzip",
			true,
			200,
			"gzip",
			false,
			"ServerID",
		},
		"deflateAcceptEncodingSends406ResponseCode": {"/about",
			"deflate",
			true,
			406,
			"identity",
			false,
			"406",
		},
		"unknownAcceptEncodingSends406ResponseCode": {"/about",
			"unknown",
			true,
			406,
			"identity",
			false,
			"406",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			DownstreamContentEncodingFullIntegrity(tc.reqUrlSlug, tc.reqAcceptEncodingHeader, tc.reqSendAcceptEncodingHeader,
				tc.wantResContentEncodingHeader, tc.wantResVaryAcceptEncodingHeader, tc.wantResBodyContent, tc.wantResStatusCode, t)
		})
	}
}

//no upstream attempt made
func TestContentEncodingPermutationsOnStatusCodeResponseInProxyHandler(t *testing.T) {
	tests := map[string]struct {
		reqUrlSlug                      string
		reqAcceptEncodingHeader         string
		reqSendAcceptEncodingHeader     bool
		wantResStatusCode               int
		wantResContentEncodingHeader    string
		wantResVaryAcceptEncodingHeader bool
		wantResBodyContent              string
	}{
		"noAcceptEncodingSendsIdentity": {"/badslug",
			"",
			false,
			404,
			"identity",
			false,
			"404",
		},
		"emptyAcceptEncodingSendsIdentity": {"/badslug",
			"",
			true,
			404,
			"identity",
			false,
			"404",
		},
		"starAcceptEncodingSendsIdentity": {"/badslug",
			"*",
			true,
			404,
			"identity",
			false,
			"404",
		},
		"starCommaGzipAcceptEncodingSendsIdentity": {"/badslug",
			"*,gzip",
			true,
			404,
			"identity",
			false,
			"404",
		},
		"identityCommaGzipAcceptEncodingSendsIdentity": {"/badslug",
			"identity,gzip",
			true,
			404,
			"identity",
			false,
			"404",
		},
		"gzipCommaIdentityAcceptEncodingSendsIdentity": {"/badslug",
			"gzip,identity",
			true,
			404,
			"identity",
			false,
			"404",
		},
		"gzipCommaStarAcceptEncodingSendsIdentity": {"/badslug",
			"gzip,*",
			true,
			404,
			"identity",
			false,
			"404",
		},
		"gzipAcceptEncodingSendsGzip": {"/badslug",
			"gzip",
			true,
			404,
			"gzip",
			false,
			"404",
		},
		"gzipCommaUnknownAcceptEncodingSendsGzip": {"/badslug",
			"gzip, unknown",
			true,
			404,
			"gzip",
			false,
			"404",
		},
		"gzipCommaBrotliAcceptEncodingSendsGzip": {"/badslug",
			"gzip, br",
			true,
			404,
			"gzip",
			false,
			"404",
		},
		"brotliAcceptEncodingSendsBrotli": {"/badslug",
			"br",
			true,
			404,
			"br",
			false,
			"404",
		},
		"brotliCommaGzipAcceptEncodingSendsGzip": {"/badslug",
			"br,gzip",
			true,
			404,
			"gzip",
			false,
			"404",
		},
		"deflateAcceptEncodingSends406ResponseCode": {"/badslug",
			"deflate",
			true,
			406,
			"identity",
			false,
			"406",
		},
		"unknownAcceptEncodingSends406ResponseCode": {"/badslug",
			"unknown",
			true,
			406,
			"identity",
			false,
			"406",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			DownstreamContentEncodingFullIntegrity(tc.reqUrlSlug, tc.reqAcceptEncodingHeader, tc.reqSendAcceptEncodingHeader,
				tc.wantResContentEncodingHeader, tc.wantResVaryAcceptEncodingHeader, tc.wantResBodyContent, tc.wantResStatusCode, t)
		})
	}
}

func DownstreamContentEncodingFullIntegrity(reqUrlSlug string, reqAcceptEncodingHeader string, reqSendAcceptEncodingHeader bool,
	wantResContentEncodingHeader string, wantResVaryAcceptEncodingHeader bool, wantResBodyContent string,
	wantResStatusCode int, t *testing.T) []byte {

	client := &http.Client{
		Transport: &http.Transport{DisableCompression: true},
	}
	req, _ := http.NewRequest("GET", "http://localhost:8080"+reqUrlSlug, nil)
	if reqSendAcceptEncodingHeader {
		req.Header.Add(j8a.AcceptEncodingS, reqAcceptEncodingHeader)
	} else {
		req.Header.Del(j8a.AcceptEncodingS)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("error connecting to server, cause: %s", err)
	}

	gotStatusCode := resp.StatusCode
	if gotStatusCode != wantResStatusCode {
		t.Errorf("want status code %v, but got %v instead", wantResStatusCode, gotStatusCode)
	}

	gotce := resp.Header.Get("Content-Encoding")
	if gotce != wantResContentEncodingHeader {
		t.Errorf("want content encoding %s, but got %s instead", wantResContentEncodingHeader, gotce)
	}

	gotVary := resp.Header.Get("Vary")
	if wantResVaryAcceptEncodingHeader == true {
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
	if len(wantResBodyContent) > 0 {
		if wantResContentEncodingHeader == "br" {
			body = *j8a.BrotliDecode(body)
		} else if wantResContentEncodingHeader == "gzip" {
			body = *j8a.Gunzip(body)
		} else if wantResContentEncodingHeader == "deflate" {
			body, _ = ioutil.ReadAll(flate.NewReader(bytes.NewBuffer(body)))
		}

		if !strings.Contains(string(body), wantResBodyContent) {
			t.Errorf("want body response %v, but got (decoded) %v", wantResBodyContent, string(body))
		}
	}
	return body
}
