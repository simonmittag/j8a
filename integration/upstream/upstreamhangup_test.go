package upstream

import (
	"github.com/simonmittag/j8a/integration"
	"testing"
)

func TestServer2UpstreamHangup(t *testing.T) {
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
			Name:                    "Sends502ForGETDuringHeader",
			TestMethod:              "/hangupduringheader",
			WantUpstreamWaitSeconds: 2,
			WantTotalWaitSeconds:    8,
			WantStatusCode:          502,
			ServerPort:              8081,
			TlsMode:                 false,
		},
		{
			Name:                    "Sends502ForGETDuringHeader",
			TestMethod:              "/hangupafterheader",
			WantUpstreamWaitSeconds: 2,
			WantTotalWaitSeconds:    8,
			WantStatusCode:          502,
			ServerPort:              8081,
			TlsMode:                 false,
		},
		{
			Name:                    "Sends502ForGETDuringHeader",
			TestMethod:              "/hangupduringbody",
			WantUpstreamWaitSeconds: 2,
			WantTotalWaitSeconds:    8,
			WantStatusCode:          502,
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
