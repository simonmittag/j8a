package integration

import (
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func TestServerHangsUpOnDownstreamIfReadTimeoutExceeded(t *testing.T) {
	//if this test fails, check the jabba configuration for connection.downstream.ReadTimeoutSeconds
	serverReadTimeoutSeconds := 4

	//step 1 make a connection
	c, err := net.Dial("tcp", ":8080")
	if err != nil {
		t.Errorf("unable to connect to jabba server for integration test, cause: %v", err)
	}

	//step 2 we send headers but do not terminate http message.
	checkWrite(t, c, "GET /mse6/get HTTP/1.1\r\n")
	checkWrite(t, c, "Host: localhost:8080\r\n")

	//step 3 we sleep locally until we should hit timeout
	t.Logf("going to sleep for %d seconds to trigger remote jabba read timeout", serverReadTimeoutSeconds)
	time.Sleep(time.Second * time.Duration(serverReadTimeoutSeconds))

	//step 4 we try to send another header however by now the server should have hung up on us.
	b := 0
	var err2 error

	WriteLoop:
	for i:=0;i<100;i++ {
		payload := []byte("User-Agent: integration\r\n")
		b, err2 = c.Write(payload)
		if err2 != nil {
			t.Logf("exiting post sleep write loop after %d bytes", i*len(payload))
			break WriteLoop
		}
	}

	if err2 == nil {
		t.Errorf("error: %v ", err2)
		t.Errorf("bytes written: %d", b)
		t.Error("expected jabba server to hang up on us after 3s, but it didn't. check downstream read timeout")
	} else {
		t.Logf("jabba hung up as expected with error: %v", err2)
	}
}

func TestServerConnectsNormallyWithoutHangingUp(t *testing.T) {
	//step 1 we connect to Jabba with net.dial
	c, err := net.Dial("tcp", ":8080")
	if err != nil {
		t.Errorf("unable to connect to jabba server for integration test, cause: %v", err)
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

func checkWrite(t *testing.T, c net.Conn, msg string) {
	j, err2 := c.Write([]byte(msg))
	if j == 0 || err2 != nil {
		t.Errorf("uh oh, unable to send data to jabba for integration test. bytes %v, err: %v", j, err2)
	} else {
		fmt.Printf("sent %v bytes to jabba, content %v", j, msg)
	}
}
