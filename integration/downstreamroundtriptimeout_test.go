package integration

import (
	"net"
	"strings"
	"testing"
	"time"
)

func TestServer2HangsUpOnDownstreamIfRoundTripTimeoutExceeded(t *testing.T) {
	//if this test fails, check the jabba configuration for connection.downstream.ReadTimeoutSeconds
	grace := 1
	serverRoundTripTimeoutSeconds := 20
	//assumes upstreamReadTimeoutSeconds := 30 so it doesn't fire before serverRoundTripTimeoutSeconds
	wait := serverRoundTripTimeoutSeconds + grace

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

	//step 4 we read a response into buffer which returns 504
	buf := make([]byte, 16)
	b, err2 := c.Read(buf)
	t.Logf("normal. jabba responded with %v bytes", b)
	if err2 != nil || !strings.Contains(string(buf), "504") {
		t.Errorf("test failure. after timeout we should first experience a 504")
	}

	//step 5 now wait for the grace period, then re-read. the connection must now be closed.
	//note you must read in a loop with small buffer because golang's reader has cached data
	time.Sleep(time.Duration(grace) * time.Second)
	for i := 0; i < 32; i++ {
		b, err2 = c.Read(buf)
	}
	if err2 != nil && err2.Error() != "EOF" {
		t.Errorf("test failure. expected jabba server to hang up on us after %ds, but it didn't. check downstream roundtrip timeout", serverRoundTripTimeoutSeconds)
	} else {
		t.Logf("normal. jabba hung up connection as expected after grace period with error: %v", err2)
	}
}

func TestServer1RoundTripNormalWithoutHangingUp(t *testing.T) {
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

func TestServer2DownstreamRoundTripTimeoutFireWithSlowHeader31S(t *testing.T) {
	performJabbaTest(t,
		"/slowheader",
		31,
		20,
		504,
		8081,
		false)
}

func TestServer3TLSDownstreamRoundTripTimeoutFireWithSlowHeader31S(t *testing.T) {
	performJabbaTest(t,
		"/slowheader",
		31,
		20,
		504,
		8443,
		true)
}

func TestServer2DownstreamRoundTripTimeoutFireWithSlowBody31S(t *testing.T) {
	performJabbaTest(t,
		"/slowbody",
		31,
		20,
		504,
		8081,
		false)
}

func TestServer3TLSDownstreamRoundTripTimeoutFireWithSlowBody31S(t *testing.T) {
	performJabbaTest(t,
		"/slowbody",
		31,
		20,
		504,
		8443,
		true)
}

func TestServer2DownstreamRoundTripTimeoutFireWithSlowHeader25S(t *testing.T) {
	performJabbaTest(t,
		"/slowheader",
		25,
		20,
		504,
		8081,
		false)
}

func TestServer3TLSDownstreamRoundTripTimeoutFireWithSlowHeader25S(t *testing.T) {
	performJabbaTest(t,
		"/slowheader",
		25,
		20,
		504,
		8443,
		true)
}

func TestServer2DownstreamRoundTripTimeoutFireWithSlowBody25S(t *testing.T) {
	performJabbaTest(t,
		"/slowbody",
		25,
		20,
		504,
		8081,
		false)
}

func TestServer3TLSDownstreamRoundTripTimeoutFireWithSlowBody25S(t *testing.T) {
	performJabbaTest(t,
		"/slowbody",
		25,
		20,
		504,
		8443,
		true)
}

func TestServer2DownstreamRoundTripTimeoutNotFireWithSlowHeader4S(t *testing.T) {
	performJabbaTest(t,
		"/slowheader",
		4,
		4,
		200,
		8081,
		false)
}

func TestServer3TLSDownstreamRoundTripTimeoutNotFireWithSlowHeader4S(t *testing.T) {
	performJabbaTest(t,
		"/slowheader",
		4,
		4,
		200,
		8443,
		true)
}

func TestServer2DownstreamRoundTripTimeoutNotFireWithSlowBody4S(t *testing.T) {
	performJabbaTest(t,
		"/slowbody",
		4,
		4,
		200,
		8081,
		false)
}

func TestServer3TLSDownstreamRoundTripTimeoutNotFireWithSlowBody4S(t *testing.T) {
	performJabbaTest(t,
		"/slowbody",
		4,
		4,
		200,
		8443,
		true)
}

func TestServer2DownstreamRoundTripTimeoutNotFireWithSlowHeader2S(t *testing.T) {
	performJabbaTest(t,
		"/slowheader",
		2,
		2,
		200,
		8081,
		false)
}

func TestServer3TLSDownstreamRoundTripTimeoutNotFireWithSlowHeader2S(t *testing.T) {
	performJabbaTest(t,
		"/slowheader",
		2,
		2,
		200,
		8443,
		true)
}

func TestServer2DownstreamRoundTripTimeoutNotFireWithSlowBody2S(t *testing.T) {
	performJabbaTest(t,
		"/slowbody",
		2,
		2,
		200,
		8081,
		false)
}

func TestServer3DownstreamRoundTripTimeoutNotFireWithSlowBody2S(t *testing.T) {
	performJabbaTest(t,
		"/slowbody",
		2,
		2,
		200,
		8443,
		true)
}
