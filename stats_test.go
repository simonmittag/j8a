package j8a

import (
	"github.com/shirou/gopsutil/process"
	"os"
	"testing"
	"time"
)

func TestStats(t *testing.T) {
	pid := os.Getpid()
	proc, _ := process.NewProcess(int32(pid))

	s := getSample(proc)
	if s.pid != int32(os.Getpid()) {
		t.Error("pid not correctly set")
	}
	if s.rssBytes == 0 {
		t.Error("rss memory bytes cannot be zero")
	}

	if s.mPc == 0 {
		t.Error("memory percent cannot be zero")
	}
	//we run this just to look for runtime errors
	s.log()
}

func TestSamplesAppend(t *testing.T) {
	pid := os.Getpid()
	proc, _ := process.NewProcess(int32(pid))
	var history samples = make(samples, historyMaxSamples)

	for i:=0;i<50;i++ {
		sample := getSample(proc)
		time.Sleep(time.Millisecond*50)
		t.Logf("appending %d sample", i)
		history.append(sample)
	}

	got := cap(history)
	want := historyMaxSamples*2
	if got > want {
		t.Errorf("samples should be limited capacity want %d got %d", want, got)
	} else {
		t.Logf("normal. samples capacity %d", got)
	}

	if len(history) != historyMaxSamples {
		t.Errorf("samples should be length want %d got %d", historyMaxSamples, len(history))
	}

}