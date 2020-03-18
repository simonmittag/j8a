package server

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const X_REQUEST_ID = "X-REQUEST-ID"

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
				sendStatusCodeAsJSON(response, request, 503)
			}
			break
		}
	}
	if !matched {
		sendStatusCodeAsJSON(response, request, 404)
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

func parseIncomingRequest(request *http.Request) (*url.URL, string, []byte) {
	url := request.URL
	method := request.Method
	body, _ := ioutil.ReadAll(request.Body)
	log.Trace().
		Str("url", url.Path).
		Str("method", method).
		Int("bodyBytes", len(body)).
		Str(X_REQUEST_ID, request.Header.Get(X_REQUEST_ID)).
		Msg("parsed request")

	return url, method, body
}

// handleUpstreamRequest the meaty part of proxying happens here.
func handleUpstreamRequest(response http.ResponseWriter, request *http.Request, upstream *Upstream, label string) {
	//parse incoming request for URL, method and body
	url, method, _ := parseIncomingRequest(request)

	writeStandardResponseHeaders(response, request, 200)

	//we need a parameterized http client to send the request upstream.
	httpClient := scaffoldHTTPClient()

	switch method {
	case "GET":
		upstreamResponse, err := httpClient.Get(upstream.String() + request.RequestURI)
		if err != nil {
			//we need to work out what to do about retries.
		}
		defer upstreamResponse.Body.Close()

		upstreamResponseBody, err2 := ioutil.ReadAll(upstreamResponse.Body)
		if err2 != nil {
			//we need to work out what to do about retries.
		}

		log.Info().
			Str("url", url.Path).
			Str("method", method).
			Str("upstream", upstream.String()).
			Str("label", label).
			Str("userAgent", request.Header.Get("User-Agent")).
			Str(X_REQUEST_ID, request.Header.Get(X_REQUEST_ID)).
			Int("upstreamResponseCode", 200).
			Int("downstreamResponseCode", 200).
			Msgf("request served")

		response.Write([]byte(upstreamResponseBody))
	}
	return
}

func decorateRequest(r *http.Request) *http.Request {
	r.Header.Set(X_REQUEST_ID, xRequestID())
	return r
}

func xRequestID() string {
	uuid, _ := uuid.NewRandom()
	return fmt.Sprintf("XR-%s-%s", ID, uuid)
}
