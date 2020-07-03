package jabba

import (
	"fmt"
	"testing"
	"time"
)

func TestAbortAllUpstreamAttempts(t *testing.T) {
	Runner = mockRuntime()

	want := true
	got := false

	mockAtmpt := func() Atmpt {
		return Atmpt{
			URL:            nil,
			Label:          "",
			Count:          1,
			StatusCode:     0,
			isGzip:         false,
			resp:           nil,
			respBody:       nil,
			CompleteHeader: nil,
			CompleteBody:   nil,
			Aborted:        nil,
			AbortedFlag:    false,
			CancelFunc: func() {
				fmt.Println("cancelfunc called")
				got = true
			},
			startDate: time.Now(),
		}
	}

	atmpt := mockAtmpt()

	proxy := Proxy{
		XRequestID:    "",
		XRequestDebug: false,
		Up: Up{
			Atmpts: []Atmpt{mockAtmpt()},
			Atmpt:  &atmpt,
		},
		Dwn: Down{
			Req:         nil,
			Resp:        Resp{},
			Method:      "",
			Path:        "",
			URI:         "",
			UserAgent:   "",
			Body:        nil,
			Aborted:     nil,
			AbortedFlag: false,
			startDate:   time.Now(),
		},
	}

	proxy.abortAllUpstreamAttempts()

	if want != got {
		t.Errorf("cancel func on proxy upstream attempt not triggered")
	}
}
