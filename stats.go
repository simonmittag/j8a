package jabba

import (
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/process"
)

const cpuSampleMilliSeconds = 2000
const samplerSleepSeconds = 60

func stats(pid int) {
	proc, _ := process.NewProcess(int32(pid))
	for {
		cpuPc, _ := proc.Percent(time.Millisecond * cpuSampleMilliSeconds)
		mPc, _ := proc.MemoryPercent()
		mInfo, _ := proc.MemoryInfo()
		log.Debug().Int("pid", pid).Float64("cpuPercent", cpuPc).Float32("memPercent", mPc).
			Uint64("rssBytes", mInfo.RSS).Uint64("vmsBytes", mInfo.VMS).Uint64("swapBytes", mInfo.Swap).
			Msg("server performance stats sample")

		time.Sleep(time.Second * samplerSleepSeconds)
	}
}
