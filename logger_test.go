package jabba

import (
	"github.com/rs/zerolog"
	"os"
	"testing"
)

//TestDefaultDownstreamReadTimeout
func TestServerID(t *testing.T) {
	os.Setenv("HOSTNAME", "localhost")
	os.Setenv("VERSION", "v0.0.0")
	initServerID()
	want := "f47f7b28"
	if ID != want {
		t.Errorf("serverID did not properly compute, got %v, want %v", ID, want)
	}
}

func TestDefaultLogLevelInit(t *testing.T) {
	os.Setenv("LOGLEVEL", "not set")
	InitLogger()
	got := zerolog.GlobalLevel().String()
	want := "info"
	if got != want {
		t.Errorf("default log level not properly initialised, got %v, want %v", got, want)
	}
}

func TestTraceLogLevelInit(t *testing.T) {
	os.Setenv("LOGLEVEL", "TRACE")
	InitLogger()
	got := zerolog.GlobalLevel().String()
	want := "trace"
	if got != want {
		t.Errorf("default log level not properly initialised, got %v, want %v", got, want)
	}
}

func TestDebugLogLevelInit(t *testing.T) {
	os.Setenv("LOGLEVEL", "DEBUG")
	InitLogger()
	got := zerolog.GlobalLevel().String()
	want := "debug"
	if got != want {
		t.Errorf("log level not properly initialised, got %v, want %v", got, want)
	}
}

func TestInfoLogLevelInit(t *testing.T) {
	os.Setenv("LOGLEVEL", "INFO")
	InitLogger()
	got := zerolog.GlobalLevel().String()
	want := "info"
	if got != want {
		t.Errorf("log level not properly initialised, got %v, want %v", got, want)
	}
}

func TestWarnLogLevelInit(t *testing.T) {
	os.Setenv("LOGLEVEL", "WARN")
	InitLogger()
	got := zerolog.GlobalLevel().String()
	want := "warn"
	if got != want {
		t.Errorf("log level not properly initialised, got %v, want %v", got, want)
	}
}
