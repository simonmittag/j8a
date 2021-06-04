package j8a

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/rs/zerolog"
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
	TLS12         TLSType = "1.2"
	TLS13         TLSType = "1.3"
	TLS_UNKNOWN   TLSType = "unknown"
	TLS_NONE      TLSType = "none"
	Authorization         = "Authorization"
	Sep                   = " "
)

//RFC7231 4.2.1
var httpSafeMethods []string = []string{"GET", "HEAD", "OPTIONS", "TRACE"}

//RFC7231 4.2.2
var httpIdempotentMethods []string = []string{"PUT", "DELETE"}
var httpRepeatableMethods = append(httpSafeMethods, httpIdempotentMethods...)

//RFC7231 4.3
var httpLegalMethods []string = append(httpRepeatableMethods, []string{"POST", "CONNECT"}...)

type proxyfunc func(*Proxy)

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
	Port        int
	Listener    string
}

// Proxy wraps data for a single downstream request/response with multiple upstream HTTP request/response cycles.
type Proxy struct {
	XRequestID   string
	XRequestInfo bool
	Up           Up
	Dwn          Down
	Route        *Route
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
		uri = proxy.Up.Atmpt.URL.String() + strings.Replace(proxy.Dwn.URI, proxy.Route.Path, t, 1)
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

const headerBodyParsed = "downstream request header and body successfully parsed"
const bodyBytes = "bodyBytes"
const method = "method"
const path = "path"

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

	proxy.XRequestInfo = parseXRequestInfo(request)
	proxy.Dwn.Path = request.URL.EscapedPath()
	proxy.Dwn.URI = request.URL.RequestURI()
	proxy.Dwn.HttpVer = parseHTTPVer(request)
	proxy.Dwn.TlsVer = parseTlsVersion(request)
	proxy.Dwn.UserAgent = parseUserAgent(request)
	proxy.Dwn.Method = parseMethod(request)
	proxy.Dwn.Listener = parseListener(request)
	proxy.Dwn.Port = parsePort(request)

	proxy.parseRequestBody(request)
	proxy.Dwn.Req = request

	proxy.Dwn.AbortedFlag = false
	proxy.Dwn.Resp.SendGzip = strings.Contains(request.Header.Get("Accept-Encoding"), "gzip")

	var ev *zerolog.Event
	if proxy.XRequestInfo {
		ev = log.Info()
	} else {
		ev = log.Trace()
	}

	ev.Str(path, proxy.Dwn.Path).
		Str(method, proxy.Dwn.Method).
		Int(bodyBytes, len(proxy.Dwn.Body)).
		Int64(dwnElpsdMicros, time.Since(proxy.Dwn.startDate).Microseconds()).
		Str(XRequestID, proxy.XRequestID).
		Msg(headerBodyParsed)
	return proxy
}

func parsePort(request *http.Request) int {
	if request.TLS == nil {
		return Runner.Connection.Downstream.Http.Port
	} else {
		return Runner.Connection.Downstream.Tls.Port
	}
}

func parseListener(request *http.Request) string {
	if request.TLS == nil {
		return HTTP
	} else {
		return TLS
	}
}

const dwnHeaderContentLengthZero = "downstream request has content-length 0"
const dwnBodyContentLengthExceedsMaxBytes = "downstream request body content-length %d exceeds max allowed bytes %d, refuse reading body"
const dwnBodyTooLarge = "downstream request body too large. %d body bytes > server max %d"
const dwnBodyReadAbort = "downstream request body aborting read, cause: %v"
const dwnBodyRead = "downstream request body read (%d/%d) bytes/content-length"

func (proxy *Proxy) parseRequestBody(request *http.Request) {
	//content length 0, do not read just go back
	if request.ContentLength == 0 {
		var ev *zerolog.Event
		if proxy.XRequestInfo {
			ev = log.Info()
		} else {
			ev = log.Trace()
		}
		ev.Int64(dwnElpsdMicros, time.Since(proxy.Dwn.startDate).Microseconds()).
			Str(XRequestID, proxy.XRequestID).Msg(dwnHeaderContentLengthZero)
		return
	}

	//only try to parse the request if supplied content-length is within limits
	if request.ContentLength >= Runner.Connection.Downstream.MaxBodyBytes {
		proxy.Dwn.ReqTooLarge = true
		log.Trace().
			Str(XRequestID, proxy.XRequestID).
			Int64(dwnElpsdMicros, time.Since(proxy.Dwn.startDate).Microseconds()).
			Msgf(dwnBodyContentLengthExceedsMaxBytes, request.ContentLength, Runner.Connection.Downstream.MaxBodyBytes)
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
			Int64(dwnElpsdMicros, time.Since(proxy.Dwn.startDate).Microseconds()).
			Msgf(dwnBodyTooLarge, n, Runner.Connection.Downstream.MaxBodyBytes)
	} else if err != nil && err != io.EOF {
		log.Trace().
			Str(XRequestID, proxy.XRequestID).
			Int64(dwnElpsdMicros, time.Since(proxy.Dwn.startDate).Microseconds()).
			Msgf(dwnBodyReadAbort, err)
	} else {
		proxy.Dwn.Body = buf
		log.Trace().
			Str(XRequestID, proxy.XRequestID).
			Int64(dwnElpsdMicros, time.Since(proxy.Dwn.startDate).Microseconds()).
			Msgf(dwnBodyRead, n, request.ContentLength)
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

const xRequestInfo = "X-REQUEST-INFO"
const xRequestDebug = "X-REQUEST-DEBUG"
const trueStr = "true"

func parseXRequestInfo(request *http.Request) bool {
	h := request.Header.Get(xRequestInfo)
	h2 := request.Header.Get(xRequestDebug)
	return (len(h) > 0 && strings.ToLower(h) == trueStr) ||
		(len(h2) > 0 && strings.ToLower(h2) == trueStr)
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

const upstreamAttemptInitialized = "upstream attempt initialized"

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
		Str(upResource, URL.String()).
		Msg(upstreamAttemptInitialized)

	return proxy
}

const upAtmptCnt = "upAtmptCnt"

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
		Int(upAtmptCnt, proxy.Up.Count).
		Str(upResource, next.URL.String()).
		Msg(upstreamAttemptInitialized)
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

const upstreamEncodeGzip = "upstream response body re-encoded with gzip before passing downstream"
const upstreamDecodeGzip = "upstream response body decoded from gzip before passing downstream"
const upstreamCopyNoRecode = "upstream response body copied without re-coding before passing downstream"
const upstreamResponseNoBody = "upstream response has no body, nothing to copy before passing downstream"

func (proxy *Proxy) encodeUpstreamResponseBody() {
	atmpt := *proxy.Up.Atmpt
	if atmpt.respBody != nil && len(*atmpt.respBody) > 0 {
		if proxy.shouldGzipEncodeResponseBody() {
			proxy.Dwn.Resp.Body = Gzip(*proxy.Up.Atmpt.respBody)

			scaffoldUpAttemptLog(proxy).
				Msg(upstreamEncodeGzip)
		} else {
			if proxy.shouldGzipDecodeResponseBody() {
				proxy.Dwn.Resp.Body = Gunzip(*proxy.Up.Atmpt.respBody)
				scaffoldUpAttemptLog(proxy).
					Msg(upstreamDecodeGzip)
			} else {
				proxy.Dwn.Resp.Body = proxy.Up.Atmpt.respBody
				scaffoldUpAttemptLog(proxy).
					Msgf(upstreamCopyNoRecode)
			}
		}
	} else {
		//just in case golang tries to use this value downstream.
		nobody := make([]byte, 0)
		proxy.Dwn.Resp.Body = &nobody
		scaffoldUpAttemptLog(proxy).
			Msg(upstreamResponseNoBody)
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

//get bearer token from request. feed into lib. check signature. check expiry. return true || false.
func (proxy *Proxy) validateJwt() bool {
	var token string = ""
	var err error
	ok := false

	ev := log.Trace().
		Str("dwnReqPath", proxy.Dwn.Path).
		Str(XRequestID, proxy.XRequestID)

	auth := proxy.Dwn.Req.Header.Get(Authorization)
	bearer := strings.Split(auth, Sep)

	if len(bearer) > 1 {
		token = bearer[1]
		routeSec := Runner.Jwt[proxy.Route.Jwt]
		alg := *new(jwa.SignatureAlgorithm)
		alg.Accept(routeSec.Alg)

		var parsed jwt.Token

		switch alg {
		case jwa.RS256, jwa.RS384, jwa.RS512, jwa.PS256, jwa.PS384, jwa.PS512:
			parsed, err = proxy.verifyJwtSignature(token, routeSec.RSAPublic, alg, ev)
		case jwa.ES256, jwa.ES384, jwa.ES512:
			parsed, err = proxy.verifyJwtSignature(token, routeSec.ECDSAPublic, alg, ev)
		case jwa.HS256, jwa.HS384, jwa.HS512:
			parsed, err = proxy.verifyJwtSignature(token, routeSec.Secret, alg, ev)
		case jwa.NoSignature:
			parsed, err = jwt.Parse([]byte(token))
		default:
			parsed, err = jwt.Parse([]byte(token))
		}

		//date claims are verified separately to signature including skew
		skew, _ := strconv.Atoi(routeSec.AcceptableSkewSeconds)
		if parsed != nil && err == nil {
			err = verifyDateClaims(token, skew, ev)
		}

		if parsed != nil && err == nil {
			err = proxy.verifyMandatoryJwtClaims(parsed, ev)
		}

		if parsed != nil {
			logDateClaims(parsed, ev)
		}

		ok = parsed != nil && err == nil
	} else {
		err = errors.New("jwt bearer token not present")
	}

	if ok {
		ev.Int64("dwnElapsedMicros", time.Since(proxy.Dwn.startDate).Microseconds()).
			Msg("jwt token validated")
	} else {
		ev.Int64("dwnElapsedMicros", time.Since(proxy.Dwn.startDate).Microseconds()).
			Msgf("jwt token rejected, cause: %v", err)
	}
	return ok
}

func (proxy *Proxy) verifyMandatoryJwtClaims(token jwt.Token, ev *zerolog.Event) error {
	var err error
	jwtc := Runner.Jwt[proxy.Route.Jwt]

	if jwtc.hasMandatoryClaims() {
		err = errors.New("failed to match any claims required by route")
		ev.Bool("jwtClaimsMatchRequiredAny", false)
		ev.Bool("jwtClaimsHasRequiredAny", true)
	} else {
		ev.Bool("jwtClaimsHasRequiredAny", false)
	}

	for i, claim := range jwtc.Claims {
		if len(claim) > 0 {
			lk := "jwtClaimsMatchRequired[" + claim + "]"
			ev.Bool(lk, false)

			json, _ := token.AsMap(context.Background())
			iter := jwtc.claimsVal[i].Run(json)
			value, ok := iter.Next()
			if value != nil {
				if _, nok := value.(error); nok {
					err = value.(error)
				} else if ok {
					ev.Bool("jwtClaimsMatchRequiredAny", true)
					ev.Bool(lk, ok)
					return nil
				} else {
					err = errors.New(fmt.Sprintf("claim not matched %s", claim))
				}
			}
		}
	}
	return err
}

func (proxy *Proxy) verifyJwtSignature(token string, keySet KeySet, alg jwa.SignatureAlgorithm, ev *zerolog.Event) (jwt.Token, error) {
	var msg *jws.Message
	var err error
	var parsed jwt.Token

	msg, err = jws.Parse([]byte(token))
	if len(msg.Signatures()) > 0 {
		//first we try to validate by a key with the kid parameter to match.
		kid := extractKid(token)
		var key interface{}
		if len(kid) > 0 {
			ev.Str("jwtKid", kid)

			key = keySet.Find(kid)
			if key != nil {
				parsed, err = jwt.Parse([]byte(token),
					jwt.WithVerify(alg, key))
			} else {
				proxy.triggerKeyRotationCheck(kid)
			}
		}

		//TODO: try this with x5t SHA1 thumbprint on previously loaded keys to augment kid. If you're reading this
		//TODO: comment feel free to get in touch with a github issue.

		//if it didn't validate above, we try other keys, provided there are any
		if len(kid) == 0 ||
			key == nil ||
			(err != nil && len(keySet) > 1) {

			for _, kp := range keySet {
				parsed, err = jwt.Parse([]byte(token),
					jwt.WithVerify(alg, kp.Key))
				if err == nil {
					break
				}
			}
		}
	} else {
		err = errors.New("no signature found on jwt token")
	}
	return parsed, err
}

func (proxy *Proxy) triggerKeyRotationCheck(kid string) {
	route := proxy.Route
	routeSec := Runner.Jwt[route.Jwt]
	if len(routeSec.JwksUrl) > 0 {
		//MUST run async since it will block on loading remote JWKS key
		go routeSec.LoadJwks()
		log.Debug().
			Str("route", route.Path).
			Str("jwt", route.Jwt).
			Str(XRequestID, proxy.XRequestID).
			Msgf("unmatched kid [%v] on incoming req triggered background key rotation search for route [%v] jwt [%v]", kid, route.Path, route.Jwt)
	}
}

func logDateClaims(parsed jwt.Token, ev *zerolog.Event) {
	if parsed.IssuedAt().Unix() > 1 {
		ev.Bool("jwtClaimsIat", true)
		ev.Str("jwtIatUtcIso", parsed.IssuedAt().Format(time.RFC3339))
		ev.Str("jwtIatLclIso", parsed.IssuedAt().Local().Format(time.RFC3339))
		ev.Int64("jwtIatUnix", parsed.IssuedAt().Unix())
	} else {
		ev.Bool("jwtClaimsIat", false)
	}

	if parsed.NotBefore().Unix() > 1 {
		ev.Bool("jwtClaimsNbf", true)
		ev.Str("jwtNbfUtcIso", parsed.NotBefore().Format(time.RFC3339))
		ev.Str("jwtNbfLclIso", parsed.NotBefore().Local().Format(time.RFC3339))
		ev.Int64("jwtNbfUnix", parsed.NotBefore().Unix())
	} else {
		ev.Bool("jwtClaimsNbf", false)
	}

	if parsed.Expiration().Unix() > 1 {
		ev.Bool("jwtClaimsExp", true)
		ev.Str("jwtExpUtcIso", parsed.Expiration().Format(time.RFC3339))
		ev.Str("jwtExpLclIso", parsed.Expiration().Local().Format(time.RFC3339))
		ev.Int64("jwtExpUnix", parsed.Expiration().Unix())
	} else {
		ev.Bool("jwtClaimsExp", false)
	}
}

func verifyDateClaims(token string, skew int, ev *zerolog.Event) error {
	//arghh i need a deep copy of this token so i can modify it, but it's an interface wrapping a package private jwt.stdToken
	//so i need to parse it again.
	skewed, err := jwt.Parse([]byte(token))

	if skewed.IssuedAt().Unix() > int64(skew*1000) {
		ev.Bool("jwtClaimsIatValidated", true)
		skewed.Set("iat", skewed.IssuedAt().Add(-time.Second*time.Duration(skew)))
	}
	if skewed.NotBefore().Unix() > int64(skew*1000) {
		ev.Bool("jwtClaimsNbfValidated", true)
		skewed.Set("nbf", skewed.NotBefore().Add(-time.Second*time.Duration(skew)))
	}
	if skewed.Expiration().Unix() > 1 {
		ev.Bool("jwtClaimsExpValidated", true)
		skewed.Set("exp", skewed.Expiration().Add(time.Second*time.Duration(skew)))
	}

	if skewed != nil {
		err = jwt.Validate(skewed)
	}

	if err != nil && strings.Contains(err.Error(), "iat") {
		ev.Bool("jwtClaimsIatValidated", false)
	}
	if err != nil && strings.Contains(err.Error(), "nbf") {
		ev.Bool("jwtClaimsNbfValidated", false)
	}
	if err != nil && strings.Contains(err.Error(), "exp") {
		ev.Bool("jwtClaimsExpValidated", false)
	}

	return err
}

func extractKid(token string) string {
	header := strings.Split(token, ".")[0]
	var decoded []byte
	decoded, err := base64.RawURLEncoding.DecodeString(header)
	if err != nil {
		return ""
	}

	var jsonToken map[string]interface{} = make(map[string]interface{})
	err = json.Unmarshal(decoded, &jsonToken)
	if err != nil {
		return ""
	}

	kid := jsonToken["kid"]

	switch kid.(type) {
	case string:
		return kid.(string)
	default:
		return ""
	}
}
