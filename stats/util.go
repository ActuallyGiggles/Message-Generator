package stats

import (
	"Message-Generator/markov"
)

type Stats struct {
	Markov         markov.Statistics
	InputsPerHour  int
	OutputsPerHour int

	System SystemStatistics

	WebsiteHits  int
	SentenceHits int

	Logs []string
}

type SystemStatistics struct {
	CPU        float64
	Memory     float64
	Goroutines int
}
