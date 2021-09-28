package downstreamHttp

import (
	"github.com/simonmittag/j8a/integration"
	"net"
	"testing"
	"time"
)

//you cannot observe the 499 the remote end tries to send because we hang up socket. check server logs
func TestCancelDownstreamBeforeUpstreamConnection(t *testing.T) {
	//step 1 we connect to j8a with net.dial
	c, err := net.Dial("tcp", ":8081")
	if err != nil {
		t.Errorf("unable to connect to j8a server for integration test, cause: %v", err)
		return
	}

	//step 2 we send headers and terminate HTTP message.
	integration.CheckWrite(t, c, "GET /mse6/slowbody?wait=5 HTTP/1.1\r\n")
	integration.CheckWrite(t, c, "Host: localhost:8081\r\n")
	integration.CheckWrite(t, c, "User-Agent: integration\r\n")
	integration.CheckWrite(t, c, "Accept: */*\r\n")
	integration.CheckWrite(t, c, "\r\n")

	c.Close()
	t.Logf("normal. we are now closing the connection to J8a. it should drop it's connection to upstream immediately")
}

//you cannot observe the 499 the remote end tries to send because we hang up socket. check server logs
func TestCancelDownstreamWhileUpstreamStillServing(t *testing.T) {
	//step 1 we connect to j8a with net.dial
	c, err := net.Dial("tcp", ":8081")
	if err != nil {
		t.Errorf("unable to connect to j8a server for integration test, cause: %v", err)
		return
	}

	//step 2 we send headers and terminate HTTP message.
	integration.CheckWrite(t, c, "GET /mse6/slowbody?wait=5 HTTP/1.1\r\n")
	integration.CheckWrite(t, c, "Host: localhost:8081\r\n")
	integration.CheckWrite(t, c, "User-Agent: integration\r\n")
	integration.CheckWrite(t, c, "Accept: */*\r\n")
	integration.CheckWrite(t, c, "\r\n")

	//wait 1 second before killing connection, enough for j8a to make upstream attempt.
	time.Sleep(2 * time.Second)

	c.Close()
	t.Logf("normal. we are now closing the connection to J8a. it should drop it's connection to upstream immediately")
}
