package j8a

import (
	"context"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"net"
	"net/http"
	"time"
)

const upConDialed = "upstream websocket connection dialed"
const upConClosed = "upstream websocket connection closed"
const upErr = "error writing upstream data, cause: "
const upConWsFail = "upstream failed websocket upgrade, cause: %s"
const dwnConClosed = "downstream websocket connection closed"
const dwnConUpgraded = "downstream upgraded to websocket connection"
const dwnConWsFail = "downstream failed websocket upgrade, cause: %s"
const dwnErr = "error reading downstream data, cause: "

const opCode = "opCode"
const msgBytes = "msgBytes"
const upBytesWritten = "%d bytes written to upstream"

func websocketHandler(response http.ResponseWriter, request *http.Request) {
	proxyHandler(response, request, upgradeWebsocket)
}

func (proxy *Proxy) logstub(e *zerolog.Event) *zerolog.Event {
	return e.Str(XRequestID, proxy.XRequestID).
		Str(dwnReqRemoteAddr, proxy.Dwn.Req.RemoteAddr).
		Int64(dwnElapsedMicros, time.Since(proxy.Dwn.startDate).Microseconds()).
		Str(dwnRegUserAgent, proxy.Dwn.UserAgent).
		Str(dwnRegHttpVer, proxy.Dwn.HttpVer)
}

func upgradeWebsocket(proxy *Proxy) {
	dwnCon, _, _, dwnErr := ws.UpgradeHTTP(proxy.Dwn.Req, proxy.Dwn.Resp.Writer)
	if dwnErr != nil {
		msg := fmt.Sprintf(dwnConWsFail, dwnErr)
		proxy.logstub(log.Warn()).Msg(msg)
		sendStatusCodeAsJSON(proxy.respondWith(400, msg))
	} else {
		proxy.logstub(log.Trace()).Msg(dwnConUpgraded)
	}

	upCon, _, _, upErr := ws.DefaultDialer.Dial(context.Background(), proxy.resolveUpstreamURI())
	if upErr != nil {
		msg := fmt.Sprintf(upConWsFail, upErr)
		proxy.logstub(log.Warn()).Msg(msg)
		sendStatusCodeAsJSON(proxy.respondWith(400, msg))
	} else {
		proxy.logstub(log.Trace()).
			Str(upReqURI, proxy.resolveUpstreamURI()).
			Msg(upConDialed)
	}

	go readDwnWebsocket(dwnCon, upCon, proxy)

}

func readDwnWebsocket(dwnCon net.Conn, upCon net.Conn, proxy *Proxy) {

ReadDs:
	for {
		msg, op, err := wsutil.ReadClientData(dwnCon)
		if err == nil {
			err = wsutil.WriteClientMessage(upCon, op, msg)
			if err != nil {
				var ulm string
				if io.EOF == err {
					ulm = upConClosed
				} else {
					ulm = upErr + err.Error()
				}
				proxy.logstub(log.Warn()).
					Int8(opCode, int8(op)).
					Msg(ulm)
				break ReadDs
			} else {
				proxy.logstub(log.Trace()).
					Int8(opCode, int8(op)).
					Int(msgBytes, len(msg)).
					Msgf(upBytesWritten, len(msg))
			}
		} else {
			if io.EOF == err {
				proxy.logstub(log.Trace()).
					Int8(opCode, int8(op)).
					Msg(dwnConClosed)
			} else {
				proxy.logstub(log.Warn()).
					Int8(opCode, int8(op)).
					Msg(dwnErr + err.Error())
			}
			break ReadDs
		}
	}

	upCon.Close()
	proxy.logstub(log.Trace()).Msg(upConClosed)

	dwnCon.Close()
	proxy.logstub(log.Trace()).Msg(dwnConClosed)
}
