package headers

import (
	"github.com/simonmittag/j8a/integration"
	"net"
	"strings"
	"testing"
)

func TestGlobalOptionsHandlerForAllowHeadersPresent(t *testing.T) {
	//step 1 we connect to j8a with net.dial
	c, err := net.Dial("tcp", ":8080")
	if err != nil {
		t.Errorf("unable to connect to j8a server for integration test, cause: %v", err)
		return
	}

	//step 2 we send headers and terminate HTTP message.
	integration.CheckWrite(t, c, "OPTIONS * HTTP/1.1\r\n")
	integration.CheckWrite(t, c, "Host: localhost:8080\r\n")
	integration.CheckWrite(t, c, "Accept: */*\r\n")
	integration.CheckWrite(t, c, "\r\n")

	//step 3 read response headers and check allow content
	buf := make([]byte, 512)
	b, err2 := c.Read(buf)
	t.Logf("normal. j8a responded with %v bytes", b)
	if err2 != nil || !strings.Contains(string(buf), "Allow") {
		t.Errorf("test failure. Allow header not present")
	} else {
		t.Logf("normal. received response %s", string(buf))
	}

	want := []string{"GET", "HEAD", "OPTIONS", "TRACE", "PUT", "DELETE", "POST", "CONNECT"}
	for _, m := range want {
		if !strings.Contains(string(buf), m) {
			t.Errorf("expected server to allow %v, but not found as response header", m)
		} else {
			t.Logf("normal. response contains Allow header for method %v", m)
		}
	}
}
