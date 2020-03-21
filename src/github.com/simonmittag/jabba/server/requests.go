package server

import (
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/rs/zerolog/log"
)

var httpUpstreamMaxAttempts int

//RFC7231 4.2.1
var httpSafeMethods []string = []string{"GET", "HEAD", "OPTIONS", "TRACE"}

//RFC7231 4.2.2
var httpIdempotentMethods []string = []string{"PUT", "DELETE"}
var httpRepeatableMethods = append(httpSafeMethods, httpIdempotentMethods...)

// UpstreamAttempt wraps connection attempts to specific upstreams that are already mapped by label
type UpstreamAttempt struct {
	Upstream   *Upstream
	Label      string
	Count      int
	StatusCode int
}

// ProxyRequest wraps data for a single downstream request/response with multiple upstream HTTP request/response cycles.
type ProxyRequest struct {
	Method          string
	URI             string
	Body            []byte
	UpstreamAttempt UpstreamAttempt
}

func (proxyRequest *ProxyRequest) resolveUpstreamURI() string {
	return proxyRequest.UpstreamAttempt.Upstream.String() + proxyRequest.URI
}

// ShouldRepeat tells us if we can safely repeat the upstream request
func (proxyRequest *ProxyRequest) shouldRepeat() bool {
	for _, method := range httpRepeatableMethods {
		if proxyRequest.Method == method {
			if proxyRequest.UpstreamAttempt.Count < getHTTPMaxUpstreamAttempts() {
				return true
			}
			return false
		}
	}
	return false
}

// ParseIncoming is a factory method for a new ProxyRequest
func (proxyRequest *ProxyRequest) parseIncoming(request *http.Request) *ProxyRequest {
	url := request.URL
	method := request.Method
	body, _ := ioutil.ReadAll(request.Body)
	log.Trace().
		Str("url", url.Path).
		Str("method", method).
		Int("bodyBytes", len(body)).
		Str(XRequestID, request.Header.Get(XRequestID)).
		Msg("parsed request")

	proxyRequest.URI = request.URL.EscapedPath()
	proxyRequest.Method = request.Method
	proxyRequest.Body = body

	return proxyRequest
}

func getHTTPMaxUpstreamAttempts() int {
	if httpUpstreamMaxAttempts == 0 {
		httpUpstreamMaxAttempts, _ = strconv.Atoi(os.Getenv("HTTP_UPSTREAM_MAX_ATTEMPTS"))
		if httpUpstreamMaxAttempts == 0 {
			httpUpstreamMaxAttempts = 2
		}
	}
	return httpUpstreamMaxAttempts
}
