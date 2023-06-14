package j8a

import (
	"github.com/rs/zerolog"
	"os"
	"testing"
	"time"
)

func TestServerID(t *testing.T) {
	os.Setenv("HOSTNAME", "localhost")
	Version = "v0.0.0"
	initServerID()
	want := "f47f7b28"
	if ID != want {
		t.Errorf("serverID did not properly compute, got %v, want %v", ID, want)
	}
}

func TestDefaultLogLevelInit(t *testing.T) {
	initLogger()
	got := zerolog.GlobalLevel().String()
	want := "info"
	if got != want {
		t.Errorf("default log level not properly initialised, got %v, want %v", got, want)
	}
}

func TestLogLevelReset(t *testing.T) {
	tests := []struct {
		n string
		l string
	}{
		{"trace", "trace"},
		{"debug", "debug"},
		{"info", "info"},
		{"warn", "warn"},
	}

	for _, tt := range tests {
		t.Run(tt.n, func(t *testing.T) {
			c := Config{
				LogLevel: tt.l,
			}
			initLogger()
			c.validateLogLevel()

			Runner = &Runtime{
				Config:       c,
				StateHandler: NewStateHandler(),
			}
			Runner.StateHandler.setState(Daemon)
			Runner.resetLogLevel()

			time.Sleep(time.Millisecond * 1000)

			got := zerolog.GlobalLevel().String()
			want := tt.l
			if got != want {
				t.Errorf("log level not properly initialised, got %v, want %v", got, want)
			}
		})
	}
}

func TestBadLogLevelPanic(t *testing.T) {
	c := Config{
		LogLevel: "blah",
	}
	initLogger()

	shouldPanic(t, c.validateLogLevel)
}
