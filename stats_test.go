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

	rt := mockRuntime()

	s := rt.getSample(proc)
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

	rt := mockRuntime()

	//now using same sample for append test, it makes the test so much faster
	sample := rt.getSample(proc)
	for i := 0; i < 50; i++ {
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

	rt := mockRuntime()

	sample := rt.getSample(proc)

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
	rt := mockRuntime()
	rt.logRuntimeStats(proc)
	time.Sleep(time.Duration(time.Second * 3))
}

func TestRuntime_LookUpResourceIps(t *testing.T) {
	rt := mockRuntime()
	ips := rt.LookUpResourceIps()

	got1 := ips["localhost"][0].String()
	want1 := "::1"
	want11 := "127.0.0.1"
	if got1 != want1 && got1 != want11 {
		t.Errorf("invalid ip lookup for host, want %v, got %v", want1, got1)
	}
	got2 := ips["127.0.0.1"][0].String()
	want2 := "127.0.0.1"
	if got2 != want2 {
		t.Errorf("invalid ip lookup for ipv4, want %v, got %v", want2, got2)
	}
	got3 := ips["[::1]"][0].String()
	want3 := "::1"
	if got3 != want3 {
		t.Errorf("invalid ip lookup for ipv6, want %v, got %v", want3, got3)
	}
	got4 := ips["::1"][0].String()
	want4 := "::1"
	if got4 != want4 {
		t.Errorf("invalid ip lookup for ipv6, want %v, got %v", want4, got4)
	}
}

func BenchmarkRuntime_FindUpConns(b *testing.B) {
	rt := mockRuntime()
	for i := 0; i < b.N; i++ {
		rt.FindUpConns()
	}
}
