package connection

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/shirou/gopsutil/process"
	"github.com/simonmittag/j8a"
	"github.com/simonmittag/j8a/integration"
	"github.com/simonmittag/procspy"
	"golang.org/x/net/http2"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

type CloseListener struct {
	net.Conn   // embed the original conn
	ClosedFlag bool
}

func (c CloseListener) Read(b []byte) (n int, err error) {
	return c.Conn.Read(b)
}

func (c CloseListener) Write(b []byte) (n int, err error) {
	return c.Conn.Write(b)
}

func (c CloseListener) LocalAddr() net.Addr {
	return c.Conn.LocalAddr()
}

func (c CloseListener) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

func (c CloseListener) SetDeadline(t time.Time) error {
	return c.Conn.SetDeadline(t)
}

func (c CloseListener) SetReadDeadline(t time.Time) error {
	return c.Conn.SetReadDeadline(t)
}

func (c CloseListener) SetWriteDeadline(t time.Time) error {
	return c.Conn.SetWriteDeadline(t)
}

func (c *CloseListener) Close() error {
	err := c.Conn.Close()
	c.ClosedFlag = true
	return err
}

func TestConnection_100ConcurrentTCPConnectionsUsingHTTP11(t *testing.T) {
	ConcurrentHTTP11ConnectionsSucceed(100, t)
}

// this test covers all codes >=400 we just use 404 cause it's easy to evoke.
func TestConnection_404ResponseClosesDownstreamConnectionUsingHTTP11(t *testing.T) {
	//step 1 we connect to j8a with net.dial
	c, err := net.Dial("tcp", ":8080")
	if err != nil {
		t.Errorf("unable to connect to j8a server for integration test, cause: %v", err)
		return
	}
	defer c.Close()

	//step 2 we send headers and terminate HTTP message.
	integration.CheckWrite(t, c, "GET /mse6/send?code=404 HTTP/1.1\r\n")
	integration.CheckWrite(t, c, "Host: localhost:8080\r\n")
	integration.CheckWrite(t, c, "User-Agent: integration\r\n")
	integration.CheckWrite(t, c, "Accept: */*\r\n")
	integration.CheckWrite(t, c, "\r\n")

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
	} else {
		t.Logf("normal. server did respond with j8a header")
	}

	//step 4 we check for the connection close header
	if !strings.Contains(response, "Connection: close") {
		t.Error("j8a did not send Connection: close for HTTP/1.1 GET 404")
	} else {
		t.Logf("normal. server did respond with Connection: close header")
	}

	//step 5 we check the status of the connection which should be closed
	var e error
	for i := 0; i < 10; i++ {
		_, e = c.Write([]byte("test1234567890"))
		time.Sleep(time.Millisecond * 250)
		if e != nil {
			break
		}
	}
	if e == nil {
		t.Errorf("connection write should have failed after close")
	} else {
		t.Logf("normal. connection closed")
	}
}

func TestConnection_404ResponseClosesDownstreamConnectionUsingHTTP2(t *testing.T) {
	var conn *tls.Conn
	var cl *CloseListener
	var err error

	cAPem := "-----BEGIN CERTIFICATE-----\nMIIE0zCCAzugAwIBAgIQB2bsiI7SUtxu+HwBxuNtpDANBgkqhkiG9w0BAQsFADCB\ngTEeMBwGA1UEChMVbWtjZXJ0IGRldmVsb3BtZW50IENBMSswKQYDVQQLDCJzaW1v\nbm1pdHRhZ0B0cm9vcGVyIChTaW1vbiBNaXR0YWcpMTIwMAYDVQQDDClta2NlcnQg\nc2ltb25taXR0YWdAdHJvb3BlciAoU2ltb24gTWl0dGFnKTAeFw0yMDA1MDEyMTE2\nNDNaFw0zMDA1MDEyMTE2NDNaMIGBMR4wHAYDVQQKExVta2NlcnQgZGV2ZWxvcG1l\nbnQgQ0ExKzApBgNVBAsMInNpbW9ubWl0dGFnQHRyb29wZXIgKFNpbW9uIE1pdHRh\nZykxMjAwBgNVBAMMKW1rY2VydCBzaW1vbm1pdHRhZ0B0cm9vcGVyIChTaW1vbiBN\naXR0YWcpMIIBojANBgkqhkiG9w0BAQEFAAOCAY8AMIIBigKCAYEAzivKfp5OiWpT\n362cVgbw9DBqwMP0pO32aP79Y4UYeAxCfaWQDdqQEatBdraShtZcvUX8vZ9jvgHE\noGMGSJb/DIVRxIDfhdvhh4qGQgbbSLwDkfLJTkpGMdONa/5yDC54fNZjF095YZn7\niPmsFbvYUfTwpM8qrP+jZzobByrTO4rG3Ps080gIR08RCA0E+uLg58rTpnsdBKZ0\nK2uuE4B4lVAs2AeS4KPMrH/rnCjSZz4KRwnaGqh+wiAjO0PHAfrbrhNsFB6P1/Zk\nCqzclj3TXdkMDaXhSvt0qJPEpNIPQMkvj9GROom7hExZUT7t7LPOZwODtiR2VjM3\nDDehfLqpNPRrxU3aOR7b4lFVtEL1+9NXKc3rnR5T2xPVVvBxx8FqYAxFmQtkGqpA\nYlRxImBONBreIr5/fdkr5xqd/S0s1pb8ubuK7x5COfqf0Mv++j+UjMptBQ3kYvOh\ntNrbnEI1q/7kvHNB8ETtJ4hqXikl9EHMYWdOo4nyGd4P8jo9jmGVAgMBAAGjRTBD\nMA4GA1UdDwEB/wQEAwICBDASBgNVHRMBAf8ECDAGAQH/AgEAMB0GA1UdDgQWBBTD\nRHJaFeI4PnqJweaJrib0F4qokTANBgkqhkiG9w0BAQsFAAOCAYEAb+K3HO2AlDed\nS2yT7GnxD75Hcjnv1tMvMIlh1EOmRMHrzbsi7jv3Z7SDe2R5s1qRku3nxbVWj8i8\noRBi5GeRE+q/HkVloi4WPmgFGxUUbkWszAFSSGN5TAs72e5sCG/wMyEa0Gj8cOO1\ndK5SH3thP8+OjSpgQXToYfOimILlk7Hj7EgKE5Y8YX8UV+41LhGkzeK2UX9dBZn1\nof9qBc0dAQVlAA/O3dOgXorgiDbNT38cjignWEwVYzjeuJCYB91Ixf0CfHJZKHZR\nZCdIAHTJqW1tx7vsbrcl0PVAMgm+rkHLL0Dh9cp4fvONXWygVSjbqKM1s8UI9bFA\nbWU5Z3MhEn25wZCXLQDIq0uC+FwCxyS9e/exL4wmYpCLmRKVCp2gUa78Rlr/FJNa\nH9kfvP41Ya+fLzDWNKAlYQgizpZJmZuhPZu7O6n0UusaI+0WTKblCFUQJkx4aKEv\nio8QmLzoedmvVpO9Zp44Lyabmc7VnjoYTOcZczx4ECwEdKH/jswc\n-----END CERTIFICATE-----\n"
	url := "https://localhost:8443/mse6/send?code=404"

	certpool, _ := x509.SystemCertPool()
	ok := certpool.AppendCertsFromPEM([]byte(cAPem))
	if !ok {
		t.Errorf("no certs appended using system certs only")
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS12,
		RootCAs:    certpool,
	}

	client := &http.Client{
		Transport: &http2.Transport{
			TLSClientConfig: tlsConfig,
			DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
				conn, err = tls.Dial(network, addr, cfg)
				cl = &CloseListener{conn, false}
				return cl, err
			},
		},
	}

	//step 1: make a get request over HTTP/2. it will result in a 404 and a GOAWAY frame being sent.
	response, err := client.Get(url)

	if err == nil {
		defer response.Body.Close()

		//step 2: read the first response
		body, _ := ioutil.ReadAll(response.Body)
		if !strings.Contains(string(body), "404") {
			t.Errorf("unable to read server 404, body response was: %v", string(body))
		}

		//step 3: check the response is HTTP/2.0
		if response.Proto != "HTTP/2.0" {
			t.Error("connection protocol was not HTTP/2.0")
		} else {
			t.Logf("normal. connection protocol is HTTP/2.0")
		}

		//ugly wait for conn close. this has to be GOAWAY frame from server
		//j8a3 has idletimeout of 30s which should not fire
		time.Sleep(time.Second * 7)

		//step 4: test connection is now closed
		if cl.ClosedFlag == false {
			t.Error("connection was not closed")
		} else {
			t.Log("normal. connection was closed")
		}
	} else {
		t.Errorf("got unexpected error during first get request %v", err)
	}
}

func TestConnection_200ResponseLeavesConnectionOptionUsingHTTP2(t *testing.T) {
	var conn *tls.Conn
	var cl *CloseListener
	var err error

	cAPem := "-----BEGIN CERTIFICATE-----\nMIIE0zCCAzugAwIBAgIQB2bsiI7SUtxu+HwBxuNtpDANBgkqhkiG9w0BAQsFADCB\ngTEeMBwGA1UEChMVbWtjZXJ0IGRldmVsb3BtZW50IENBMSswKQYDVQQLDCJzaW1v\nbm1pdHRhZ0B0cm9vcGVyIChTaW1vbiBNaXR0YWcpMTIwMAYDVQQDDClta2NlcnQg\nc2ltb25taXR0YWdAdHJvb3BlciAoU2ltb24gTWl0dGFnKTAeFw0yMDA1MDEyMTE2\nNDNaFw0zMDA1MDEyMTE2NDNaMIGBMR4wHAYDVQQKExVta2NlcnQgZGV2ZWxvcG1l\nbnQgQ0ExKzApBgNVBAsMInNpbW9ubWl0dGFnQHRyb29wZXIgKFNpbW9uIE1pdHRh\nZykxMjAwBgNVBAMMKW1rY2VydCBzaW1vbm1pdHRhZ0B0cm9vcGVyIChTaW1vbiBN\naXR0YWcpMIIBojANBgkqhkiG9w0BAQEFAAOCAY8AMIIBigKCAYEAzivKfp5OiWpT\n362cVgbw9DBqwMP0pO32aP79Y4UYeAxCfaWQDdqQEatBdraShtZcvUX8vZ9jvgHE\noGMGSJb/DIVRxIDfhdvhh4qGQgbbSLwDkfLJTkpGMdONa/5yDC54fNZjF095YZn7\niPmsFbvYUfTwpM8qrP+jZzobByrTO4rG3Ps080gIR08RCA0E+uLg58rTpnsdBKZ0\nK2uuE4B4lVAs2AeS4KPMrH/rnCjSZz4KRwnaGqh+wiAjO0PHAfrbrhNsFB6P1/Zk\nCqzclj3TXdkMDaXhSvt0qJPEpNIPQMkvj9GROom7hExZUT7t7LPOZwODtiR2VjM3\nDDehfLqpNPRrxU3aOR7b4lFVtEL1+9NXKc3rnR5T2xPVVvBxx8FqYAxFmQtkGqpA\nYlRxImBONBreIr5/fdkr5xqd/S0s1pb8ubuK7x5COfqf0Mv++j+UjMptBQ3kYvOh\ntNrbnEI1q/7kvHNB8ETtJ4hqXikl9EHMYWdOo4nyGd4P8jo9jmGVAgMBAAGjRTBD\nMA4GA1UdDwEB/wQEAwICBDASBgNVHRMBAf8ECDAGAQH/AgEAMB0GA1UdDgQWBBTD\nRHJaFeI4PnqJweaJrib0F4qokTANBgkqhkiG9w0BAQsFAAOCAYEAb+K3HO2AlDed\nS2yT7GnxD75Hcjnv1tMvMIlh1EOmRMHrzbsi7jv3Z7SDe2R5s1qRku3nxbVWj8i8\noRBi5GeRE+q/HkVloi4WPmgFGxUUbkWszAFSSGN5TAs72e5sCG/wMyEa0Gj8cOO1\ndK5SH3thP8+OjSpgQXToYfOimILlk7Hj7EgKE5Y8YX8UV+41LhGkzeK2UX9dBZn1\nof9qBc0dAQVlAA/O3dOgXorgiDbNT38cjignWEwVYzjeuJCYB91Ixf0CfHJZKHZR\nZCdIAHTJqW1tx7vsbrcl0PVAMgm+rkHLL0Dh9cp4fvONXWygVSjbqKM1s8UI9bFA\nbWU5Z3MhEn25wZCXLQDIq0uC+FwCxyS9e/exL4wmYpCLmRKVCp2gUa78Rlr/FJNa\nH9kfvP41Ya+fLzDWNKAlYQgizpZJmZuhPZu7O6n0UusaI+0WTKblCFUQJkx4aKEv\nio8QmLzoedmvVpO9Zp44Lyabmc7VnjoYTOcZczx4ECwEdKH/jswc\n-----END CERTIFICATE-----\n"
	url := "https://localhost:8443/mse6/get"

	certpool, _ := x509.SystemCertPool()
	ok := certpool.AppendCertsFromPEM([]byte(cAPem))
	if !ok {
		t.Errorf("no certs appended using system certs only")
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS12,
		RootCAs:    certpool,
	}

	client := &http.Client{
		Transport: &http2.Transport{
			TLSClientConfig: tlsConfig,
			DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
				conn, err = tls.Dial(network, addr, cfg)
				cl = &CloseListener{conn, false}
				return cl, err
			},
		},
	}

	//step 1: make a get request over HTTP/2. it will result in a 404 and a GOAWAY frame being sent.
	response, err := client.Get(url)

	if err == nil {
		defer response.Body.Close()

		//step 2: read the first response
		body, _ := ioutil.ReadAll(response.Body)
		if !strings.Contains(string(body), "get endpoint") {
			t.Errorf("unable to read server 200, body response was: %v", string(body))
		} else {
			t.Logf("normal. 200 OK")
		}

		//step 3: check the response is HTTP/2.0
		if response.Proto != "HTTP/2.0" {
			t.Error("connection protocol was not HTTP/2.0")
		} else {
			t.Logf("normal. connection protocol is HTTP/2.0")
		}

		//ugly wait for conn. j8a3 has idletimeout of 30s which should not fire
		time.Sleep(time.Second * 7)

		//step 4: test connection remains open
		if cl.ClosedFlag == false {
			t.Log("normal. connection remains open")
		} else {
			t.Error("connection was closed")
		}
	} else {
		t.Errorf("got unexpected error during first get request %v", err)
	}
}

func ConcurrentHTTP11ConnectionsSucceed(total int, t *testing.T) {
	good := 0
	bad := 0

	R200 := 0
	N200 := 0

	wg := sync.WaitGroup{}
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        2000,
			MaxIdleConnsPerHost: 2000,
			//disable HTTP/2 support for TLS
			TLSNextProto: map[string]func(authority string, c *tls.Conn) http.RoundTripper{},
		},
	}

	for i := 0; i < total; i++ {
		wg.Add(1)

		go func(j int) {
			serverPort := 8080
			resp, err := client.Get(fmt.Sprintf("http://localhost:%d/mse6/slowbody?wait=2", serverPort))
			if err != nil {
				t.Errorf("received upstream error for GET request: %v", err)
				bad++
			} else if resp != nil && resp.Body != nil {
				defer resp.Body.Close()
				if resp.Status != "200 OK" {
					t.Logf("goroutine %d, received non 200 status but normal server response: %v", j, resp.Status)
					good++
					N200++
				} else {
					t.Logf("goroutine %d, received status 200 OK", j)
					good++
					R200++
				}
			}

			wg.Done()
		}(i)
	}

	wg.Wait()
	t.Logf("done! good HTTP response: %d, 200s: %d, non 200s: %d, connection errors: %d", good, R200, N200, bad)
}

// this test is a little different to other integration tests. it's not a unit test
// because of the dependency on site hosted on external domain. it checks that procspy
// connections from within j8a to that site are properly counted as if it were an
// upstream resource configured within j8a.
func TestRuntime_CountResourceIps(t *testing.T) {
	proc, _ := process.NewProcess(int32(os.Getpid()))
	ad, _ := net.LookupIP("jsonplaceholder.typicode.com")
	ips := map[string][]net.IP{"jsonplaceholder.typicode.com": ad}

	c := http.Client{Transport: &http.Transport{IdleConnTimeout: 0}}

	resp, _ := c.Get("https://jsonplaceholder.typicode.com/")
	if resp != nil {
		defer resp.Body.Close()
	}

	ps, _ := procspy.Connections(true)

	rt := j8a.Runtime{
		Config: j8a.Config{
			Resources: map[string][]j8a.ResourceMapping{
				"jsonplaceholder.typicode.com": []j8a.ResourceMapping{{
					URL: j8a.URL{
						Host: "jsonplaceholder.typicode.com",
						Port: "443",
					},
				}},
			}},
	}
	got := rt.CountUpConns(proc, ps, ips)
	want := 1
	if got != want {
		t.Errorf("want %v connections for adyntest.com, got %v", want, got)
	}
	c.CloseIdleConnections()
}
