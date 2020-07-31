package integration

import (
	"crypto/tls"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"testing"
)

var (
	versions = map[uint16]string{
		tls.VersionSSL30: "SSL",
		tls.VersionTLS10: "TLS 1.0",
		tls.VersionTLS11: "TLS 1.1",
		tls.VersionTLS12: "TLS 1.2",
		tls.VersionTLS13: "TLS 1.3",
	}
)

func TestHTTP11GetOverTLS12(t *testing.T) {
	HTTP11GetOverTlsVersion(t, tls.VersionTLS12)
}

func TestHTTP11GetOverTLS13(t *testing.T) {
	HTTP11GetOverTlsVersion(t, tls.VersionTLS13)
}

func HTTP11GetOverTlsVersion(t *testing.T, tlsVersion uint16) {
	var conn *tls.Conn
	var err error

	url := "https://localhost:8443/about"
	tlsConfig := &tls.Config{
		MinVersion: tlsVersion,
		MaxVersion: tlsVersion,
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
			//disable HTTP/2 support for TLS
			TLSNextProto: map[string]func(authority string, c *tls.Conn) http.RoundTripper{},
			DialTLS: func(network, addr string) (net.Conn, error) {
				conn, err = tls.Dial(network, addr, tlsConfig)
				return conn, err
			},
		},
	}
	response, err := client.Get(url)
	defer response.Body.Close()
	if err != nil {
		t.Errorf("unable to establish GET, cause %v", err)
	}

	body, _ := ioutil.ReadAll(response.Body)
	if !strings.Contains(string(body), "Jabba") {
		t.Errorf("unable to establish GET, body response: %v", string(body))
	}

	wantTls := versions[tlsVersion]
	gotTls := versions[conn.ConnectionState().Version]
	if gotTls != wantTls {
		t.Errorf("illegal TLS version want %v, got %v", wantTls, gotTls)
	}
}
