package integration

import (
	"fmt"
	"net/http"
	"testing"
)

func TestPathTransform(t *testing.T) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	want := 201
	resp, err := client.Get(fmt.Sprintf("http://localhost:8080/mse7/send?code=%d", want))
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		t.Errorf("error  connecting to server, cause: %v", err)
	}

	got := resp.StatusCode
	if got != want {
		t.Errorf("URL transform expected http response want %d, got %d", want, got)
	}
}
