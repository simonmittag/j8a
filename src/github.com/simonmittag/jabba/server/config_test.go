package server

import (
	"testing"
)

//TestDefaultDownstreamReadTimeout checks that config is overridden
func TestDefaultDownstreamReadTimeout(t *testing.T) {
	config := new(Config).setDefaultValues()
	got := config.Connection.Downstream.ReadTimeoutSeconds
	want := 120
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

//TestDefaultDownstreamIdleTimeout checks that config is overridden
func TestDefaultDownstreamIdleTimeout(t *testing.T) {
	config := new(Config).setDefaultValues()
	got := config.Connection.Downstream.IdleTimeoutSeconds
	want := 120
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

//TestDefaultDownstreamRoundtripTimeout checks that config is overridden
func TestDefaultDownstreamRoundtripTimeout(t *testing.T) {
	config := new(Config).setDefaultValues()
	got := config.Connection.Downstream.RoundTripTimeoutSeconds
	want := 240
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

//TestDefaultUpstreamSocketTimeout checks that config is overridden
func TestDefaultUpstreamSocketTimeout(t *testing.T) {
	config := new(Config).setDefaultValues()
	got := config.Connection.Upstream.SocketTimeoutSeconds
	want := 3
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

//TestDefaultUpstreamSocketTimeout checks that config is overridden
func TestDefaultUpstreamReadTimeout(t *testing.T) {
	config := new(Config).setDefaultValues()
	got := config.Connection.Upstream.ReadTimeoutSeconds
	want := 120
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

//TestDefaultUpstreamIdleTimeout checks that config is overridden
func TestDefaultUpstreamIdleTimeout(t *testing.T) {
	config := new(Config).setDefaultValues()
	got := config.Connection.Upstream.IdleTimeoutSeconds
	want := 120
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

//TestDefaultUpstreamConnectionPoolSize checks that config is overridden
func TestDefaultUpstreamConnectionPoolSize(t *testing.T) {
	config := new(Config).setDefaultValues()
	got := config.Connection.Upstream.PoolSize
	want := 32768
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

//TestDefaultUpstreamConnectionMaxAttempts checks that config is overridden
func TestDefaultUpstreamConnectionMaxAttempts(t *testing.T) {
	config := new(Config).setDefaultValues()
	got := config.Connection.Upstream.MaxAttempts
	want := 1
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}
