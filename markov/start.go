package markov

import (
	"sync"
	"time"
)

var (
	instructions    StartInstructions
	writeInterval   = 10 * time.Minute
	zipInterval     = 6 * time.Hour
	defluffInterval = 12 * time.Hour

	busy sync.Mutex

	stats Statistics
)

// Start starts markov based on instructions pDuration
func Start(sI StartInstructions) {
	instructions = sI

	createFolders()
	loadStats()
	go tickerLoops(instructions.ErrorChannel)
}

func tickerLoops(errCh chan error) {
	var writingTicker *time.Ticker
	var zippingTicker *time.Ticker
	var defluffTicker *time.Ticker

	writingTicker = writeTicker()

	if instructions.ShouldZip {
		zippingTicker = time.NewTicker(zipInterval)
		stats.NextZipTime = time.Now().Add(zipInterval)
	}

	if instructions.ShouldDefluff {
		defluffTicker = determineDefluffTime(true)
	}

	for {
		select {
		case <-writingTicker.C:
			go writeLoop(errCh)
		case <-zippingTicker.C:
			go zipChains()
		case <-defluffTicker.C:
			defluffTicker = determineDefluffTime(false)
			go defluff()
		}
	}
}
