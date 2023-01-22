package stats

import (
	"markov-generator/markov"
)

type Stats struct {
	Markov         markov.Statistics
	InputsPerHour  int
	OutputsPerHour int

	System SystemStatistics `json:"system"`

	WebsiteHits  int
	SentenceHits int

	Logs []string `json:"logs"`
}

type SystemStatistics struct {
	CPU        float64 `json:"cpu"`
	Memory     float64 `json:"memory"`
	Goroutines int     `json:"goroutines"`
}
