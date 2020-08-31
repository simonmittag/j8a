package integration

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"testing"
	"time"
)

var tlsConfig *tls.Config

func TestServer1UpstreamReadTimeoutFireWithSlowHeader31S(t *testing.T) {
	performj8aTest(t,
		"/slowheader",
		31,
		12,
		504,
		8080,
		false)
}

func TestServer1UpstreamReadTimeoutFireWithSlowBody31S(t *testing.T) {
	performj8aTest(t,
		"/slowbody",
		31,
		12,
		504,
		8080,
		false)
}

func TestServer1UpstreamReadTimeoutFireWithSlowHeader25S(t *testing.T) {
	performj8aTest(t,
		"/slowheader",
		25,
		12,
		504,
		8080,
		false)
}

func TestServer1UpstreamReadTimeoutFireWithSlowBody25S(t *testing.T) {
	performj8aTest(t,
		"/slowbody",
		25,
		12,
		504,
		8080,
		false)
}

func TestServer1UpstreamReadTimeoutFireWithSlowHeader4S(t *testing.T) {
	performj8aTest(t,
		"/slowheader",
		4,
		12,
		504,
		8080,
		false)
}

func TestServer1UpstreamReadTimeoutFireWithSlowBody4S(t *testing.T) {
	performj8aTest(t,
		"/slowbody",
		4,
		12,
		504,
		8080,
		false)
}

func TestServer1UpstreamReadTimeoutNotFireWithSlowHeader2S(t *testing.T) {
	performJ8aTest(t,
		"/slowheader",
		2,
		2,
		200,
		8080,
		false)
}

func TestServer1UpstreamReadTimeoutNotFireWithSlowBody2S(t *testing.T) {
	performJ8aTest(t,
		"/slowbody",
		2,
		2,
		200,
		8080,
		false)
}

func performJ8aTest(t *testing.T, testMethod string, wantUpstreamWaitSeconds int, wantTotalWaitSeconds int, wantStatusCode int, serverPort int, tlsMode bool) {
	start := time.Now()
	scheme := "http"

	var client *http.Client
	if tlsMode {
		scheme = "https"
		cAPem := "-----BEGIN CERTIFICATE-----\nMIIE0zCCAzugAwIBAgIQB2bsiI7SUtxu+HwBxuNtpDANBgkqhkiG9w0BAQsFADCB\ngTEeMBwGA1UEChMVbWtjZXJ0IGRldmVsb3BtZW50IENBMSswKQYDVQQLDCJzaW1v\nbm1pdHRhZ0B0cm9vcGVyIChTaW1vbiBNaXR0YWcpMTIwMAYDVQQDDClta2NlcnQg\nc2ltb25taXR0YWdAdHJvb3BlciAoU2ltb24gTWl0dGFnKTAeFw0yMDA1MDEyMTE2\nNDNaFw0zMDA1MDEyMTE2NDNaMIGBMR4wHAYDVQQKExVta2NlcnQgZGV2ZWxvcG1l\nbnQgQ0ExKzApBgNVBAsMInNpbW9ubWl0dGFnQHRyb29wZXIgKFNpbW9uIE1pdHRh\nZykxMjAwBgNVBAMMKW1rY2VydCBzaW1vbm1pdHRhZ0B0cm9vcGVyIChTaW1vbiBN\naXR0YWcpMIIBojANBgkqhkiG9w0BAQEFAAOCAY8AMIIBigKCAYEAzivKfp5OiWpT\n362cVgbw9DBqwMP0pO32aP79Y4UYeAxCfaWQDdqQEatBdraShtZcvUX8vZ9jvgHE\noGMGSJb/DIVRxIDfhdvhh4qGQgbbSLwDkfLJTkpGMdONa/5yDC54fNZjF095YZn7\niPmsFbvYUfTwpM8qrP+jZzobByrTO4rG3Ps080gIR08RCA0E+uLg58rTpnsdBKZ0\nK2uuE4B4lVAs2AeS4KPMrH/rnCjSZz4KRwnaGqh+wiAjO0PHAfrbrhNsFB6P1/Zk\nCqzclj3TXdkMDaXhSvt0qJPEpNIPQMkvj9GROom7hExZUT7t7LPOZwODtiR2VjM3\nDDehfLqpNPRrxU3aOR7b4lFVtEL1+9NXKc3rnR5T2xPVVvBxx8FqYAxFmQtkGqpA\nYlRxImBONBreIr5/fdkr5xqd/S0s1pb8ubuK7x5COfqf0Mv++j+UjMptBQ3kYvOh\ntNrbnEI1q/7kvHNB8ETtJ4hqXikl9EHMYWdOo4nyGd4P8jo9jmGVAgMBAAGjRTBD\nMA4GA1UdDwEB/wQEAwICBDASBgNVHRMBAf8ECDAGAQH/AgEAMB0GA1UdDgQWBBTD\nRHJaFeI4PnqJweaJrib0F4qokTANBgkqhkiG9w0BAQsFAAOCAYEAb+K3HO2AlDed\nS2yT7GnxD75Hcjnv1tMvMIlh1EOmRMHrzbsi7jv3Z7SDe2R5s1qRku3nxbVWj8i8\noRBi5GeRE+q/HkVloi4WPmgFGxUUbkWszAFSSGN5TAs72e5sCG/wMyEa0Gj8cOO1\ndK5SH3thP8+OjSpgQXToYfOimILlk7Hj7EgKE5Y8YX8UV+41LhGkzeK2UX9dBZn1\nof9qBc0dAQVlAA/O3dOgXorgiDbNT38cjignWEwVYzjeuJCYB91Ixf0CfHJZKHZR\nZCdIAHTJqW1tx7vsbrcl0PVAMgm+rkHLL0Dh9cp4fvONXWygVSjbqKM1s8UI9bFA\nbWU5Z3MhEn25wZCXLQDIq0uC+FwCxyS9e/exL4wmYpCLmRKVCp2gUa78Rlr/FJNa\nH9kfvP41Ya+fLzDWNKAlYQgizpZJmZuhPZu7O6n0UusaI+0WTKblCFUQJkx4aKEv\nio8QmLzoedmvVpO9Zp44Lyabmc7VnjoYTOcZczx4ECwEdKH/jswc\n-----END CERTIFICATE-----\n"
		certpool, _ := x509.SystemCertPool()
		ok := certpool.AppendCertsFromPEM([]byte(cAPem))
		if !ok {
			t.Errorf("no certs appended using system certs only")
		}
		tlsConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			MaxVersion: tls.VersionTLS13,
			RootCAs:    certpool,
		}

		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
				//disable HTTP/2 support for TLS
				TLSNextProto: map[string]func(authority string, c *tls.Conn) http.RoundTripper{},
			},
		}
	} else {
		client = &http.Client{}
	}

	resp, err := client.Get(fmt.Sprintf("%s://localhost:%d/mse6%s?wait=%d", scheme, serverPort, testMethod, wantUpstreamWaitSeconds))
	gotTotalWait := time.Since(start)

	if err != nil {
		t.Errorf("error connecting to upstream for port %d, testMethod %s, cause: %v", serverPort, testMethod, err)
	}
	gotStatusCode := resp.StatusCode

	if !okayTimeDrift(gotTotalWait, wantTotalWaitSeconds) {
		t.Errorf("bad time drift for port %d, testMethod %s, want seconds %d, got %f", serverPort, testMethod, wantTotalWaitSeconds, gotTotalWait.Seconds())
	}

	if gotStatusCode != wantStatusCode {
		t.Errorf("bad status code for port %d, testMethod %s, want statusCode %d, got %d", serverPort, testMethod, wantStatusCode, gotStatusCode)
	}
}

func okayTimeDrift(elapsed time.Duration, waitSeconds int) bool {
	fmin := 1.0
	fmax := 1.1
	elapsedSeconds := elapsed.Seconds()

	if elapsedSeconds > fmax*float64(waitSeconds) {
		return false
	}
	if elapsedSeconds < fmin*float64(waitSeconds) {
		return false
	}
	return true
}
