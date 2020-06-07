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
	serverRoundTripTimeoutSeconds := 20
	//assumes upstreamReadTimeoutSeconds := 30 so it doesn't fire before serverRoundTripTimeoutSeconds
	wait := serverRoundTripTimeoutSeconds+grace

	//step 1 make a connection
	c, err := net.Dial("tcp", ":8081")
	if err != nil {
		t.Errorf("test failure. unable to connect to jabba server for integration test, cause: %v", err)
	}

	//step 2 we send headers and terminate http message so Jabba sends request upstream.
	checkWrite(t, c, "GET /mse6/slowbody?wait=21 HTTP/1.1\r\n")
	checkWrite(t, c, "Host: localhost:8081\r\n")
	checkWrite(t, c, "\r\n")

	//step 3 we sleep locally until we should hit timeout
	t.Logf("normal. going to sleep for %d seconds to trigger remote jabba roundtrip timeout", wait)
	time.Sleep(time.Second * time.Duration(wait))


	//step 4 we read a response into buffer after timeout which has to fail immediately
	//because the server already hung up on us
	buf := make([]byte, 1024)
	b, err2 := c.Read(buf)
	t.Logf("normal. jabba responded with %v bytes and error code %v", b, err)
	t.Logf("normal. jabba partial response (this should be empty): %v", string(buf))

	if err2.Error() != "EOF" {
		t.Errorf("test failure. error: %v ", err2)
		t.Errorf("test failure. bytes written: %d", b)
		t.Errorf("test failure. expected jabba server to hang up on us after %ds, but it didn't. check downstream roundtrip timeout", serverRoundTripTimeoutSeconds)
	} else {
		t.Logf("normal. jabba hung up as expected with error: %v", err2)
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
	checkWrite(t, c, "Host: localhost:8081\r\n")
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
