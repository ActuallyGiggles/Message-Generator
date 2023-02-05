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
	previousIntakeTotal int
	previousOutputTotal int
	logs                []string
)

func Start() {
	go intakePerHour()
}

func intakePerHour() {
	for range time.Tick(10 * time.Second) {
		mStats := markov.Stats()

		fmt.Println("session input", mStats.SessionInputs)
		stats.InputsPerHour = mStats.SessionInputs - previousIntakeTotal
		fmt.Println("input per hour", stats.InputsPerHour)
		fmt.Println("previous intake", previousIntakeTotal)
		previousIntakeTotal = mStats.SessionInputs

		previousOutputTotal = mStats.SessionOutputs
		stats.OutputsPerHour = mStats.SessionOutputs - previousOutputTotal
	}
}

func Log(message ...string) {
	t := time.Now()
	for _, m := range message {
		logs = append(logs, fmt.Sprintf("[%d/%d/%d %d:%d] %s %s", int(t.Month()), t.Day(), t.Year(), t.Hour(), t.Minute(), "|", m))
	}
}

func GetStats() (stats Stats) {
	stats.Markov = markov.Stats()
	stats.System = SystemStats()
	stats.Logs = logs
	return stats
}

// SystemStats provides statistics on CPU, Memory, and GoRoutines.
func SystemStats() (s SystemStatistics) {
	CPUUsage(&s)
	GoroutineUsage(&s)
	MemoryUsage(&s)
	return s
}
func CPUUsage(s *SystemStatistics) {
	percentage, err := cpu.Percent(0, false)
	if err != nil {
		panic(err)
	}
	s.CPU = percentage[0]
}

func MemoryUsage(s *SystemStatistics) {
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		panic(err)
	}
	s.Memory = vmStat.UsedPercent
}

func GoroutineUsage(s *SystemStatistics) {
	s.Goroutines = runtime.NumGoroutine()
}
