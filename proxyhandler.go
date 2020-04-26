package jabba

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
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

	if upstreamError == nil {
		//this is required, else we leak TCP connections.
		defer upstreamResponse.Body.Close()
		upstreamResponseBody, bodyError := parseUpstreamResponse(upstreamResponse, proxy)
		if shouldSendDownstreamResponse(upstreamResponse, bodyError) {
			writeStandardResponseHeaders(proxy)
			copyUpstreamResponseHeaders(proxy, upstreamResponse)
			resetContentLengthHeader(proxy, upstreamResponseBody)
			setContentEncodingHeader(proxy, upstreamResponse)
			writeStatusCodeHeader(proxy.respondWith(upstreamResponse.StatusCode, "none"))
			copyUpstreamResponseBody(proxy, upstreamResponseBody)
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

func shouldSendDownstreamResponse(upstreamResponse *http.Response, bodyError error) bool {
	return bodyError == nil && upstreamResponse.StatusCode < 500
}

func parseUpstreamResponse(upstreamResponse *http.Response, proxy *Proxy) ([]byte, error) {
	upstreamResponseBody, bodyError := ioutil.ReadAll(upstreamResponse.Body)
	if c := bytes.Compare(upstreamResponseBody[0:2], gzipMagicBytes); c == 0 {
		proxy.Up.Atmpt.isGzip = true
	}
	return upstreamResponseBody, bodyError
}

func resetContentLengthHeader(proxy *Proxy, upstreamResponseBody []byte) {
	if proxy.Dwn.Method == "HEAD" || len(upstreamResponseBody) == 0 {
		proxy.Dwn.Resp.Writer.Header().Set(contentLength, "0")
	}
}

func setContentEncodingHeader(proxy *Proxy, upstreamResponse *http.Response) {
	if proxy.Dwn.Resp.SendGzip {
		proxy.Dwn.Resp.Writer.Header().Set(contentEncoding, "gzip")
	} else {
		if proxy.shouldGzipDecodeResponseBody() {
			proxy.Dwn.Resp.Writer.Header().Set(contentEncoding, "identity")
		} else {
			ce := upstreamResponse.Header[contentEncoding]
			if len(ce) > 0 {
				proxy.Dwn.Resp.Writer.Header().Set(contentEncoding, strings.Join(ce, " "))
			}
		}
	}
}

func copyUpstreamResponseBody(proxy *Proxy, upstreamResponseBody []byte) {
	start := time.Now()
	if proxy.shouldGzipEncodeResponseBody() {
		proxy.Dwn.Resp.Writer.Write(Gzip(upstreamResponseBody))
		elapsed := time.Since(start)
		log.Trace().Msgf("copying upstream body with gzip re-encoding took %s", elapsed)
	} else {
		if proxy.shouldGzipDecodeResponseBody() {
			proxy.Dwn.Resp.Writer.Write(Gunzip([]byte(upstreamResponseBody)))
			elapsed := time.Since(start)
			log.Trace().Msgf("copying upstream body with gzip re-decoding took %s", elapsed)
		} else {
			proxy.Dwn.Resp.Writer.Write([]byte(upstreamResponseBody))
			elapsed := time.Since(start)
			log.Trace().Msgf("copying upstream body without coding took %s", elapsed)
		}
	}


}

func copyUpstreamResponseHeaders(proxy *Proxy, upstreamResponse *http.Response) {
	proxy.Up.Atmpt.StatusCode = upstreamResponse.StatusCode
	for key, values := range upstreamResponse.Header {
		if shouldRewrite(key) {
			for _, mval := range values {
				proxy.Dwn.Resp.Writer.Header().Set(key, mval)
			}
		}
	}
}

//status code must be last, no headers may be written after this one.
func writeStatusCodeHeader(proxy *Proxy) {
	proxy.Dwn.Resp.Writer.WriteHeader(proxy.Dwn.Resp.StatusCode)
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
