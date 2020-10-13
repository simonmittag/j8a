package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestDownstreamChunkedRequestisProxiedUpstream(t *testing.T) {
	client := &http.Client{}
	serverPort := 8080
	wantDownstreamStatusCode := 200

	jsonData := map[string]string{"firstname": "firstname", "lastname": "lastname", "rank": "general", "color": "green"}
	jsonValue, _ := json.Marshal(jsonData)
	buf := bytes.NewBuffer(jsonValue)

	url := fmt.Sprintf("http://localhost:%d/mse6/put", serverPort)
	req, _ := http.NewRequest("PUT", url, buf)
	req.Header.Set("Accept-Encoding", "identity")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Encoding", "identity")
	req.Header.Add("Transfer-Encoding", "chunked")

	resp, err := client.Do(req)
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
		t.Errorf("chunked PUT should return ok, want %d, got %d", wantDownstreamStatusCode, gotDownstreamStatusCode)
	}
}

func TestUpstreamChunkedRequestisProxiedDownstream(t *testing.T) {
	client := &http.Client{}
	serverPort := 8080
	wantDownstreamStatusCode := 200

	url := fmt.Sprintf("http://localhost:%d/mse6/chunked?wait=1", serverPort)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept-Encoding", "identity")

	resp, err := client.Do(req)
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
		t.Errorf("chunked GET upstream should return ok, want %d, got %d", wantDownstreamStatusCode, gotDownstreamStatusCode)
	}

	want2 := "119"
	got2 := resp.Header.Get("Content-Length")
	if got2 != want2 {
		t.Errorf("chunked GET upstream should return Content-Length, want %s, got %s", want2, got2)
	}

	want3 := ""
	got3 := resp.Header.Get("Transfer-Encoding")
	if got3 != want3 {
		t.Errorf("chunked GET upstream should not have Transfer-Encoding header, got %s", got3)
	}
}
