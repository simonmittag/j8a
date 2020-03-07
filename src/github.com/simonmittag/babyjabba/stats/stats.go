package stats

import (
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/process"
)

func BootStrap() {
	go stats(os.Getpid())
}

func stats(pid int) {
	proc, _ := process.NewProcess(int32(pid))
	for {

		cpuPc, _ := proc.Percent(time.Millisecond * 2000)
		mPc, _ := proc.MemoryPercent()
		mInfo, _ := proc.MemoryInfo()
		log.Debug().Int("pid", pid).Float64("cpuPercent", cpuPc).Float32("memPercent", mPc).
			Uint64("rssBytes", mInfo.RSS).Uint64("vmsBytes", mInfo.VMS).Uint64("swapBytes", mInfo.Swap).
			Msg("server performance stats sample")

		time.Sleep(time.Minute * 1)
	}
}
