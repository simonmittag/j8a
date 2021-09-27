package downstream

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestServer2HangsUpOnDownstreamIfRoundTripTimeoutExceeded(t *testing.T) {
	//if this test fails, check the j8a configuration for connection.downstream.ReadTimeoutSeconds
	grace := 1
	serverRoundTripTimeoutSeconds := 20
	//assumes upstreamReadTimeoutSeconds := 30 so it doesn't fire before serverRoundTripTimeoutSeconds
	wait := serverRoundTripTimeoutSeconds + grace

	//step 1 make a connection
	c, err := net.Dial("tcp", ":8081")
	if err != nil {
		t.Errorf("test failure. unable to connect to j8a server for integration test, cause: %v", err)
	}

	//step 2 we send headers and terminate http message so j8a sends request upstream.
	checkWrite(t, c, "GET /mse6/slowbody?wait=21 HTTP/1.1\r\n")
	checkWrite(t, c, "Host: localhost:8081\r\n")
	checkWrite(t, c, "\r\n")

	//step 3 we sleep locally until we should hit timeout
	t.Logf("normal. going to sleep for %d seconds to trigger remote j8a roundtrip timeout", wait)
	time.Sleep(time.Second * time.Duration(wait))

	//step 4 we read a response into buffer which returns 504
	buf := make([]byte, 16)
	b, err2 := c.Read(buf)
	t.Logf("normal. j8a responded with %v bytes", b)
	if err2 != nil || !strings.Contains(string(buf), "504") {
		t.Errorf("test failure. after timeout we should first experience a 504")
	}

	//step 5 now wait for the grace period, then re-read. the connection must now be closed.
	//note you must read in a loop with small buffer because golang's reader has cached data
	time.Sleep(time.Duration(grace) * time.Second)
	for i := 0; i < 32; i++ {
		b, err2 = c.Read(buf)
	}
	if err2 != nil && err2.Error() != "EOF" {
		t.Errorf("test failure. expected j8a server to hang up on us after %ds, but it didn't. check downstream roundtrip timeout", serverRoundTripTimeoutSeconds)
	} else {
		t.Logf("normal. j8a hung up connection as expected after grace period with error: %v", err2)
	}
}

func TestServer1RoundTripNormalWithoutHangingUp(t *testing.T) {
	//step 1 we connect to j8a with net.dial
	c, err := net.Dial("tcp", ":8080")
	if err != nil {
		t.Errorf("unable to connect to j8a server for integration test, cause: %v", err)
		return
	}
	defer c.Close()

	//step 2 we send headers and terminate HTTP message.
	checkWrite(t, c, "GET /mse6/slowbody?wait=3 HTTP/1.1\r\n")
	checkWrite(t, c, "Host: localhost:8081\r\n")
	checkWrite(t, c, "User-Agent: integration\r\n")
	checkWrite(t, c, "Accept: */*\r\n")
	checkWrite(t, c, "\r\n")

	//step 3 we try to read the server response. Warning this isn't a proper http client
	//i.e. doesn't include parsing content length, reading response properly.
	buf := make([]byte, 1024)
	l, err := c.Read(buf)
	t.Logf("j8a responded with %v bytes and error code %v", l, err)
	t.Logf("j8a partial response: %v", string(buf))
	if l == 0 {
		t.Error("j8a did not respond, 0 bytes read")
	}
	response := string(buf)
	if !strings.Contains(response, "Server: j8a") {
		t.Error("j8a did not respond, server information not found on response ")
	}
}

func TestServer2DownstreamRoundTripTimeoutFireWithSlowHeader31S(t *testing.T) {
	performJ8aTest(t,
		"/slowheader",
		31,
		20,
		504,
		8081,
		false)
}

func TestServer3TLSDownstreamRoundTripTimeoutFireWithSlowHeader31S(t *testing.T) {
	performJ8aTest(t,
		"/slowheader",
		31,
		20,
		504,
		8443,
		true)
}

func TestServer2DownstreamRoundTripTimeoutFireWithSlowBody31S(t *testing.T) {
	performJ8aTest(t,
		"/slowbody",
		31,
		20,
		504,
		8081,
		false)
}

func TestServer3TLSDownstreamRoundTripTimeoutFireWithSlowBody31S(t *testing.T) {
	performJ8aTest(t,
		"/slowbody",
		31,
		20,
		504,
		8443,
		true)
}

func TestServer2DownstreamRoundTripTimeoutFireWithSlowHeader25S(t *testing.T) {
	performJ8aTest(t,
		"/slowheader",
		25,
		20,
		504,
		8081,
		false)
}

func TestServer3TLSDownstreamRoundTripTimeoutFireWithSlowHeader25S(t *testing.T) {
	performJ8aTest(t,
		"/slowheader",
		25,
		20,
		504,
		8443,
		true)
}

//works as expected on b164
func TestServer2DownstreamRoundTripTimeoutFireWithSlowBody25S(t *testing.T) {
	performJ8aTest(t,
		"/slowbody",
		25,
		20,
		504,
		8081,
		false)
}

func TestServer3TLSDownstreamRoundTripTimeoutFireWithSlowBody25S(t *testing.T) {
	performJ8aTest(t,
		"/slowbody",
		25,
		20,
		504,
		8443,
		true)
}

func TestServer2DownstreamRoundTripTimeoutNotFireWithSlowHeader4S(t *testing.T) {
	performJ8aTest(t,
		"/slowheader",
		4,
		4,
		200,
		8081,
		false)
}

func TestServer3TLSDownstreamRoundTripTimeoutNotFireWithSlowHeader4S(t *testing.T) {
	performJ8aTest(t,
		"/slowheader",
		4,
		4,
		200,
		8443,
		true)
}

func TestServer2DownstreamRoundTripTimeoutNotFireWithSlowBody4S(t *testing.T) {
	performJ8aTest(t,
		"/slowbody",
		4,
		4,
		200,
		8081,
		false)
}

func TestServer3TLSDownstreamRoundTripTimeoutNotFireWithSlowBody4S(t *testing.T) {
	performJ8aTest(t,
		"/slowbody",
		4,
		4,
		200,
		8443,
		true)
}

func TestServer2DownstreamRoundTripTimeoutNotFireWithSlowHeader2S(t *testing.T) {
	performJ8aTest(t,
		"/slowheader",
		2,
		2,
		200,
		8081,
		false)
}

func TestServer3TLSDownstreamRoundTripTimeoutNotFireWithSlowHeader2S(t *testing.T) {
	performJ8aTest(t,
		"/slowheader",
		2,
		2,
		200,
		8443,
		true)
}

func TestServer2DownstreamRoundTripTimeoutNotFireWithSlowBody2S(t *testing.T) {
	performJ8aTest(t,
		"/slowbody",
		2,
		2,
		200,
		8081,
		false)
}

func TestServer3DownstreamRoundTripTimeoutNotFireWithSlowBody2S(t *testing.T) {
	performJ8aTest(t,
		"/slowbody",
		2,
		2,
		200,
		8443,
		true)
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
		tlsConfig := &tls.Config{
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

