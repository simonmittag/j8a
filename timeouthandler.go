package jabba

import (
	"net/http"
	"time"
)

type downstreamTimeoutHandler struct {
	delegate http.Handler
	timeout  time.Duration
}

func (h *downstreamTimeoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	proxyHandled := make(chan struct{})
	timedout := make(chan struct{})

	time.AfterFunc(h.timeout, func() {

		close(timedout)
	})

	go func() {
		h.delegate.ServeHTTP(w, r)
		close(proxyHandled)
	}()

	select {
	case <-proxyHandled:
		//do nothing
	case <-timedout:
		//TODO: check proxyHandler has not already started writing response
		w.WriteHeader(504)
		w.Write([]byte("504"))
	}
}

func DownstreamTimeoutHandler(delegate http.Handler, timeout time.Duration) http.Handler {
	return &downstreamTimeoutHandler{
		delegate: delegate,
		timeout:  timeout,
	}
}
