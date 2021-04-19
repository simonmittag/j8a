package j8a

import (
	"context"
	"fmt"
	"github.com/simonmittag/ws"
	"github.com/simonmittag/ws/wsutil"
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
const dwnConWsFail = "downstream connection closed, failed websocket upgrade, cause: %s"
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

type WebsocketTx struct {
	UpBytesRead   int64
	UpBytesWrite  int64
	DwnBytesRead  int64
	DwnBytesWrite int64
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
	var status = make(chan WebsocketStatus)
	var tx *WebsocketTx = &WebsocketTx{}

	//upCon has to run first. if it fails we still want to send a 50x HTTP response.
	upCon, _, _, upErr := ws.DefaultDialer.Dial(context.Background(), proxy.resolveUpstreamURI())
	defer func() {
		if upCon != nil {
			ws.WriteFrame(upCon, ws.NewCloseFrame(ws.NewCloseFrameBody(ws.StatusNormalClosure, upConClosed)))
			upCon.Close()
			proxy.scaffoldWebsocketLog(log.Trace()).
				Int64(upBytesRead, tx.UpBytesRead).
				Int64(upBytesWrite, tx.UpBytesWrite).
				Msg(upConClosed)
		}
	}()

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
			}
		} else if wse && 400 <= int(wsStatusErr) && 599 >= int(wsStatusErr) {
			uev.Msg(upWebsocketConnectionFailed)
		} else {
			uev.Msgf(upWebsocketUnspecifiedNetworkEvent, upErr)
		}

		sendStatusCodeAsJSON(proxy.respondWith(502, upConWsFail))
		return
	} else {
		uev.Msg(upConDialed)
	}


	dwnCon, _, _, dwnErr := ws.DefaultHTTPUpgrader.Upgrade(proxy.Dwn.Req, proxy.Dwn.Resp.Writer)
	if dwnErr != nil {
		rce, ok := dwnErr.(*ws.RejectConnectionErrorType)
		if ok {
			log.Warn().
				Int(dwnResCode, rce.Code()).
				Msgf("unable to upgrade downstream connection, cause: %v", rce)
		}
		msg := fmt.Sprintf(dwnConWsFail, dwnErr)
		proxy.respondWith(400, msg)
		proxy.scaffoldWebsocketLog(log.Warn()).
			Int16(dwnResCode, 400).
			Msg(msg)
		//gobwas/ws has sent a HTTP 426 across the hijacked connection already
		return
	} else {
		proxy.scaffoldWebsocketLog(log.Info()).Msg(dwnConUpgraded)
	}

	go readDwnWebsocket(dwnCon, upCon, proxy, status, tx)
	go readUpWebsocket(dwnCon, upCon, proxy, status, tx)
	_ = <-status
}

func readDwnWebsocket(dwnCon net.Conn, upCon net.Conn, proxy *Proxy, status chan<- WebsocketStatus, tx *WebsocketTx) {
ReadDwn:
	for {
		msg, op, dre := wsutil.ReadClientData(dwnCon)
		if dre == nil {
			lm := int64(len(msg))
			tx.DwnBytesRead += lm

			uwe := wsutil.WriteClientMessage(upCon, op, msg)
			if uwe == nil {
				tx.UpBytesWrite += lm
				proxy.scaffoldWebsocketLog(log.Trace()).
					Int8(opCode, int8(op)).
					Int64(msgBytes, lm).
					Msgf(upBytesWritten, lm)
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
			if _, closed := dre.(wsutil.ClosedError); closed {
				status <- WebsocketStatus{
					DwnExit: dre,
				}
			} else if io.EOF == dre {
				status <- WebsocketStatus{
					DwnExit: dre,
				}
			} else if _, netop := dre.(*net.OpError); netop {
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

func readUpWebsocket(dwnCon net.Conn, upCon net.Conn, proxy *Proxy, status chan<- WebsocketStatus, tx *WebsocketTx) {
ReadUp:
	for {
		msg, op, ure := wsutil.ReadServerData(upCon)
		if ure == nil {
			lm := int64(len(msg))
			tx.UpBytesRead += lm
			dwe := wsutil.WriteServerMessage(dwnCon, op, msg)
			if dwe == nil {
				tx.DwnBytesWrite += lm
				proxy.scaffoldWebsocketLog(log.Trace()).
					Int8(opCode, int8(op)).
					Int64(msgBytes, lm).
					Msgf(dwnBytesWritten, lm)
			} else {
				//TODO this doesn't cover everything readDwnWebsocket does. it needs the same
				//error handling but we don't want code duplication.
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
			if _, closed := ure.(wsutil.ClosedError); closed {
				status <- WebsocketStatus{
					UpExit: ure,
				}
			} else if io.EOF == ure {
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

