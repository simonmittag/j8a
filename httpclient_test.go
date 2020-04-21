package jabba

import (
	"os"
	"testing"
	"time"
)

//TestDefaultDownstreamReadTimeout
func TestGetTcpCntAndKeepAliveIntervalDuration(t *testing.T) {
	Runner = &Runtime{
		Config: Config{
			Connection: Connection{
				Upstream: Upstream{
					IdleTimeoutSeconds: 120,
				},
			},
		},
	}

	os.Setenv("OS", "linux")
	got := getTCPKeepCnt()
	want := 9
	if got != want {
		t.Errorf("incorrect linux tcp cnt interval for socket timeout test, got %v, want %v", got, want)
	}
	gotKeepAlive := getKeepAliveIntervalDuration()
	wantKeepAlive := time.Duration(120 / float64(got) * 1000000000)
	if gotKeepAlive != wantKeepAlive {
		t.Errorf("incorrect linux keepalive interval duration, got %v, want %v", gotKeepAlive, wantKeepAlive)
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
