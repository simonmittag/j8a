package routing

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestRouting(t *testing.T) {
	var tests = []struct {
		name         string
		url          string
		responseCode int
		echoport     int
	}{
		{"with no host and path match, path transform", "http://localhost:8080/mse7/echoport", 200, 60083},
		{"with no host and path match", "http://localhost:8080/jwtes256/echoport", 401, 0},
		{"with no host and path match", "http://localhost:8080/jwths256/echoport", 401, 0},
		{"with no host and path match", "http://localhost:8080/jwtnone/echoport", 401, 0},
		{"with no host and path match", "http://localhost:8080/mse6jwt", 401, 0},
		{"with no host and path match", "http://localhost:8080/mse6/echoport", 200, 60083},
		{"with unicode host and path match", "http://aaa😊😊😊.com:8080/mse6/echoport", 200, 60084},
		{"with unicode host and path match", "http://aaa😊😊😊.com:8080/mse66/echoport", 200, 60083},
		{"with unicode host match but route doesn't match", "http://aaa😊😊😊.com:8080/", 404, 0},
		{"with punycode host and path match", "http://xn--aaa-yi33baa.com:8080/mse6/echoport", 200, 60084},
		{"with subdomain punycode host and path match to route", "http://sub.xn--bbb-yi33baa.com:8080/mse6/echoport", 200, 60084},
		{"with subdomain host and path match to route", "http://sub.bbb😊😊😊.com:8080/mse6/echoport", 200, 60084},
		{"with punycode host and no path match", "http://sub.xn--bbb-yi33baa.com:8080/noroute", 404, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &http.Client{}
			resp, _ := client.Get(tt.url)
			body := ""
			if resp != nil && resp.Body != nil {
				b, _ := io.ReadAll(resp.Body)
				body = string(b)
				defer resp.Body.Close()
			}
			if resp == nil {
				t.Error("no response")
			} else if resp.StatusCode != tt.responseCode {
				t.Errorf("url %v want response %v, got %v", tt.url, tt.responseCode, resp.StatusCode)
			}

			if tt.echoport > 0 && !strings.Contains(body, fmt.Sprintf("%v", tt.echoport)) {
				t.Errorf("wanted echo port %v, but got %v", tt.echoport, body)
			}
		})
	}
}
