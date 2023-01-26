package markov

import (
	"fmt"
	"sync"
	"time"
)

var (
	instructions    StartInstructions
	writeInterval   = 1 * time.Minute
	zipInterval     = 6 * time.Hour
	defluffInterval = 24 * time.Hour

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
	go tickerLoops()
}

func tickerLoops() {
	var writingTicker *time.Ticker
	var zippingTicker *time.Ticker
	var defluffTicker *time.Ticker

	writingTicker = writeTicker()
	if instructions.ShouldZip {
		zippingTicker = time.NewTicker(zipInterval)
		stats.NextZipTime = time.Now().Add(zipInterval)
	}
	if instructions.ShouldDefluff {
		defluffTicker = time.NewTicker(defluffInterval)
		stats.NextDefluffTime = time.Now().Add(defluffInterval)
	}

	for {
		select {
		case <-writingTicker.C:
			// fmt.Println("write ticker went off")
			go writeLoop()
			stats.NextWriteTime = time.Now().Add(writeInterval)
		case <-zippingTicker.C:
			fmt.Println("zip ticker went off")
			//go zipChains()
			stats.NextZipTime = time.Now().Add(zipInterval)
		case <-defluffTicker.C:
			fmt.Println("defluff ticker went off")
			//go defluff()
			stats.NextDefluffTime = time.Now().Add(defluffInterval)
		}
	}
}
