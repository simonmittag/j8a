package upstream

import (
	"crypto/tls"
	"github.com/simonmittag/j8a/integration"
	"testing"
)

var tlsConfig *tls.Config

func TestServer1UpstreamReadTimeout(t *testing.T) {
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
			Name:                    "FireWithSlowHeader31S",
			TestMethod:              "/slowheader",
			WantUpstreamWaitSeconds: 31,
			WantTotalWaitSeconds:    12,
			WantStatusCode:          504,
			ServerPort:              8080,
			TlsMode:                 false,
		},
		{
			Name:                    "FireWithSlowBody31S",
			TestMethod:              "/slowbody",
			WantUpstreamWaitSeconds: 31,
			WantTotalWaitSeconds:    12,
			WantStatusCode:          504,
			ServerPort:              8080,
			TlsMode:                 false,
		},
		{
			Name:                    "FireWithSlowHeader25S",
			TestMethod:              "/slowheader",
			WantUpstreamWaitSeconds: 25,
			WantTotalWaitSeconds:    12,
			WantStatusCode:          504,
			ServerPort:              8080,
			TlsMode:                 false,
		},
		{
			Name:                    "FireWithSlowBody25S",
			TestMethod:              "/slowbody",
			WantUpstreamWaitSeconds: 25,
			WantTotalWaitSeconds:    12,
			WantStatusCode:          504,
			ServerPort:              8080,
			TlsMode:                 false,
		},
		{
			Name:                    "FireWithSlowHeader4S",
			TestMethod:              "/slowheader",
			WantUpstreamWaitSeconds: 4,
			WantTotalWaitSeconds:    12,
			WantStatusCode:          504,
			ServerPort:              8080,
			TlsMode:                 false,
		},
		{
			Name:                    "FireWithSlowBody4S",
			TestMethod:              "/slowheader",
			WantUpstreamWaitSeconds: 4,
			WantTotalWaitSeconds:    12,
			WantStatusCode:          504,
			ServerPort:              8080,
			TlsMode:                 false,
		},
		{
			Name:                    "NotFireWithSlowHeader2S",
			TestMethod:              "/slowheader",
			WantUpstreamWaitSeconds: 2,
			WantTotalWaitSeconds:    2,
			WantStatusCode:          200,
			ServerPort:              8080,
			TlsMode:                 false,
		},
		{
			Name:                    "NotFireWithSlowBody2S",
			TestMethod:              "/slowbody",
			WantUpstreamWaitSeconds: 2,
			WantTotalWaitSeconds:    2,
			WantStatusCode:          200,
			ServerPort:              8080,
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
