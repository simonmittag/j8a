package j8a

import (
	"net"
	"net/http"
	"sync/atomic"
)

type ConnectionWatcher struct {
	dwnOpenConns    int64
	dwnMaxOpenConns int64
	upOpenConns     int64
	upMaxOpenConns  int64
}

// OnStateChange records open connections in response to connection
// state changes. Set net/http Server.ConnState to this method
// as value.
func (cw *ConnectionWatcher) OnStateChange(conn net.Conn, state http.ConnState) {
	switch state {
	case http.StateNew:
		cw.AddDwn(1)
	case http.StateHijacked, http.StateClosed:
		cw.AddDwn(-1)
	}
	cw.UpdateMaxDwn(cw.DwnCount())
}

// Count returns the number of connections at the time
// the call.
func (cw *ConnectionWatcher) DwnCount() uint64 {
	return uint64(atomic.LoadInt64(&cw.dwnOpenConns))
}

func (cw *ConnectionWatcher) DwnMaxCount() uint64 {
	return uint64(atomic.LoadInt64(&cw.dwnMaxOpenConns))
}

// Add adds c to the number of active connections.
func (cw *ConnectionWatcher) AddDwn(c int64) {
	atomic.AddInt64(&cw.dwnOpenConns, c)
}

// Sets the maximum number of active connections observed
func (cw *ConnectionWatcher) UpdateMaxDwn(c uint64) {
	if c > cw.DwnMaxCount() {
		atomic.StoreInt64(&cw.dwnMaxOpenConns, int64(c))
	}
}

func (cw *ConnectionWatcher) UpCount() uint64 {
	return uint64(atomic.LoadInt64(&cw.upOpenConns))
}

func (cw *ConnectionWatcher) UpMaxCount() uint64 {
	return uint64(atomic.LoadInt64(&cw.upMaxOpenConns))
}

func (cw *ConnectionWatcher) AddUp(c int64) {
	atomic.AddInt64(&cw.upOpenConns, c)
}

func (cw *ConnectionWatcher) SetUp(c uint64) {
	atomic.StoreInt64(&cw.upOpenConns, int64(c))
}

func (cw *ConnectionWatcher) UpdateMaxUp(c uint64) {
	if c > cw.UpMaxCount() {
		atomic.StoreInt64(&cw.upMaxOpenConns, int64(c))
	}
}
