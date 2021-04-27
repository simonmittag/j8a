package j8a

import (
	"context"
	"fmt"
	"github.com/hako/durafmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/simonmittag/ws"
	"github.com/simonmittag/ws/wsutil"
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
const upBytesWritten = "upstream websocket %d bytes written"

const dwnConClosed = "downstream websocket connection closed after %s"
const dwnConUpgraded = "downstream upgraded to websocket connection"
const dwnConWsFail = "downstream connection closed, failed websocket upgrade, cause: %s"
const dwnReadErr = "error reading from downstream websocket, cause: "
const dwnWriteErr = "error writing to downstream websocket, cause: "
const dwnBytesWritten = "downstream websocket %d bytes written"

const opCode = "opCode"
const msgBytes = "msgBytes"

type WebsocketStatus struct {
	DwnOpCode ws.OpCode
	DwnExit   error
	DwnErr    error
	UpOpCode  ws.OpCode
	UpExit    error
	UpErr     error
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

//use elapsed to pass zero or *one* time exactly
func (proxy *Proxy) scaffoldWebsocketLog(e *zerolog.Event, elapsed ...int64) *zerolog.Event {
	e.Str(XRequestID, proxy.XRequestID).
		Str(dwnReqRemoteAddr, proxy.Dwn.Req.RemoteAddr)

	if len(elapsed) > 0 {
		e.Int64(dwnElapsedMicros, elapsed[0])
	} else {
		e.Int64(dwnElapsedMicros, time.Since(proxy.Dwn.startDate).Microseconds())
	}

	return e.Str(dwnReqUserAgent, proxy.Dwn.UserAgent).
		Str(dwnReqHttpVer, proxy.Dwn.HttpVer).
		Str(dwnReqPath, proxy.Dwn.Path).
		Str(upReqURI, proxy.resolveUpstreamURI())
}

const upWebsocketConnectionFailed = "upstream websocket connection failed"
const upWebsocketUnspecifiedNetworkEvent = "upstream websocket unspecified network event: %s"
const webSocketTimeout = " websocket connection idle timeout fired after %d seconds"
const upWebsocketTimeoutFired = "upstream" + webSocketTimeout
const dwnWebsocketTimeoutFired = "downstream" + webSocketTimeout
const webSocketHangup = " websocket connection hung up TCP socket on us by remote end"
const upWebSocketHangup = "upstream" + webSocketHangup
const dwnWebSocketHangup = "downstream" + webSocketHangup
const webSocketClosed = " websocket connection close requested by remote end"
const upWebSocketClosed = "upstream" + webSocketClosed
const dwnWebSocketClosed = "downstream" + webSocketClosed

const connect = "connect"
const iotimeout = "i/o timeout"

func upgradeWebsocket(proxy *Proxy) {
	var status = make(chan WebsocketStatus)
	var tx *WebsocketTx = &WebsocketTx{}

	//upCon has to run first. if it fails we still want to send a 50x HTTP response from within j8a.
	upCon, _, _, upErr := ws.DefaultDialer.Dial(context.Background(), proxy.resolveUpstreamURI())
	defer func() {
		if upCon != nil {
			ws.WriteFrame(upCon, ws.NewCloseFrame(ws.NewCloseFrameBody(ws.StatusNormalClosure, "")))

			//after sending close frame we are not expected to process any other frames and tear down socket.
			//See: https://tools.ietf.org/html/rfc6455#section-5.5.1
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

	dwnCon, _, _, dwnErr := scaffoldHTTPUpgrader(proxy).Upgrade(proxy.Dwn.Req, proxy.Dwn.Resp.Writer)
	defer func() {
		if dwnCon != nil && dwnErr == nil {
			ws.WriteFrame(dwnCon, ws.NewCloseFrame(ws.NewCloseFrameBody(ws.StatusNormalClosure, "")))

			//after sending close frame we are not expected to process any other frames and tear down socket.
			//See: https://tools.ietf.org/html/rfc6455#section-5.5.1
			dwnCon.Close()

			elapsed := time.Since(proxy.Dwn.startDate)
			ev := proxy.scaffoldWebsocketLog(log.Info(), elapsed.Microseconds())
			ev.Int64(dwnBytesRead, tx.DwnBytesRead).
				Int64(dwnBytesWrite, tx.DwnBytesWrite).
				Msgf(dwnConClosed, durafmt.Parse(elapsed).LimitFirstN(2).String())
		}
	}()

	if dwnErr != nil {
		msg := fmt.Sprintf(dwnConWsFail, dwnErr)
		rce, rcet := dwnErr.(*ws.RejectConnectionErrorType)
		ev := proxy.scaffoldWebsocketLog(log.Warn())
		//below only logs response code it was already send by simonmittag/ws on hijacked conn.
		if rcet {
			proxy.respondWith(rce.Code(), msg)
			ev.Int(dwnResCode, rce.Code())
		}
		ev.Msg(msg)

		return
	} else {
		proxy.scaffoldWebsocketLog(log.Info()).Msg(dwnConUpgraded)
	}

	go readDwnWebsocket(dwnCon, upCon, proxy, status, tx)
	go readUpWebsocket(dwnCon, upCon, proxy, status, tx)

	proxy.logWebsocketConnectionExitStatus(<-status)
}

const EOF = "EOF"

func (proxy *Proxy) logWebsocketConnectionExitStatus(conStat WebsocketStatus) {
	isTimeout := func(err error) bool {
		noe, noet := err.(*net.OpError)
		return noet && noe.Err != nil && noe.Err.Error() == iotimeout
	}
	isHangup := func(err error) bool {
		return err != nil && err.Error() == EOF
	}
	isCloseRequested := func(err error) bool {
		ce, cet := err.(wsutil.ClosedError)
		return cet && ce.Code == 1000
	}

	ev := proxy.scaffoldWebsocketLog(log.Trace())
	if conStat.UpExit != nil {
		if isTimeout(conStat.UpExit) {
			ev.Msgf(upWebsocketTimeoutFired, Runner.Connection.Upstream.IdleTimeoutSeconds)
		} else if isHangup(conStat.UpExit) {
			ev.Msg(upWebSocketHangup)
		} else if isCloseRequested(conStat.UpExit) {
			ev.Msg(upWebSocketClosed)
		} else {
			ev.Msg(conStat.UpExit.Error())
		}
	}
	if conStat.DwnExit != nil {
		if isTimeout(conStat.DwnExit) {
			ev.Msgf(dwnWebsocketTimeoutFired, Runner.Connection.Downstream.IdleTimeoutSeconds)
		} else if isHangup(conStat.DwnExit) {
			ev.Msg(dwnWebSocketHangup)
		} else if isCloseRequested(conStat.DwnExit) {
			ev.Msg(dwnWebSocketClosed)
		} else {
			ev.Msg(conStat.DwnExit.Error())
		}
	}
}

func scaffoldHTTPUpgrader(proxy *Proxy) ws.HTTPUpgrader {
	var h = make(map[string][]string)
	h[Server] = []string{serverVersion()}
	h[XRequestID] = []string{proxy.XRequestID}
	if Runner.isTLSOn() {
		h[strictTransportSecurity] = []string{maxAge31536000}
	}

	upg := ws.HTTPUpgrader{
		Timeout: time.Second * time.Duration(Runner.Connection.Downstream.ReadTimeoutSeconds),
		Header:  h,
	}
	return upg
}

func readDwnWebsocket(dwnCon net.Conn, upCon net.Conn, proxy *Proxy, status chan<- WebsocketStatus, tx *WebsocketTx) {
ReadDwn:
	for {
		dwnCon.SetDeadline(time.Now().Add(time.Second * time.Duration(Runner.Connection.Downstream.IdleTimeoutSeconds)))
		msg, op, dre := wsutil.ReadClientData(dwnCon)
		if dre == nil {
			lm := int64(len(msg))
			tx.DwnBytesRead += lm

			upCon.SetDeadline(time.Now().Add(time.Second * time.Duration(Runner.Connection.Upstream.IdleTimeoutSeconds)))
			uwe := wsutil.WriteClientMessage(upCon, op, msg)
			if uwe == nil {
				tx.UpBytesWrite += lm
				proxy.scaffoldWebsocketLog(log.Trace()).
					Int8(opCode, int8(op)).
					Int64(msgBytes, lm).
					Msgf(upBytesWritten, lm)
			} else {
				if !isExit(uwe) {
					proxy.scaffoldWebsocketLog(log.Warn()).
						Int8(opCode, int8(op)).
						Msg(upWriteErr + uwe.Error())
				}
				status <- WebsocketStatus{UpExit: uwe, UpOpCode: op}
				break ReadDwn
			}
		} else {
			if !isExit(dre) {
				proxy.scaffoldWebsocketLog(log.Warn()).
					Int8(opCode, int8(op)).
					Msg(dwnReadErr + dre.Error())
			}
			status <- WebsocketStatus{DwnExit: dre, DwnOpCode: op}
			break ReadDwn
		}
	}
}

func readUpWebsocket(dwnCon net.Conn, upCon net.Conn, proxy *Proxy, status chan<- WebsocketStatus, tx *WebsocketTx) {
ReadUp:
	for {
		upCon.SetDeadline(time.Now().Add(time.Second * time.Duration(Runner.Connection.Upstream.IdleTimeoutSeconds)))
		msg, op, ure := wsutil.ReadServerData(upCon)
		if ure == nil {
			lm := int64(len(msg))
			tx.UpBytesRead += lm

			//we must set both deadlines inside the loop to keep updating timeouts
			dwnCon.SetDeadline(time.Now().Add(time.Second * time.Duration(Runner.Connection.Downstream.IdleTimeoutSeconds)))
			dwe := wsutil.WriteServerMessage(dwnCon, op, msg)
			if dwe == nil {
				tx.DwnBytesWrite += lm
				proxy.scaffoldWebsocketLog(log.Trace()).
					Int8(opCode, int8(op)).
					Int64(msgBytes, lm).
					Msgf(dwnBytesWritten, lm)
			} else {
				if !isExit(dwe) {
					proxy.scaffoldWebsocketLog(log.Warn()).
						Int8(opCode, int8(op)).
						Msg(dwnWriteErr + dwe.Error())
				}
				status <- WebsocketStatus{DwnExit: dwe, DwnOpCode: op}
				break ReadUp
			}
		} else {
			if !isExit(ure) {
				proxy.scaffoldWebsocketLog(log.Warn()).
					Int8(opCode, int8(op)).
					Msg(upReadErr + ure.Error())
			}
			status <- WebsocketStatus{UpExit: ure, UpOpCode: op}
			break ReadUp
		}
	}
}

func isExit(err error) bool {
	_, closed := err.(wsutil.ClosedError)
	_, netop := err.(*net.OpError)
	return closed || netop || err == io.EOF
}
