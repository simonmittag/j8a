package j8a

import (
	"bytes"
	"context"
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
const contentLength = "Content-Length"
const date = "Date"
const server = "Server"

//httpClient is the global user agent for upstream requests
var httpClient HTTPClient

//httpResponseHeadersNoRewrite contains a list of headers that are not copied from upstream to downstream to avoid bugs.
var httpResponseHeadersNoRewrite []string = []string{date, contentLength, contentEncoding, server}

func proxyHandler(response http.ResponseWriter, request *http.Request) {
	matched := false

	//preprocess incoming request in proxy object
	proxy := new(Proxy).
		parseIncoming(request).
		setOutgoing(response)

	//all malformed requests are rejected here and we return a 400
	if !validate(proxy) {
		sendStatusCodeAsJSON(proxy.respondWith(400, "bad or malformed request"))
		return
	}

	//once a route is matched, it needs to be mapped to an upstream resource via a policy
	for _, route := range Runner.Routes {
		if matched = route.matchURI(request); matched {
			url, label, mapped := route.mapURL(proxy)
			if mapped {
				//mapped requests are sent to httpclient
				handle(proxy.firstAttempt(url, label))
			} else {
				//unmapped request means an internal configuration error in server
				sendStatusCodeAsJSON(proxy.respondWith(503, "unable to map upstream resource"))
				return
			}
			break
		}
	}

	//unmatched paths means we have no route for this and always return a 404
	if !matched {
		sendStatusCodeAsJSON(proxy.respondWith(404, "upstream resource not found"))
	}
}

func validate(proxy *Proxy) bool {
	return proxy.hasLegalHTTPMethod()
}

func handle(proxy *Proxy) {
	upstreamResponse, upstreamError := performUpstreamRequest(proxy)
	if upstreamResponse != nil && upstreamResponse.Body != nil {
		defer upstreamResponse.Body.Close()
	}

	if !processUpstreamResponse(proxy, upstreamResponse, upstreamError) {
		if proxy.shouldRetryUpstreamAttempt() {
			handle(proxy.nextAttempt())
		} else {
			//sends 504 for downstream timeout, 504 for upstream timeout, 502 in all other cases
			if proxy.hasDownstreamAborted() {
				sendStatusCodeAsJSON(proxy.respondWith(504, "gateway timeout triggered by downstream event"))
			} else if proxy.hasUpstreamAttemptAborted() {
				sendStatusCodeAsJSON(proxy.respondWith(504, "gateway timeout triggered by upstream attempt"))
			} else {
				sendStatusCodeAsJSON(proxy.respondWith(502, "bad gateway triggered. unable to process upstream response"))
			}
		}
	}
}

func scaffoldUpstreamRequest(proxy *Proxy) *http.Request {
	//this context is used to time out the upstream request
	ctx, cancel := context.WithCancel(context.TODO())

	//remember the cancelFunc, we may need to call it earlier from the outside
	proxy.Up.Atmpt.CancelFunc = cancel

	//will call the cancel func in it's own goroutine after timeout seconds.
	time.AfterFunc(time.Duration(Runner.Connection.Upstream.ReadTimeoutSeconds)*time.Second, func() {
		cancel()
	})

	upstreamRequest, _ := http.NewRequestWithContext(ctx,
		proxy.Dwn.Method,
		proxy.resolveUpstreamURI(),
		proxy.bodyReader())

	proxy.Up.Atmpt.Aborted = upstreamRequest.Context().Done()

	//TODO: test if upstream request headers are reprocessed correctly
	for key, values := range proxy.Dwn.Req.Header {
		upstreamRequest.Header.Set(key, strings.Join(values, " "))
	}
	upstreamRequest.Header.Set(XRequestID, proxy.XRequestID)

	return upstreamRequest
}

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
				log.Trace().
					Str("error", fmt.Sprintf("error: %v", err)).
					Str(XRequestID, proxy.XRequestID).
					Int64("upAtmptElapsedMicros", time.Since(proxy.Up.Atmpt.startDate).Microseconds()).
					Str("upAtmpt", proxy.Up.Atmpt.print()).
					Msgf("safe to ignore. recovered internally from closed header success channel after request already handled.")
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
			Int("upReadTimeoutSecs", Runner.Connection.Upstream.ReadTimeoutSeconds).
			Msg("upstream connection read timeout fired, aborting upstream response header processing.")
	case <-proxy.Dwn.Aborted:
		proxy.abortAllUpstreamAttempts()
		proxy.Dwn.AbortedFlag = true
		scaffoldUpAttemptLog(proxy).
			Msg("aborting upstream response header processing. downstream connection read timeout fired or user cancelled request")
	case <-proxy.Up.Atmpt.CompleteHeader:
		scaffoldUpAttemptLog(proxy).
			Msg("upstream response headers processed")
	}

	return upstreamResponse, upstreamError
}

func parseUpstreamResponse(upstreamResponse *http.Response, proxy *Proxy) ([]byte, error) {
	var upstreamResponseBody []byte
	var bodyError error

	//get a reference to this before any race conditions may occur
	attemptIndex := proxy.Up.Count - 1

	go func() {
		proxy.Up.Atmpt.StatusCode = upstreamResponse.StatusCode
		upstreamResponseBody, bodyError = ioutil.ReadAll(upstreamResponse.Body)
		if c := bytes.Compare(upstreamResponseBody[0:2], gzipMagicBytes); c == 0 {
			proxy.Up.Atmpt.isGzip = true
		}

		defer func() {
			if err := recover(); err != nil {
				scaffoldUpAttemptLog(proxy).
					Msgf("safe to ignore. recovered internally from closed body success channel after request already handled.")
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
			Int("upReadTimeoutSecs", Runner.Connection.Upstream.ReadTimeoutSeconds).
			Msg("upstream connection read timeout fired, aborting upstream response body processing")
	case <-proxy.Dwn.Aborted:
		proxy.abortAllUpstreamAttempts()
		proxy.Dwn.AbortedFlag = true
		scaffoldUpAttemptLog(proxy).
			Msg("aborting upstream response body processing. downstream connection read timeout fired or user cancelled request")
	case <-proxy.Up.Atmpt.CompleteBody:
		scaffoldUpAttemptLog(proxy).
			Msg("upstream response body processed")
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
				proxy.prepareDownstreamResponseHeaders()
				proxy.sendDownstreamStatusCodeHeader()
				proxy.copyUpstreamResponseBody()
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
	return !proxy.hasDownstreamAborted() &&
		!proxy.hasUpstreamAttemptAborted() &&
		bodyError == nil &&
		proxy.Up.Atmpt.resp.StatusCode < 500
}

func shouldProxyHeader(header string) bool {
	for _, dont := range httpResponseHeadersNoRewrite {
		if header == dont {
			return false
		}
	}
	return true
}

func scaffoldUpAttemptLog(proxy *Proxy) *zerolog.Event {
	var ev *zerolog.Event
	if proxy.XRequestDebug {
		ev = log.Debug()
	} else {
		ev = log.Trace()
	}

	return ev.
		Str(XRequestID, proxy.XRequestID).
		Int64("upAtmptElapsedMicros", time.Since(proxy.Up.Atmpt.startDate).Microseconds()).
		Int64("dwnElapsedMicros", time.Since(proxy.Dwn.startDate).Microseconds()).
		Str("upAtmpt", proxy.Up.Atmpt.print())
}

func logHandledDownstreamRoundtrip(proxy *Proxy) {
	elapsed := time.Since(proxy.Dwn.startDate)
	msg := "downstream response served"
	ev := log.Info()

	if proxy.Dwn.Resp.StatusCode > 399 {
		msg = "downstream error response served"
		ev = log.Warn()
		ev = ev.Str("dwnResErrMsg", proxy.Dwn.Resp.Message)
	}

	ev = ev.Str("dwnReqPath", proxy.Dwn.Path).
		Str("dwnReqMethod", proxy.Dwn.Method).
		Str("dwnReqUserAgent", proxy.Dwn.UserAgent).
		Str("dwnHttpVer", proxy.Dwn.HttpVer).
		Int("dwnResCode", proxy.Dwn.Resp.StatusCode).
		Str("dwnResContentEnc", proxy.contentEncoding()).
		Int64("dwnElapsedMicros", elapsed.Microseconds()).
		Str(XRequestID, proxy.XRequestID)

	if Runner.isTLSMode() {
		ev = ev.Str("dwnTlsVer", proxy.Dwn.TlsVer)
	}

	if proxy.hasMadeUpstreamAttempt() {
		ev = ev.Str("upURI", proxy.resolveUpstreamURI()).
			Str("upLabel", proxy.Up.Atmpt.Label).
			Int("upAtmptResCode", proxy.Up.Atmpt.StatusCode).
			Int64("upAtmptElapsedMicros", time.Since(proxy.Up.Atmpt.startDate).Microseconds()).
			Bool("upAtmptAbort", proxy.Up.Atmpt.AbortedFlag).
			Str("upAtmpt", proxy.Up.Atmpt.print())
	}

	ev.Msg(msg)
}

func logSuccessfulUpstreamAttempt(proxy *Proxy, upstreamResponse *http.Response) {
	scaffoldUpAttemptLog(proxy).
		Int("upAtmptResCode", upstreamResponse.StatusCode).
		Msgf("upstream attempt successful")
}

func logUnsuccessfulUpstreamAttempt(proxy *Proxy, upstreamResponse *http.Response, upstreamError error) {
	ev := scaffoldUpAttemptLog(proxy)
	if upstreamResponse != nil && upstreamResponse.StatusCode > 0 {
		ev = ev.Int("upAtmptResCode", upstreamResponse.StatusCode)
	}
	ev.Msgf("upstream attempt unsuccessful")
}
