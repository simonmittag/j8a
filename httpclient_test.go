package j8a

import (
	"os"
	"testing"
	"time"
)

func TestHttpClientSocketTimeout(t *testing.T) {
	Runner = &Runtime{
		Config: Config{
			Connection: Connection{
				Upstream: Upstream{
					SocketTimeoutSeconds: 1,
				},
			},
		},
		Start: time.Now(),
	}

	client := scaffoldHTTPClient(*Runner)
	start := time.Now()
	_, err := client.Get("http://10.73.124.255:2/uri")
	elapsed := time.Since(start)
	want := time.Duration(1 * time.Second)
	if elapsed < want {
		t.Errorf("socket timeout was not respected, client aborted too early, wanted > %v, got %v", want, elapsed)
	}
	if err == nil {
		t.Errorf("uh oh http client not meant to resolve got no error for non existing URL")
	}
}

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
		Start: time.Now(),
	}

	os.Setenv("OS", "linux")
	got := getTCPKeepCnt()
	want := 9
	if got != want {
		t.Errorf("incorrect linux tcp cnt interval for socket timeout test, got %v, want %v", got, want)
	}
	gotKeepAlive := getKeepAliveIntervalDuration().Nanoseconds()
	//nanos for keepAlive interval
	wantKeepAlive := int64(120 / float64(got) * 1000000000)
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
