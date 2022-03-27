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

func TestContentEncodingAboutPermutations(t *testing.T) {
	tests := map[string]struct {
		reqUrlSlug                      string
		reqAcceptEncodingHeader         string
		reqSendAcceptEncodingHeader     bool
		wantResStatusCode               int
		wantResContentEncodingHeader    string
		wantResVaryAcceptEncodingHeader bool
		wantResBodyContent              string
	}{
		"noAcceptEncoding": {"/about",
			"",
			false,
			200,
			"identity",
			false,
			"ServerID",
		},
		"emptyAcceptEncoding": {"/about",
			"",
			true,
			200,
			"identity",
			false,
			"ServerID",
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
		}
		if wantResContentEncodingHeader == "gzip" {
			body = *j8a.Gunzip(body)
		}
		if wantResContentEncodingHeader == "deflate" {
			body, _ = ioutil.ReadAll(flate.NewReader(bytes.NewBuffer(body)))
		}
		if !strings.Contains(string(body), wantResBodyContent) {
			t.Errorf("want body response %v, but got (decoded) %v", wantResBodyContent, string(body))
		}
	}
	return body
}
