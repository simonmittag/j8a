package server

import (
	"testing"
)

//TestDefaultDownstreamReadTimeout
func TestDefaultDownstreamReadTimeout(t *testing.T) {
	config := new(Config).setDefaultValues()
	got := config.Connection.Downstream.ReadTimeoutSeconds
	want := 120
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

//TestDefaultDownstreamIdleTimeout
func TestDefaultDownstreamIdleTimeout(t *testing.T) {
	config := new(Config).setDefaultValues()
	got := config.Connection.Downstream.IdleTimeoutSeconds
	want := 120
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

//TestDefaultDownstreamRoundtripTimeout
func TestDefaultDownstreamRoundtripTimeout(t *testing.T) {
	config := new(Config).setDefaultValues()
	got := config.Connection.Downstream.RoundTripTimeoutSeconds
	want := 240
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

//TestDefaultUpstreamSocketTimeout
func TestDefaultUpstreamSocketTimeout(t *testing.T) {
	config := new(Config).setDefaultValues()
	got := config.Connection.Upstream.SocketTimeoutSeconds
	want := 3
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

//TestDefaultUpstreamSocketTimeout
func TestDefaultUpstreamReadTimeout(t *testing.T) {
	config := new(Config).setDefaultValues()
	got := config.Connection.Upstream.ReadTimeoutSeconds
	want := 120
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

//TestDefaultUpstreamIdleTimeout
func TestDefaultUpstreamIdleTimeout(t *testing.T) {
	config := new(Config).setDefaultValues()
	got := config.Connection.Upstream.IdleTimeoutSeconds
	want := 120
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

//TestDefaultUpstreamConnectionPoolSize
func TestDefaultUpstreamConnectionPoolSize(t *testing.T) {
	config := new(Config).setDefaultValues()
	got := config.Connection.Upstream.PoolSize
	want := 32768
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

//TestDefaultUpstreamConnectionMaxAttempts
func TestDefaultUpstreamConnectionMaxAttempts(t *testing.T) {
	config := new(Config).setDefaultValues()
	got := config.Connection.Upstream.MaxAttempts
	want := 1
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}
