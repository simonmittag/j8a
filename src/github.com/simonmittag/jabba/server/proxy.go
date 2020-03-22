package server

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

var httpUpstreamMaxAttempts int

//RFC7231 4.2.1
var httpSafeMethods []string = []string{"GET", "HEAD", "OPTIONS", "TRACE"}

//RFC7231 4.2.2
var httpIdempotentMethods []string = []string{"PUT", "DELETE"}
var httpRepeatableMethods = append(httpSafeMethods, httpIdempotentMethods...)

// Attempt wraps connection attempts to specific upstreams that are already mapped by label
type Attempt struct {
	Upstream   *Upstream
	Label      string
	Count      int
	StatusCode int
}

// Downstream request and response writer
type Downstream struct {
	Request    *http.Request
	Response   http.ResponseWriter
	StatusCode int
	Message    string
}

// Proxy wraps data for a single downstream request/response with multiple upstream HTTP request/response cycles.
type Proxy struct {
	XRequestID string
	Method     string
	URI        string
	Body       []byte
	Attempt    Attempt
	Downstream Downstream
}

func (proxy *Proxy) resolveUpstreamURI() string {
	return proxy.Attempt.Upstream.String() + proxy.URI
}

// ShouldRepeat tells us if we can safely repeat the upstream request
func (proxy *Proxy) shouldRepeat() bool {
	for _, method := range httpRepeatableMethods {
		if proxy.Method == method {
			if proxy.Attempt.Count < getHTTPMaxUpstreamAttempts() {
				return true
			}
			return false
		}
	}
	return false
}

// ParseIncoming is a factory method for a new ProxyRequest, embeds the incoming request.
func (proxy *Proxy) parseIncoming(request *http.Request) *Proxy {
	url := request.URL
	method := request.Method
	body, _ := ioutil.ReadAll(request.Body)
	log.Trace().
		Str("url", url.Path).
		Str("method", method).
		Int("bodyBytes", len(body)).
		Str(XRequestID, proxy.XRequestID).
		Msg("parsed request")

	proxy.URI = request.URL.EscapedPath()
	proxy.Method = request.Method
	proxy.Body = body
	proxy.Downstream = Downstream{
		Request: request,
	}
	return proxy
}

func (proxy *Proxy) setOutgoing(response http.ResponseWriter) *Proxy {
	proxy.Downstream.Response = response
	return proxy
}

func (proxy Proxy) bodyReader() io.Reader {
	if len(proxy.Body) > 0 {
		return bytes.NewReader(proxy.Body)
	}
	return nil
}

func (proxy *Proxy) firstAttempt(upstream *Upstream, label string) *Proxy {
	proxy.Attempt = Attempt{
		Label:    label,
		Upstream: upstream,
		Count:    1,
	}
	return proxy
}

func (proxy *Proxy) initXRequestID() *Proxy {
	uuid, _ := uuid.NewRandom()
	xr := fmt.Sprintf("XR-%s-%s", ID, uuid)
	proxy.XRequestID = xr
	return proxy
}

func (proxy *Proxy) respondWith(statusCode int, message string) *Proxy {
	proxy.Downstream.StatusCode = statusCode
	proxy.Downstream.Message = message
	return proxy
}
