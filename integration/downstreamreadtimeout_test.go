package integration

import (
	"net"
	"strings"
	"testing"
	"time"
)

func TestServerHangsUpOnDownstreamIfReadTimeoutExceeded(t *testing.T) {
	c, err := net.Dial("tcp", ":8080")
	if err != nil {
		t.Errorf("unable to connect to jabba server for integration test, cause: %v", err)
	}

	//we are a slow client. send headers then sleep
	c.Write([]byte("GET /mse6/some/get HTTP/1.1\n"))
	time.Sleep(time.Second * 3)

	//now try to send more headers
	j, err2 := c.Write([]byte("Host: localhost\n"))
	if j > 0 || err2 == nil {
		t.Error("expected jabba server to hang up on us after 3s, but it didn't. check downstream read timeout")
	}
}

func TestServerConnectsNormallyWithoutHangingUp(t *testing.T) {
	//step 1 we connect to Jabba with net.dial
	c, err := net.Dial("tcp", ":8087")
	if err != nil {
		t.Errorf("unable to connect to jabba server for integration test, cause: %v", err)
		return
	}
	defer c.Close()

	//step 2 we send headers and terminate HTTP message.
	checkWrite(t, c, "GET /mse6/some/get HTTP/1.1\r\n")
	checkWrite(t, c, "Host: localhost:8087\r\n")
	checkWrite(t, c, "User-Agent: integrationTest\r\n")
	checkWrite(t, c, "Accept: */*\r\n")
	checkWrite(t, c, "\r\n")

	//step 3 we try to read the server response. Warning this isn't a proper http client
	//i.e. doesn't include parsing content length, reading response properly.
	buf := make([]byte, 1024)
	l, err := c.Read(buf)
	t.Logf("jabba responded with %v bytes and error code %v", l, err)
	t.Logf("jabba partial response: %v", string(buf))
	if l == 0 {
		t.Error("jabba did not respond")
	}
	response := string(buf)
	if !strings.Contains(response, "Server: Jabba") {
		t.Error("jabba did not respond")
	}
}

func checkWrite(t *testing.T, c net.Conn, msg string) {
	j, err2 := c.Write([]byte(msg))
	if j == 0 || err2 != nil {
		t.Errorf("uh oh, unable to send data to jabba for integration test. bytes %v, err: %v", j, err2)
	}
}
