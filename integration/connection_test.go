package integration

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
)

func Test100ConcurrentTCPConnectionsUsingHTTP11(t *testing.T) {
	ConcurrentHTTP11ConnectionsSucceed(100, t)
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

func Test404ResponseClosesDownstreamConnection(t *testing.T) {
	//step 1 we connect to j8a with net.dial
	c, err := net.Dial("tcp", ":8080")
	if err != nil {
		t.Errorf("unable to connect to j8a server for integration test, cause: %v", err)
		return
	}
	defer c.Close()

	//step 2 we send headers and terminate HTTP message.
	checkWrite(t, c, "GET /mse6/xyz HTTP/1.1\r\n")
	checkWrite(t, c, "Host: localhost:8080\r\n")
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
