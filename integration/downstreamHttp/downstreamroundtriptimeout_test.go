package downstreamHttp

import (
	"github.com/simonmittag/j8a/integration"
	"net"
	"strings"
	"testing"
	"time"
)

func TestServer1RoundTripNormalWithoutHangingUp(t *testing.T) {
	//step 1 we connect to j8a with net.dial
	c, err := net.Dial("tcp", ":8080")
	if err != nil {
		t.Errorf("unable to connect to j8a server for integration test, cause: %v", err)
		return
	}
	defer c.Close()

	//step 2 we send headers and terminate HTTP message.
	integration.CheckWrite(t, c, "GET /mse6/slowbody?wait=3 HTTP/1.1\r\n")
	integration.CheckWrite(t, c, "Host: localhost:8081\r\n")
	integration.CheckWrite(t, c, "User-Agent: integration\r\n")
	integration.CheckWrite(t, c, "Accept: */*\r\n")
	integration.CheckWrite(t, c, "\r\n")

	//step 3 we try to read the server response. Warning this isn't a proper http client
	//i.e. doesn't include parsing content length, reading response properly.
	buf := make([]byte, 1024)
	l, err := c.Read(buf)
	t.Logf("j8a responded with %v bytes and error code %v", l, err)
	t.Logf("j8a partial response: %v", string(buf))
	if l == 0 {
		t.Error("j8a did not respond, 0 bytes read")
	}
	response := string(buf)
	if !strings.Contains(response, "Server: j8a") {
		t.Error("j8a did not respond, server information not found on response ")
	}
}

func TestServer2HangsUpOnDownstreamIfRoundTripTimeoutExceeded(t *testing.T) {
	//if this test fails, check the j8a configuration for connection.downstream.ReadTimeoutSeconds
	grace := 1
	serverRoundTripTimeoutSeconds := 20
	//assumes upstreamReadTimeoutSeconds := 30 so it doesn't fire before serverRoundTripTimeoutSeconds
	wait := serverRoundTripTimeoutSeconds + grace

	//step 1 make a connection
	c, err := net.Dial("tcp", ":8081")
	if err != nil {
		t.Errorf("test failure. unable to connect to j8a server for integration test, cause: %v", err)
	}

	//step 2 we send headers and terminate http message so j8a sends request upstream.
	integration.CheckWrite(t, c, "GET /mse6/slowbody?wait=21 HTTP/1.1\r\n")
	integration.CheckWrite(t, c, "Host: localhost:8081\r\n")
	integration.CheckWrite(t, c, "\r\n")

	//step 3 we sleep locally until we should hit timeout
	t.Logf("normal. going to sleep for %d seconds to trigger remote j8a roundtrip timeout", wait)
	time.Sleep(time.Second * time.Duration(wait))

	//step 4 we read a response into buffer which returns 504
	buf := make([]byte, 16)
	b, err2 := c.Read(buf)
	t.Logf("normal. j8a responded with %v bytes", b)
	if err2 != nil || !strings.Contains(string(buf), "504") {
		t.Errorf("test failure. after timeout we should first experience a 504")
	}

	//step 5 now wait for the grace period, then re-read. the connection must now be closed.
	//note you must read in a loop with small buffer because golang's client reader in this test reads cached data
	time.Sleep(time.Duration(grace) * time.Second)
	for i := 0; i < 32; i++ {
		b, err2 = c.Read(buf)
	}
	if err2 != nil && err2.Error() != "EOF" {
		t.Errorf("test failure. expected j8a server to hang up on us after %ds, but it didn't. check downstream roundtrip timeout", serverRoundTripTimeoutSeconds)
	} else {
		t.Logf("normal. j8a hung up connection as expected after grace period with error: %v", err2)
	}
}

func TestServer2DownstreamRoundTripTimeout(t *testing.T) {
	var tcs = []struct {
		Name                    string
		TestMethod              string
		WantUpstreamWaitSeconds int
		WantTotalWaitSeconds    int
		WantStatusCode          int
		ServerPort              int
		TlsMode                 bool
	}{
		{
			Name:                    "FireWithSlowHeader31s",
			TestMethod:              "/slowheader",
			WantUpstreamWaitSeconds: 31,
			WantTotalWaitSeconds:    20,
			WantStatusCode:          504,
			ServerPort:              8081,
			TlsMode:                 false,
		},
		{
			Name:                    "FireWithSlowBody31s",
			TestMethod:              "/slowbody",
			WantUpstreamWaitSeconds: 31,
			WantTotalWaitSeconds:    20,
			WantStatusCode:          504,
			ServerPort:              8081,
			TlsMode:                 false,
		},
		{
			Name:                    "FireWithSlowHeader25S",
			TestMethod:              "/slowheader",
			WantUpstreamWaitSeconds: 25,
			WantTotalWaitSeconds:    20,
			WantStatusCode:          504,
			ServerPort:              8081,
			TlsMode:                 false,
		},
		{
			Name:                    "NotFireWithSlowHeader4S",
			TestMethod:              "/slowheader",
			WantUpstreamWaitSeconds: 4,
			WantTotalWaitSeconds:    4,
			WantStatusCode:          200,
			ServerPort:              8081,
			TlsMode:                 false,
		},
		{
			Name:                    "NotFireWithSlowBody4S",
			TestMethod:              "/slowbody",
			WantUpstreamWaitSeconds: 4,
			WantTotalWaitSeconds:    4,
			WantStatusCode:          200,
			ServerPort:              8081,
			TlsMode:                 false,
		},
		{
			Name:                    "NotFireWithSlowHeader2S",
			TestMethod:              "/slowheader",
			WantUpstreamWaitSeconds: 2,
			WantTotalWaitSeconds:    2,
			WantStatusCode:          200,
			ServerPort:              8081,
			TlsMode:                 false,
		},
		{
			Name:                    "NotFireWithSlowBody2S",
			TestMethod:              "/slowbody",
			WantUpstreamWaitSeconds: 2,
			WantTotalWaitSeconds:    2,
			WantStatusCode:          200,
			ServerPort:              8081,
			TlsMode:                 false,
		},
	}
	for _, tc := range tcs {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			integration.PerformJ8aTest(t,
				tc.TestMethod,
				tc.WantUpstreamWaitSeconds,
				tc.WantTotalWaitSeconds,
				tc.WantStatusCode,
				tc.ServerPort,
				tc.TlsMode)
		})
	}
}
