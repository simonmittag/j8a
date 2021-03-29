package j8a

import (
	"context"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

func websocketHandler(response http.ResponseWriter, request *http.Request) {
	proxyHandler(response, request, upgradeWebsocket)
}

func upgradeWebsocket(proxy *Proxy) {
	wslog := func(e *zerolog.Event) *zerolog.Event {
		return e.Str(XRequestID, proxy.XRequestID).
			Str("dwnReqRemoteAddr", proxy.Dwn.Req.RemoteAddr).
			Int64("dwnElapsedMicros", time.Since(proxy.Dwn.startDate).Microseconds()).
			Str("dwnReqUserAgent", proxy.Dwn.UserAgent).
			Str("dwnReqHttpVer", proxy.Dwn.HttpVer)
	}

	dsconn, _, _, dserr := ws.UpgradeHTTP(proxy.Dwn.Req, proxy.Dwn.Resp.Writer)
	if dserr != nil {
		msg := fmt.Sprintf("downstream failed websocket upgrade, cause: %s", dserr)
		wslog(log.Warn()).Msg(msg)
		sendStatusCodeAsJSON(proxy.respondWith(400, msg))
	} else {
		wslog(log.Trace()).
			Msg("downstream upgraded to websocket connection")
	}

	usconn, _, _, userr := ws.DefaultDialer.Dial(context.Background(), proxy.resolveUpstreamURI())
	if userr != nil {
		msg := fmt.Sprintf("upstream failed websocket upgrade, cause: %s", userr)
		wslog(log.Warn()).Msg(msg)
		sendStatusCodeAsJSON(
			proxy.respondWith(400, msg))
	} else {
		wslog(log.Trace()).
			Str("upReqURI", proxy.resolveUpstreamURI()).
			Msg("upstream upgraded to websocket connection")
	}

	go func() {
		defer dsconn.Close()
		defer usconn.Close()

	ReadDs:
		for {
			msg, op, err := wsutil.ReadClientData(dsconn)
			if err != nil {
				var msg string
				if "EOF" == err.Error() {
					msg = "downstream websocket connection closed"
				} else {
					msg = "error reading ds client data: " + err.Error()
				}
				wslog(log.Trace()).Msg(msg)
				break ReadDs
			} else {
				err = wsutil.WriteServerMessage(usconn, op, msg)
				if err != nil {
					msg := "error writing us server data: " + err.Error()
					wslog(log.Warn()).Msg(msg)
					wsutil.WriteClientMessage(dsconn, op, []byte(msg))
					break ReadDs
				}
			}
		}
	}()

	//go func() {
	//	defer dsconn.Close()
	//	defer usconn.Close()
	//
	//	ReadUs:
	//	for {
	//		msg, op, err := wsutil.ReadClientData(usconn)
	//		if err != nil {
	//			msg := "error reading us server data: " +err.Error()
	//			log.Trace().Msg(msg)
	//			wsutil.WriteClientMessage(dsconn, op, []byte(msg))
	//			break ReadUs
	//		}
	//		err = wsutil.WriteServerMessage(dsconn, op, msg)
	//		if err != nil {
	//			msg := "error writing ds client data: " +err.Error()
	//			log.Trace().Msg(msg)
	//			wsutil.WriteClientMessage(dsconn, op, []byte((msg)))
	//			break ReadUs
	//		}
	//	}
	//}()
}
