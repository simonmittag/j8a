package integration

import (
	"net"
	"testing"
	"time"
)


func TestServerHangsUpOnDownstreamIfReadTimeoutExceeded(t *testing.T) {
	c, err := net.Dial("tcp", ":8080")
	if err!=nil {
		t.Errorf("unable to connect to jabba server for integration test, cause: %v", err)
	}

	//we are a slow client. send headers then sleep
	c.Write([]byte("GET /mse6/some/get HTTP/1.1\n"))
	time.Sleep(time.Second*3)

	//now try to send more headers
	j, err2 := c.Write([]byte("Host: localhost\n"))
	if j>0 || err2 == nil {
		t.Error("expected jabba server to hang up on us after 3s, but it didn't. check downstream read timeout")
	}
}

func TestServerConnectsNormallyWithoutHangingUp(t *testing.T) {
	//step 1 we connect to Jabba with net.dial
	c, err := net.Dial("tcp", ":8080")
	if err!=nil {
		t.Errorf("unable to connect to jabba server for integration test, cause: %v", err)
		return
	}
	defer c.Close()

	//step 2 we send first header: GET /path HTTP/1.1
	j, err2 := c.Write([]byte("GET /mse6/some/get HTTP/1.1\r\n"))
	if j==0 || err2 != nil {
		t.Errorf("uh oh, unable to send data to jabba for integration test, cause: %v", err2)
		return
	}

	//step 3 we send second header 'Host'
	k, err3 := c.Write([]byte("Host: localhost\r\n"))
	if k==0 || err3 != nil {
		t.Errorf("uh oh, unable to send data to jabba for integration test, cause: %v", err3)
		return
	}

	//step 3.1 we send an extra carriage return line feed
	k1, err31 := c.Write([]byte("\r\n"))
	if k1==0 || err31 != nil {
		t.Errorf("uh oh, unable to send data to jabba for integration test, cause: %v", err31)
		return
	}

	//step 4 we try to read the server response
	var buf []byte
	for ;; {
		l, err4 := c.Read(buf)
		if l==0 || err4 != nil {
			t.Errorf("cannot read data from jabba socket, cause: %v", err4)
			return
		} else {
			t.Logf("jabba responded with: %v", buf)
		}
	}

}