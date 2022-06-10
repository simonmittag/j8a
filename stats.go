package j8a

import (
	"fmt"
	"github.com/hako/durafmt"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/process"
	"github.com/simonmittag/procspy"
	"net"
	"os"
	"runtime/pprof"
	"sync"
	"time"
)

type sample struct {
	pid                int32
	cpuPc              float64
	mPc                float32
	rssBytes           uint64
	vmsBytes           uint64
	swapBytes          uint64
	time               time.Time
	dwnOpenTcpConns    uint64
	dwnMaxOpenTcpConns uint64
	upOpenTcpConns     uint64
	upMaxOpenTcpConns  uint64
	threads            int
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
const pidDwnOpenTcpConns = "pidDwnOpenTcpConns"
const pidDwnMaxOpenTcpConns = "pidDwnMaxOpenTcpConns"
const pidUpOpenTcpConns = "pidUpOpenTcpConns"
const pidUpMaxOpenTcpConns = "pidUpMaxOpenTcpConns"
const pidOSThreads = "pidOSThreads"
const serverPerformance = "server performance"
const pcd2f = "%.2f"

const rssMemIncrease = "RSS memory increase for previous %s with high factor >=%s, monitor actively."

func (s sample) log() {
	log.Warn().
		Int32(pid, s.pid).
		Str(pidCPUCorePct, fmt.Sprintf(pcd2f, s.cpuPc)).
		Str(pidMemPct, fmt.Sprintf(pcd2f, s.mPc)).
		Uint64(pidDwnOpenTcpConns, s.dwnOpenTcpConns).
		Uint64(pidDwnMaxOpenTcpConns, s.dwnMaxOpenTcpConns).
		Uint64(pidUpOpenTcpConns, s.upOpenTcpConns).
		Uint64(pidUpMaxOpenTcpConns, s.upMaxOpenTcpConns).
		Uint64(pidRssBytes, s.rssBytes).
		Uint64(pidVmsBytes, s.vmsBytes).
		Uint64(pidSwapBytes, s.swapBytes).
		Int(pidOSThreads, s.threads).
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

	var threadProfile = pprof.Lookup("threadcreate")

	cpuPc, _ := proc.Percent(time.Millisecond * cpuSampleMilliSeconds)
	mPc, _ := proc.MemoryPercent()
	mInfo, _ := proc.MemoryInfo()

	cs, e := rt.FindUpConns()
	if e != nil {
		//if this fails, skip upconns and continue with other sample data.
		_ = e
	} else {
		d := rt.CountUpConns(proc, cs, rt.LookUpResourceIps())
		rt.ConnectionWatcher.SetUp(uint64(d))
		rt.ConnectionWatcher.UpdateMaxUp(uint64(d))
	}

	procStatsLock.Unlock()
	return sample{
		pid:                proc.Pid,
		cpuPc:              cpuPc,
		mPc:                mPc,
		rssBytes:           mInfo.RSS,
		vmsBytes:           mInfo.VMS,
		swapBytes:          mInfo.Swap,
		time:               time.Now(),
		dwnOpenTcpConns:    rt.ConnectionWatcher.DwnCount(),
		dwnMaxOpenTcpConns: rt.ConnectionWatcher.DwnMaxCount(),
		upOpenTcpConns:     rt.ConnectionWatcher.UpCount(),
		upMaxOpenTcpConns:  rt.ConnectionWatcher.UpMaxCount(),
		threads:            threadProfile.Count(),
	}
}

func (rt *Runtime) FindUpConns() (procspy.ConnIter, error) {
	cs, e := procspy.Connections(true)
	return cs, e
}

func (rt *Runtime) CountUpConns(proc *process.Process, cs procspy.ConnIter, ips map[string][]net.IP) int {
	d := 0
UpConn:
	for c := cs.Next(); c != nil; c = cs.Next() {
		if c.PID == uint(proc.Pid) {
			for _, v := range rt.Config.Resources {
				for _, r := range v {
					if c.RemotePort == r.URL.Port {
						for _, ip := range ips[r.URL.Host] {
							if ip.Equal(c.RemoteAddress) {
								d++
								continue UpConn
							}
						}
					}
				}
			}
		}
	}
	return d
}

func (rt *Runtime) LookUpResourceIps() map[string][]net.IP {
	var ips = make(map[string][]net.IP)
	for _, v := range rt.Resources {
		for _, r := range v {
			is := make([]net.IP, 1)
			is[0] = net.ParseIP(r.URL.Host)
			if is[0] == nil {
				is, _ = net.LookupIP(r.URL.Host)
			}
			ips[r.URL.Host] = is
		}
	}
	return ips
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
