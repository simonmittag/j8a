package integration

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestHeaderOrderRewriteDownstreamToUpstream(t *testing.T) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://localhost:8080/mse6/echoheader", nil)
	req.Header.Add("X", "1")
	req.Header.Add("X", "2")
	req.Header.Add("X", "3")
	req.Header.Add("X", "4")
	req.Header.Add("X", "5")
	req.Header.Add("X", "6")
	req.Header.Add("X", "7")
	req.Header.Add("X", "8")
	req.Header.Add("X", "9")
	req.Header.Add("X", "10")
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("error connecting to server, cause: %s", err)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	body, _ := ioutil.ReadAll(resp.Body)
	utf8 := string(body)
	if !strings.Contains(utf8, "X:[1 2 3 4 5 6 7 8 9 10]") {
		t.Errorf("should have sent X headers upstream in order, but sent this %s", body)
	}
}
