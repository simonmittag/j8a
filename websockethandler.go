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
const upErr = "error writing to upstream websocket, cause: "
const upConWsFail = "upstream failed websocket upgrade, cause: %s"
const upBytesWritten = "%d bytes written to upstream websocket"

const dwnConClosed = "downstream websocket connection closed"
const dwnConUpgraded = "downstream upgraded to websocket connection"
const dwnConWsFail = "downstream failed websocket upgrade, cause: %s"
const dwnErr = "error reading from downstream websocket, cause: "
const dwnBytesWritten = "%d bytes written to downstream websocket"

const opCode = "opCode"
const msgBytes = "msgBytes"

type WebsocketStatus struct {
	DwnExit error
	DwnErr  error
	UpExit  error
	UpErr   error
}

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
	//up has to run first to catch missing connections. once dwn is upgraded
	//to websocket we cannot send HTTP 400 bad request.
	upCon, _, _, upErr := ws.DefaultDialer.Dial(context.Background(), proxy.resolveUpstreamURI())
	if upErr != nil {
		msg := fmt.Sprintf(upConWsFail, upErr)
		proxy.logstub(log.Warn()).Msg(msg)
		sendStatusCodeAsJSON(proxy.respondWith(502, msg))
		return
	} else {
		proxy.logstub(log.Trace()).
			Str(upReqURI, proxy.resolveUpstreamURI()).
			Msg(upConDialed)
	}

	dwnCon, _, _, dwnErr := ws.UpgradeHTTP(proxy.Dwn.Req, proxy.Dwn.Resp.Writer)
	if dwnErr != nil {
		msg := fmt.Sprintf(dwnConWsFail, dwnErr)
		proxy.logstub(log.Warn()).Msg(msg)
		sendStatusCodeAsJSON(proxy.respondWith(400, msg))
		return
	} else {
		proxy.logstub(log.Info()).Msg(dwnConUpgraded)
	}

	var status = make(chan WebsocketStatus)

	go readDwnWebsocket(dwnCon, upCon, proxy, status)
	go readUpWebsocket(dwnCon, upCon, proxy, status)

	closer := func() {
		ue := upCon.Close()
		if ue == nil {
			proxy.logstub(log.Trace()).Msg(upConClosed)
		}

		de := dwnCon.Close()
		if de == nil {
			proxy.logstub(log.Info()).Msg(dwnConClosed)
		}
	}

	_ = <-status
	closer()
}

func readDwnWebsocket(dwnCon net.Conn, upCon net.Conn, proxy *Proxy, status chan WebsocketStatus) {
ReadDwn:
	for {
		msg, op, readDwnErr := wsutil.ReadClientData(dwnCon)
		if readDwnErr == nil {
			upWriteErr := wsutil.WriteClientMessage(upCon, op, msg)
			if upWriteErr == nil {
				proxy.logstub(log.Trace()).
					Int8(opCode, int8(op)).
					Int(msgBytes, len(msg)).
					Msgf(upBytesWritten, len(msg))
			} else {
				if io.EOF == upWriteErr {
					status <- WebsocketStatus{
						UpExit: upWriteErr,
					}
				} else {
					proxy.logstub(log.Warn()).
						Int8(opCode, int8(op)).
						Msg(upErr + upWriteErr.Error())
					status <- WebsocketStatus{
						UpErr: upWriteErr,
					}
				}
				break ReadDwn
			}
		} else {
			if io.EOF == readDwnErr {
				status <- WebsocketStatus{
					DwnExit: readDwnErr,
				}
			} else {
				proxy.logstub(log.Warn()).
					Int8(opCode, int8(op)).
					Msg(dwnErr + readDwnErr.Error())
				status <- WebsocketStatus{
					DwnErr: readDwnErr,
				}
			}
			break ReadDwn
		}
	}
}

func readUpWebsocket(dwnCon net.Conn, upCon net.Conn, proxy *Proxy, status chan WebsocketStatus) {
ReadUp:
	for {
		msg, op, readUpErr := wsutil.ReadServerData(upCon)
		if readUpErr == nil {
			writeDwnErr := wsutil.WriteServerMessage(dwnCon, op, msg)
			if writeDwnErr == nil {
				proxy.logstub(log.Trace()).
					Int8(opCode, int8(op)).
					Int(msgBytes, len(msg)).
					Msgf(dwnBytesWritten, len(msg))
			} else {
				if io.EOF == writeDwnErr {
					status <- WebsocketStatus{
						DwnExit: writeDwnErr,
					}
				} else {
					proxy.logstub(log.Warn()).
						Int8(opCode, int8(op)).
						Msg(dwnErr + writeDwnErr.Error())
					status <- WebsocketStatus{
						DwnErr: writeDwnErr,
					}
				}
				break ReadUp
			}
		} else {
			if io.EOF == readUpErr {
				status <- WebsocketStatus{
					UpExit: readUpErr,
				}
			} else {
				proxy.logstub(log.Warn()).
					Int8(opCode, int8(op)).
					Msg(upErr + readUpErr.Error())
				status <- WebsocketStatus{
					UpErr: readUpErr,
				}
			}
			break ReadUp
		}
	}
}
