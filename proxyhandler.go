package jabba

import (
	"bytes"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

//XRequestID is a per HTTP request unique identifier
const XRequestID = "X-REQUEST-ID"
const contentEncoding = "Content-Encoding"
const contentLength = "Content-Length"
const date = "Date"

//httpClient is the global user agent for upstream requests
var httpClient HTTPClient

//httpResponseHeadersNoRewrite contains a list of headers that are not copied from upstream to downstream to avoid bugs.
var httpResponseHeadersNoRewrite []string = []string{date, contentLength, contentEncoding}

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
		if matched = matchRouteInURI(route, request); matched {
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

func matchRouteInURI(route Route, request *http.Request) bool {
	matched, _ := regexp.MatchString("^"+route.Path, request.RequestURI)
	return matched
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
		proxy.Up.Atmpt.respBody = &upstreamResponseBody
		if shouldSendDownstreamResponse(proxy, bodyError) {
			proxy.processHeaders()
			proxy.copyUpstreamResponseBody()
			logHandledRequest(proxy)
			return
		}
	}
	log.Trace().
		Str(XRequestID, proxy.XRequestID).
		Int("upstreamResponseCode", upstreamResponse.StatusCode).
		Int("upstreamAttempt", proxy.Up.Atmpt.Count).
		Int("upstreamMaxAttempt", Runner.Connection.Upstream.MaxAttempts).
		Err(upstreamError).Msgf("upstream attempt did not pass exit criteria for proxying")
	if proxy.shouldAttemptRetry() {
		handle(proxy.nextAttempt())
	} else {
		sendStatusCodeAsJSON(proxy.respondWith(502, "bad gateway, unable to read upstream response"))
	}
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
	log.Info().
		Str("path", proxy.Dwn.Path).
		Str("method", proxy.Dwn.Method).
		Str("userAgent", proxy.Dwn.UserAgent).
		Int("downstreamResponseCode", proxy.Dwn.Resp.StatusCode).
		Str("downstreamContentEncoding", proxy.contentEncoding()).
		Str(XRequestID, proxy.XRequestID).
		Str("upstreamURI", proxy.resolveUpstreamURI()).
		Str("upstreamLabel", proxy.Up.Atmpt.Label).
		Int("upstreamResponseCode", proxy.Up.Atmpt.StatusCode).
		Msgf("request served")
}

func shouldRewrite(header string) bool {
	for _, dont := range httpResponseHeadersNoRewrite {
		if header == dont {
			return false
		}
	}
	return true
}
