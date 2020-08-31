package integration

import (
	"crypto/tls"
	"crypto/x509"
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

	cAPem := "-----BEGIN CERTIFICATE-----\nMIIE0zCCAzugAwIBAgIQB2bsiI7SUtxu+HwBxuNtpDANBgkqhkiG9w0BAQsFADCB\ngTEeMBwGA1UEChMVbWtjZXJ0IGRldmVsb3BtZW50IENBMSswKQYDVQQLDCJzaW1v\nbm1pdHRhZ0B0cm9vcGVyIChTaW1vbiBNaXR0YWcpMTIwMAYDVQQDDClta2NlcnQg\nc2ltb25taXR0YWdAdHJvb3BlciAoU2ltb24gTWl0dGFnKTAeFw0yMDA1MDEyMTE2\nNDNaFw0zMDA1MDEyMTE2NDNaMIGBMR4wHAYDVQQKExVta2NlcnQgZGV2ZWxvcG1l\nbnQgQ0ExKzApBgNVBAsMInNpbW9ubWl0dGFnQHRyb29wZXIgKFNpbW9uIE1pdHRh\nZykxMjAwBgNVBAMMKW1rY2VydCBzaW1vbm1pdHRhZ0B0cm9vcGVyIChTaW1vbiBN\naXR0YWcpMIIBojANBgkqhkiG9w0BAQEFAAOCAY8AMIIBigKCAYEAzivKfp5OiWpT\n362cVgbw9DBqwMP0pO32aP79Y4UYeAxCfaWQDdqQEatBdraShtZcvUX8vZ9jvgHE\noGMGSJb/DIVRxIDfhdvhh4qGQgbbSLwDkfLJTkpGMdONa/5yDC54fNZjF095YZn7\niPmsFbvYUfTwpM8qrP+jZzobByrTO4rG3Ps080gIR08RCA0E+uLg58rTpnsdBKZ0\nK2uuE4B4lVAs2AeS4KPMrH/rnCjSZz4KRwnaGqh+wiAjO0PHAfrbrhNsFB6P1/Zk\nCqzclj3TXdkMDaXhSvt0qJPEpNIPQMkvj9GROom7hExZUT7t7LPOZwODtiR2VjM3\nDDehfLqpNPRrxU3aOR7b4lFVtEL1+9NXKc3rnR5T2xPVVvBxx8FqYAxFmQtkGqpA\nYlRxImBONBreIr5/fdkr5xqd/S0s1pb8ubuK7x5COfqf0Mv++j+UjMptBQ3kYvOh\ntNrbnEI1q/7kvHNB8ETtJ4hqXikl9EHMYWdOo4nyGd4P8jo9jmGVAgMBAAGjRTBD\nMA4GA1UdDwEB/wQEAwICBDASBgNVHRMBAf8ECDAGAQH/AgEAMB0GA1UdDgQWBBTD\nRHJaFeI4PnqJweaJrib0F4qokTANBgkqhkiG9w0BAQsFAAOCAYEAb+K3HO2AlDed\nS2yT7GnxD75Hcjnv1tMvMIlh1EOmRMHrzbsi7jv3Z7SDe2R5s1qRku3nxbVWj8i8\noRBi5GeRE+q/HkVloi4WPmgFGxUUbkWszAFSSGN5TAs72e5sCG/wMyEa0Gj8cOO1\ndK5SH3thP8+OjSpgQXToYfOimILlk7Hj7EgKE5Y8YX8UV+41LhGkzeK2UX9dBZn1\nof9qBc0dAQVlAA/O3dOgXorgiDbNT38cjignWEwVYzjeuJCYB91Ixf0CfHJZKHZR\nZCdIAHTJqW1tx7vsbrcl0PVAMgm+rkHLL0Dh9cp4fvONXWygVSjbqKM1s8UI9bFA\nbWU5Z3MhEn25wZCXLQDIq0uC+FwCxyS9e/exL4wmYpCLmRKVCp2gUa78Rlr/FJNa\nH9kfvP41Ya+fLzDWNKAlYQgizpZJmZuhPZu7O6n0UusaI+0WTKblCFUQJkx4aKEv\nio8QmLzoedmvVpO9Zp44Lyabmc7VnjoYTOcZczx4ECwEdKH/jswc\n-----END CERTIFICATE-----\n"

	url := "https://localhost:8443/about"

	certpool, _ := x509.SystemCertPool()

	ok := certpool.AppendCertsFromPEM([]byte(cAPem))
	if !ok {
		t.Errorf("no certs appended using system certs only")
	}

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
		if !strings.Contains(string(body), "j8a") {
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
