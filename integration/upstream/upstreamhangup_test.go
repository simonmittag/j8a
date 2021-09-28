package upstream

import (
	"github.com/simonmittag/j8a/integration"
	"testing"
)

func TestServer2UpstreamHangupSends502ForGETDuringHeader(t *testing.T) {
	integration.PerformJ8aTest(t,
		"/hangupduringheader",
		2,
		8,
		502,
		8081,
		false)
}

func TestServer2UpstreamHangupSends502ForGETAfterHeader(t *testing.T) {
	integration.PerformJ8aTest(t,
		"/hangupafterheader",
		2,
		8,
		502,
		8081,
		false)
}

func TestServer2UpstreamHangupSends502ForGETDuringBody(t *testing.T) {
	integration.PerformJ8aTest(t,
		"/hangupduringbody",
		2,
		8,
		502,
		8081,
		false)
}
