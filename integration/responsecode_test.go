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

func TestStatusCode300fProxiedResponse(t *testing.T) {
	performOneJabbaResponseCodeTest(t, 300, 300, 8080)
}

func TestStatusCode301OfProxiedResponse(t *testing.T) {
	performOneJabbaResponseCodeTest(t, 301, 301, 8080)
}

func TestStatusCode302OfProxiedResponse(t *testing.T) {
	performOneJabbaResponseCodeTest(t, 302, 302, 8080)
}

func TestStatusCode303OfProxiedResponse(t *testing.T) {
	performOneJabbaResponseCodeTest(t, 303, 303, 8080)
}

func TestStatusCode216OfProxiedResponse(t *testing.T) {
	performOneJabbaResponseCodeTest(t, 216, 216, 8080)
}

func performOneJabbaResponseCodeTest(t *testing.T, getUpstreamStatusCode int, wantDownstreamStatusCode int, serverPort int) {
	var wg sync.WaitGroup
	wg.Add(1)
	performJabbaResponseCodeTest(&wg, t, getUpstreamStatusCode, wantDownstreamStatusCode, serverPort)
	wg.Wait()
}

func performJabbaResponseCodeTest(wg *sync.WaitGroup, t *testing.T, getUpstreamStatusCode int, wantDownstreamStatusCode int, serverPort int) {
	//for multithreaded tests we need to count them all down
	defer wg.Done()

	//test client do not follow redirects mate!
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(fmt.Sprintf("http://localhost:%d/mse6/send?code=%d", serverPort, getUpstreamStatusCode))
	if resp != nil && resp.Body != nil {
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
