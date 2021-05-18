package integration

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"testing"
)

func Test1024ConcurrentTCPConnectionsUsingHTTP11(t *testing.T) {
	ConcurrentHTTP11ConnectionsSucceed(1024, t)
}

//func Test4096ConcurrentTCPConnectionsUsingHTTP11(t *testing.T) {
//	ConcurrentHTTP11ConnectionsSucceed(4096, t)
//}

func ConcurrentHTTP11ConnectionsSucceed(total int, t *testing.T) {
	good := 0
	bad := 0

	R200 := 0
	N200 := 0

	wg := sync.WaitGroup{}
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        2000,
			MaxIdleConnsPerHost: 2000,
			//disable HTTP/2 support for TLS
			TLSNextProto: map[string]func(authority string, c *tls.Conn) http.RoundTripper{},
		},
	}

	for i := 0; i < total; i++ {
		wg.Add(1)

		go func(j int) {
			serverPort := 8080
			resp, err := client.Get(fmt.Sprintf("http://localhost:%d/mse6/slowbody?wait=2", serverPort))
			if err != nil {
				t.Errorf("received upstream error for GET request: %v", err)
				bad++
			} else if resp != nil && resp.Body != nil {
				defer resp.Body.Close()
				if resp.Status != "200 OK" {
					t.Logf("goroutine %d, received non 200 status but normal server response: %v", j, resp.Status)
					good++
					N200++
				} else {
					t.Logf("goroutine %d, received status 200 OK", j)
					good++
					R200++
				}
			}

			wg.Done()
		}(i)
	}

	wg.Wait()
	t.Logf("done! good HTTP response: %d, 200s: %d, non 200s: %d, connection errors: %d", good, R200, N200, bad)
}
