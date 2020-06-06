package integration

import (
	"net"
	"strings"
	"testing"
	"time"
)

func TestServerHangsUpOnDownstreamIfRoundTripTimeoutExceeded(t *testing.T) {
	//if this test fails, check the jabba configuration for connection.downstream.ReadTimeoutSeconds
	grace := 1
	serverRoundTripTimeoutSeconds := 20 + grace
	//note this test assumes upstreamReadTimeoutSeconds := 30, double check config if failing.

	//step 1 make a connection
	c, err := net.Dial("tcp", ":8081")
	if err != nil {
		t.Errorf("unable to connect to jabba server for integration test, cause: %v", err)
	}

	//step 2 we send headers and terminate http message so Jabba sends request upstream.
	//note this time interval needs to be < upstreamReadTimeoutSeconds to prevent another timeout to fire during test.
	checkWrite(t, c, "GET /mse6/slowbody?wait=21 HTTP/1.1\r\n")
	checkWrite(t, c, "Host: localhost:8080\r\n")
	checkWrite(t, c, "\r\n")

	//step 3 we sleep locally until we should hit timeout
	t.Logf("going to sleep for %d seconds to trigger remote jabba roundtrip timeout", serverRoundTripTimeoutSeconds)
	time.Sleep(time.Second * time.Duration(serverRoundTripTimeoutSeconds))

	buf := make([]byte, 1024)
	b, err2 := c.Read(buf)
	t.Logf("jabba responded with %v bytes and error code %v", b, err)
	t.Logf("jabba partial response: %v", string(buf))

	if err2.Error() != "EOF" {
		t.Errorf("error: %v ", err2)
		t.Errorf("bytes written: %d", b)
		t.Errorf("expected jabba server to hang up on us after %ds, but it didn't. check downstream roundtrip timeout", serverRoundTripTimeoutSeconds)
	} else {
		t.Logf("jabba hung up as expected with error: %v", err2)
	}
}

func TestServerRoundTripNormalWithoutHangingUp(t *testing.T) {
	//step 1 we connect to Jabba with net.dial
	c, err := net.Dial("tcp", ":8080")
	if err != nil {
		t.Errorf("unable to connect to jabba server for integration test, cause: %v", err)
		return
	}
	defer c.Close()

	//step 2 we send headers and terminate HTTP message.
	checkWrite(t, c, "GET /mse6/slowbody?wait=3 HTTP/1.1\r\n")
<<<<<<< HEAD
	checkWrite(t, c, "Host: localhost:8081\r\n")
=======
	checkWrite(t, c, "Host: localhost:8080\r\n")
>>>>>>> ec999d0e2bacafccf74bd5d31ac4a2d43d5b1de5
	checkWrite(t, c, "User-Agent: integration\r\n")
	checkWrite(t, c, "Accept: */*\r\n")
	checkWrite(t, c, "\r\n")

	//step 3 we try to read the server response. Warning this isn't a proper http client
	//i.e. doesn't include parsing content length, reading response properly.
	buf := make([]byte, 1024)
	l, err := c.Read(buf)
	t.Logf("jabba responded with %v bytes and error code %v", l, err)
	t.Logf("jabba partial response: %v", string(buf))
	if l == 0 {
		t.Error("jabba did not respond, 0 bytes read")
	}
	response := string(buf)
	if !strings.Contains(response, "Server: Jabba") {
		t.Error("jabba did not respond, server information not found on response ")
	}
}
