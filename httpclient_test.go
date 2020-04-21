package jabba

import (
	"os"
	"testing"
)

//TestDefaultDownstreamReadTimeout
func TestGetTcpCnt(t *testing.T) {
	os.Setenv("OS", "linux")
	got := getTCPKeepCnt()
	want := 9
	if got != want {
		t.Errorf("incorrect linux tcp cnt interval for socket timeout test, got %v, want %v", got, want)
	}

	os.Setenv("OS", "windows")
	got = getTCPKeepCnt()
	want = 5
	if got != want {
		t.Errorf("incorrect windows tcp cnt interval for socket timeout test, got %v, want %v", got, want)
	}

	os.Setenv("OS", "darwin")
	got = getTCPKeepCnt()
	want = 8
	if got != want {
		t.Errorf("incorrect darwin tcp cnt interval for socket timeout test, got %v, want %v", got, want)
	}

	os.Setenv("OS", "freebsd")
	got = getTCPKeepCnt()
	want = 8
	if got != want {
		t.Errorf("incorrect freebsd tcp cnt interval for socket timeout test, got %v, want %v", got, want)
	}

	os.Setenv("OS", "openbsd")
	got = getTCPKeepCnt()
	want = 8
	if got != want {
		t.Errorf("incorrect openbsd tcp cnt interval for socket timeout test, got %v, want %v", got, want)
	}

	os.Setenv("OS", "alpine")
	got = getTCPKeepCnt()
	want = 9
	if got != want {
		t.Errorf("incorrect other linux tcp cnt interval for socket timeout test, got %v, want %v", got, want)
	}
}
