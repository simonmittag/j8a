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

func TestSamplesAppendDoNotExceedCapacity(t *testing.T) {
	pid := os.Getpid()
	proc, _ := process.NewProcess(int32(pid))
	var history samples = make(samples, historyMaxSamples)

	for i := 0; i < 50; i++ {
		sample := getSample(proc)
		time.Sleep(time.Millisecond * 50)
		t.Logf("appending %d sample", i)
		history.append(sample)
	}

	got := cap(history)
	want := historyMaxSamples * 2
	if got > want {
		t.Errorf("samples should be limited capacity want %d got %d", want, got)
	} else {
		t.Logf("normal. samples capacity %d", got)
	}

	if len(history) != historyMaxSamples {
		t.Errorf("samples should be length want %d got %d", historyMaxSamples, len(history))
	}
}

func TestProcStatsAppendInOrder(t *testing.T) {
	pid := int32(os.Getpid())
	proc, _ := process.NewProcess(pid)
	var history samples = make(samples, historyMaxSamples)
	sample := getSample(proc)

	for i := 0; i < historyMaxSamples; i++ {
		history.append(sample)
		if history[historyMaxSamples-1-i].pid != pid {
			t.Errorf("sample not inserted")
		}
	}
}

func TestRSSGrowthRatesHigh(t *testing.T) {
	memory := samples{
		sample{
			pid:      1,
			rssBytes: 16,
			time:     time.Time{},
		},
		sample{
			pid:      1,
			rssBytes: 32,
			time:     time.Time{},
		},
	}

	growth := memory.log()

	var want float64 = 2
	for _, got := range growth {
		if got.v > want {
			t.Errorf("growth too large, want <= %f got %f", want, got.v)
		}
	}

	got2 := growth[0].v
	want2 := float64(1)
	if got2 != want2 {
		t.Errorf("illegal grow rate at index 0, want %f, got %f", want2, got2)
	}

	got21 := growth[0].high
	want21 := false
	if got21 != want21 {
		t.Errorf("illegal high mark at index 0, want %t, got %t", want21, got21)
	}

	got3 := growth[1].v
	want3 := float64(2)
	if got3 != want3 {
		t.Errorf("illegal grow rate at index 1, want %f, got %f", want3, got3)
	}

	got31 := growth[1].high
	want31 := true
	if got31 != want31 {
		t.Errorf("illegal high mark at index 1, want %t, got %t", want31, got31)
	}
}

func TestRSSGrowthRatesLow(t *testing.T) {
	memory := samples{
		sample{
			pid:      1,
			rssBytes: 16,
			time:     time.Time{},
		},
		sample{
			pid:      1,
			rssBytes: 17,
			time:     time.Time{},
		},
		sample{
			pid:      1,
			rssBytes: 27,
			time:     time.Time{},
		},
		sample{
			pid:      1,
			rssBytes: 29,
			time:     time.Time{},
		},
	}

	growth := memory.log()

	var want float64 = 2
	for i, got := range growth {
		if got.high {
			t.Errorf("growth illegally marked as high at index %d", i)
		}
		if got.v > want {
			t.Errorf("growth too large, want <= %f got %f", want, got.v)
		} else {
			t.Logf("normal. got growth %f", got.v)
		}
	}
}

func TestRSSGrowthRatesInsufficientData(t *testing.T) {
	memory := samples{
		sample{
			pid: 0,
		},
		sample{
			pid: 0,
		},
	}

	growth := memory.log()

	var want float64 = 1
	for i, got := range growth {
		if got.high {
			t.Errorf("no data, growth illegally marked as high at index %d", i)
		}
		if got.v > want {
			t.Errorf("no data, want growth %f got %f", want, got.v)
		}
	}
}

func TestLogProcStats(t *testing.T) {
	//runtime panic test only, no assertions
	proc, _ := process.NewProcess(int32(os.Getpid()))
	logProcStats(proc)
	time.Sleep(time.Duration(time.Second * 3))
}
