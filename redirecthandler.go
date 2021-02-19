package j8a

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
	"time"
)

const HTTPS = "https://"
const Q = "?"

func redirectHandler(response http.ResponseWriter, request *http.Request) {
	redirectTime := time.Now()

	//preprocess incoming request in proxy object
	proxy := new(Proxy).
		setOutgoing(response).
		parseIncoming(request)

	//all malformed requests are rejected here and we return a 400
	if !validate(proxy) {
		if proxy.Dwn.ReqTooLarge {
			sendStatusCodeAsJSON(proxy.respondWith(413, fmt.Sprintf("http request entity too large, limit is %d bytes", Runner.Connection.Downstream.MaxBodyBytes)))
		} else {
			sendStatusCodeAsJSON(proxy.respondWith(400, "bad or malformed request"))
		}
		return
	}

	httpPort := fmt.Sprintf(":%d", Runner.Connection.Downstream.Http.Port)
	tlsPort := fmt.Sprintf(":%d", Runner.Connection.Downstream.Tls.Port)
	tlsHost := strings.Replace(request.Host, httpPort, tlsPort, 1)

	target := HTTPS + tlsHost + request.URL.Path
	if len(request.URL.RawQuery) > 0 {
		target += Q + request.URL.RawQuery
	}

	log.Info().Str("dwnReqListnr", proxy.Dwn.Listener).
		Str("dwnReqPort", fmt.Sprintf("%d", proxy.Dwn.Port)).
		Str("dwnReqPath", proxy.Dwn.Path).
		Str("dwnReqRemoteAddr", ipr.extractAddr(proxy.Dwn.Req.RemoteAddr)).
		Str("dwnReqMethod", parseMethod(request)).
		Str("dwnReqUserAgent", parseUserAgent(request)).
		Str("dwnReqHttpVer", parseHTTPVer(request)).
		Int("dwnResCode", http.StatusPermanentRedirect).
		Int64("dwnResElpsdMicros", time.Since(redirectTime).Microseconds()).
		Str(XRequestID, proxy.XRequestID).
		Msg("global HTTP redirect to TLS")

	http.Redirect(response, request, target, http.StatusPermanentRedirect)
}
