package markov

import (
	"fmt"
	"sync"
	"time"
)

var (
	instructions    StartInstructions
	writeInterval   = 10 * time.Minute
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
	// go checkForDefluffDate(false)
	go tickerLoops()
}

func tickerLoops() {
	var writingTicker *time.Ticker
	var zippingTicker *time.Ticker
	var defluffTicker *time.Ticker

	if instructions.WriteInterval == 0 {
		writingTicker = writeTicker()
	}
	if instructions.ShouldZip {
		zippingTicker = time.NewTicker(6 * time.Hour)
		stats.NextZipTime = time.Now().Add(time.Duration(6*time.Hour) * unit)
	}
	if instructions.ShouldDefluff {
		defluffTicker = time.NewTicker(24 * time.Hour)
		stats.NextDefluffTime = time.Now().Add(time.Duration(24*time.Hour) * unit)
	}

	for {
		select {
		case <-writingTicker.C:
			fmt.Println("write ticker went off")
			go writeLoop()
			stats.NextWriteTime = time.Now().Add(writeInterval)
		case <-zippingTicker.C:
			fmt.Println("zip ticker went off")
			go zipChains()
			stats.NextZipTime = time.Now().Add(zipInterval)
		case <-defluffTicker.C:
			fmt.Println("defluff ticker went off")
			go defluff()
			stats.NextDefluffTime = time.Now().Add(defluffInterval)
		}
	}
}
