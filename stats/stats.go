package stats

import (
	"Twitch-Message-Generator/markov"
	"fmt"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

var (
	StartTime           time.Time
	InputsPerHour       int
	previousIntakeTotal int
	OutputsPerHour      int
	previousOutputTotal int
	Logs                []string

	rts SystemStatistics
)

func Start() {
	StartTime = time.Now()

	go intakePerHour()
}

func intakePerHour() {
	for range time.Tick(1 * time.Hour) {
		stats := markov.Stats()

		InputsPerHour = stats.SessionInputs - previousIntakeTotal
		previousIntakeTotal = stats.SessionInputs

		OutputsPerHour = stats.SessionOutputs - previousOutputTotal
		previousOutputTotal = stats.SessionOutputs
	}
}

func Log(message ...string) {
	ct := time.Now()
	year, month, day := ct.Date()
	hour := ct.Hour()
	minute := ct.Minute()
	second := ct.Second()
	Logs = append(Logs, fmt.Sprintf("%d/%d/%d %d:%d:%d %s", year, int(month), day, hour, minute, second, message))
}

func GetStats() (stats Stats) {
	stats.Markov = markov.Stats()

	stats.InputsPerHour = InputsPerHour
	stats.OutputsPerHour = OutputsPerHour

	stats.System = SystemStats()

	stats.Logs = Logs

	return stats
}

// SystemStats provides statistics on CPU, Memory, and GoRoutines.
func SystemStats() SystemStatistics {
	CPUUsage(&rts)
	GoroutineUsage(&rts)
	MemoryUsage(&rts)

	return rts
}
func CPUUsage(rts *SystemStatistics) {
	percentage, err := cpu.Percent(0, false)
	if err != nil {
		panic(err)
	}

	rts.CPU = percentage[0]
}

func MemoryUsage(rts *SystemStatistics) {
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		panic(err)
	}

	rts.Memory = vmStat.UsedPercent
}

func GoroutineUsage(rts *SystemStatistics) {
	rts.Goroutines = runtime.NumGoroutine()
}

// func MemUsage(rts *RuntimeStatistics) {
// 	var m runtime.MemStats
// 	runtime.ReadMemStats(&m)
// 	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
// 	rts.MemoryAllocated = bToMb(m.Alloc)
// 	rts.MemoryTotalAllocated = bToMb(m.TotalAlloc)
// 	rts.MemorySystem = bToMb(m.Sys)
// }

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
