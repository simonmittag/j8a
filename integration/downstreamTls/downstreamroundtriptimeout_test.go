package downstreamTls

import (
	"github.com/simonmittag/j8a/integration"
	"testing"
)

func TestServer3DownstreamRoundTripTimeout(t *testing.T) {
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
			Name:                    "NotFireWithSlowBody2S",
			TestMethod:              "/slowbody",
			WantUpstreamWaitSeconds: 2,
			WantTotalWaitSeconds:    2,
			WantStatusCode:          200,
			ServerPort:              8443,
			TlsMode:                 true,
		},
		{
			Name:                    "FireWithSlowHeader31S",
			TestMethod:              "/slowheader",
			WantUpstreamWaitSeconds: 31,
			WantTotalWaitSeconds:    20,
			WantStatusCode:          504,
			ServerPort:              8443,
			TlsMode:                 true,
		},
		{
			Name:                    "FireWithSlowBody31S",
			TestMethod:              "/slowbody",
			WantUpstreamWaitSeconds: 31,
			WantTotalWaitSeconds:    20,
			WantStatusCode:          504,
			ServerPort:              8443,
			TlsMode:                 true,
		},
		{
			Name:                    "FireWithSlowHeader25S",
			TestMethod:              "/slowheader",
			WantUpstreamWaitSeconds: 25,
			WantTotalWaitSeconds:    20,
			WantStatusCode:          504,
			ServerPort:              8443,
			TlsMode:                 true,
		},
		{
			Name:                    "FireWithSlowBody25S",
			TestMethod:              "/slowbody",
			WantUpstreamWaitSeconds: 25,
			WantTotalWaitSeconds:    20,
			WantStatusCode:          504,
			ServerPort:              8443,
			TlsMode:                 true,
		},
		{
			Name:                    "NotFireWithSlowHeader4S",
			TestMethod:              "/slowheader",
			WantUpstreamWaitSeconds: 4,
			WantTotalWaitSeconds:    4,
			WantStatusCode:          200,
			ServerPort:              8443,
			TlsMode:                 true,
		},
		{
			Name:                    "NotFireWithSlowBody4S",
			TestMethod:              "/slowbody",
			WantUpstreamWaitSeconds: 4,
			WantTotalWaitSeconds:    4,
			WantStatusCode:          200,
			ServerPort:              8443,
			TlsMode:                 true,
		},
		{
			Name:                    "NotFireWithSlowHeader2S",
			TestMethod:              "/slowheader",
			WantUpstreamWaitSeconds: 2,
			WantTotalWaitSeconds:    2,
			WantStatusCode:          200,
			ServerPort:              8443,
			TlsMode:                 true,
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
