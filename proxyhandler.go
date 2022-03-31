package j8a

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

//XRequestID is a per HTTP request unique identifier
const XRequestID = "X-Request-Id"
const contentEncoding = "Content-Encoding"
const transferEncoding = "Transfer-Encoding"
const acceptEncoding = "Accept-Encoding"
const contentLength = "Content-Length"
const date = "Date"
const server = "Server"

//httpClient is the global user agent for upstream requests
var httpClient HTTPClient

//httpHeadersNoRewrite contains a list of headers that are not copied in either direction. they must be set by the
//server or are ignored.
var httpHeadersNoRewrite []string = []string{connectionS, date, contentLength, acceptEncoding, transferEncoding, server, varyS}

//extract IPs for stdout. thread safe.
var ipr iprex = iprex{}

func httpHandler(response http.ResponseWriter, request *http.Request) {
	proxyHandler(response, request, handleHTTP)
}

const badOrMalFormedRequest = "bad or malformed request"
const jwtBearerTokenMissing = "jwt bearer token missing, invalid, expired or unauthorized"
const unableToMapUpstreamResource = "unable to map upstream resource"
const upstreamResourceNotFound = "upstream resource not found"
const httpRequestEntityTooLarge = "http request entity too large, limit is %d bytes"
const downstreamRequestAbortedBeforeFirstUpstream = "downstream request aborted or timed out before first upstream attempt"

var httpRequestInvalidAcceptEncoding string

func formatInvalidAcceptEncoding() string {
	if len(httpRequestInvalidAcceptEncoding) == 0 {
		httpRequestInvalidAcceptEncoding = fmt.Sprintf("http request content encoding header must contain one of %v", SupportedContentEncodings.Print())
	}
	return httpRequestInvalidAcceptEncoding
}

func proxyHandler(response http.ResponseWriter, request *http.Request, exec proxyfunc) {
	matched := false

	//preprocess incoming request in proxy object
	proxy := new(Proxy).
		setOutgoing(response).
		parseIncoming(request)

	//all malformed requests are rejected here and we return a 400
	if !validate(proxy) {
		if proxy.Dwn.ReqTooLarge {
			sendStatusCodeAsJSON(proxy.respondWith(413, fmt.Sprintf(httpRequestEntityTooLarge, Runner.Connection.Downstream.MaxBodyBytes)))
		} else if !proxy.Dwn.AcceptEncoding.hasAtLeastOneValidEncoding() {
			sendStatusCodeAsJSON(proxy.respondWith(406, formatInvalidAcceptEncoding()))
		} else {
			sendStatusCodeAsJSON(proxy.respondWith(400, badOrMalFormedRequest))
		}
		return
	}

	//if we timed out or aborted during downstream request parsing we stop the handler before the first upstream attempt.
	if proxy.hasDownstreamAbortedOrTimedout() {
		infoOrTraceEv(proxy).Str(path, proxy.Dwn.Path).
			Str(method, proxy.Dwn.Method).
			Int64(dwnElpsdMicros, time.Since(proxy.Dwn.startDate).Microseconds()).
			Str(XRequestID, proxy.XRequestID).
			Msg(downstreamRequestAbortedBeforeFirstUpstream)
		sendStatusCodeAsJSON(proxy)
		return
	}

	//once a route is matched, it needs to be mapped to an upstream resource via a policy
	for _, route := range Runner.Routes {
		if matched = route.matchURI(request); matched {
			proxy.setRoute(&route)
			if route.hasJwt() && !proxy.validateJwt() {
				sendStatusCodeAsJSON(proxy.respondWith(401, jwtBearerTokenMissing))
				return
			}
			url, label, mapped := route.mapURL(proxy)
			if mapped {
				//mapped requests are sent to proxyfuncs.
				exec(proxy.firstAttempt(url, label))
			} else {
				//unmapped request means an internal configuration error in server
				sendStatusCodeAsJSON(proxy.respondWith(503, unableToMapUpstreamResource))
				return
			}
			break
		}
	}

	//unmatched paths means we have no route for this and always return a 404
	if !matched {
		sendStatusCodeAsJSON(proxy.respondWith(404, upstreamResourceNotFound))
	}
}

func validate(proxy *Proxy) bool {
	return proxy.hasLegalHTTPMethod() &&
		!proxy.Dwn.ReqTooLarge &&
		proxy.Dwn.AcceptEncoding.hasAtLeastOneValidEncoding()
}

const connectionClosedByRemoteUserAgent = "downstream remote user agent aborted request or closed connection"
const gatewayTimeoutTriggeredByDownstreamEvent = "gateway timeout triggered by downstream timeout"
const gatewayTimeoutTriggeredByUpstreamEvent = "gateway timeout triggered by upstream attempt"
const badGatewayTriggeredUnableToProcessUpstreamResponse = "bad gateway triggered. unable to process upstream response"

func handleHTTP(proxy *Proxy) {
	upstreamResponse, upstreamError := performUpstreamRequest(proxy)
	if upstreamResponse != nil && upstreamResponse.Body != nil {
		defer upstreamResponse.Body.Close()
	}

	if !processUpstreamResponse(proxy, upstreamResponse, upstreamError) {
		if proxy.shouldRetryUpstreamAttempt() {
			handleHTTP(proxy.nextAttempt())
		} else {
			//sends 504 for downstream timeout, 504 for upstream timeout, 499 for downstream remote hangup,
			//502 in all other cases
			if proxy.Dwn.TimeoutFlag == true {
				sendStatusCodeAsJSON(proxy.respondWith(504, gatewayTimeoutTriggeredByDownstreamEvent))
			} else if proxy.Dwn.AbortedFlag == true {
				sendStatusCodeAsJSON(proxy.respondWith(499, connectionClosedByRemoteUserAgent))
			} else if proxy.hasUpstreamAttemptAborted() {
				sendStatusCodeAsJSON(proxy.respondWith(504, gatewayTimeoutTriggeredByUpstreamEvent))
			} else {
				sendStatusCodeAsJSON(proxy.respondWith(502, badGatewayTriggeredUnableToProcessUpstreamResponse))
			}
		}
	}
}

const upstreamURIResolved = "upstream URI resolved"
const keepAlive = "Keep-Alive"

func scaffoldUpstreamRequest(proxy *Proxy) *http.Request {
	//this context is used to time out the upstream request
	ctx, cancel := context.WithCancel(context.TODO())

	//remember the cancelFunc, we may need to call it earlier from the outside
	proxy.Up.Atmpt.CancelFunc = cancel

	//will call the cancel func in it's own goroutine after timeout seconds.
	time.AfterFunc(time.Duration(Runner.Connection.Upstream.ReadTimeoutSeconds)*time.Second, func() {
		cancel()
	})

	upURI := proxy.resolveUpstreamURI()

	upstreamRequest, _ := http.NewRequestWithContext(ctx,
		proxy.Dwn.Method,
		upURI,
		proxy.bodyReader())

	var ev *zerolog.Event
	if proxy.XRequestInfo {
		ev = log.Info()
	} else {
		ev = log.Trace()
	}

	ev.Str(dwnReqPath, proxy.Dwn.Path).
		Int64(dwnElpsdMicros, time.Since(proxy.Dwn.startDate).Microseconds()).
		Str(XRequestID, proxy.XRequestID).
		Str(upReqURI, upURI).
		Msg(upstreamURIResolved)

	proxy.Up.Atmpt.Aborted = upstreamRequest.Context().Done()

	//this contains all accept encodings for content negotiation but is guaranteed to have one valid value.
	upstreamRequest.Header.Add(acceptEncoding, proxy.Dwn.AcceptEncoding.Print())

	//set upstream headers
	for key, values := range proxy.Dwn.Req.Header {
		if shouldProxyHeader(key) {
			for _, value := range values {
				upstreamRequest.Header.Add(key, value)
			}
		}
	}

	//this is redundant for HTTP/1.1, spec ref: https://datatracker.ietf.org/doc/html/rfc2616#section-8.1.3
	//upstreamRequest.Header.Set(connectionS, keepAlive)
	upstreamRequest.Header.Set(XRequestID, proxy.XRequestID)

	return upstreamRequest
}

const upResHeaders = "upResHeaders"
const upstreamResHeaderAborted = "upstream response header processing aborted"
const upstreamResHeadersProcessed = "upstream response headers processed"
const upConReadTimeoutFired = "upstream connection read timeout fired, aborting upstream response header processing."
const safeToIgnoreFailedHeaderChannelClosure = "safe to ignore. recovered internally from closed header success channel after request already handled."

const downstreamRtFired = "downstream roundtrip timeout fired"
const downstreamReqAborted = "downstream request aborted"

func performUpstreamRequest(proxy *Proxy) (*http.Response, error) {
	//get a reference to this before any race conditions may occur
	attemptIndex := proxy.Up.Count - 1
	req := scaffoldUpstreamRequest(proxy)
	var upstreamResponse *http.Response
	var upstreamError error

	go func() {
		//this blocks until upstream headers come in
		upstreamResponse, upstreamError = httpClient.Do(req)
		proxy.Up.Atmpt.resp = upstreamResponse

		defer func() {
			if err := recover(); err != nil {
				//it's safe to ignore this.
			}
		}()

		if proxy.Up.Atmpts[attemptIndex].CompleteHeader != nil && !proxy.Up.Atmpts[attemptIndex].AbortedFlag && !proxy.Dwn.AbortedFlag {
			close(proxy.Up.Atmpts[attemptIndex].CompleteHeader)
		}
	}()

	//race for upstream headers complete, upstream timeout or downstream abort (timeout or cancellation)
	select {

	case <-proxy.Up.Atmpt.Aborted:
		proxy.Up.Atmpt.AbortedFlag = true
		proxy.Up.Atmpt.StatusCode = 0
		scaffoldUpAttemptLog(proxy).
			Int(upReadTimeoutSecs, Runner.Connection.Upstream.ReadTimeoutSeconds).
			Msg(upConReadTimeoutFired)
	case <-proxy.Dwn.Timeout:
		proxy.Dwn.TimeoutFlag = true
		scaffoldUpAttemptLog(proxy).
			Msg(downstreamRtFired)

		proxy.abortAllUpstreamAttempts()
		scaffoldUpAttemptLog(proxy).
			Msg(upstreamResHeaderAborted)
	case <-proxy.Dwn.Aborted:
		proxy.Dwn.AbortedFlag = true
		scaffoldUpAttemptLog(proxy).
			Msg(downstreamReqAborted)

		proxy.abortAllUpstreamAttempts()
		scaffoldUpAttemptLog(proxy).
			Msg(upstreamResHeaderAborted)
	case <-proxy.Up.Atmpt.CompleteHeader:
		scaffoldUpAttemptLog(proxy).
			RawJSON(upResHeaders, jsonifyUpstreamHeaders(proxy)).
			Msg(upstreamResHeadersProcessed)
	}

	return upstreamResponse, upstreamError
}

const upReadTimeoutSecs = "upReadTimeoutSecs"
const safeToIgnoreFailedBodyChannelClosure = "safe to ignore. recovered internally from closed body success channel after request already handled."
const upstreamConReadTimeoutFired = "upstream connection read timeout fired, aborting upstream response body processing"

const upResBodyBytes = "upResBodyBytes"
const upstreamResBodyAbort = "upstream response body processing aborted"
const upstreamResBodyProcessed = "upstream response body processed"
const emptyJSON = "{}"

func jsonifyUpstreamHeaders(proxy *Proxy) []byte {
	if proxy.Up.Atmpt == nil || proxy.Up.Atmpt.resp == nil || proxy.Up.Atmpt.resp.Header == nil {
		return []byte(emptyJSON)
	}
	//catch all
	jsonb, err := json.Marshal(proxy.Up.Atmpt.resp.Header)
	if err != nil {
		jsonb = []byte(emptyJSON)
	}
	return jsonb
}

const upAtmptResBodyTrunc = "upAtmptResBodyTrunc"
const upAtmptCntntEnc = "upAtmptCntntEnc"
const more = "..."
const moreEncoded = " [encoded]"

func parseUpstreamResponse(upstreamResponse *http.Response, proxy *Proxy) ([]byte, error) {
	var upstreamResponseBody []byte
	var bodyError error

	//get a reference to this before any race conditions may occur
	attemptIndex := proxy.Up.Count - 1

	go func() {
		proxy.Up.Atmpt.StatusCode = upstreamResponse.StatusCode
		proxy.Up.Atmpt.ContentEncoding = NewContentEncoding(proxy.Up.Atmpt.resp.Header.Get(contentEncoding))

		upstreamResponseBody, bodyError = ioutil.ReadAll(upstreamResponse.Body)

		defer func() {
			if err := recover(); err != nil {
				scaffoldUpAttemptLog(proxy).
					Msg(safeToIgnoreFailedBodyChannelClosure)
			}
		}()

		//this is ok, see: https://stackoverflow.com/questions/8593645/is-it-ok-to-leave-a-channel-open#:~:text=5%20Answers&text=It's%20OK%20to%20leave%20a,it%20will%20be%20garbage%20collected.&text=Closing%20the%20channel%20is%20a,that%20no%20more%20data%20follows.
		if proxy.Up.Atmpt.CompleteBody != nil && !proxy.Up.Atmpt.AbortedFlag && !proxy.Dwn.AbortedFlag {
			close(proxy.Up.Atmpts[attemptIndex].CompleteBody)
		}
	}()

	select {
	case <-proxy.Up.Atmpt.Aborted:
		proxy.Up.Atmpt.AbortedFlag = true
		scaffoldUpAttemptLog(proxy).
			Int(upReadTimeoutSecs, Runner.Connection.Upstream.ReadTimeoutSeconds).
			Msg(upstreamConReadTimeoutFired)
	case <-proxy.Dwn.Timeout:
		proxy.Dwn.TimeoutFlag = true
		scaffoldUpAttemptLog(proxy).
			Msg(downstreamRtFired)

		proxy.abortAllUpstreamAttempts()
		scaffoldUpAttemptLog(proxy).
			Msg(upstreamResBodyAbort)
	case <-proxy.Dwn.Aborted:
		proxy.Dwn.AbortedFlag = true
		scaffoldUpAttemptLog(proxy).
			Msg(downstreamReqAborted)

		proxy.abortAllUpstreamAttempts()
		scaffoldUpAttemptLog(proxy).
			Msg(upstreamResBodyAbort)
	case <-proxy.Up.Atmpt.CompleteBody:
		ul := scaffoldUpAttemptLog(proxy)

		//truncate body for logging
		var t []byte
		if len(upstreamResponseBody) > 25 {
			t = append(t, upstreamResponseBody[0:25]...)
		} else {
			t = upstreamResponseBody
		}

		//and show what is necessary depending on encoding
		if proxy.Up.Atmpt.ContentEncoding.isEncoded() {
			s := hex.EncodeToString(t)
			if len(s) == 50 {
				s += more
			}
			ul.Str(upAtmptResBodyTrunc, s+moreEncoded)
		} else {
			s := string(t)
			if len(s) == 25 {
				s += more
			}
			ul.Str(upAtmptResBodyTrunc, s)
		}

		if len(proxy.Up.Atmpt.ContentEncoding) > 0 {
			ul.Str(upAtmptCntntEnc, proxy.Up.Atmpt.ContentEncoding.print())
		}

		ul.Int(upResBodyBytes, len(upstreamResponseBody)).
			Msg(upstreamResBodyProcessed)
	}

	return upstreamResponseBody, bodyError
}

func processUpstreamResponse(proxy *Proxy, upstreamResponse *http.Response, upstreamError error) bool {
	//process only if we can work with upstream attempt
	if upstreamResponse != nil && upstreamError == nil && !proxy.hasUpstreamAttemptAborted() {
		//j8a blocks here when waiting for upstream body
		upstreamResponseBody, bodyError := parseUpstreamResponse(upstreamResponse, proxy)
		upstreamError = bodyError
		proxy.Up.Atmpt.respBody = &upstreamResponseBody
		if shouldProxyUpstreamResponse(proxy, bodyError) {
			logSuccessfulUpstreamAttempt(proxy, upstreamResponse)
			if isUpstreamClientError(proxy) {
				proxy.copyUpstreamStatusCodeHeader()
				sendStatusCodeAsJSON(proxy)
			} else {
				proxy.writeStandardResponseHeaders()
				proxy.copyUpstreamResponseHeaders()
				proxy.copyUpstreamStatusCodeHeader()
				proxy.encodeUpstreamResponseBody()
				proxy.setContentLengthHeader()
				proxy.sendDownstreamStatusCodeHeader()
				proxy.pipeDownstreamResponse()
				logHandledDownstreamRoundtrip(proxy)
			}
			return true
		}
	}
	//now log unsuccessful and retry or exit with status Code.
	logUnsuccessfulUpstreamAttempt(proxy, upstreamResponse, upstreamError)
	return false
}

func isUpstreamClientError(proxy *Proxy) bool {
	return proxy.Up.Atmpt.StatusCode > 399 && proxy.Up.Atmpt.StatusCode < 500
}

func shouldProxyUpstreamResponse(proxy *Proxy, bodyError error) bool {
	return !proxy.hasDownstreamAbortedOrTimedout() &&
		!proxy.hasUpstreamAttemptAborted() &&
		bodyError == nil &&
		proxy.Up.Atmpt.resp.StatusCode < 500
}

func shouldProxyHeader(header string) bool {
	for _, illegal := range httpHeadersNoRewrite {
		if strings.EqualFold(header, illegal) {
			return false
		}
	}
	return true
}

func scaffoldUpAttemptLog(proxy *Proxy) *zerolog.Event {
	var ev *zerolog.Event
	if proxy.XRequestInfo {
		ev = log.Info()
	} else {
		ev = log.Trace()
	}

	return ev.
		Str(XRequestID, proxy.XRequestID).
		Int64(upAtmtpElpsdMicros, time.Since(proxy.Up.Atmpt.startDate).Microseconds()).
		Int64(dwnElpsdMicros, time.Since(proxy.Dwn.startDate).Microseconds()).
		Str(upAtmpt, proxy.Up.Atmpt.print())
}

const downstreamResponseServed = "downstream HTTP response served"
const downstreamErrorResponseServed = "downstream HTTP error response served"

const pdS = "%d"

func logHandledDownstreamRoundtrip(proxy *Proxy) {
	elapsed := time.Since(proxy.Dwn.startDate)
	msg := downstreamResponseServed
	ev := log.Info()

	if proxy.hasMadeUpstreamAttempt() {
		ev = ev.Str(upReqURI, proxy.resolveUpstreamURI()).
			Str(upLabel, proxy.Up.Atmpt.Label).
			Int(upAtmptResCode, proxy.Up.Atmpt.StatusCode).
			Int(upAtmptResBodyBytes, len(*proxy.Up.Atmpt.respBody)).
			Int64(upAtmptElpsdMicros, time.Since(proxy.Up.Atmpt.startDate).Microseconds()).
			Bool(upAtmptAbort, proxy.Up.Atmpt.AbortedFlag).
			Str(upAtmpt, proxy.Up.Atmpt.print())
	}

	if proxy.Dwn.Resp.StatusCode > 399 {
		msg = downstreamErrorResponseServed
		ev = log.Warn()
		ev = ev.Str(dwnResErrMsg, proxy.Dwn.Resp.Message)
	}

	ev = ev.Str(dwnReqListnr, proxy.Dwn.Listener).
		Str(dwnReqPort, fmt.Sprintf(pdS, proxy.Dwn.Port)).
		Str(dwnReqPath, proxy.Dwn.Path).
		Str(dwnReqRemoteAddr, ipr.extractAddr(proxy.Dwn.Req.RemoteAddr)).
		Str(dwnReqMethod, proxy.Dwn.Method).
		Str(dwnReqUserAgent, proxy.Dwn.UserAgent).
		Str(dwnReqHttpVer, proxy.Dwn.HttpVer).
		Int(dwnResCode, proxy.Dwn.Resp.StatusCode).
		Int64(dwnResCntntLen, proxy.Dwn.Resp.ContentLength).
		Str(dwnResCntntEnc, string(proxy.Dwn.Resp.ContentEncoding)).
		Int64(dwnResElpsdMicros, elapsed.Microseconds()).
		Str(XRequestID, proxy.XRequestID)

	if Runner.isTLSOn() {
		ev = ev.Str(dwnReqTlsVer, proxy.Dwn.TlsVer)
	}

	ev.Msg(msg)
}

const upstreamAttemptSuccessful = "upstream attempt successful"
const upstreamAttemptUnsuccessful = "upstream attempt unsuccessful, cause: "

func logSuccessfulUpstreamAttempt(proxy *Proxy, upstreamResponse *http.Response) {
	scaffoldUpAttemptLog(proxy).
		Int(upAtmptResCode, upstreamResponse.StatusCode).
		Msg(upstreamAttemptSuccessful)
}

const undeterminedUpstreamError = "undetermined but raw error was: %v"
const upstreamHangup = "upstream TCP socket hung up on us remotely"
const eofS = "EOF"

func logUnsuccessfulUpstreamAttempt(proxy *Proxy, upstreamResponse *http.Response, upstreamError error) {
	ev := scaffoldUpAttemptLog(proxy)
	if upstreamResponse != nil && upstreamResponse.StatusCode > 0 {
		ev = ev.Int(upAtmptResCode, upstreamResponse.StatusCode)
	}
	if upstreamError != nil && strings.Contains(upstreamError.Error(), eofS) {
		ev.Msg(upstreamAttemptUnsuccessful + upstreamHangup)
	} else {
		ev.Msg(upstreamAttemptUnsuccessful + fmt.Sprintf(undeterminedUpstreamError, upstreamError))
	}
}
