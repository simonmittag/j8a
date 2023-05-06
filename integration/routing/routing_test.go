package routing

import (
	"net/http"
	"testing"
)

func TestRouting(t *testing.T) {
	var tests = []struct {
		name         string
		url          string
		responseCode int
	}{
		{"with unicode host and path match", "http://aaa😊😊😊.com:8080/mse6/get", 200},
		{"with unicode host match but route doesn't match", "http://aaa😊😊😊.com:8080/", 404},
		{"with punycode host and path match", "http://xn--aaa-yi33baa.com:8080/mse6/get", 200},
		{"with subdomain punycode host and path match to route", "http://sub.xn--bbb-yi33baa.com:8080/mse6/get", 200},
		{"with punycode host and no path match", "http://sub.xn--bbb-yi33baa.com:8080/noroute", 404},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &http.Client{}
			resp, _ := client.Get(tt.url)
			if resp != nil && resp.Body != nil {
				defer resp.Body.Close()
			}
			if resp == nil {
				t.Error("no response")
			} else if resp.StatusCode != tt.responseCode {
				t.Errorf("url %v want response %v, got %v", tt.url, tt.responseCode, resp.StatusCode)
			}
		})
	}
}
