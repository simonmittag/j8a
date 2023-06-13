package j8a

import (
	"testing"
	"time"
)

func TestStateLesser(t *testing.T) {
	tests := []struct {
		n string
		a State
		b State
		v bool
	}{
		{n: "Bootstrap not lesser Bootstrap", a: Bootstrap, b: Bootstrap, v: false},
		{n: "Bootstrap lesser Daemon", a: Bootstrap, b: Daemon, v: true},
		{n: "Bootstrap lesser Shutdown", a: Bootstrap, b: Shutdown, v: true},
		{n: "Daemon not lesser Bootstrap", a: Daemon, b: Bootstrap, v: false},
		{n: "Daemon not lesser Daemon", a: Daemon, b: Daemon, v: false},
		{n: "Daemon lesser Shutdown", a: Daemon, b: Shutdown, v: true},
		{n: "Shutdown not lesser Daemon", a: Shutdown, b: Daemon, v: false},
		{n: "Shutdown not lesser Bootstrap", a: Shutdown, b: Bootstrap, v: false},
		{n: "Shutdown not lesser Shutdown", a: Shutdown, b: Shutdown, v: false},
	}

	for _, tt := range tests {
		t.Run(tt.n, func(t *testing.T) {
			want := tt.v
			got := tt.a.Lesser(tt.b)
			if want != got {
				t.Errorf("%v failed, want %v got %v", tt.n, want, got)
			}
		})
	}
}

// Tests pass under the timeout threshold if a status is lesser or equal the current status.
// results greater the treshold means the StateHandler would otherwise wait indefinitely for the status.
func TestStateHandlerWaitForStatus(t *testing.T) {
	tests := []struct {
		n            string
		current      State
		waitingFor   State
		delaySeconds int
		greater      bool
	}{
		{"Bootstrap waiting for Bootstrap", Bootstrap, Bootstrap, 1, false},
		{"Bootstrap waiting for Daemon", Bootstrap, Daemon, 1, true},
		{"Bootstrap waiting for Shutdown", Bootstrap, Shutdown, 1, true},
		{"Daemon waiting for Bootstrap", Daemon, Bootstrap, 1, false},
		{"Daemon waiting for Daemon", Daemon, Daemon, 1, false},
		{"Daemon waiting for Shutdown", Daemon, Shutdown, 1, true},
		{"Shutdown waiting for Bootstrap", Shutdown, Bootstrap, 1, false},
		{"Shutdown waiting for Daemon", Shutdown, Daemon, 1, false},
		{"Shutdown waiting for Shutdown", Shutdown, Shutdown, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.n, func(t *testing.T) {
			sh := NewStateHandler()
			//TODO this blocks
			sh.setState(tt.current)

			before := time.Now()
			sh.waitState(tt.waitingFor, tt.delaySeconds)
			after := time.Now().Sub(before)
			if after > time.Second*time.Duration(tt.delaySeconds) == tt.greater {
				t.Logf("normal. current status %v waiting for %v delay %v", tt.current, tt.waitingFor, after)
			} else {
				t.Errorf("current status %v waiting for %v delay %v", tt.current, tt.waitingFor, after)
			}
		})
	}
}
