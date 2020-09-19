package j8a

// Connection Params
type Connection struct {
	Downstream Downstream
	Upstream   Upstream
}

// Downstream params for the HTTP or TLS server that j8a exposes
type Downstream struct {
	// ReadTimeoutSeconds is the maximum duration for reading the entire
	// request, including the body, the downstream user agent sends to us.
	ReadTimeoutSeconds int

	// WriteTimeoutSeconds is the maximum duration round trip time in seconds any
	// single request spends in the server, this includes the time to read the request,
	// processing upstream attempts and writing the response into downstream socket.
	RoundTripTimeoutSeconds int

	// IdleTimeoutSeconds is the maximum duration, a downstream idle socket connection is kept open
	// before the server hangs up on the downstream user agent.
	IdleTimeoutSeconds int

	// Serving Mode, can be "TLS"
	Mode string

	// Serving on this port
	Port int

	// TLS x509 certificate
	Cert string

	// TLS secret key
	Key string
}

// Upstream connection params for remote servers that are being proxied
type Upstream struct {

	// PoolSize is the maximum size of the client socket connection pool for idle connections
	PoolSize int

	// IdleTimeoutSeconds is the total wait period in seconds before we hang up on an idle upstream connection.
	IdleTimeoutSeconds int

	// SocketTimeoutSeconds is the wait period to establish socket connection with an upstream server.
	// This setting controls roundtrip time for simple TCP connections, combined with handshake time for TLS
	// if applicable.
	SocketTimeoutSeconds int

	// ReadTimeoutSeconds is the wait period to read the entire upstream response once connection was established
	// before an individual upstream request is aborted
	ReadTimeoutSeconds int

	// MaxAttempts is the maximum allowable number of request attempts to obtain a successful response for repeatable
	// HTTP requests.
	MaxAttempts int

	// TlsInsecureSkipVerify skips the host name validation and certificate chain verification of upstream connections
	// using TLS. Use this only for testing or if you know what you are doing. Defaults to false
	TlsInsecureSkipVerify bool
}
