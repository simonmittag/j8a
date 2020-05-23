package integration

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"unicode"
)

func TestServerDoesNotExceedConnectionPoolSize(t *testing.T) {
	for j:=0; j<100;j++ {
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

	//double check connection pool size value in jabba config if the test fails.
	want := 8
	pid := getJabbaPid()
	got := osConnsWithLsof(pid, want)
	if got > want {
		t.Errorf("jabba pid %s too many upstream tcp connections, want %d, got %d", pid, want, got)
	}
}

//This is a brittle helper method to get the number of upstream connections for a pid.
//If this test starts failing check your os output of lsof.
//TODO do no rely on lsof output formatting
func osConnsWithLsof(pid string, want int) int {
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
	log.Info().Msgf("tried 16x100 rotating upstream connections, "+
		"process %s expected %d max in pool, found %d using lsof", pid, want, got)
	return got
}

func getJabbaPid() string {
	pidc := exec.Command("pgrep", "jabba")
	outpidc, _ := pidc.CombinedOutput()
	pid := string(outpidc)
	pid = strings.TrimFunc(pid, func(r rune) bool {
		return !unicode.IsGraphic(r)
	})
	log.Info().Msgf("pid for jabba %s", pid)
	return pid
}
