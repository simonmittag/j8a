package j8a

import (
	"fmt"
	"github.com/hako/durafmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/process"
)

const cpuSampleMilliSeconds = 2000
const samplerSleepSeconds = 60

type sample struct {
	pid       int
	cpuPc     float64
	mPc       float32
	rssBytes  uint64
	vmsBytes  uint64
	swapBytes uint64
}

//infinite loop. run this in a goroutine to hold on to the process object
func stats(pid int) {
	proc, _ := process.NewProcess(int32(pid))
	for {
		logSample(getSample(pid, proc))
		time.Sleep(time.Second * samplerSleepSeconds)
	}
}

func uptime() {
	for {
		upNanos := time.Since(Runner.Start)
		uptime := durafmt.Parse(upNanos).LimitFirstN(2).String()
		log.Debug().
			Int64("uptimeMicros", int64(upNanos/1000)).
			Msgf(fmt.Sprintf("server uptime %s", uptime))
		time.Sleep(time.Hour * 24)
	}
}

func getSample(pid int, proc *process.Process) sample {
	cpuPc, _ := proc.Percent(time.Millisecond * cpuSampleMilliSeconds)
	mPc, _ := proc.MemoryPercent()
	mInfo, _ := proc.MemoryInfo()

	return sample{
		pid:       pid,
		cpuPc:     cpuPc,
		mPc:       mPc,
		rssBytes:  mInfo.RSS,
		vmsBytes:  mInfo.VMS,
		swapBytes: mInfo.Swap,
	}
}

func logSample(s sample) {
	log.Debug().Int("pid", s.pid).Str("cpuCore"+
		"Pct", fmt.Sprintf("%.2f", s.cpuPc)).Str("memPct", fmt.Sprintf("%.2f", s.mPc)).
		Uint64("rssBytes", s.rssBytes).Uint64("vmsBytes", s.vmsBytes).Uint64("swapBytes", s.swapBytes).
		Msg("server performance stats sample")
}
