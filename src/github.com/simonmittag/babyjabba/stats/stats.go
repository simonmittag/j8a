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
		log.Debug().Int("pid", pid).Float64("cpuPercent", cpuPc).Float32("memPercent", mPc).Msg("server performance stats sample")

		time.Sleep(time.Second * 5)
	}
}
