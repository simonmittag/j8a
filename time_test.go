package jabba

import (
	"os"
	"testing"
)

func TestTimezoneSet(t *testing.T) {
	want := "Australia/Sydney"
	os.Setenv("TZ", want)
	initTime()
	got := TZ
	if want != got {
		t.Errorf("timezone want %v, got %v", want, got)
	}
}

func TestTimezoneNotSet(t *testing.T) {
	want := "UTC"
	os.Setenv("TZ", "")
	initTime()
	got := TZ
	if want != got {
		t.Errorf("timezone want %v, got %v", want, got)
	}
}
