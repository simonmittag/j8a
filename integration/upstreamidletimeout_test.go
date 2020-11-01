package integration

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"
	"unicode"
)

func TestServerDoesNotExceedConnectionPoolSize(t *testing.T) {
	for j := 0; j < 10; j++ {
		for i := 1; i <= 16; i++ {
			url := fmt.Sprintf("http://localhost:8080/s%02d/get", i)
			resp, err := http.Get(url)
			//do this so the test does not leak file descriptors
			resp.Body.Close()

			if err != nil {
				t.Errorf("error connecting to upstream, cause: %v", err)
			}

			if resp.StatusCode != 200 {
				t.Errorf("server does not return ok from working upstream, want 200, got %v", resp.StatusCode)
			}
		}
	}

	//does the connection pool max out properly?
	//double check connection pool size value in j8a config if the test fails.
	wantConns := 8
	pid := getj8aPid()
	gotConns := osConnsWithLsof(pid)
	log.Info().Msgf("tried 16x10 rotating upstream connections, "+
		"j8a pid %s expected %d max in pool, found %d using lsof", pid, wantConns, gotConns)
	if gotConns > wantConns {
		t.Errorf("j8a pid %s too many idle upstream tcp connections in pool, want %d, got %d", pid, wantConns, gotConns)
	}

	//we also want to know if the pool empties after timeout passes.
	waitPeriodSeconds := 10
	grace := 1
	time.Sleep(time.Second * time.Duration(waitPeriodSeconds+grace))
	gotConns = osConnsWithLsof(pid)
	log.Info().Msgf("j8a pid %s after timeout want pool 0, found %d using lsof", pid, gotConns)
	if gotConns > 0 {
		t.Errorf("j8a pid %s connection pool not empty after timeout %d, want 0, got %d", pid, waitPeriodSeconds, gotConns)
	}
}

//This is a brittle helper method to get the number of upstream connections for a pid.
//If this test starts failing check your os output of lsof.
func osConnsWithLsof(pid string) int {
	lsofc := exec.Command("lsof", "-ai", "-p", pid)
	outlsofc, _ := lsofc.CombinedOutput()
	conns := strings.Split(string(outlsofc), "\n")
	r, _ := regexp.Compile("->localhost:60[0-1][0-9]{2}")
	got := 0
	for _, c := range conns {
		if r.MatchString(c) {
			got++
		}
	}
	return got
}

func getj8aPid() string {
	pidc := exec.Command("pgrep", "j8a")
	outpidc, _ := pidc.CombinedOutput()
	pid := string(outpidc)
	pid = strings.TrimFunc(pid, func(r rune) bool {
		return !unicode.IsGraphic(r)
	})
	log.Info().Msgf("pid for j8a %s", pid)
	return pid
}
