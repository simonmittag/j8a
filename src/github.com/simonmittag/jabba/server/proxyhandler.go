package server

import (
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

//XRequestID is a per HTTP request unique identifier
const XRequestID = "X-REQUEST-ID"

var httpClient *http.Client
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
			upstream, label, mapped := route.mapUpstream()
			if mapped {
				//mapped requests get sent upstream
				handle(proxy.firstAttempt(upstream, label))
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

func scaffoldHTTPClient() *http.Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: time.Duration(Runner.
						Connection.
						Client.
						ConnectTimeoutSeconds) * time.Second,
					KeepAlive: time.Duration(Runner.
						Connection.
						Client.
						TCPConnectionKeepAliveSeconds) * time.Second,
				}).Dial,
				//TLS handshake timeout is the same as connection timeout
				TLSHandshakeTimeout: time.Duration(Runner.
					Connection.
					Client.
					ConnectTimeoutSeconds) * time.Second,
			},
		}
	}
	return httpClient
}

func scaffoldUpstreamRequest(proxy *Proxy) *http.Request {
	upstreamRequest, _ := http.NewRequest(proxy.Method,
		proxy.resolveUpstreamURI(),
		proxy.bodyReader())
	//TODO: test if upstream request headers are reproced correctly
	for key, values := range proxy.Downstream.Request.Header {
		upstreamRequest.Header.Set(key, strings.Join(values, " "))
	}
	return upstreamRequest
}

// handle the proxy request
func handle(proxy *Proxy) {
	upstreamRequest := scaffoldUpstreamRequest(proxy)
	upstreamResponse, upstreamError := scaffoldHTTPClient().Do(upstreamRequest)

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
		proxy.Downstream.Response.Header().Set("Content-Length", "0")
	}
}

func copyUpstreamResponseBody(proxy *Proxy, upstreamResponseBody []byte) {
	proxy.Downstream.Response.Write([]byte(upstreamResponseBody))
}

func copyUpstreamResponseHeaders(proxy *Proxy, upstreamResponse *http.Response) {
	for key, values := range upstreamResponse.Header {
		if shouldRewrite(key) {
			proxy.Downstream.Response.Header().Set(key, strings.Join(values, " "))
		}
	}
}

//status code must be last, no headers may be written after this one.
func writeStatusCodeHeader(proxy *Proxy) {
	proxy.Downstream.Response.WriteHeader(proxy.Downstream.StatusCode)
}

func logHandledRequest(proxy *Proxy) {
	log.Info().
		Str("url", proxy.URI).
		Str("method", proxy.Method).
		Str("upstream", proxy.Attempt.Upstream.String()).
		Str("label", proxy.Attempt.Label).
		Str("userAgent", proxy.Downstream.Request.Header.Get("User-Agent")).
		Str(XRequestID, proxy.XRequestID).
		Int("upstreamResponseCode", proxy.Attempt.StatusCode).
		Int("downstreamResponseCode", proxy.Downstream.StatusCode).
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
