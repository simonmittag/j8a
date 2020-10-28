package j8a

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

var httpUpstreamMaxAttempts int

type TLSType string

const (
	TLS12       TLSType = "TLS1.2"
	TLS13       TLSType = "TLS1.3"
	TLS_UNKNOWN TLSType = "unknown"
	TLS_NONE    TLSType = "none"
)

//RFC7231 4.2.1
var httpSafeMethods []string = []string{"GET", "HEAD", "OPTIONS", "TRACE"}

//RFC7231 4.2.2
var httpIdempotentMethods []string = []string{"PUT", "DELETE"}
var httpRepeatableMethods = append(httpSafeMethods, httpIdempotentMethods...)

//RFC7231 4.3
var httpLegalMethods []string = append(httpRepeatableMethods, []string{"POST", "CONNECT"}...)

// Atmpt wraps connection attempts to specific upstreams that are already mapped by label
type Atmpt struct {
	URL            *URL
	Label          string
	Count          int
	StatusCode     int
	isGzip         bool
	resp           *http.Response
	respBody       *[]byte
	CompleteHeader chan struct{}
	CompleteBody   chan struct{}
	Aborted        <-chan struct{}
	AbortedFlag    bool
	CancelFunc     func()
	startDate      time.Time
}

func (atmpt Atmpt) print() string {
	return fmt.Sprintf("%d/%d", atmpt.Count, Runner.Connection.Upstream.MaxAttempts)
}

// Resp wraps downstream http response writer and data
type Resp struct {
	Writer        http.ResponseWriter
	StatusCode    int
	Message       string
	SendGzip      bool
	Body          *[]byte
	ContentLength int64
}

//Up wraps upstream
type Up struct {
	Atmpt  *Atmpt
	Atmpts []Atmpt
	Count  int
}

//Down wraps downstream exchange
type Down struct {
	Req         *http.Request
	Resp        Resp
	Method      string
	Path        string
	URI         string
	UserAgent   string
	Body        []byte
	Aborted     <-chan struct{}
	AbortedFlag bool
	ReqTooLarge bool
	startDate   time.Time
	HttpVer     string
	TlsVer      string
}

// Proxy wraps data for a single downstream request/response with multiple upstream HTTP request/response cycles.
type Proxy struct {
	XRequestID    string
	XRequestDebug bool
	Up            Up
	Dwn           Down
	Route         *Route
}

// TODO downstream aborted needs to cover both timeouts and user aborted requests.
func (proxy *Proxy) hasDownstreamAborted() bool {

	//non blocking read if request context was aborted
	select {
	case <-proxy.Dwn.Aborted:
		proxy.Dwn.AbortedFlag = true
	default:
	}
	if proxy.Dwn.AbortedFlag == true {
		proxy.respondWith(504, "gateway timeout triggered after downstream roundtripTimeoutSeconds")
	}
	return proxy.Dwn.AbortedFlag
}

func (proxy *Proxy) resolveUpstreamURI() string {
	uri := proxy.Up.Atmpt.URL.String() + proxy.Dwn.URI
	if len(proxy.Route.Transform) > 0 {
		t := proxy.Route.Transform
		if t == "/" {
			t = ""
		}
		uri = proxy.Up.Atmpt.URL.String() + string(proxy.Route.PathRegex.ReplaceAll([]byte(proxy.Dwn.URI), []byte(t)))
	}
	return uri
}

func (proxy *Proxy) abortAllUpstreamAttempts() {
	for _, atmpt := range proxy.Up.Atmpts {
		atmpt.AbortedFlag = true
		if atmpt.CancelFunc != nil {
			atmpt.CancelFunc()
			scaffoldUpAttemptLog(proxy).
				Msgf("aborted upstream attempt after prior downstream abort.")
		}
	}
}

func (proxy *Proxy) hasUpstreamAttemptAborted() bool {
	//non blocking read if request context was aborted
	select {
	case <-proxy.Up.Atmpt.Aborted:
		proxy.Up.Atmpt.AbortedFlag = true
	default:
	}
	return proxy.Up.Atmpt.AbortedFlag
}

// tells us if we can safely retry with another upstream attempt
func (proxy *Proxy) shouldRetryUpstreamAttempt() bool {

	// part one is checking for repeatable methods. we don't retry i.e. POST
	retry := false
Retry:
	for _, method := range httpRepeatableMethods {
		if proxy.Dwn.Method == method {
			if proxy.Up.Atmpt.Count < Runner.Connection.Upstream.MaxAttempts {
				retry = true
				break Retry
			}
			retry = false
		}
	}

	// once downstream context has signalled, do not re-attempt upstream
	if proxy.hasDownstreamAborted() {
		retry = false
	}

	if !retry {
		scaffoldUpAttemptLog(proxy).
			Msg("upstream retries stopped after upstream attempt")
	}

	return retry
}

func (proxy *Proxy) hasMadeUpstreamAttempt() bool {
	return proxy.Up.Atmpt != nil && proxy.Up.Atmpt.resp != nil
}

// ParseIncoming is a factory method for a new ProxyRequest, embeds the incoming request.
func (proxy *Proxy) parseIncoming(request *http.Request) *Proxy {
	proxy.Dwn.startDate = time.Now()
	proxy.XRequestID = createXRequestID(request)

	//set request context and initialise timeout func
	ctx, cancel := context.WithCancel(context.TODO())
	proxy.Dwn.Aborted = ctx.Done()
	time.AfterFunc(Runner.getDownstreamRoundTripTimeoutDuration(), func() {
		cancel()
	})

	proxy.XRequestDebug = parseXRequestDebug(request)
	proxy.Dwn.Path = request.URL.EscapedPath()
	proxy.Dwn.URI = request.URL.RequestURI()
	proxy.Dwn.HttpVer = parseHTTPVer(request)
	proxy.Dwn.TlsVer = parseTlsVersion(request)
	proxy.Dwn.UserAgent = parseUserAgent(request)
	proxy.Dwn.Method = parseMethod(request)

	proxy.parseRequestBody(request)
	proxy.Dwn.Req = request

	proxy.Dwn.AbortedFlag = false
	proxy.Dwn.Resp.SendGzip = strings.Contains(request.Header.Get("Accept-Encoding"), "gzip")

	log.Trace().
		Str("path", proxy.Dwn.Path).
		Str("method", proxy.Dwn.Method).
		Int("bodyBytes", len(proxy.Dwn.Body)).
		Int64("dwnElapsedMicros", time.Since(proxy.Dwn.startDate).Microseconds()).
		Str(XRequestID, proxy.XRequestID).
		Msg("parsed downstream request header and body")
	return proxy
}

func (proxy *Proxy) parseRequestBody(request *http.Request) {
	//content length 0, do not read just go back
	if request.ContentLength == 0 {
		log.Trace().
			Int64("dwnElapsedMicros", time.Since(proxy.Dwn.startDate).Microseconds()).
			Str(XRequestID, proxy.XRequestID).Msg("downstream request header content-length 0, don't attempt reading body")
		return
	}

	//only try to parse the request if supplied content-length is within limits
	if request.ContentLength >= Runner.Connection.Downstream.MaxBodyBytes {
		proxy.Dwn.ReqTooLarge = true
		log.Trace().
			Str(XRequestID, proxy.XRequestID).
			Int64("dwnElapsedMicros", time.Since(proxy.Dwn.startDate).Microseconds()).
			Msgf("downstream request body content-length %d exceeds max allowed bytes %d, refuse reading body", request.ContentLength, Runner.Connection.Downstream.MaxBodyBytes)
		return
	}

	//create buffered reader so we can fetch chunks of request as they come.
	//No need to close request.Body of type io.ReadCloser, see: https://golang.org/pkg/net/http/#Request
	bodyReader := bufio.NewReader(http.MaxBytesReader(proxy.Dwn.Resp.Writer,
		request.Body,
		Runner.Connection.Downstream.MaxBodyBytes))

	var err error
	var buf []byte

	//read body. knows how to deal with transfer encoding: chunked, identity

	buf, err = ioutil.ReadAll(bodyReader)
	n := len(buf)
	if int64(n) > Runner.Connection.Downstream.MaxBodyBytes {
		proxy.Dwn.ReqTooLarge = true
		log.Trace().
			Str(XRequestID, proxy.XRequestID).
			Int64("dwnElapsedMicros", time.Since(proxy.Dwn.startDate).Microseconds()).
			Msgf("downstream request body too large. %d body bytes > server max %d", n, Runner.Connection.Downstream.MaxBodyBytes)
	} else if err != nil && err != io.EOF {
		log.Trace().
			Str(XRequestID, proxy.XRequestID).
			Int64("dwnElapsedMicros", time.Since(proxy.Dwn.startDate).Microseconds()).
			Msgf("downstream request body aborting read, cause: %v", err)
	} else {
		proxy.Dwn.Body = buf
		log.Trace().
			Str(XRequestID, proxy.XRequestID).
			Int64("dwnElapsedMicros", time.Since(proxy.Dwn.startDate).Microseconds()).
			Msgf("downstream request body read (%d/%d) bytes/content-length", n, request.ContentLength)
	}

}

func parseMethod(request *http.Request) string {
	return strings.ToUpper(request.Method)
}

func parseUserAgent(request *http.Request) string {
	ua := request.Header.Get("User-Agent")
	if len(ua) == 0 {
		ua = "unknown"
	}
	return ua
}

func parseHTTPVer(request *http.Request) string {
	return fmt.Sprintf("%d.%d", request.ProtoMajor, request.ProtoMinor)
}

func parseXRequestDebug(request *http.Request) bool {
	h := request.Header.Get("X-REQUEST-DEBUG")
	return len(h) > 0 && strings.ToLower(h) == "true"
}

func parseTlsVersion(request *http.Request) string {
	if request.TLS != nil {
		if request.TLS.Version == tls.VersionTLS12 {
			return string(TLS12)
		}
		if request.TLS.Version == tls.VersionTLS13 {
			return string(TLS13)
		}
		return string(TLS_UNKNOWN)
	} else {
		return string(TLS_NONE)
	}
}

func createXRequestID(request *http.Request) string {
	//matches case insensitive
	xr := request.Header.Get(XRequestID)
	if len(xr) == 0 {
		uuid, _ := uuid.NewRandom()
		xr = fmt.Sprintf("XR-%s-%s", ID, uuid)
	}
	return xr
}

func (proxy *Proxy) setOutgoing(out http.ResponseWriter) *Proxy {
	proxy.Dwn.Resp = Resp{
		Writer: out,
	}
	return proxy
}

func (proxy Proxy) bodyReader() io.Reader {
	if len(proxy.Dwn.Body) > 0 {
		return bytes.NewReader(proxy.Dwn.Body)
	}
	return nil
}

func (proxy *Proxy) firstAttempt(URL *URL, label string) *Proxy {
	first := Atmpt{
		Label:          label,
		URL:            URL,
		Count:          1,
		resp:           nil,
		respBody:       nil,
		CompleteHeader: make(chan struct{}),
		CompleteBody:   make(chan struct{}),
		Aborted:        make(chan struct{}),
		CancelFunc:     nil,
		startDate:      time.Now(),
	}
	proxy.Up.Atmpts = []Atmpt{first}
	proxy.Up.Atmpt = &proxy.Up.Atmpts[0]
	proxy.Up.Count = 1

	scaffoldUpAttemptLog(proxy).
		Msg("first upstream attempt initialized")

	return proxy
}

func (proxy *Proxy) nextAttempt() *Proxy {
	next := Atmpt{
		URL:            proxy.Up.Atmpt.URL,
		Label:          proxy.Up.Atmpt.Label,
		Count:          proxy.Up.Atmpt.Count + 1,
		StatusCode:     0,
		isGzip:         false,
		resp:           nil,
		respBody:       nil,
		CompleteHeader: make(chan struct{}),
		CompleteBody:   make(chan struct{}),
		Aborted:        make(chan struct{}),
		AbortedFlag:    false,
		CancelFunc:     nil,
		startDate:      time.Now(),
	}
	proxy.Up.Atmpts = append(proxy.Up.Atmpts, next)
	proxy.Up.Count = next.Count
	proxy.Up.Atmpt = &proxy.Up.Atmpts[len(proxy.Up.Atmpts)-1]

	scaffoldUpAttemptLog(proxy).
		Msg("next upstream attempt initialized")
	return proxy
}

func (proxy *Proxy) writeContentEncodingHeader() {
	if proxy.Dwn.Resp.ContentLength > 0 {
		proxy.Dwn.Resp.Writer.Header().Set(contentEncoding, proxy.contentEncoding())
	}
}

func (proxy *Proxy) copyUpstreamResponseHeaders() {
	for key, values := range proxy.Up.Atmpt.resp.Header {
		if shouldProxyHeader(key) {
			for _, value := range values {
				proxy.Dwn.Resp.Writer.Header().Add(key, value)
			}
		}
	}
}

func (proxy *Proxy) encodeUpstreamResponseBody() {
	if *proxy.Up.Atmpt.respBody != nil && len(*proxy.Up.Atmpt.respBody) > 0 {
		if proxy.shouldGzipEncodeResponseBody() {
			proxy.Dwn.Resp.Body = Gzip(*proxy.Up.Atmpt.respBody)
			scaffoldUpAttemptLog(proxy).
				Msg("copying upstream body with gzip re-encoding")
		} else {
			if proxy.shouldGzipDecodeResponseBody() {
				proxy.Dwn.Resp.Body = Gunzip(*proxy.Up.Atmpt.respBody)
				scaffoldUpAttemptLog(proxy).
					Msg("copying upstream body with gzip re-decoding")
			} else {
				proxy.Dwn.Resp.Body = proxy.Up.Atmpt.respBody
				scaffoldUpAttemptLog(proxy).
					Msgf("copying upstream body as is")
			}
		}
	} else {
		//just in case golang tries to use this value downstream.
		nobody := make([]byte, 0)
		proxy.Dwn.Resp.Body = &nobody
		scaffoldUpAttemptLog(proxy).
			Msgf("skipping empty upstream body")
	}
}

func (proxy *Proxy) contentEncoding() string {
	ce := "identity"
	if proxy.Dwn.Resp.SendGzip {
		ce = "gzip"
	} else if proxy.hasMadeUpstreamAttempt() && !proxy.shouldGzipDecodeResponseBody() {
		ceA := proxy.Up.Atmpt.resp.Header[contentEncoding]
		if len(ceA) > 0 {
			ce = strings.Join(ceA, " ")
		}
	}

	return ce
}

func (proxy *Proxy) setRoute(route *Route) {
	proxy.Route = route
}

//RFC7230, section 3.3.2
func (proxy *Proxy) setContentLengthHeader() {
	proxy.Dwn.Resp.ContentLength = int64(len(*proxy.Dwn.Resp.Body))

	if te := proxy.Dwn.Resp.Writer.Header().Get(transferEncoding); len(te) != 0 ||
		//we set 0 for status code 204 because of RFC7230, 4.3.7, see: https://tools.ietf.org/html/rfc7231#page-31
		//however golang removes this in it's own implementation.
		//Spec ambiguous, see Errata: https://www.rfc-editor.org/errata/eid5806
		//overall there is little harm done by absent header. J8a tests distinguish between
		//Content-Length==0 and no header present to detect when/if future golang version changes behavior.
		proxy.Dwn.Resp.StatusCode == 204 ||
		(proxy.Dwn.Resp.StatusCode >= 100 && proxy.Dwn.Resp.StatusCode < 200) ||
		proxy.Dwn.Method == "CONNECT" {
		proxy.Dwn.Resp.Writer.Header().Set(contentLength, "0")
	} else if proxy.Dwn.Method == "HEAD" {
		//special case for upstream HEAD response with intact content-length we do copy
		//see RFC7231 4.3.2: https://tools.ietf.org/html/rfc7231#page-25
		cl := proxy.Up.Atmpt.resp.Header.Get(contentLength)
		_, err := strconv.ParseInt(cl, 10, 64)
		if len(cl) > 0 && err == nil {
			proxy.Dwn.Resp.Writer.Header().Set(contentLength, cl)
		} else {
			proxy.Dwn.Resp.Writer.Header().Set(contentLength, "0")
		}
	} else {
		proxy.Dwn.Resp.Writer.Header().Set(contentLength, fmt.Sprintf("%d", proxy.Dwn.Resp.ContentLength))
	}
}

func (proxy *Proxy) pipeDownstreamResponse() {
	proxy.Dwn.Resp.Writer.Write(*proxy.Dwn.Resp.Body)
}

//status Code must be last, no headers may be written after this one.
func (proxy *Proxy) copyUpstreamStatusCodeHeader() {
	proxy.respondWith(proxy.Up.Atmpt.StatusCode, "none")
}

func (proxy *Proxy) sendDownstreamStatusCodeHeader() {
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
