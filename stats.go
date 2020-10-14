package j8a

import (
	"fmt"
	"github.com/hako/durafmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/process"
)

type sample struct {
	pid       int32
	cpuPc     float64
	mPc       float32
	rssBytes  uint64
	vmsBytes  uint64
	swapBytes uint64
	time      time.Time
}

type samples []sample

const cpuSampleMilliSeconds = 2000
const logSamplerSleepSeconds = 60
const historySamplerSleepHours = 24
const historyMaxSamples int = 7

var procStatsLock sync.Mutex
var procHistory samples

func (s sample) log() {
	log.Debug().Int32("pid", s.pid).Str("cpuCore"+
		"Pct", fmt.Sprintf("%.2f", s.cpuPc)).Str("memPct", fmt.Sprintf("%.2f", s.mPc)).
		Uint64("rssBytes", s.rssBytes).Uint64("vmsBytes", s.vmsBytes).Uint64("swapBytes", s.swapBytes).
		Msg("server performance")
}

func (samples *samples) append(s sample) {
	l := len(*samples)
	if l >= historyMaxSamples {
		(*samples)[0] = *new(sample)
		*samples = (*samples)[1:]
	}
	*samples = append(*samples, s)
}

func getSample(proc *process.Process) sample {
	procStatsLock.Lock()

	cpuPc, _ := proc.Percent(time.Millisecond * cpuSampleMilliSeconds)
	mPc, _ := proc.MemoryPercent()
	mInfo, _ := proc.MemoryInfo()

	procStatsLock.Unlock()
	return sample{
		pid:       proc.Pid,
		cpuPc:     cpuPc,
		mPc:       mPc,
		rssBytes:  mInfo.RSS,
		vmsBytes:  mInfo.VMS,
		swapBytes: mInfo.Swap,
		time:      time.Now(),
	}
}

//log proc samples infinite loop
func logProcStats(proc *process.Process) {
	go func() {
		for {
			getSample(proc).log()
			time.Sleep(time.Second * logSamplerSleepSeconds)
		}
	}()

	go func() {
		procHistory = make(samples, historyMaxSamples)
		for {
			procHistory.append(getSample(proc))
			time.Sleep(time.Hour * historySamplerSleepHours)
		}
	}()
}

func logUptime() {
	for {
		upNanos := time.Since(Runner.Start)
		uptime := durafmt.Parse(upNanos).LimitFirstN(2).String()
		log.Debug().
			Int64("uptimeMicros", int64(upNanos/1000)).
			Msgf(fmt.Sprintf("server uptime is %s", uptime))
		time.Sleep(time.Hour * 24)
	}
}
