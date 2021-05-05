package integration

import (
	"bytes"
	"context"
	"github.com/simonmittag/ws"
	"github.com/simonmittag/ws/wsutil"
	"net"
	"testing"
)

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
		t.Errorf("unable to TCP socket connection, cause: %v", e5)
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
