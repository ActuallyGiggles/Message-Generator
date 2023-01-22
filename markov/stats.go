package markov

import (
	"encoding/json"
	"os"
	"time"
)

func Stats() (statistics Statistics) {
	stats.LifetimeUptime = time.Now().Sub(stats.LifetimeStartTime)
	stats.SessionUptime = time.Now().Sub(stats.SessionStartTime)
	stats.TimeUntilWrite = stats.NextWriteTime.Sub(time.Now())
	stats.Workers = len(CurrentWorkers())
	return stats
}

func saveStats() {
	Stats()

	statsData, err := json.MarshalIndent(stats, "", " ")
	if err != nil {
		debugLog(err)
	}

	f, err := os.OpenFile("./markov-chains/stats/stats.json", os.O_CREATE, 0666)
	if err != nil {
		debugLog(err)
	}

	_, err = f.Write(statsData)
	defer f.Close()

	if err != nil {
		debugLog(err)
	} else {
	}
}

func loadStats() {
	f, err := os.OpenFile("./markov-chains/stats/stats.json", os.O_CREATE, 0666)
	if err != nil {
		debugLog("Failed reading stats:", err)
	}
	defer f.Close()

	fS, _ := f.Stat()
	if fS.Size() == 0 {
		stats.LifetimeStartTime = time.Now()
		stats.SessionStartTime = time.Now()
		return
	}

	err = json.NewDecoder(f).Decode(&stats)
	if err != nil {
		debugLog("Error when unmarshalling stats:", "\n", err)
	}

	stats.SessionStartTime = time.Now()
	stats.SessionInputs = 0
	stats.SessionOutputs = 0
	stats.Durations = nil
}

func track(process string) (string, time.Time) {
	return process, time.Now()
}

func duration(process string, start time.Time) {
	duration := time.Since(start).Round(1 * time.Second)
	debugLog(process + ": " + duration.String())

	var exists bool

	for _, d := range stats.Durations {
		if d.ProcessName == process {
			exists = true
			d.Duration = duration.String()
		}
	}

	if !exists {
		stats.Durations = append(stats.Durations, report{
			ProcessName: process,
			Duration:    duration.String(),
		})
	}
}

func ReportDurations() []report {
	return stats.Durations
}
