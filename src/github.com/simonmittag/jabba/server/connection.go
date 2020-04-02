package server

// Connection Params
type Connection struct {
	Server Server
	Client Client
}

// Server params for the HTTP or TLS server that Jabba exposes
type Server struct {
	// ReadTimeoutSeconds is the maximum duration for reading the entire
	// request, including the body, the downstream user agent sends to us.
	ReadTimeoutSeconds int

	// WriteTimeoutSeconds is the maximum duration round trip time in seconds any
	// single request spends in the server, this includes the time to read the request,
	// processing upstream attempts and writing the response.
	RoundTripTimeoutSeconds int
}

// Client params for connections to upstream servers that are being proxied
type Client struct {

	// PoolSize is the maximum size of the socket connection pool for idle connections
	PoolSize int

	// KeepAliveTimeoutSeconds is the total wait period in seconds before we give up on an idle upstream connection.
	KeepAliveTimeoutSeconds int

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
}
