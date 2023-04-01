package j8a

import (
	"github.com/rs/zerolog"
	"os"
	"testing"
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

func TestTraceLogLevelInit(t *testing.T) {
	c := Config{
		LogLevel: "trace",
	}
	initLogger()
	c.validateLogLevel()
	got := zerolog.GlobalLevel().String()
	want := "trace"
	if got != want {
		t.Errorf("log level not properly initialised, got %v, want %v", got, want)
	}
}

func TestDebugLogLevelInit(t *testing.T) {
	c := Config{
		LogLevel: "debug",
	}
	initLogger()
	c.validateLogLevel()
	got := zerolog.GlobalLevel().String()
	want := "debug"
	if got != want {
		t.Errorf("log level not properly initialised, got %v, want %v", got, want)
	}
}

func TestInfoLogLevelInit(t *testing.T) {
	c := Config{
		LogLevel: "INFO",
	}
	initLogger()
	c.validateLogLevel()
	got := zerolog.GlobalLevel().String()
	want := "info"
	if got != want {
		t.Errorf("log level not properly initialised, got %v, want %v", got, want)
	}
}

func TestWarnLogLevelInit(t *testing.T) {
	c := Config{
		LogLevel: "warn",
	}
	initLogger()
	c.validateLogLevel()
	got := zerolog.GlobalLevel().String()
	want := "warn"
	if got != want {
		t.Errorf("log level not properly initialised, got %v, want %v", got, want)
	}
}

func TestBadLogLevelPanic(t *testing.T) {
	c := Config{
		LogLevel: "blah",
	}
	initLogger()

	shouldPanic(t, c.validateLogLevel)
}
