package integration

import (
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func TestServerHangsUpOnDownstreamIfReadTimeoutExceeded(t *testing.T) {
	//if this test fails, check the j8a configuration for connection.downstream.ReadTimeoutSeconds
	grace := 1
	serverReadTimeoutSeconds := 3
	wait := serverReadTimeoutSeconds + grace

	//step 1 make a connection
	c, err := net.Dial("tcp", ":8080")
	if err != nil {
		t.Errorf("test failure. unable to connect to j8a server for integration test, cause: %v", err)
	}

	//step 2 we send headers but do not terminate http message. this will cause j8a to wait for more data
	checkWrite(t, c, "GET /mse6/get HTTP/1.1\r\n")
	checkWrite(t, c, "Host: localhost:8080\r\n")

	//step 3 we sleep locally until we should hit timeout
	t.Logf("normal. going to sleep for %d seconds to trigger remote j8a read timeout", wait)
	time.Sleep(time.Second * time.Duration(wait))

	//step 4 we try to send another header however by now the server should have hung up on us.
	//this needs to be a loop of multiple writes because golang or the OS may buffer writes and not fail
	//immediately
	b := 0
	var err2 error

WriteLoop:
	for i := 0; i < 100; i++ {
		payload := []byte("User-Agent: integration\r\n")
		b, err2 = c.Write(payload)
		if err2 != nil {
			t.Logf("normal: exiting post sleep write loop after %d bytes", i*len(payload))
			break WriteLoop
		}
	}

	if err2 == nil {
		t.Errorf("error: %v ", err2)
		t.Errorf("bytes written: %d", b)
		t.Error("test failure. expected j8a server to hang up on us after 3s, but it didn't. check downstream read timeout")
	} else {
		t.Logf("normal: j8a hung up as expected with error: %v", err2)
	}
}

func TestServerConnectsNormallyWithoutHangingUp(t *testing.T) {
	//step 1 we connect to j8a with net.dial
	c, err := net.Dial("tcp", ":8080")
	if err != nil {
		t.Errorf("test failure. unable to connect to j8a server for integration test, cause: %v", err)
		return
	}
	defer c.Close()

	//step 2 we send headers and terminate HTTP message.
	checkWrite(t, c, "GET /mse6/get HTTP/1.1\r\n")
	checkWrite(t, c, "Host: localhost:8080\r\n")
	checkWrite(t, c, "User-Agent: integration\r\n")
	checkWrite(t, c, "Accept: */*\r\n")
	checkWrite(t, c, "\r\n")

	//step 3 we try to read the server response. Warning this isn't a proper http client
	//i.e. doesn't include parsing content length, nor reading response properly.
	buf := make([]byte, 1024)
	l, err := c.Read(buf)
	t.Logf("normal. j8a responded with %v bytes and error code %v", l, err)
	t.Logf("normal. j8a partial response: %v", string(buf))
	if l == 0 {
		t.Error("test failure. j8a did not respond, 0 bytes read")
	}
	response := string(buf)
	if !strings.Contains(response, "Server: j8a") {
		t.Error("test failure. j8a did not respond, server information not found on response ")
	}
}

func checkWrite(t *testing.T, c net.Conn, msg string) {
	j, err2 := c.Write([]byte(msg))
	if j == 0 || err2 != nil {
		t.Errorf("test failure. uh oh, unable to send data to j8a for integration test. bytes %v, err: %v", j, err2)
	} else {
		fmt.Printf("normal. sent %v bytes to j8a, content %v", j, msg)
	}
}
