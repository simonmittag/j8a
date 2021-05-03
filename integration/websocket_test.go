package integration

import (
	"context"
	"github.com/simonmittag/ws"
	"github.com/simonmittag/ws/wsutil"
	"testing"
)

func TestWSConnectionEstablishedAndEchoMessageWithCleanExit(t *testing.T) {
	con, _, _, e := ws.DefaultDialer.Dial(context.Background(), "ws://localhost:8080/websocket?n=1")
	if e != nil {
		t.Errorf("unable to connect to ws, cause: %v", e)
	}

	want := []byte("hello world")
	e2 := wsutil.WriteClientMessage(con, ws.OpText, want)
	if e2 != nil {
		t.Errorf("unable to write ws message, cause: %v", e2)
	}

	msg, op, e3 := wsutil.ReadServerData(con)
	if op != ws.OpText {
		t.Errorf("opCode should be text, was: %v", op)
	}
	t.Logf("normal. received message: %v", string(msg))
	if e3 != nil {
		t.Errorf("unable to read back ws echo message, cause: %v", e3)
	}

	//so this is how you orderly close a WS connection in gobwas.
	cf := ws.NewCloseFrame(ws.NewCloseFrameBody(
		ws.StatusNormalClosure, "unit test close requested",
	))
	cf = ws.MaskFrameInPlace(cf)
	e4 := ws.WriteFrame(con, cf)
	if e4 != nil {
		t.Errorf("unable to close ws protocol connection, cause: %v", e4)
	}

	e5 := con.Close()
	if e5 != nil {
		t.Errorf("unable to TCP socket connection, cause: %v", e5)
	}

}
