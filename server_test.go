package jabba

import (
	"github.com/rs/zerolog/log"
	"net/http"
	"sync"
	"testing"
	"time"
)

var mse6Wait sync.WaitGroup = sync.WaitGroup{}

func TestServerBootStrap(t *testing.T) {
	resp, _ := http.Get("http://localhost:8080/about")
	if resp.StatusCode != 200 {
		t.Errorf("server does not return ok status response after starting, want 200, got %v", resp.StatusCode)
	}
}

func TestServerMakesSuccessfulUpstreamConnection(t *testing.T) {
	resp, err := http.Get("http://localhost:8080/v2/billing")

	if err!=nil {
		t.Errorf("error connecting to upstream, cause: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("server does not return ok from working upstream, want 200, got %v", resp.StatusCode)
	}
}

//func TestServerUpstreamReadTimeout(t *testing.T) {
//	resp, err := http.Get("http://localhost:8080/about")
//	resp, err = http.Get("http://localhost:8080/v2/billing")
//	resp, err = http.Get("http://localhost:8080/v2/slowheader")
//
//	if err!=nil {
//		t.Errorf("error connecting to upstream, cause: %v", err)
//	}
//
//	if resp.StatusCode != 502 {
//		t.Errorf("slow header writes from server > upstream timeout should not return ok and fail after max attempts, want 502, got %v", resp.StatusCode)
//	}
//}

func setupJabbaWithMse6() {
	//start mse6 our upstream server
	mse6Wait.Add(1)
	go mse6Listen()
	mse6Wait.Wait()

	//start jabba
	Boot.Add(1)
	go BootStrap()
	Boot.Wait()
}

func mse6Listen() {
	log.Trace().Msg("starting Upstream http server MSE6")

	http.HandleFunc("/v2/billing", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "identity")
		w.WriteHeader(200)
		w.Write([]byte(`{"MSE6":"Hello from the billing endpoint"}`))
	})

	http.HandleFunc("/v2/posting", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "identity")
		w.WriteHeader(200)
		w.Write([]byte(`{"MSE6":"Hello from the posting endpoint"}`))
	})

	http.HandleFunc("/v2/slowbody", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "identity")
		w.WriteHeader(200)
		time.Sleep(time.Second*2)
		w.Write([]byte(`{"MSE6":"Hello from the posting endpoint"}`))
	})

	http.HandleFunc("/v2/slowheader", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Second*2)
		w.Header().Set("Content-Encoding", "identity")
		w.WriteHeader(200)
		w.Write([]byte(`{"MSE6":"Hello from the posting endpoint"}`))
	})

	http.HandleFunc("/v2/gzip", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		w.WriteHeader(200)
		w.Write(Gzip([]byte(`{"MSE6":"Hello from the posting endpoint"}`)))
	})

	mse6Wait.Done()
	err := http.ListenAndServe(":60083", nil)
	if err != nil {
		panic(err.Error())
	}
}
