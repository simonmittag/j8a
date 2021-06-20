package integration

import (
	"net"
	"testing"
)

func TestCancelDownstreamWhileUpstreamStillServing(t *testing.T) {
	//step 1 we connect to j8a with net.dial
	c, err := net.Dial("tcp", ":8081")
	if err != nil {
		t.Errorf("unable to connect to j8a server for integration test, cause: %v", err)
		return
	}

	//step 2 we send headers and terminate HTTP message.
	checkWrite(t, c, "GET /mse6/slowbody?wait=5 HTTP/1.1\r\n")
	checkWrite(t, c, "Host: localhost:8081\r\n")
	checkWrite(t, c, "User-Agent: integration\r\n")
	checkWrite(t, c, "Accept: */*\r\n")
	checkWrite(t, c, "\r\n")

	c.Close()
	t.Logf("normal. we are now closing the connection to J8a. it should drop it's connection to upstream immediately")
}