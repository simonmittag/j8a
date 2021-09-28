package upstream

import (
	"crypto/tls"
	"github.com/simonmittag/j8a/integration"
	"testing"
)

var tlsConfig *tls.Config

func TestServer1UpstreamReadTimeoutFireWithSlowHeader31S(t *testing.T) {
	integration.PerformJ8aTest(t,
		"/slowheader",
		31,
		12,
		504,
		8080,
		false)
}

func TestServer1UpstreamReadTimeoutFireWithSlowBody31S(t *testing.T) {
	integration.PerformJ8aTest(t,
		"/slowbody",
		31,
		12,
		504,
		8080,
		false)
}

func TestServer1UpstreamReadTimeoutFireWithSlowHeader25S(t *testing.T) {
	integration.PerformJ8aTest(t,
		"/slowheader",
		25,
		12,
		504,
		8080,
		false)
}

func TestServer1UpstreamReadTimeoutFireWithSlowBody25S(t *testing.T) {
	integration.PerformJ8aTest(t,
		"/slowbody",
		25,
		12,
		504,
		8080,
		false)
}

func TestServer1UpstreamReadTimeoutFireWithSlowHeader4S(t *testing.T) {
	integration.PerformJ8aTest(t,
		"/slowheader",
		4,
		12,
		504,
		8080,
		false)
}

func TestServer1UpstreamReadTimeoutFireWithSlowBody4S(t *testing.T) {
	integration.PerformJ8aTest(t,
		"/slowbody",
		4,
		12,
		504,
		8080,
		false)
}

func TestServer1UpstreamReadTimeoutNotFireWithSlowHeader2S(t *testing.T) {
	integration.PerformJ8aTest(t,
		"/slowheader",
		2,
		2,
		200,
		8080,
		false)
}

func TestServer1UpstreamReadTimeoutNotFireWithSlowBody2S(t *testing.T) {
	integration.PerformJ8aTest(t,
		"/slowbody",
		2,
		2,
		200,
		8080,
		false)
}
