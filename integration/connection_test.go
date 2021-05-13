package integration

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"testing"
)

func Test100ConcurrentTCPConnectionsUsingHTTP11(t *testing.T) {
	total := 100
	good := 0
	bad := 0

	wg := sync.WaitGroup{}
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        1000,
			MaxIdleConnsPerHost: 1000,
			//disable HTTP/2 support for TLS
			TLSNextProto: map[string]func(authority string, c *tls.Conn) http.RoundTripper{},
		},
	}

	for i := 0; i < total; i++ {
		wg.Add(1)

		go func(j int) {
			serverPort := 8080
			t.Logf("goroutine %d, sending request", j)
			resp, err := client.Get(fmt.Sprintf("http://localhost:%d/mse6/slowbody?wait=2", serverPort))
			if err != nil {
				t.Errorf("received upstream error instead of 200: %v", err)
				bad++
			} else if resp != nil && resp.Body != nil {
				defer resp.Body.Close()
				if resp.Status != "200 OK" {
					t.Errorf("goroutine %d, received non 200 status: %v", j, resp.Status)
					bad++
				} else {
					t.Logf("goroutine %d, received status 200 OK", j)
					good++
				}
			}

			wg.Done()
		}(i)
	}

	wg.Wait()
	t.Logf("done! good: %d, bad: %d", good, bad)
}
