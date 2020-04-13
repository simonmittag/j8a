package server

import (
	"testing"
)

//TestDefaultTimeouts checks that config is overridden with proper defaults
func TestDefaultDownstreamReadTimeout(t *testing.T) {
	config := new(Config).setDefaultTimeouts()
	t1 := config.Connection.Downstream.ReadTimeoutSeconds
	if t1 != 120 {
		t.Errorf("default config.Connection.Downstream.ReadTimeoutSeconds got %d, want %d", t1, 120)
	}
}

func TestDefaultDownstreamIdleTimeout(t *testing.T) {
	config := new(Config).setDefaultTimeouts()
	t1 := config.Connection.Downstream.IdleTimeoutSeconds
	if t1 != 120 {
		t.Errorf("default config.Connection.Downstream.IdleTimeoutSeconds got %d, want %d", t1, 120)
	}
}

func TestDefaultDownstreamRoundtripTimeout(t *testing.T) {
	config := new(Config).setDefaultTimeouts()
	t1 := config.Connection.Downstream.RoundTripTimeoutSeconds
	if t1 != 240 {
		t.Errorf("default config.Connection.Downstream.IdleTimeoutSeconds got %d, want %d", t1, 240)
	}
}
