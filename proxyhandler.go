package jabba

import (
	"bytes"
	"context"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

//XRequestID is a per HTTP request unique identifier
const XRequestID = "X-REQUEST-ID"
const contentEncoding = "Content-Encoding"
const contentLength = "Content-Length"
const date = "Date"
const server = "Server"

//httpClient is the global user agent for upstream requests
var httpClient HTTPClient

//httpResponseHeadersNoRewrite contains a list of headers that are not copied from upstream to downstream to avoid bugs.
var httpResponseHeadersNoRewrite []string = []string{date, contentLength, contentEncoding, server}

// main proxy handling
func proxyHandler(response http.ResponseWriter, request *http.Request) {
	matched := false

	//preprocess incoming request in proxy object
	proxy := new(Proxy).
		initXRequestID().
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
				//unmapped request mean an internal configuration error in server
				sendStatusCodeAsJSON(proxy.respondWith(503, "unable to map upstream resource"))
				return
			}
			break
		}
	}

	//unmatched paths means we have no route for this and always return a 404
	if !matched {
		sendStatusCodeAsJSON(proxy.respondWith(404, "upstream resource not found"))
		return
	}
}

func validate(proxy *Proxy) bool {
	return proxy.hasLegalHTTPMethod()
}

func scaffoldUpstreamRequest(proxy *Proxy) *http.Request {
	ctx, cancel := context.WithCancel(context.TODO())
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
	return upstreamRequest
}

// handle the proxy request
func handle(proxy *Proxy) {
	upstreamResponse, upstreamError := performUpstreamRequest(proxy)
	if upstreamResponse != nil && upstreamResponse.Body != nil {
		defer upstreamResponse.Body.Close()
	}

	if !processUpstreamResponse(proxy, upstreamResponse, upstreamError) {
		if proxy.shouldAttemptRetry() {
			handle(proxy.nextAttempt())
		} else {
			//sends 502 if no more retries possible
			sendStatusCodeAsJSON(proxy.respondWith(502, "bad gateway, unable to read upstream response"))
		}
	}
}

func processUpstreamResponse(proxy *Proxy, upstreamResponse *http.Response, upstreamError error) bool {
	//process only if we can work with upstream attempt
	if upstreamResponse != nil && upstreamError == nil && !proxy.hasUpstreamAtmptAborted() {
		//jabba blocks here when waiting for upstream body
		upstreamResponseBody, bodyError := parseUpstreamResponse(upstreamResponse, proxy)
		upstreamError = bodyError
		proxy.Up.Atmpt.respBody = &upstreamResponseBody
		if shouldProxyUpstreamResponse(proxy, bodyError) {
			//sends proxied response, 200-498
			proxy.processHeaders()
			proxy.copyUpstreamResponseBody()
			logHandledRequest(proxy)
			return true
		} else if proxy.hasDownstreamAborted() {
			//sends 503 because of abort, but not here. handled by timeouthandler
			//TODO all 503s in jabba today should be 504
			logHandledRequest(proxy)
			return true
		}
	}
	//now log unsuccessful and retry or exit with status code.
	logUnsuccessfulUpstreamAttempt(proxy, upstreamResponse, upstreamError)
	return false
}

func performUpstreamRequest(proxy *Proxy) (*http.Response, error) {
	req := scaffoldUpstreamRequest(proxy)

	var upstreamResponse *http.Response
	var upstreamError error

	go func() {
		//this blocks until upstream headers come in
		upstreamResponse, upstreamError = httpClient.Do(req)
		proxy.Up.Atmpt.resp = upstreamResponse

		//defer func() {
		//	if err := recover(); err != nil {
		//		log.Trace().
		//			Str(XRequestID, proxy.XRequestID).
		//			Str("upstreamAttempt", proxy.Up.Atmpt.print()).
		//			Msgf("recovered internally from closed header success channel after request already handled. safe to ignore")
		//	}
		//}()

		if proxy.Up.Atmpt.CompleteHeader != nil && !proxy.Up.Atmpt.AbortedFlag && !proxy.Dwn.AbortedFlag {
			close(proxy.Up.Atmpt.CompleteHeader)
		}
	}()

	//race for upstream headers complete, upstream timeout or downstream abort (timeout or cancellation)
	select {

	case <-proxy.Up.Atmpt.Aborted:
		proxy.Up.Atmpt.AbortedFlag = true
		proxy.Up.Atmpt.StatusCode = 0
		log.Trace().
			Str(XRequestID, proxy.XRequestID).
			Str("upstreamAttempt", proxy.Up.Atmpt.print()).
			Int("upstreamReadTimeoutSeconds", Runner.Connection.Upstream.ReadTimeoutSeconds).
			Msg("upstream connection read timeout fired, aborting upstream response header processing.")
	case <-proxy.Dwn.Aborted:
		proxy.Dwn.AbortedFlag = true
		log.Trace().
			Str(XRequestID, proxy.XRequestID).
			Str("upstreamAttempt", proxy.Up.Atmpt.print()).
			Msg("aborting upstream response header processing. downstream connection read timeout fired or user cancelled request")
	case <-proxy.Up.Atmpt.CompleteHeader:
		log.Trace().
			Str(XRequestID, proxy.XRequestID).
			Str("upstreamAttempt", proxy.Up.Atmpt.print()).
			Msg("upstream response headers processed")
	}

	return upstreamResponse, upstreamError
}

func logUnsuccessfulUpstreamAttempt(proxy *Proxy, upstreamResponse *http.Response, upstreamError error) {
	ev := log.Trace().
		Str(XRequestID, proxy.XRequestID).
		Str("upstreamAttempt", proxy.Up.Atmpt.print())
	if upstreamResponse != nil && upstreamResponse.StatusCode > 0 {
		ev = ev.Int("upstreamResponseCode", upstreamResponse.StatusCode)
	}
	//if upstreamError != nil {
	//	ev = ev.Err(upstreamError)
	//}
	ev.Msg("upstream attempt unsuccessful")
}

func shouldProxyUpstreamResponse(proxy *Proxy, bodyError error) bool {
	return !proxy.hasDownstreamAborted() &&
		!proxy.hasUpstreamAtmptAborted() &&
		bodyError == nil &&
		proxy.Up.Atmpt.resp.StatusCode < 500
}

func parseUpstreamResponse(upstreamResponse *http.Response, proxy *Proxy) ([]byte, error) {
	var upstreamResponseBody []byte
	var bodyError error

	go func() {
		upstreamResponseBody, bodyError = ioutil.ReadAll(upstreamResponse.Body)
		if c := bytes.Compare(upstreamResponseBody[0:2], gzipMagicBytes); c == 0 {
			proxy.Up.Atmpt.isGzip = true
		}

		//defer func() {
		//	if err := recover(); err != nil {
		//		log.Trace().
		//			Str(XRequestID, proxy.XRequestID).
		//			Str("upstreamAttempt", proxy.Up.Atmpt.print()).
		//			Msgf("recovered internally from closed body success channel after request already handled. safe to ignore")
		//	}
		//}()

		//this is ok, see: https://stackoverflow.com/questions/8593645/is-it-ok-to-leave-a-channel-open#:~:text=5%20Answers&text=It's%20OK%20to%20leave%20a,it%20will%20be%20garbage%20collected.&text=Closing%20the%20channel%20is%20a,that%20no%20more%20data%20follows.
		if proxy.Up.Atmpt.CompleteBody != nil && !proxy.Up.Atmpt.AbortedFlag && !proxy.Dwn.AbortedFlag {
			close(proxy.Up.Atmpt.CompleteBody)
		}
	}()

	select {
	case <-proxy.Up.Atmpt.Aborted:
		proxy.Up.Atmpt.AbortedFlag = true
		log.Trace().
			Str(XRequestID, proxy.XRequestID).
			Str("upstreamAttempt", proxy.Up.Atmpt.print()).
			Int("upstreamReadTimeoutSeconds", Runner.Connection.Upstream.ReadTimeoutSeconds).
			Msg("upstream connection read timeout fired, aborting upstream response body processing")
	case <-proxy.Dwn.Aborted:
		proxy.Dwn.AbortedFlag = true
		log.Trace().
			Str(XRequestID, proxy.XRequestID).
			Str("upstreamAttempt", proxy.Up.Atmpt.print()).
			Msg("aborting upstream response body processing. downstream connection read timeout fired or user cancelled request")
	case <-proxy.Up.Atmpt.CompleteBody:
		log.Trace().
			Str(XRequestID, proxy.XRequestID).
			Str("upstreamAttempt", proxy.Up.Atmpt.print()).
			Msg("upstream response body processed")
	}

	return upstreamResponseBody, bodyError
}

func logHandledRequest(proxy *Proxy) {
	msg := "downstream response served"
	ev := log.Info()

	if proxy.Dwn.Resp.StatusCode > 399 {
		msg = "downstream error response served"
		ev = log.Warn()
	}

	ev = ev.Str("path", proxy.Dwn.Path).
		Str("method", proxy.Dwn.Method).
		Str("userAgent", proxy.Dwn.UserAgent).
		Int("downstreamResponseCode", proxy.Dwn.Resp.StatusCode).
		Str("downstreamContentEncoding", proxy.contentEncoding()).
		Str(XRequestID, proxy.XRequestID)

	if proxy.hasMadeUpstreamAttempt() {
		ev = ev.Str("upstreamURI", proxy.resolveUpstreamURI()).
			Str("upstreamLabel", proxy.Up.Atmpt.Label).
			Int("upstreamResponseCode", proxy.Up.Atmpt.StatusCode).
			Str("upstreamAttempt", proxy.Up.Atmpt.print())
	}

	ev.Msg(msg)
}

func shouldRewrite(header string) bool {
	for _, dont := range httpResponseHeadersNoRewrite {
		if header == dont {
			return false
		}
	}
	return true
}
