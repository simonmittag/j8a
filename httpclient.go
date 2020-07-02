package jabba

import (
	"net"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/rs/zerolog/log"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
	Get(uri string) (*http.Response, error)
}

// scaffoldHTTPClient is a factory method that applies connection params to the transport layer of http.Client
func scaffoldHTTPClient(runtime Runtime) HTTPClient {
	idleConnTimeoutDuration := time.Duration(runtime.Connection.Upstream.IdleTimeoutSeconds) * time.Second
	tLSHandshakeTimeoutDuration := time.Duration(runtime.Connection.Upstream.SocketTimeoutSeconds) * time.Second
	socketTimeoutDuration := time.Duration(runtime.Connection.Upstream.SocketTimeoutSeconds) * time.Second
	readTimeoutDuration := time.Duration(runtime.Connection.Upstream.ReadTimeoutSeconds) * time.Second

	httpClient = &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   socketTimeoutDuration,
				KeepAlive: getKeepAliveIntervalDuration(),
			}).Dial,
			//TLS handshake timeout is the same as connection timeout
			TLSHandshakeTimeout: tLSHandshakeTimeoutDuration,
			MaxIdleConns:        runtime.Connection.Upstream.PoolSize,
			MaxIdleConnsPerHost: runtime.Connection.Upstream.PoolSize,
			IdleConnTimeout:     idleConnTimeoutDuration,
		},
		//Timeout: readTimeoutDuration,
	}

	log.Debug().
		Int("upMaxIdleConns", runtime.Connection.Upstream.PoolSize).
		Int("upMaxIdleConnsPerHost", runtime.Connection.Upstream.PoolSize).
		Float64("upTransportDialTimeoutSecs", socketTimeoutDuration.Seconds()).
		Float64("upTlsHandshakeTimeoutSecs", tLSHandshakeTimeoutDuration.Seconds()).
		Float64("upIdleConnTimeoutSecs", idleConnTimeoutDuration.Seconds()).
		Float64("upReadTimeoutSecs", readTimeoutDuration.Seconds()).
		Float64("upTransportDialKeepAliveIntervalSecs", getKeepAliveIntervalDuration().Seconds()).
		Msg("server derived upstream params")

	return httpClient
}

// getKeepAliveIntervalSecondsDuration. KeepAlive is effectively: initial delay + interval * TCP_KEEPCNT (9 on linux, 8 ox OSX).
// The KeepAliveIntervalSecondsDuration here defines interval, i.e. default 15s * 9 = 135s on linux
// See: https://github.com/golang/go/issues/23459#issuecomment-374777402
// The OS uses zero payload TCP segments to attempt to keep the connection alive.
// after the total number of unacknowledged TCP_KEEPCNT is reached, the dialer kills the
// connection.
func getKeepAliveIntervalDuration() time.Duration {
	return time.Duration(float64(Runner.
		Connection.
		Upstream.
		IdleTimeoutSeconds) / float64(getTCPKeepCnt()) * float64(time.Second))
}

func getTCPKeepCnt() int {
	os := os.Getenv("OS")
	if len(os) == 0 {
		os = runtime.GOOS
	}
	switch os {
	case "windows":
		return 5
	case "darwin", "freebsd", "openbsd":
		return 8
	case "linux":
		return 9
	//if we don't know, assume some kind of linux
	default:
		return 9
	}
}
