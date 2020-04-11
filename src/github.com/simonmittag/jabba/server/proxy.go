package server

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

var httpUpstreamMaxAttempts int

//RFC7231 4.2.1
var httpSafeMethods []string = []string{"GET", "HEAD", "OPTIONS", "TRACE"}

//RFC7231 4.2.2
var httpIdempotentMethods []string = []string{"PUT", "DELETE"}
var httpRepeatableMethods = append(httpSafeMethods, httpIdempotentMethods...)

//RFC7231 4.3
var httpLegalMethods []string = append(httpRepeatableMethods, []string{"POST", "CONNECT"}...)

// Attempt wraps connection attempts to specific upstreams that are already mapped by label
type Attempt struct {
	URL        *URL
	Label      string
	Count      int
	StatusCode int
}

// Response writer and data
type Response struct {
	Writer     http.ResponseWriter
	StatusCode int
	Message    string
}

// Proxy wraps data for a single downstream request/response with multiple upstream HTTP request/response cycles.
type Proxy struct {
	//downstream request params
	Request    *http.Request
	XRequestID string
	Method     string
	Path       string
	URI        string
	UserAgent  string
	Gzip       bool
	Body       []byte

	//upstream attempt
	Attempt Attempt

	//downstream response
	Response Response
}

func (proxy *Proxy) resolveUpstreamURI() string {
	return proxy.Attempt.URL.String() + proxy.URI
}

// ShouldRepeat tells us if we can safely repeat the upstream request
func (proxy *Proxy) shouldAttemptRetry() bool {
	for _, method := range httpRepeatableMethods {
		if proxy.Method == method {
			if proxy.Attempt.Count < Runner.Connection.Upstream.MaxAttempts {
				return true
			}
			return false
		}
	}
	return false
}

// ParseIncoming is a factory method for a new ProxyRequest, embeds the incoming request.
func (proxy *Proxy) parseIncoming(request *http.Request) *Proxy {
	//TODO: we are not processing downstream body reading errors, i.e. illegal content length
	body, _ := ioutil.ReadAll(request.Body)
	proxy.Path = request.URL.EscapedPath()
	proxy.URI = request.URL.RequestURI()

	proxy.UserAgent = request.Header.Get("User-Agent")
	if len(proxy.UserAgent) == 0 {
		proxy.UserAgent = "unknown"
	}

	proxy.Method = strings.ToUpper(request.Method)
	proxy.Body = body
	proxy.Request = request
	proxy.Response = Response{}

	proxy.Gzip = strings.Contains(request.Header.Get("Accept-Encoding"), "gzip")

	log.Trace().
		Str("path", proxy.Path).
		Str("method", proxy.Method).
		Int("bodyBytes", len(proxy.Body)).
		Str(XRequestID, proxy.XRequestID).
		Msg("parsed request")
	return proxy
}

func (proxy *Proxy) setOutgoing(out http.ResponseWriter) *Proxy {
	proxy.Response.Writer = out
	return proxy
}

func (proxy Proxy) bodyReader() io.Reader {
	if len(proxy.Body) > 0 {
		return bytes.NewReader(proxy.Body)
	}
	return nil
}

func (proxy *Proxy) firstAttempt(URL *URL, label string) *Proxy {
	proxy.Attempt = Attempt{
		Label: label,
		URL:   URL,
		Count: 1,
	}
	return proxy
}

func (proxy *Proxy) nextAttempt() *Proxy {
	proxy.Attempt.Count++
	proxy.Attempt.StatusCode = 0
	return proxy
}

func (proxy *Proxy) initXRequestID() *Proxy {
	uuid, _ := uuid.NewRandom()
	xr := fmt.Sprintf("XR-%s-%s", ID, uuid)
	proxy.XRequestID = xr
	return proxy
}

func (proxy *Proxy) contentEncoding() string {
	if proxy.Gzip {
		return "gzip"
	}
	return "identity"
}

func (proxy *Proxy) respondWith(statusCode int, message string) *Proxy {
	proxy.Response.StatusCode = statusCode
	proxy.Response.Message = message
	return proxy
}

func (proxy *Proxy) hasLegalHTTPMethod() bool {
	for _, legal := range httpLegalMethods {
		if proxy.Method == legal {
			return true
		}
	}
	return false
}
