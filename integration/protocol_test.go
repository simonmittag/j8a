package integration

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"net/http"
	"os"
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

const golangTlsProtoUnsupported = "remote error: tls: protocol version not supported"
const golangNoSupportedVersions = "tls: no supported versions satisfy MinVersion and MaxVersion"
const weDontWantAnyError = ""

func TestHTTP11GetOverSSL30(t *testing.T) {
	HTTP11GetOverTlsVersion(t, tls.VersionSSL30, golangNoSupportedVersions)
}

func TestHTTP11GetOverTLS10(t *testing.T) {
	HTTP11GetOverTlsVersion(t, tls.VersionTLS10, golangTlsProtoUnsupported)
}

func TestHTTP11GetOverTLS11(t *testing.T) {
	HTTP11GetOverTlsVersion(t, tls.VersionTLS11, golangTlsProtoUnsupported)
}

func TestHTTP11GetOverTLS12(t *testing.T) {
	HTTP11GetOverTlsVersion(t, tls.VersionTLS12, weDontWantAnyError)
}

func TestHTTP11GetOverTLS13(t *testing.T) {
	HTTP11GetOverTlsVersion(t, tls.VersionTLS13, weDontWantAnyError)
}

func HTTP11GetOverTlsVersion(t *testing.T, tlsVersion uint16, wantErr string) {
	var conn *tls.Conn
	var err error

	url := "https://localhost:8443/about"

	certpool := x509.NewCertPool()
	caPemFile, _ := os.Open("./certs/rootCA.pem")
	caPem, _ := ioutil.ReadAll(caPemFile)
	certpool.AppendCertsFromPEM(caPem)

	tlsConfig := &tls.Config{
		MinVersion: tlsVersion,
		MaxVersion: tlsVersion,
		RootCAs:    certpool,
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
	response, err2 := client.Get(url)

	if err == nil && err2 == nil {
		defer response.Body.Close()
		body, _ := ioutil.ReadAll(response.Body)
		if !strings.Contains(string(body), "Jabba") {
			t.Errorf("unable to establish GET, body response: %v", string(body))
		}

		wantTls := versions[tlsVersion]
		gotTls := versions[conn.ConnectionState().Version]
		if gotTls != wantTls {
			t.Errorf("illegal TLS version want %v, got %v", wantTls, gotTls)
		}
	} else {
		if len(wantErr) == 0 {
			t.Errorf("connection error, cause %v or %v", err, err2)
		} else {
			if wantErr == err.Error() || wantErr == err2.Error() {
				t.Logf("ok. got expected error: %v", wantErr)
			} else {
				t.Errorf("got unexpected error, want %v, got %v and %v", wantErr, err, err2)
			}
		}
	}
}
