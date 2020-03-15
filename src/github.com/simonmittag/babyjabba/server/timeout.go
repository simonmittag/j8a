package server

//Timeout describes various HTTP timeouts both directions
type Timeout struct {
	// DownstreamReadTimeoutSeconds is the maximum duration for reading the entire
	// request, including the body.
	DownstreamReadTimeoutSeconds int

	DownstreamWriteTimeoutSeconds int

	UpstreamConnectTimeoutSeconds int

	UpstreamReadTimeoutSeconds int
}
