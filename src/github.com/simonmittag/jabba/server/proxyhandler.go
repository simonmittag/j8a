package server

import (
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

//XRequestID is a per HTTP request unique identifier
const XRequestID = "X-REQUEST-ID"

var httpClient *http.Client

func proxyHandler(response http.ResponseWriter, request *http.Request) {
	matched := false
	proxy := new(Proxy).
		initXRequestID().
		parseIncoming(request).
		setOutgoing(response)
	for _, route := range Runner.Routes {
		if matched = matchRouteInURI(route, request); matched {
			upstream, label, mapped := route.mapUpstream()
			if mapped {
				handle(proxy.firstAttempt(upstream, label))
			} else {
				//matched, but non mapped requests == configuration error in server
				sendStatusCodeAsJSON(proxy.respondWith(503, "unable to map upstream resource"))
			}
			break
		}
	}
	if !matched {
		sendStatusCodeAsJSON(proxy.respondWith(404, "upstream resource not found"))
	}
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
		defer upstreamResponse.Body.Close()
		upstreamResponseBody, bodyError := ioutil.ReadAll(upstreamResponse.Body)
		if bodyError == nil {
			writeStandardResponseHeaders(proxy.respondWith(200, "OK"))
			//TODO: what is the upstream content type? and other special headers from server...we need to copy relevant headers back down.
			proxy.Downstream.Response.Write([]byte(upstreamResponseBody))
			logUpstreamAttempt(proxy)
		} else {
			log.Warn().Err(bodyError).Msgf("error encountered parsing body")
		}
	} else {
		log.Warn().Err(upstreamError).Msgf("error encountered upstream")
	}
}

func logUpstreamAttempt(proxy *Proxy) {
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

func getHTTPMaxUpstreamAttempts() int {
	if httpUpstreamMaxAttempts == 0 {
		httpUpstreamMaxAttempts, _ = strconv.Atoi(os.Getenv("HTTP_UPSTREAM_MAX_ATTEMPTS"))
		if httpUpstreamMaxAttempts == 0 {
			httpUpstreamMaxAttempts = 2
		}
	}
	return httpUpstreamMaxAttempts
}
