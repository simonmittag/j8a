package server

// Connection Params
type Connection struct {
	Server Server
	Client Client
}

// Server params for the HTTP or TLS server that Jabba exposes
type Server struct {
	// ReadTimeoutSeconds is the maximum duration for reading the entire
	// request, including the body, the remote user agent sends to us.
	ReadTimeoutSeconds int

	// WriteTimeoutSeconds is the maximum duration for reading the entire
	// request, including the body, then sending the full response to the remote user agent.
	// Value cannot be less than DownstreamReadTimeoutSeconds.
	WriteTimeoutSeconds int
}

// Client params for connections to upstream servers that are being proxied
type Client struct {

	// ConnectionPoolSize is the Size of the connection pool for idle connections
	TCPConnectionPoolSize int

	// TCPConnectionKeepAliveSeconds is the wait period before a connection is closed.
	TCPConnectionKeepAliveSeconds int

	// ConnectTimeoutSeconds is the wait period to establish socket connection with an upstream server.
	// This setting controls simple TCP connect for HTTP upstream servers and handshake time for TLS if applicable.
	ConnectTimeoutSeconds int

	// ReadTimeoutSeconds is the wait period to read the entire upstream response once connection was established
	ReadTimeoutSeconds int

	// MaxAttempts is the maximum number of request attempts to obtain a successful response for repeatable requests.
	MaxAttempts int
}
