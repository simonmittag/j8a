package j8a

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"math"
	"time"
)

type State string

const (
	Bootstrap State = "Bootstrap"
	Daemon    State = "Daemon"
	Shutdown  State = "Shutdown"
)

func (s State) Lesser(t State) bool {
	if s == Bootstrap && (t == Daemon || t == Shutdown) {
		return true
	}
	if s == Daemon && t == Shutdown {
		return true
	}
	return false
}

type StateHandler struct {
	Current State
	Update  chan State
}

func NewStateHandler() *StateHandler {
	return &StateHandler{
		Current: Bootstrap,
		Update:  make(chan State),
	}
}

func (sh *StateHandler) waitState(s State, timeoutSeconds ...int) {
	if s == sh.Current || s.Lesser(sh.Current) {
		return
	} else {
		to := time.Duration(math.MaxInt64)
		if len(timeoutSeconds) > 0 {
			to = time.Second * time.Duration(timeoutSeconds[0])
		}
		for {
			select {
			case ev := <-sh.Update:
				if s == ev || s.Lesser(ev) {
					return
				}
			case <-time.After(to):
				return
			}
		}
	}
}

func (sh *StateHandler) setState(s State) {
	// == matters because we may want to retrigger the state for waiting goroutines.
	if sh.Current == s || sh.Current.Lesser(s) {
		sh.Current = s

		msg := fmt.Sprintf("server state now %v", sh.Current)
		log.Info().Msg(msg)

		//needs to be async else setState blocks
		go func() {
			sh.Update <- s
		}()
	}
}
