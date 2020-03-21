package server

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

//XRequestID is a per HTTP request unique identifier
const XRequestID = "X-REQUEST-ID"

var httpClient *http.Client

func proxyHandler(response http.ResponseWriter, request *http.Request) {
	matched := false
	decorateRequest(request)
	for _, route := range Runner.Routes {
		if matched = matchRouteInURI(route, request); matched {
			upstream, label, mapped := route.mapUpstream()
			if mapped {
				handleUpstreamRequest(response, request, upstream, label)
			} else {
				//matched, but non mapped requests == configuration error in server
				sendStatusCodeAsJSON(response, request, 503, "unable to map upstream resource")
			}
			break
		}
	}
	if !matched {
		sendStatusCodeAsJSON(response, request, 404, "upstream resource not found")
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

// handleUpstreamRequest the meaty part of proxying happens here.
func handleUpstreamRequest(response http.ResponseWriter, request *http.Request, upstream *Upstream, label string) {
	//parse incoming request and build ProxyRequest object
	proxyRequest := new(ProxyRequest).
		parseIncoming(request).
		firstAttempt(upstream, label)

	//make the actual HTTP request. TODO: cannot do post right now because of nil body reader
	upstreamRequest, _ := http.NewRequest(proxyRequest.Method, proxyRequest.resolveUpstreamURI(), nil)
	for key, values := range request.Header {
		upstreamRequest.Header.Set(key, strings.Join(values, ""))
	}
	upstreamResponse, upstreamError := scaffoldHTTPClient().Do(upstreamRequest)

	if upstreamError == nil {
		defer upstreamResponse.Body.Close()
		upstreamResponseBody, bodyError := ioutil.ReadAll(upstreamResponse.Body)
		if bodyError == nil {
			writeStandardResponseHeaders(response, request, 200)
			//what is the upstream content type?
			response.Write([]byte(upstreamResponseBody))
			log.Info().
				Str("url", proxyRequest.URI).
				Str("method", proxyRequest.Method).
				Str("upstream", upstream.String()).
				Str("label", proxyRequest.UpstreamAttempt.Label).
				Str("userAgent", request.Header.Get("User-Agent")).
				Str(XRequestID, request.Header.Get(XRequestID)).
				Int("upstreamResponseCode", 200).
				Int("downstreamResponseCode", 200).
				Msgf("request served")
		} else {
			log.Warn().Err(bodyError).Msgf("error encountered parsing body")
		}
	} else {
		log.Warn().Err(upstreamError).Msgf("error encountered upstream")
	}
}

func decorateRequest(r *http.Request) *http.Request {
	r.Header.Set(XRequestID, xRequestID())
	return r
}

func xRequestID() string {
	uuid, _ := uuid.NewRandom()
	return fmt.Sprintf("XR-%s-%s", ID, uuid)
}
