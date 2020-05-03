package jabba

import (
	"github.com/rs/zerolog/log"
	"net/http"
	"testing"
	"time"
)

func TestServerBootStrap(t *testing.T) {
	resp, _ := http.Get("http://localhost:8080/about")
	if resp.StatusCode != 200 {
		t.Errorf("server does not return ok status response after starting, want 200, got %v", resp.StatusCode)
	}
}

func TestServerMakesSuccessfulUpstreamConnection(t *testing.T) {
	resp, err := http.Get("http://localhost:8080/v2/billing")

	if err != nil {
		t.Errorf("error connecting to upstream, cause: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("server does not return ok from working upstream, want 200, got %v", resp.StatusCode)
	}
}

func TestServerUpstreamReadTimeout(t *testing.T) {
	resp, err := http.Get("http://localhost:8080/v2/slowheader")
	if err != nil {
		t.Errorf("error connecting to upstream, cause: %v", err)
	}

	if resp.StatusCode != 502 {
		t.Errorf("slow header writes from server > upstream timeout should not return ok and fail after max attempts, want 502, got %v", resp.StatusCode)
	}
}

func setupJabbaWithMse6() {
	Boot.Add(1)
	go mse6Listen()

	Boot.Add(1)
	go BootStrap()
	Boot.Wait()
}

func mse6Listen() {
	log.Trace().Msg("starting Upstream http server MSE6")

	http.HandleFunc("/v2/billing", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "identity")
		w.WriteHeader(200)
		w.Write([]byte(`{"MSE6":"Hello from the v2/billing endpoint"}`))
	})

	http.HandleFunc("/v2/posting", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "identity")
		w.WriteHeader(200)
		w.Write([]byte(`{"MSE6":"Hello from the v2/posting endpoint"}`))
	})

	http.HandleFunc("/v2/slowbody", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "identity")
		w.Header().Set("Transfer-Encoding", "identity")
		w.WriteHeader(200)
		hj, _ := w.(http.Hijacker)
		conn, bufrw, _ := hj.Hijack()
		defer conn.Close()
		time.Sleep(time.Second * 3)
		bufrw.WriteString(`[{"mse6":"Hello from the slowbody endpoint"}`)
		bufrw.Flush()
		time.Sleep(time.Second * 3)
		bufrw.WriteString(`,{"mse6":"and some more data from the slowbody endpoint"}]`)
		bufrw.Flush()
	})

	http.HandleFunc("/v2/slowheader", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Second * 3)
		w.Header().Set("Content-Encoding", "identity")
		w.WriteHeader(202)
		w.Write([]byte(`{"MSE6":"Hello from the v2/slowheader endpoint"}`))
	})

	http.HandleFunc("/v2/gzip", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		w.WriteHeader(200)
		w.Write(Gzip([]byte(`{"MSE6":"Hello from the v2/gzip endpoint"}`)))
	})

	Boot.Done()
	err := http.ListenAndServe(":60083", nil)
	if err != nil {
		panic(err.Error())
	}
}
