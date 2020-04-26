package jabba

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

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

// Atmpt wraps connection attempts to specific upstreams that are already mapped by label
type Atmpt struct {
	URL        *URL
	Label      string
	Count      int
	StatusCode int
	isGzip     bool
	resp       *http.Response
	respBody   *[]byte
}

// Resp wraps downstream http response writer and data
type Resp struct {
	Writer     http.ResponseWriter
	StatusCode int
	Message    string
	SendGzip   bool
}

//Up wraps upstream
type Up struct {
	Atmpt Atmpt
}

//Down wraps downstream exchange
type Down struct {
	Req       *http.Request
	Method    string
	Path      string
	URI       string
	UserAgent string
	Body      []byte

	Resp Resp
}

// Proxy wraps data for a single downstream request/response with multiple upstream HTTP request/response cycles.
type Proxy struct {
	XRequestID string
	Up         Up
	Dwn        Down
}

func (proxy *Proxy) resolveUpstreamURI() string {
	return proxy.Up.Atmpt.URL.String() + proxy.Dwn.URI
}

// ShouldRepeat tells us if we can safely repeat the upstream request
func (proxy *Proxy) shouldAttemptRetry() bool {
	for _, method := range httpRepeatableMethods {
		if proxy.Dwn.Method == method {
			if proxy.Up.Atmpt.Count < Runner.Connection.Upstream.MaxAttempts {
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
	proxy.Dwn.Path = request.URL.EscapedPath()
	proxy.Dwn.URI = request.URL.RequestURI()

	proxy.Dwn.UserAgent = request.Header.Get("User-Agent")
	if len(proxy.Dwn.UserAgent) == 0 {
		proxy.Dwn.UserAgent = "unknown"
	}

	proxy.Dwn.Method = strings.ToUpper(request.Method)
	proxy.Dwn.Body = body
	proxy.Dwn.Req = request
	proxy.Dwn.Resp = Resp{}
	proxy.Dwn.Resp.SendGzip = strings.Contains(request.Header.Get("Accept-Encoding"), "gzip")

	log.Trace().
		Str("path", proxy.Dwn.Path).
		Str("method", proxy.Dwn.Method).
		Int("bodyBytes", len(proxy.Dwn.Body)).
		Str(XRequestID, proxy.XRequestID).
		Msg("parsed request")
	return proxy
}

func (proxy *Proxy) setOutgoing(out http.ResponseWriter) *Proxy {
	proxy.Dwn.Resp.Writer = out
	return proxy
}

func (proxy Proxy) bodyReader() io.Reader {
	if len(proxy.Dwn.Body) > 0 {
		return bytes.NewReader(proxy.Dwn.Body)
	}
	return nil
}

func (proxy *Proxy) firstAttempt(URL *URL, label string) *Proxy {
	proxy.Up.Atmpt = Atmpt{
		Label:    label,
		URL:      URL,
		Count:    1,
		resp:     nil,
		respBody: nil,
	}
	return proxy
}

func (proxy *Proxy) nextAttempt() *Proxy {
	proxy.Up.Atmpt.Count++
	proxy.Up.Atmpt.StatusCode = 0
	proxy.Up.Atmpt.isGzip = false
	proxy.Up.Atmpt.resp = nil
	proxy.Up.Atmpt.respBody = nil
	return proxy
}

func (proxy *Proxy) initXRequestID() *Proxy {
	uuid, _ := uuid.NewRandom()
	xr := fmt.Sprintf("XR-%s-%s", ID, uuid)
	proxy.XRequestID = xr
	return proxy
}

func (proxy *Proxy) writeContentEncodingHeader() {
	proxy.Dwn.Resp.Writer.Header().Set(contentEncoding, proxy.contentEncoding())
}

func (proxy *Proxy) copyUpstreamResponseHeaders() {
	proxy.Up.Atmpt.StatusCode = proxy.Up.Atmpt.resp.StatusCode
	for key, values := range proxy.Up.Atmpt.resp.Header {
		if shouldRewrite(key) {
			for _, mval := range values {
				proxy.Dwn.Resp.Writer.Header().Set(key, mval)
			}
		}
	}
}

func (proxy *Proxy) copyUpstreamResponseBody() {
	start := time.Now()
	if proxy.shouldGzipEncodeResponseBody() {
		proxy.Dwn.Resp.Writer.Write(Gzip(*proxy.Up.Atmpt.respBody))
		elapsed := time.Since(start)
		log.Trace().Msgf("copying upstream body with gzip re-encoding took %s", elapsed)
	} else {
		if proxy.shouldGzipDecodeResponseBody() {
			proxy.Dwn.Resp.Writer.Write(Gunzip([]byte(*proxy.Up.Atmpt.respBody)))
			elapsed := time.Since(start)
			log.Trace().Msgf("copying upstream body with gzip re-decoding took %s", elapsed)
		} else {
			proxy.Dwn.Resp.Writer.Write([]byte(*proxy.Up.Atmpt.respBody))
			elapsed := time.Since(start)
			log.Trace().Msgf("copying upstream body without coding took %s", elapsed)
		}
	}
}

func (proxy *Proxy) contentEncoding() string {
	if proxy.Dwn.Resp.SendGzip {
		return "gzip"
	} else {
		if proxy.shouldGzipDecodeResponseBody() {
			return "identity"
		} else {
			ce := proxy.Up.Atmpt.resp.Header[contentEncoding]
			if len(ce) > 0 {
				return strings.Join(ce, " ")
			} else {
				return "identity"
			}
		}
	}
}

func (proxy *Proxy) processHeaders() {
	proxy.writeStandardResponseHeaders()
	proxy.copyUpstreamResponseHeaders()
	proxy.resetContentLengthHeader()
	proxy.writeContentEncodingHeader()
	proxy.writeStatusCodeHeader()
}

func (proxy *Proxy) resetContentLengthHeader() {
	if proxy.Dwn.Method == "HEAD" || len(*proxy.Up.Atmpt.respBody) == 0 {
		proxy.Dwn.Resp.Writer.Header().Set(contentLength, "0")
	}
}

//status code must be last, no headers may be written after this one.
func (proxy *Proxy) writeStatusCodeHeader() {
	proxy.respondWith(proxy.Up.Atmpt.StatusCode, "none")
	proxy.Dwn.Resp.Writer.WriteHeader(proxy.Dwn.Resp.StatusCode)
}

func (proxy *Proxy) respondWith(statusCode int, message string) *Proxy {
	proxy.Dwn.Resp.StatusCode = statusCode
	proxy.Dwn.Resp.Message = message
	return proxy
}

func (proxy *Proxy) hasLegalHTTPMethod() bool {
	for _, legal := range httpLegalMethods {
		if proxy.Dwn.Method == legal {
			return true
		}
	}
	return false
}

func (proxy *Proxy) shouldGzipEncodeResponseBody() bool {
	return proxy.Dwn.Resp.SendGzip && !proxy.Up.Atmpt.isGzip
}

func (proxy *Proxy) shouldGzipDecodeResponseBody() bool {
	return !proxy.Dwn.Resp.SendGzip && proxy.Up.Atmpt.isGzip
}
