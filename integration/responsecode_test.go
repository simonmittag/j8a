package integration

import (
	"fmt"
	"net/http"
	"testing"
)

func TestStatusCodeOfProxiedResponses100To200(t *testing.T) {
	for i:=100;i<200;i++ {
		performJabbaResponseCodeTest(t,i, i, 8080)
	}
}

func TestStatusCodeOfProxiedResponses200To400(t *testing.T) {
	for i:=200;i<400;i++ {
		performJabbaResponseCodeTest(t,i, i, 8080)
	}
}

func TestStatusCodeOfProxiedResponses500To599(t *testing.T) {
	for i:=500;i<600;i++ {
		performJabbaResponseCodeTest(t,i, 502, 8080)
	}
}

func performJabbaResponseCodeTest(t *testing.T, getUpstreamStatusCode, wantDownstreamStatusCode int, serverPort int) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/mse6/send?code=%d", serverPort, getUpstreamStatusCode))
	gotDownstreamStatusCode := resp.StatusCode

	//if wantDownstreamStatusCode>399&&wantDownstreamStatusCode<500 {
	//	resp
	//}

	if err != nil {
		t.Errorf("error connecting to upstream for port %d, /send, cause: %v", serverPort, err)
	}

	if gotDownstreamStatusCode != wantDownstreamStatusCode {
		t.Errorf("bad status code for port %d, testMethod /send, want statusCode %d, got %d", serverPort, wantDownstreamStatusCode, gotDownstreamStatusCode)
	} else {
		t.Logf("normal status code %d for port %d, testMethod /send, upstream statusCode %d, want downstream statusCode %d", gotDownstreamStatusCode, serverPort, getUpstreamStatusCode, wantDownstreamStatusCode)
	}
}
