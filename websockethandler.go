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
	"os"
	"time"
)

const upConDialed = "upstream websocket connection dialed"
const upConClosed = "upstream websocket connection closed"
const upWriteErr = "error writing to upstream websocket, cause: "
const upReadErr = "error reading from upstream websocket, cause: "
const upConWsFail = "upstream failed websocket upgrade"
const upBytesWritten = "%d bytes written to upstream websocket"

const dwnConClosed = "downstream websocket connection closed"
const dwnConUpgraded = "downstream upgraded to websocket connection"
const dwnConWsFail = "downstream failed websocket upgrade, cause: %s"
const dwnReadErr = "error reading from downstream websocket, cause: "
const dwnWriteErr = "error writing to downstream websocket, cause: "
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

func (proxy *Proxy) scaffoldWebsocketLog(e *zerolog.Event) *zerolog.Event {
	return e.Str(XRequestID, proxy.XRequestID).
		Str(dwnReqRemoteAddr, proxy.Dwn.Req.RemoteAddr).
		Int64(dwnElapsedMicros, time.Since(proxy.Dwn.startDate).Microseconds()).
		Str(dwnReqUserAgent, proxy.Dwn.UserAgent).
		Str(dwnReqHttpVer, proxy.Dwn.HttpVer).
		Str(dwnReqPath, proxy.Dwn.Path).
		Str(upReqURI, proxy.resolveUpstreamURI())
}

const upWebsocketConnectionFailed = "upstream websocket connection failed"
const upWebsocketUnspecifiedNetworkEvent = "upstream websocket unspecified network event: %s"
const connect = "connect"

func upgradeWebsocket(proxy *Proxy) {
	//upCon has to run first. if it fails we still want to send a 50x HTTP response.
	upCon, _, _, upErr := ws.DefaultDialer.Dial(context.Background(), proxy.resolveUpstreamURI())
	uev := proxy.scaffoldWebsocketLog(log.Trace())
	if upErr != nil {
		netOpErr, noe := upErr.(*net.OpError)
		wsStatusErr, wse := upErr.(ws.StatusError)
		if noe {
			syscallErr, sce := netOpErr.Err.(*os.SyscallError)
			if sce && syscallErr.Syscall == connect {
				uev.Msg(upWebsocketConnectionFailed)
			} else {
				uev.Msgf(upWebsocketUnspecifiedNetworkEvent, upErr)
			} //TODO: brittle
		} else if wse && 404 == int(wsStatusErr) {
			uev.Msg(upWebsocketConnectionFailed)
		} else {
			uev.Msgf(upWebsocketUnspecifiedNetworkEvent, upErr)
		}

		sendStatusCodeAsJSON(proxy.respondWith(502, upConWsFail))
		return
	} else {
		uev.Msg(upConDialed)
	}

	dwnCon, _, _, dwnErr := ws.UpgradeHTTP(proxy.Dwn.Req, proxy.Dwn.Resp.Writer)
	if dwnErr != nil {
		msg := fmt.Sprintf(dwnConWsFail, dwnErr)
		proxy.scaffoldWebsocketLog(log.Trace()).Msg(msg)
		sendStatusCodeAsJSON(proxy.respondWith(400, msg))
		return
	} else {
		proxy.scaffoldWebsocketLog(log.Info()).Msg(dwnConUpgraded)
	}

	var status = make(chan WebsocketStatus)

	go readDwnWebsocket(dwnCon, upCon, proxy, status)
	go readUpWebsocket(dwnCon, upCon, proxy, status)
	_ = <-status

	ue := upCon.Close()
	if ue == nil {
		proxy.scaffoldWebsocketLog(log.Trace()).Msg(upConClosed)
	}

	de := dwnCon.Close()
	if de == nil {
		proxy.scaffoldWebsocketLog(log.Info()).Msg(dwnConClosed)
	}
}

func readDwnWebsocket(dwnCon net.Conn, upCon net.Conn, proxy *Proxy, status chan WebsocketStatus) {
ReadDwn:
	for {
		msg, op, dre := wsutil.ReadClientData(dwnCon)
		if dre == nil {
			uwe := wsutil.WriteClientMessage(upCon, op, msg)
			if uwe == nil {
				proxy.scaffoldWebsocketLog(log.Trace()).
					Int8(opCode, int8(op)).
					Int(msgBytes, len(msg)).
					Msgf(upBytesWritten, len(msg))
			} else {
				if io.EOF == uwe {
					status <- WebsocketStatus{
						UpExit: uwe,
					}
				} else {
					proxy.scaffoldWebsocketLog(log.Warn()).
						Int8(opCode, int8(op)).
						Msg(upWriteErr + uwe.Error())
					status <- WebsocketStatus{
						UpErr: uwe,
					}
				}
				break ReadDwn
			}
		} else {
			if io.EOF == dre {
				status <- WebsocketStatus{
					DwnExit: dre,
				}
			} else {
				proxy.scaffoldWebsocketLog(log.Warn()).
					Int8(opCode, int8(op)).
					Msg(dwnReadErr + dre.Error())
				status <- WebsocketStatus{
					DwnErr: dre,
				}
			}
			break ReadDwn
		}
	}
}

func readUpWebsocket(dwnCon net.Conn, upCon net.Conn, proxy *Proxy, status chan WebsocketStatus) {
ReadUp:
	for {
		msg, op, ure := wsutil.ReadServerData(upCon)
		if ure == nil {
			dwe := wsutil.WriteServerMessage(dwnCon, op, msg)
			if dwe == nil {
				proxy.scaffoldWebsocketLog(log.Trace()).
					Int8(opCode, int8(op)).
					Int(msgBytes, len(msg)).
					Msgf(dwnBytesWritten, len(msg))
			} else {
				if io.EOF == dwe {
					status <- WebsocketStatus{
						DwnExit: dwe,
					}
				} else {
					proxy.scaffoldWebsocketLog(log.Warn()).
						Int8(opCode, int8(op)).
						Msg(dwnWriteErr + dwe.Error())
					status <- WebsocketStatus{
						DwnErr: dwe,
					}
				}
				break ReadUp
			}
		} else {
			if io.EOF == ure {
				status <- WebsocketStatus{
					UpExit: ure,
				}
			} else if _, netop := ure.(*net.OpError); netop {
				status <- WebsocketStatus{
					UpExit: ure,
				}
			} else {
				proxy.scaffoldWebsocketLog(log.Warn()).
					Int8(opCode, int8(op)).
					Msg(upReadErr + ure.Error())
				status <- WebsocketStatus{
					UpErr: ure,
				}
			}
			break ReadUp
		}
	}
}
