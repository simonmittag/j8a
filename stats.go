package jabba

import (
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
	log.Debug().Int("pid", s.pid).Float64("cpuPct", s.cpuPc).Float32("memPct", s.mPc).
		Uint64("rssBytes", s.rssBytes).Uint64("vmsBytes", s.vmsBytes).Uint64("swapBytes", s.swapBytes).
		Msg("server performance stats sample")
}
