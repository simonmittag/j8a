package server

// Connection Params for BabyJabba
type Connection struct {
	Server Server
	Client Client
}

// Server params for the HTTP or TLS server that BabyJabba exposes
type Server struct {
	// ReadTimeoutSeconds is the maximum duration for reading the entire
	// request, including the body, the remote user agent sends to us.
	ReadTimeoutSeconds int

	// WriteTimeoutSeconds is the maximum duration for reading the entire
	// request, including the body, then sending the full response to the remote user agent.
	// Value cannot be less than DownstreamReadTimeoutSeconds.
	WriteTimeoutSeconds int
}

// Client params for connections to upstream servers that are being proxied by BabyJabba
type Client struct {

	//ConnectTimeoutSeconds is the wait period to establish socket connection with an upstream server
	ConnectTimeoutSeconds int

	//ReadTimeoutSeconds is the wait period to read the entire upstream response
	ReadTimeoutSeconds int
}
