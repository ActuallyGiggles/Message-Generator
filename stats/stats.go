package stats

import (
	"Message-Generator/markov"
	"fmt"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

var (
	stats               Stats
	StartTime           time.Time
	InputsPerHour       int
	previousIntakeTotal int
	OutputsPerHour      int
	previousOutputTotal int
	Logs                []string
)

func Start() {
	StartTime = time.Now()
	stats.Markov.TotalStartTime = GetStats().Markov.TotalStartTime
	stats.Markov.TotalUptime = GetStats().Markov.TotalUptime

	go intakePerHour()
}

func intakePerHour() {
	for range time.Tick(1 * time.Hour) {
		mStats := markov.Stats()

		stats.InputsPerHour = mStats.SessionInputs - previousIntakeTotal
		previousIntakeTotal = mStats.SessionInputs

		stats.OutputsPerHour = mStats.SessionOutputs - previousOutputTotal
		previousOutputTotal = mStats.SessionOutputs
	}
}

func Log(message ...string) {
	t := time.Now()
	stats.Logs = append(stats.Logs, fmt.Sprintf("[%d/%d/%d %d:%d] %s", int(t.Month()), t.Day(), t.Year(), t.Hour(), t.Minute(), message))
}

func GetStats() (stats Stats) {
	stats.Markov = markov.Stats()
	return stats
}

// SystemStats provides statistics on CPU, Memory, and GoRoutines.
func SystemStats() {
	CPUUsage(&stats.System)
	GoroutineUsage(&stats.System)
	MemoryUsage(&stats.System)
}
func CPUUsage(rts *SystemStatistics) {
	percentage, err := cpu.Percent(0, false)
	if err != nil {
		panic(err)
	}
	stats.System.CPU = percentage[0]
}

func MemoryUsage(rts *SystemStatistics) {
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		panic(err)
	}
	stats.System.Memory = vmStat.UsedPercent
}

func GoroutineUsage(rts *SystemStatistics) {
	stats.System.Goroutines = runtime.NumGoroutine()
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
