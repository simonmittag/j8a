package jabba

import (
	"github.com/shirou/gopsutil/process"
	"os"
	"testing"
)

func TestStats(t *testing.T) {
	pid := os.Getpid()
	proc, _ := process.NewProcess(int32(pid))

	s := getSample(pid, proc)
	if s.pid != os.Getpid() {
		t.Error("pid not correctly set")
	}
	if s.rssBytes == 0 {
		t.Error("rss memory bytes cannot be zero")
	}
	if s.vmsBytes == 0 {
		t.Error("vms memory bytes should not be zero. in the unlikely case this test ever fails on your OS, please ignore")
	}
	if s.cpuPc == 0 {
		t.Error("cpu percent cannot be zero")
	}
	if s.mPc == 0 {
		t.Error("memory percent cannot be zero")
	}
	//we run this just to look for runtime errors
	logSample(s)
}