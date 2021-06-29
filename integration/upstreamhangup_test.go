package integration

import "testing"

func TestServer2UpstreamHangupSends502ForGETDuringHeader(t *testing.T) {
	performJ8aTest(t,
		"/hangupduringheader",
		2,
		8,
		502,
		8081,
		false)
}

func TestServer2UpstreamHangupSends502ForGETAfterHeader(t *testing.T) {
	performJ8aTest(t,
		"/hangupafterheader",
		2,
		8,
		502,
		8081,
		false)
}

func TestServer2UpstreamHangupSends502ForGETDuringBody(t *testing.T) {
	performJ8aTest(t,
		"/hangupduringbody",
		2,
		8,
		502,
		8081,
		false)
}
