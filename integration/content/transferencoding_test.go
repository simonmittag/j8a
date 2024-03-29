package content

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/simonmittag/j8a/integration"
	"net"
	"net/http"
	"strings"
	"testing"
)

func TestDownstreamTransferEncodingUnsupported(t *testing.T) {
	//illegal. Since go v1.16 identity is not supported
	DownstreamTransferEncodingNoBodyResponsds501Unsupported("identity", t)
	DownstreamTransferEncodingNoBodyResponsds501Unsupported("fugazi", t)
	DownstreamTransferEncodingNoBodyResponsds501Unsupported("deflate", t)
	DownstreamTransferEncodingNoBodyResponsds501Unsupported("compress", t)
	DownstreamTransferEncodingNoBodyResponsds501Unsupported("gzip", t)
}

func DownstreamTransferEncodingNoBodyResponsds501Unsupported(enc string, t *testing.T) {
	//if this test fails, check the j8a configuration for connection.downstream.ReadTimeoutSeconds
	rCode := "501"

	//step 1 make a connection
	c, err := net.Dial("tcp", ":8080")
	if err != nil {
		t.Errorf("test failure. unable to connect to j8a server for integration test, cause: %v", err)
	}

	//step 2 we send headers and terminate http message so j8a sends request upstream.
	integration.CheckWrite(t, c, "GET /mse6/get HTTP/1.1\r\n")
	integration.CheckWrite(t, c, "Host: localhost:8087\r\n")
	integration.CheckWrite(t, c, "Accept-Encoding: identity\r\n")
	integration.CheckWrite(t, c, fmt.Sprintf("Transfer-Encoding: %s\r\n", enc))
	integration.CheckWrite(t, c, "\r\n")

	//step 3 we read a response into buffer which returns 501
	buf := make([]byte, 16)
	b, err2 := c.Read(buf)
	t.Logf("normal. j8a responded with %v bytes", b)
	if err2 != nil || !strings.Contains(string(buf), rCode) {
		t.Errorf("test failure. want response code %s but got %s", rCode, string(buf))
	} else {
		t.Logf("normal. received response code %s", rCode)
	}
}

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
