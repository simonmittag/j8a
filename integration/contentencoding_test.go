package integration

import (
	"fmt"
	"net"
	"strings"
	"testing"
)

func TestIdentityEncodingOn404(t *testing.T) {
	DownstreamAcceptEncodingHTTP11("identity", "/", t)
}

func TestGzipEncodingOn404(t *testing.T) {
	DownstreamAcceptEncodingHTTP11("gzip", "/", t)
}

func DownstreamAcceptEncodingHTTP11(enc string, slug string, t *testing.T) {
	//if this test fails, check the j8a configuration for connection.downstream.ReadTimeoutSeconds

	//step 1 make a connection
	c, err := net.Dial("tcp", ":8080")
	if err != nil {
		t.Errorf("test failure. unable to connect to j8a server for integration test, cause: %v", err)
	}

	//step 2 we send headers and terminate http message so j8a sends request upstream.
	checkWrite(t, c, fmt.Sprintf("GET %s HTTP/1.1\r\n", slug))
	checkWrite(t, c, "Host: localhost:8080\r\n")
	checkWrite(t, c, fmt.Sprintf("Accept-Encoding: %v\r\n", enc))
	checkWrite(t, c, "\r\n")

	//step 3 we read a response into buffer which returns 501
	buf := make([]byte, 128)
	b, err2 := c.Read(buf)
	t.Logf("normal. j8a responded with %v bytes", b)
	if err2 != nil || !strings.Contains(string(buf), enc) {
		t.Errorf("test failure. want response content encoding %s but got %s", enc, string(buf))
	} else {
		t.Logf("normal. received response %s", string(buf))
	}
}
