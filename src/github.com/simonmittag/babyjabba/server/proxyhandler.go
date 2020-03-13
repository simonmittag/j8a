package server

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const X_REQUEST_ID = "X-REQUEST-ID"

func proxyHandler(response http.ResponseWriter, request *http.Request) {
	matched := false
	decorateRequest(request)
	for _, route := range Runtime.Routes {
		if matchRouteInURI(route, request) {
			upstream, label, mapped := route.mapUpstream()
			if mapped {
				handleUpstreamRequest(response, request, upstream, label)
			} else {
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

func handleUpstreamRequest(response http.ResponseWriter, request *http.Request, upstream *Upstream, label string) {
	//handle request by sending upstream
	url := request.URL
	method := request.Method
	// body, err := ioutil.ReadAll(r.Body)
	// if err != nil {
	// 	log.Debug().Msgf("unable to parse request body for %url", url)
	// }
	writeStandardResponseHeaders(response, request, 200)
	response.Write([]byte(fmt.Sprintf("proxy request for upstream %v", *upstream)))

	log.Info().
		Str("url", url.Path).
		Str("method", method).
		Str("upstream", upstream.String()).
		Str("label", label).
		Str(X_REQUEST_ID, request.Header.Get(X_REQUEST_ID)).
		Int("upstreamResponseCode", 200).
		Int("downstreamResponseCode", 200).
		Msgf("request served")
}

func decorateRequest(r *http.Request) *http.Request {
	r.Header.Set(X_REQUEST_ID, xRequestID())
	return r
}

func xRequestID() string {
	uuid, _ := uuid.NewRandom()
	return fmt.Sprintf("XR-%s-%s", ID, uuid)
}
