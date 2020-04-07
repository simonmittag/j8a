package server

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

//XRequestID is a per HTTP request unique identifier
const XRequestID = "X-REQUEST-ID"

//httpClient is the global user agent for upstream requests
var httpClient *http.Client

//httpResponseHeadersNoRewrite contains a list of headers that are not copied from upstream to downstream to avoid bugs.
var httpResponseHeadersNoRewrite []string = []string{"Date", "Content-Length"}

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
	upstreamRequest, _ := http.NewRequest(proxy.Method,
		proxy.resolveUpstreamURI(),
		proxy.bodyReader())
	//TODO: test if upstream request headers are reprocessed correctly
	for key, values := range proxy.Request.Header {
		upstreamRequest.Header.Set(key, strings.Join(values, " "))
	}
	return upstreamRequest
}

// handle the proxy request
func handle(proxy *Proxy) {
	upstreamRequest := scaffoldUpstreamRequest(proxy)
	upstreamResponse, upstreamError := httpClient.Do(upstreamRequest)
	proxy.Attempt.StatusCode = upstreamResponse.StatusCode

	if upstreamError == nil {
		//this is required, else we leak TCP connections.
		defer upstreamResponse.Body.Close()
		upstreamResponseBody, bodyError := ioutil.ReadAll(upstreamResponse.Body)
		if bodyError == nil {
			writeStandardResponseHeaders(proxy)
			copyUpstreamResponseHeaders(proxy, upstreamResponse)
			resetContentLengthHeader(proxy, upstreamResponseBody)
			writeStatusCodeHeader(proxy.respondWith(200, "OK"))
			copyUpstreamResponseBody(proxy, upstreamResponseBody)
			logHandledRequest(proxy)
			return
		}
		log.Debug().Err(bodyError).Msgf("error encountered parsing body")
	}
	log.Debug().Err(upstreamError).Msgf("error encountered upstream")
	if proxy.shouldAttemptRetry() {
		handle(proxy.nextAttempt())
	} else {
		sendStatusCodeAsJSON(proxy.respondWith(502, "bad gateway, unable to read upstream response"))
	}
}

func resetContentLengthHeader(proxy *Proxy, upstreamResponseBody []byte) {
	if proxy.Method == "HEAD" || len(upstreamResponseBody) == 0 {
		proxy.Response.Writer.Header().Set("Content-Length", "0")
	}
}

func copyUpstreamResponseBody(proxy *Proxy, upstreamResponseBody []byte) {
	proxy.Response.Writer.Write([]byte(upstreamResponseBody))
}

func copyUpstreamResponseHeaders(proxy *Proxy, upstreamResponse *http.Response) {
	for key, values := range upstreamResponse.Header {
		if shouldRewrite(key) {
			proxy.Response.Writer.Header().Set(key, strings.Join(values, " "))
		}
	}
}

//status code must be last, no headers may be written after this one.
func writeStatusCodeHeader(proxy *Proxy) {
	proxy.Response.Writer.WriteHeader(proxy.Response.StatusCode)
}

func logHandledRequest(proxy *Proxy) {
	log.Info().
		Str("path", proxy.URI).
		Str("method", proxy.Method).
		Str("userAgent", proxy.Request.Header.Get("User-Agent")).
		Int("downstreamResponseCode", proxy.Response.StatusCode).
		Str(XRequestID, proxy.XRequestID).
		Str("upstreamURI", proxy.resolveUpstreamURI()).
		Str("upstreamLabel", proxy.Attempt.Label).
		Int("upstreamResponseCode", proxy.Attempt.StatusCode).
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
