package j8a

import (
	"fmt"
	"github.com/hako/durafmt"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/process"
)

type sample struct {
	pid         int32
	cpuPc       float64
	mPc         float32
	rssBytes    uint64
	vmsBytes    uint64
	swapBytes   uint64
	time        time.Time
	openTcpCons uint64
}

type growthRate struct {
	v    float64
	high bool
}

type samples []sample

const cpuSampleMilliSeconds = 2000
const logSamplerSleepSeconds = 60
const historySamplerSleepSeconds = 3600
const historyMaxSamples = 24
const growthRateThreshold float64 = 2.0

var procStatsLock sync.Mutex
var procHistory samples

const pid = "pid"
const pidCPUCorePct = "pidCpuCore"
const pidMemPct = "pidMemPct"
const pidRssBytes = "pidRssBytes"
const pidVmsBytes = "pidVmsBytes"
const pidSwapBytes = "pidSwapBytes"
const pidOpenTcpCons = "pidOpenTcpCons"
const serverPerformance = "server performance"
const pcd2f = "%.2f"

const rssMemIncrease = "RSS memory increase for previous %s with high factor >=%s, monitor actively."

func (s sample) log() {
	log.Debug().
		Int32(pid, s.pid).
		Str(pidCPUCorePct, fmt.Sprintf(pcd2f, s.cpuPc)).
		Str(pidMemPct, fmt.Sprintf(pcd2f, s.mPc)).
		Uint64(pidOpenTcpCons, s.openTcpCons).
		Uint64(pidRssBytes, s.rssBytes).
		Uint64(pidVmsBytes, s.vmsBytes).
		Uint64(pidSwapBytes, s.swapBytes).
		Msg(serverPerformance)
}

func (samples *samples) append(s sample) {
	l := len(*samples)
	if l >= historyMaxSamples {
		(*samples)[0] = *new(sample)
		*samples = (*samples)[1:]
	}
	*samples = append(*samples, s)
}

func (samples *samples) log() []growthRate {
	var growthRates = make([]growthRate, len(*samples))

	for l := len(*samples) - 1; l >= 0; l-- {
		if (*samples)[l].pid == 0 {
			//effectively does nothing to log here because of insufficient data.
			return growthRates
		}
	}

	for l := len(*samples) - 1; l >= 0; l-- {
		growthRates[l].v = float64((*samples)[l].rssBytes) / float64((*samples)[0].rssBytes)
		if growthRates[l].v >= growthRateThreshold {
			growthRates[l].high = true
		}
	}

High:
	for m := len(*samples) - 1; m >= 0; m-- {
		if growthRates[m].high {
			log.Debug().
				Msgf(rssMemIncrease,
					durafmt.Parse(time.Duration(time.Second*historySamplerSleepSeconds*historyMaxSamples)).LimitFirstN(1).String(),
					fmt.Sprintf("%.1f", growthRateThreshold))
			break High
		}
	}

	return growthRates
}

func (rt *Runtime) getSample(proc *process.Process) sample {
	procStatsLock.Lock()

	cpuPc, _ := proc.Percent(time.Millisecond * cpuSampleMilliSeconds)
	mPc, _ := proc.MemoryPercent()
	mInfo, _ := proc.MemoryInfo()

	procStatsLock.Unlock()
	return sample{
		pid:         proc.Pid,
		cpuPc:       cpuPc,
		mPc:         mPc,
		rssBytes:    mInfo.RSS,
		vmsBytes:    mInfo.VMS,
		swapBytes:   mInfo.Swap,
		time:        time.Now(),
		openTcpCons: rt.ConnectionWatcher.Count(),
	}
}

//log proc samples infinite loop
func (rt *Runtime) logRuntimeStats(proc *process.Process) {
	go func() {
		for {
			rt.getSample(proc).log()
			time.Sleep(time.Second * logSamplerSleepSeconds)
		}
	}()

	go func() {
		procHistory = make(samples, historyMaxSamples)
		lazy := rt.getSample(proc)
		for k := 0; k < historyMaxSamples; k++ {
			procHistory[k] = lazy
		}
		for {
			procHistory.append(rt.getSample(proc))
			time.Sleep(time.Second * historySamplerSleepSeconds)
		}
	}()

	go func() {
		for {
			time.Sleep(time.Second * historySamplerSleepSeconds)
			procHistory.log()
		}
	}()
}

const uptimeMicros = "uptimeMicros"

func (rt *Runtime) logUptime() {
	go func() {
		for {
			upNanos := time.Since(rt.Start)
			if upNanos > time.Second*10 {
				uptime := durafmt.Parse(upNanos).LimitFirstN(1).String()
				log.Debug().
					Int(pid, os.Getpid()).
					Int64(uptimeMicros, int64(upNanos/1000)).
					Msgf(fmt.Sprintf("server upTime is %s", uptime))
			}
			time.Sleep(time.Hour * 24)
		}
	}()
}
