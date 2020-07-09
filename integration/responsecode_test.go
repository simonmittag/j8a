package integration

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
)

//Test continue / informational responses
func TestStatusCodeOfProxiedResponses100To103(t *testing.T) {
	var wg sync.WaitGroup
	for i := 100; i <= 103; i++ {
		wg.Add(1)
		go performJabbaResponseCodeTest(&wg, t, i, i, 8080)
	}
	wg.Wait()
}

//Test normal responses
func TestStatusCodeOfProxiedResponses200To226(t *testing.T) {
	var wg1 sync.WaitGroup
	for i := 200; i <= 226; i++ {
		wg1.Add(1)
		go performJabbaResponseCodeTest(&wg1, t, i, i, 8080)
	}
	wg1.Wait()
}

//Test redirects
func TestStatusCodeOfProxiedResponses300To308(t *testing.T) {
	var wg1 sync.WaitGroup
	for i := 300; i <= 308; i++ {
		wg1.Add(1)
		go performJabbaResponseCodeTest(&wg1, t, i, i, 8080)
	}
	wg1.Wait()
}

//Test client errors
func TestStatusCodeOfProxiedResponses400To499(t *testing.T) {
	var wg1 sync.WaitGroup
	for i := 400; i <= 499; i++ {
		wg1.Add(1)
		go performJabbaResponseCodeTest(&wg1, t, i, i, 8080)
	}
	wg1.Wait()
}

//Test server errors
func TestStatusCodeOfProxiedResponses500To599(t *testing.T) {
	var wg2 sync.WaitGroup
	for i := 500; i <= 599; i++ {
		wg2.Add(1)
		performJabbaResponseCodeTest(&wg2, t, i, 502, 8080)
	}
	wg2.Wait()
}

func TestStatusCode216OfProxiedResponse(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	performJabbaResponseCodeTest(&wg, t, 216, 216, 8080)
	wg.Wait()
}

func performJabbaResponseCodeTest(wg *sync.WaitGroup, t *testing.T, getUpstreamStatusCode, wantDownstreamStatusCode int, serverPort int) {
	defer wg.Done()
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/mse6/send?code=%d", serverPort, getUpstreamStatusCode))
	if resp!=nil && resp.Body!=nil {
		defer resp.Body.Close()
	}

	gotDownstreamStatusCode := 0
	if err != nil {
		t.Errorf("error connecting to upstream for port %d, /send, cause: %v", serverPort, err)
		return
	} else {
		gotDownstreamStatusCode = resp.StatusCode
	}

	if gotDownstreamStatusCode != wantDownstreamStatusCode {
		t.Errorf("bad. port %d, testMethod /send, up code %d, want dwn code %d, but got %d", serverPort,
			getUpstreamStatusCode, wantDownstreamStatusCode, gotDownstreamStatusCode)
	} else {
		t.Logf("normal. port %d, testMethod /send, up code %d, want dwn code %d, got %d", serverPort,
			getUpstreamStatusCode, wantDownstreamStatusCode, gotDownstreamStatusCode)
	}
}
