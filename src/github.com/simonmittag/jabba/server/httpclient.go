package server

import (
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/rs/zerolog/log"
)

func scaffoldHTTPClient() *http.Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: time.Duration(Runner.
						Connection.
						Client.
						ConnectTimeoutSeconds) * time.Second,
					KeepAlive: getKeepAliveIntervalSecondsDuration(),
				}).Dial,
				//TLS handshake timeout is the same as connection timeout
				TLSHandshakeTimeout: time.Duration(Runner.
					Connection.
					Client.
					ConnectTimeoutSeconds) * time.Second,
				MaxIdleConns:        Runner.Connection.Client.TCPConnectionPoolSize,
				MaxIdleConnsPerHost: Runner.Connection.Client.TCPConnectionPoolSize,
				// IdleConnTimeout: time.Duration(Runner.
				// 	Connection.
				// 	Client.
				// 	TCPConnectionKeepAliveSeconds),
			},
		}
		log.Debug().
			Int("upstreamTlsHandshakeTimeout", Runner.Connection.Client.ConnectTimeoutSeconds).
			Int("upstreamMaxIdleConns", Runner.Connection.Client.TCPConnectionPoolSize).
			Int("upstreamMaxIdleConnsPerHost", Runner.Connection.Client.TCPConnectionPoolSize).
			Int("upstreamIdleConnTimeout", Runner.Connection.Client.TCPConnectionKeepAliveSeconds).
			Msg("http client upstream params")
	}
	return httpClient
}

// KeepAlive is effectively: initial delay + interval * TCP_KEEPCNT (9 on linux).
// The KeepAliveIntervalSecondsDuration here defines interval, i.e. default 15s * 9 = 135s on linux
// See: https://github.com/golang/go/issues/23459#issuecomment-374777402
// The OS uses zero payload TCP segments to attempt to keep the connection alive.
// after the total number of unacknowledged TCP_KEEPCNT is reached, the dialer kills the
// connection.
func getKeepAliveIntervalSecondsDuration() time.Duration {
	return time.Duration(float64(Runner.
		Connection.
		Client.
		TCPConnectionKeepAliveSeconds)/float64(getTCP_KEEPCNT())) * time.Second
}

func getTCP_KEEPCNT() int {
	switch runtime.GOOS {
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
