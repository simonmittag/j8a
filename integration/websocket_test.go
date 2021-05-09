package integration

import (
	"bytes"
	"context"
	"crypto/tls"
	"github.com/simonmittag/ws"
	"github.com/simonmittag/ws/wsutil"
	"net"
	"strings"
	"testing"
	"time"
)

const wsse = "unexpected HTTP response status: "

func TestWSConnectionEstablishedAndEchoMessageWithDownstreamExitClean(t *testing.T) {
	con, _, _, e := ws.DefaultDialer.Dial(context.Background(), "ws://localhost:8080/websocket?n=1")
	if e != nil {
		t.Errorf("unable to connect to ws, cause: %v", e)
		return
	}

	if !echoHelloWorld(t, con) {
		return
	}

	//so this is how you orderly close a WS connection in gobwas.
	cf := ws.NewCloseFrame(ws.NewCloseFrameBody(
		ws.StatusNormalClosure, "unit test close requested",
	))
	cf = ws.MaskFrameInPlace(cf)
	e4 := ws.WriteFrame(con, cf)
	if e4 != nil {
		t.Errorf("unable to close ws protocol connection, cause: %v", e4)
		return
	}

	e5 := con.Close()
	if e5 != nil {
		t.Errorf("unable to close TCP socket connection, cause: %v", e5)
	}

}

func Test101ResponseForWsUpgrade(t *testing.T) {
	c, err := net.Dial("tcp", ":8080")
	if err != nil {
		t.Errorf("test failure. unable to connect to j8a server for integration test, cause: %v", err)
		return
	}
	defer c.Close()

	//step 2 we send headers and terminate HTTP message.
	checkWrite(t, c, "GET /websocket HTTP/1.1\r\n")
	checkWrite(t, c, "Host: localhost:8080\r\n")
	checkWrite(t, c, "User-Agent: integration\r\n")
	checkWrite(t, c, "Origin: http://localhost\r\n")
	checkWrite(t, c, "Content-Type: application/json\r\n")
	checkWrite(t, c, "Connection: Upgrade\r\n")
	checkWrite(t, c, "Upgrade: websocket\r\n")
	checkWrite(t, c, "Sec-WebSocket-Key: x3JJHMbDL1EzLkh9GBhXDw==\r\n")
	checkWrite(t, c, "Sec-WebSocket-Version: 13\r\n")
	checkWrite(t, c, "\r\n")

	buf := make([]byte, 1024)
	l, err := c.Read(buf)
	t.Logf("normal. j8a responded with %v bytes and error code %v", l, err)
	t.Logf("normal. j8a partial response: %v", string(buf))
	if l == 0 {
		t.Error("test failure. j8a did not respond, 0 bytes read")
	}
	response := string(buf)
	if !strings.Contains(response, "Server: j8a") {
		t.Error("test failure. j8a did not respond, server information not found on response ")
	}
	if !strings.Contains(response, "101 Switching Protocols") {
		t.Error("test failure. j8a did not upgrade to websocket connection ")
	}
}

func Test400ResponseForBadSecHeader(t *testing.T) {
	c, err := net.Dial("tcp", ":8080")
	if err != nil {
		t.Errorf("test failure. unable to connect to j8a server for integration test, cause: %v", err)
		return
	}
	defer c.Close()

	//step 2 we send headers and terminate HTTP message.
	checkWrite(t, c, "GET /websocket HTTP/1.1\r\n")
	checkWrite(t, c, "Host: localhost:8080\r\n")
	checkWrite(t, c, "User-Agent: integration\r\n")
	checkWrite(t, c, "Origin: http://localhost\r\n")
	checkWrite(t, c, "Content-Type: application/json\r\n")
	checkWrite(t, c, "Connection: Upgrade\r\n")
	checkWrite(t, c, "Upgrade: websocket\r\n")
	checkWrite(t, c, "Sec-WebSocket-Key: XXXXXXXXXXXXXXXXXXXXXX\r\n")
	checkWrite(t, c, "Sec-WebSocket-Version: 13\r\n")
	checkWrite(t, c, "\r\n")

	buf := make([]byte, 1024)
	l, err := c.Read(buf)
	t.Logf("normal. j8a responded with %v bytes and error code %v", l, err)
	t.Logf("normal. j8a partial response: %v", string(buf))
	if l == 0 {
		t.Error("test failure. j8a did not respond, 0 bytes read")
	}
	response := string(buf)
	if !strings.Contains(response, "400 Bad Request") {
		t.Error("test failure. j8a should have responded with bad request ")
	}
}

func Test404ResponseUpstreamURLIsNotMapped(t *testing.T) {
	_, _, _, e := ws.DefaultDialer.Dial(context.Background(), "ws://localhost:8080/nourl")
	if e == nil {
		t.Errorf("bad url should return an error but did not: %v", e)
		return
	} else {
		if wse, ok := e.(ws.StatusError); ok {
			if wse.Error() != wsse+"404" {
				t.Errorf("j8a should return 404 for not found but got %s", wse.Error())
			}
		}
	}
}

func Test405ResponseForInvalidHTTPMethod(t *testing.T) {
	c, err := net.Dial("tcp", ":8080")
	if err != nil {
		t.Errorf("test failure. unable to connect to j8a server for integration test, cause: %v", err)
		return
	}
	defer c.Close()

	//step 2 we send headers and terminate HTTP message.
	checkWrite(t, c, "PUT /websocket HTTP/1.1\r\n")
	checkWrite(t, c, "Host: localhost:8080\r\n")
	checkWrite(t, c, "User-Agent: integration\r\n")
	checkWrite(t, c, "Origin: http://localhost\r\n")
	checkWrite(t, c, "Content-Type: application/json\r\n")
	checkWrite(t, c, "Connection: Upgrade\r\n")
	checkWrite(t, c, "Upgrade: websocket\r\n")
	checkWrite(t, c, "Sec-WebSocket-Key: x3JJHMbDL1EzLkh9GBhXDw==\r\n")
	checkWrite(t, c, "Sec-WebSocket-Version: 13\r\n")
	checkWrite(t, c, "\r\n")

	buf := make([]byte, 1024)
	l, err := c.Read(buf)
	t.Logf("normal. j8a responded with %v bytes and error code %v", l, err)
	t.Logf("normal. j8a partial response: %v", string(buf))
	if l == 0 {
		t.Error("test failure. j8a did not respond, 0 bytes read")
	}
	response := string(buf)
	if !strings.Contains(response, "405 Method Not Allowed") {
		t.Error("test failure. j8a should have responded with bad protocol version ")
	}
}

func Test426ResponseForBadWsProtocolVersion(t *testing.T) {
	c, err := net.Dial("tcp", ":8080")
	if err != nil {
		t.Errorf("test failure. unable to connect to j8a server for integration test, cause: %v", err)
		return
	}
	defer c.Close()

	//step 2 we send headers and terminate HTTP message.
	checkWrite(t, c, "GET /websocket HTTP/1.1\r\n")
	checkWrite(t, c, "Host: localhost:8080\r\n")
	checkWrite(t, c, "User-Agent: integration\r\n")
	checkWrite(t, c, "Origin: http://localhost\r\n")
	checkWrite(t, c, "Content-Type: application/json\r\n")
	checkWrite(t, c, "Connection: Upgrade\r\n")
	checkWrite(t, c, "Upgrade: websocket\r\n")
	checkWrite(t, c, "Sec-WebSocket-Key: x3JJHMbDL1EzLkh9GBhXDw==\r\n")
	checkWrite(t, c, "Sec-WebSocket-Version: XX\r\n")
	checkWrite(t, c, "\r\n")

	buf := make([]byte, 1024)
	l, err := c.Read(buf)
	t.Logf("normal. j8a responded with %v bytes and error code %v", l, err)
	t.Logf("normal. j8a partial response: %v", string(buf))
	if l == 0 {
		t.Error("test failure. j8a did not respond, 0 bytes read")
	}
	response := string(buf)
	if !strings.Contains(response, "426 Upgrade Required") {
		t.Error("test failure. j8a should have responded with bad protocol version ")
	}
}

func Test400ResponseForBadHTTPProtocolVersion(t *testing.T) {
	c, err := net.Dial("tcp", ":8080")
	if err != nil {
		t.Errorf("test failure. unable to connect to j8a server for integration test, cause: %v", err)
		return
	}
	defer c.Close()

	//step 2 we send headers and terminate HTTP message.
	checkWrite(t, c, "GET /websocket HTTP/2\r\n")
	checkWrite(t, c, "Host: localhost:8080\r\n")
	c.Write([]byte("User-Agent: integration\r\n"))
	c.Write([]byte("Origin: http://localhost\r\n"))
	c.Write([]byte("Content-Type: application/json\r\n"))
	c.Write([]byte("Connection: Upgrade\r\n"))
	c.Write([]byte("Upgrade: websocket\r\n"))
	c.Write([]byte("Sec-WebSocket-Key: x3JJHMbDL1EzLkh9GBhXDw==\r\n"))
	c.Write([]byte("Sec-WebSocket-Version: 13\r\n"))
	c.Write([]byte("\r\n"))

	buf := make([]byte, 1024)
	l, err := c.Read(buf)
	t.Logf("normal. j8a responded with %v bytes and error code %v", l, err)
	t.Logf("normal. j8a partial response: %v", string(buf))
	if l == 0 {
		t.Error("test failure. j8a did not respond, 0 bytes read")
	}
	response := string(buf)
	if !strings.Contains(response, "400 Bad Request") {
		t.Error("test failure. j8a should have responded with bad request ")
	}
}

func Test502ResponseUpstreamURLisUnavailable(t *testing.T) {
	_, _, _, e := ws.DefaultDialer.Dial(context.Background(), "ws://localhost:8080/websocketdown")
	if e == nil {
		t.Errorf("bad url should return an error but did not: %v", e)
		return
	} else {
		if wse, ok := e.(ws.StatusError); ok {
			if wse.Error() != wsse+"502" {
				t.Errorf("j8a should return 502 for not found but got %s", wse.Error())
			}
		}
	}
}

func Test502ResponseUpstreamURLisUnavailableAfterLongSocketTimeout(t *testing.T) {
	start := time.Now()

	dialer := ws.Dialer{
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	t.Log("begin dialling j8a")
	_, _, _, e := dialer.Dial(context.Background(), "wss://localhost:8443/websocketdown")
	t.Log("j8a returns")
	elapsed := time.Since(start)
	want := 10
	if elapsed > time.Duration(want)*time.Second {
		t.Errorf("upstream socket connection timeout exceed, want: %vs, got: %vs", want, elapsed)
	}

	if e == nil {
		t.Errorf("bad url should return an error but did not: %v", e)
		return
	} else {
		if wse, ok := e.(ws.StatusError); ok {
			if wse.Error() != wsse+"502" {
				t.Errorf("j8a should return 502 for not found but got %s", wse.Error())
			}
		}
	}
}

func TestWSConnectionEstablishedAndEchoMessageWithDownstreamExitDirtyClosingJustProtocol(t *testing.T) {
	con, _, _, e := ws.DefaultDialer.Dial(context.Background(), "ws://localhost:8080/websocket?n=1")
	if e != nil {
		t.Errorf("unable to connect to ws, cause: %v", e)
		return
	}

	if !echoHelloWorld(t, con) {
		return
	}

	//so this is how you orderly close a WS connection in gobwas.
	cf := ws.NewCloseFrame(ws.NewCloseFrameBody(
		ws.StatusNormalClosure, "unit test close requested",
	))
	cf = ws.MaskFrameInPlace(cf)
	e4 := ws.WriteFrame(con, cf)
	if e4 != nil {
		t.Errorf("unable to close ws protocol connection, cause: %v", e4)
		return
	}

	//in this case j8a closes the network connection after we forget to do so and we still should be unable to read any further.
	msg, _, e6 := wsutil.ReadServerData(con)
	if e6 == nil {
		t.Errorf("error. should get read error after connection closed but none received, instead got msg: %v", msg)
	} else {
		if wce, wcet := e6.(wsutil.ClosedError); !wcet {
			t.Errorf("error. we expect netop error here but got %s", wce)
		} else {
			t.Logf("normal. j8a closed connection with %s", e6)
		}
	}
}

func TestWSConnectionEstablishedAndEchoMessageWithDownstreamExitDirtyClosingJustSocket(t *testing.T) {
	con, _, _, e := ws.DefaultDialer.Dial(context.Background(), "ws://localhost:8080/websocket?n=1")
	if e != nil {
		t.Errorf("error. unable to connect to ws, cause: %v", e)
		return
	}

	if !echoHelloWorld(t, con) {
		return
	}

	e5 := con.Close()
	if e5 != nil {
		t.Errorf("error. unable to close TCP socket connection, cause: %v", e5)
	}

	msg, _, e6 := wsutil.ReadServerData(con)
	if e6 == nil {
		t.Errorf("error. should get read error after connection closed but none received, instead got msg: %v", msg)
	} else {
		if netop, netopt := e6.(*net.OpError); !netopt {
			t.Errorf("error. we expect netop error here but got %s", netop)
		} else {
			t.Logf("normal. j8a closed connection with %s", e6)
		}
	}
}

func TestWSConnectionEstablishedThenUpstreamTimeout(t *testing.T) {
	con, _, _, e := ws.DefaultDialer.Dial(context.Background(), "ws://localhost:8080/websocket?n=1")
	if e != nil {
		t.Errorf("unable to connect to ws, cause: %v", e)
		return
	} else {
		t.Log("normal. established connection")
	}

	if !echoHelloWorld(t, con) {
		return
	}

	t.Log("normal. sleeping for 10 seconds")
	time.Sleep(time.Second * 11)
	t.Log("normal. waking up")

	_, _, e4 := wsutil.ReadServerData(con)
	if e4 == nil {
		t.Errorf("error. upstream should have closed connection, but was nil err")
	} else {
		if wce, wcet := e4.(wsutil.ClosedError); !wcet {
			t.Errorf("error. j8a should have closed normal, but returned %s", wce)
		}
		t.Logf("normal. j8a closed connection with %s", e4)
	}
}

func TestWSSConnectionEstablishedThenDownstreamTimeout(t *testing.T) {
	dialer := ws.Dialer{
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	con, _, _, e := dialer.Dial(context.Background(), "wss://localhost:8443/websocket?n=1")
	if e != nil {
		t.Errorf("unable to connect to ws, cause: %v", e)
		return
	} else {
		t.Log("normal. established connection")
	}

	if !echoHelloWorld(t, con) {
		return
	}

	t.Log("normal. sleeping for 31 seconds")
	time.Sleep(time.Second * 31)
	t.Log("normal. waking up")

	_, _, e4 := wsutil.ReadServerData(con)
	if e4 == nil {
		t.Errorf("error. upstream should have closed connection, but was nil err")
	} else {
		t.Logf("normal. j8a closed tls connection with %s", e4)
	}
}

func TestWSConnectionEstablishedAndEchoMessageWithUpstreamExitClean(t *testing.T) {
	con, _, _, e := ws.DefaultDialer.Dial(context.Background(), "ws://localhost:8080/websocket?n=1&c")
	if e != nil {
		t.Errorf("unable to connect to ws, cause: %v", e)
		return
	}

	if !echoHelloWorld(t, con) {
		return
	}

	_, _, e4 := wsutil.ReadServerData(con)
	if e4 == nil {
		t.Errorf("upstream should have closed connection, but was nil err")
	} else {
		if wce, wcet := e4.(wsutil.ClosedError); !wcet {
			t.Errorf("j8a should have closed normal, but returned %s", wce)
		}
		t.Logf("normal. j8a closed connection with %s", e4)
	}
}

func TestWSConnectionEstablishedAndEchoMessageWithUpstreamExitProtocolOnly(t *testing.T) {
	con, _, _, e := ws.DefaultDialer.Dial(context.Background(), "ws://localhost:8080/websocket?n=1&c1")
	if e != nil {
		t.Errorf("unable to connect to ws, cause: %v", e)
		return
	}

	if !echoHelloWorld(t, con) {
		return
	}

	_, _, e4 := wsutil.ReadServerData(con)
	if e4 == nil {
		t.Errorf("upstream should have closed connection, but was nil err")
	} else {
		if wce, wcet := e4.(wsutil.ClosedError); !wcet {
			t.Errorf("j8a should have closed normal, but returned %s", wce)
		}
		t.Logf("normal. j8a closed connection with %s", e4)
	}
}

func TestWSConnectionEstablishedAndEchoMessageWithUpstreamExitSocketOnly(t *testing.T) {
	con, _, _, e := ws.DefaultDialer.Dial(context.Background(), "ws://localhost:8080/websocket?n=1&c2")
	if e != nil {
		t.Errorf("unable to connect to ws, cause: %v", e)
		return
	}

	if !echoHelloWorld(t, con) {
		return
	}

	_, _, e4 := wsutil.ReadServerData(con)
	if e4 == nil {
		t.Errorf("upstream should have closed connection, but was nil err")
	} else {
		if wce, wcet := e4.(wsutil.ClosedError); !wcet {
			t.Errorf("j8a should have closed normal, but returned %s", wce)
		}
		t.Logf("normal. j8a closed connection with %s", e4)
	}
}

func TestWSSConnectionUpstreamSucceedsTLSInsecureSkipVerifyOn(t *testing.T) {
	con, _, _, e := ws.DefaultDialer.Dial(context.Background(), "ws://localhost:8080/websocketsecure/websocket?n=1&c")
	if e != nil {
		t.Errorf("unable to connect to ws, cause: %v", e)
		return
	}

	if !echoHelloWorld(t, con) {
		return
	}

	_, _, e4 := wsutil.ReadServerData(con)
	if e4 == nil {
		t.Errorf("upstream should have closed connection, but was nil err")
	} else {
		if wce, wcet := e4.(wsutil.ClosedError); !wcet {
			t.Errorf("j8a should have closed normal, but returned %s", wce)
		}
		t.Logf("normal. j8a closed connection with %s", e4)
	}
}

func TestWSSConnectionUpstreamFailsTLSInsecureSkipVerifyOff(t *testing.T) {
	_, _, _, e := ws.DefaultDialer.Dial(context.Background(), "ws://localhost:8081/websocketsecure/websocket?n=1&c")
	if e == nil {
		t.Errorf("connection error should be 502 but was nil")
	} else {
		e1, wsset := e.(ws.StatusError)
		if !wsset {
			t.Errorf("should have received websocket status error, but got: %v", e1)
		} else if e1.Error() == wsse+"502" {
			t.Logf("normal. received status error %v", e)
		} else {
			t.Errorf("error. should have received 502 from remote but got: %v", e1)
		}
	}
}

func echoHelloWorld(t *testing.T, con net.Conn) bool {
	want := []byte("hello world")
	e2 := wsutil.WriteClientMessage(con, ws.OpText, want)
	if e2 != nil {
		t.Errorf("unable to write ws message, cause: %v", e2)
		return false
	}

	msg, op, e3 := wsutil.ReadServerData(con)
	if op != ws.OpText {
		t.Errorf("opCode should be text, was: %v", op)
		return false
	}

	if e3 != nil {
		t.Errorf("unable to read back ws echo message, cause: %v", e3)
		return false
	}
	if bytes.Compare(want, msg) != 0 {
		t.Errorf("sent %s, received wrong echo from server: %s", want, string(msg))
	} else {
		t.Logf("normal. sent and received message: %v", string(msg))
	}
	return true
}
