package markov

import (
	"sync"
	"time"
)

var (
	instructions StartInstructions

	busy sync.Mutex

	stats Statistics

	trackingProgress bool
	progressChannel  chan Progress
)

// Start starts markov based on instructions pDuration
func Start(sI StartInstructions) {
	instructions = sI

	createFolders()
	loadStats()
	// go checkForDefluffDate(false)
	go tickerLoops()
}

func tickerLoops() {
	var writingTicker *time.Ticker
	var zippingTicker *time.Ticker
	var defluffTicker *time.Ticker

	writingTicker = writeTicker()
	if instructions.ShouldZip {
		zippingTicker = time.NewTicker(6 * time.Hour)
	}
	if instructions.ShouldDefluff {
		defluffTicker = time.NewTicker(7 * (24 * time.Hour))
	}

	for {
		select {
		case <-writingTicker.C:
			go writeLoop()
		case <-zippingTicker.C:
			go zipChains()
		case <-defluffTicker.C:
			go defluff()
		}
	}
}
