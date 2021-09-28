package downstreamTls

import (
	"github.com/simonmittag/j8a/integration"
	"testing"
)

func TestServer3DownstreamRoundTripTimeoutNotFireWithSlowBody2S(t *testing.T) {
	integration.PerformJ8aTest(t,
		"/slowbody",
		2,
		2,
		200,
		8443,
		true)
}

func TestServer3TLSDownstreamRoundTripTimeoutFireWithSlowHeader31S(t *testing.T) {
	integration.PerformJ8aTest(t,
		"/slowheader",
		31,
		20,
		504,
		8443,
		true)
}

func TestServer3TLSDownstreamRoundTripTimeoutFireWithSlowBody31S(t *testing.T) {
	integration.PerformJ8aTest(t,
		"/slowbody",
		31,
		20,
		504,
		8443,
		true)
}

func TestServer3TLSDownstreamRoundTripTimeoutFireWithSlowHeader25S(t *testing.T) {
	integration.PerformJ8aTest(t,
		"/slowheader",
		25,
		20,
		504,
		8443,
		true)
}

func TestServer3TLSDownstreamRoundTripTimeoutFireWithSlowBody25S(t *testing.T) {
	integration.PerformJ8aTest(t,
		"/slowbody",
		25,
		20,
		504,
		8443,
		true)
}

func TestServer3TLSDownstreamRoundTripTimeoutNotFireWithSlowHeader4S(t *testing.T) {
	integration.PerformJ8aTest(t,
		"/slowheader",
		4,
		4,
		200,
		8443,
		true)
}

func TestServer3TLSDownstreamRoundTripTimeoutNotFireWithSlowBody4S(t *testing.T) {
	integration.PerformJ8aTest(t,
		"/slowbody",
		4,
		4,
		200,
		8443,
		true)
}

func TestServer3TLSDownstreamRoundTripTimeoutNotFireWithSlowHeader2S(t *testing.T) {
	integration.PerformJ8aTest(t,
		"/slowheader",
		2,
		2,
		200,
		8443,
		true)
}
