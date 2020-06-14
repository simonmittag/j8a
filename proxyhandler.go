package jabba

import (
	"bytes"
	"fmt"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
	"strings"
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
			url, label, mapped := route.mapURL()
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
	upstreamRequest, _ := http.NewRequest(proxy.Dwn.Method,
		proxy.resolveUpstreamURI(),
		proxy.bodyReader())
	//TODO: test if upstream request headers are reprocessed correctly
	for key, values := range proxy.Dwn.Req.Header {
		upstreamRequest.Header.Set(key, strings.Join(values, " "))
	}
	return upstreamRequest
}

// handle the proxy request
func handle(proxy *Proxy) {
	upstreamResponse, upstreamError := httpClient.Do(scaffoldUpstreamRequest(proxy))
	proxy.Up.Atmpt.resp = upstreamResponse

	if upstreamError == nil {
		//this is required, else we leak TCP connections.
		defer upstreamResponse.Body.Close()
		upstreamResponseBody, bodyError := parseUpstreamResponse(upstreamResponse, proxy)
		upstreamError = bodyError
		proxy.Up.Atmpt.respBody = &upstreamResponseBody
		if shouldSendDownstreamResponse(proxy, bodyError) {
			proxy.processHeaders()
			proxy.copyUpstreamResponseBody()
			logHandledRequest(proxy)
			return
		}
	}
	logUpstreamAttempt(proxy, upstreamResponse, upstreamError)

	if proxy.shouldAttemptRetry() {
		handle(proxy.nextAttempt())
	} else {
		sendStatusCodeAsJSON(proxy.respondWith(502, "bad gateway, unable to read upstream response"))
	}
}

func logUpstreamAttempt(proxy *Proxy, upstreamResponse *http.Response, upstreamError error) {
	ev := log.Trace().
		Str(XRequestID, proxy.XRequestID).
		Int("upstreamAttempt", proxy.Up.Atmpt.Count).
		Int("upstreamMaxAttempt", Runner.Connection.Upstream.MaxAttempts)
	if upstreamResponse != nil && upstreamResponse.StatusCode > 0 {
		ev = ev.Int("upstreamResponseCode", upstreamResponse.StatusCode)
	}
	if upstreamError != nil {
		ev = ev.Err(upstreamError)
	}
	ev.Msg("upstream attempt not proxied")
}

func shouldSendDownstreamResponse(proxy *Proxy, bodyError error) bool {
	return bodyError == nil && proxy.Up.Atmpt.resp.StatusCode < 500
}

func parseUpstreamResponse(upstreamResponse *http.Response, proxy *Proxy) ([]byte, error) {
	upstreamResponseBody, bodyError := ioutil.ReadAll(upstreamResponse.Body)
	if c := bytes.Compare(upstreamResponseBody[0:2], gzipMagicBytes); c == 0 {
		proxy.Up.Atmpt.isGzip = true
	}
	return upstreamResponseBody, bodyError
}

func logHandledRequest(proxy *Proxy) {
	msg := "request served"
	ev := log.Info()

	if proxy.Dwn.Resp.StatusCode > 399 {
		msg = "request not served"
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
			Str("upstreamAttempt", fmt.Sprintf("%d/%d", proxy.Up.Atmpt.Count, Runner.Connection.Upstream.MaxAttempts))
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
