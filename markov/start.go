package markov

import (
	"sync"
	"time"
)

var (
	instructions  StartInstructions
	writeInterval = 10 * time.Minute
	zipInterval   = 6 * time.Hour

	busy         sync.Mutex
	stats        Statistics
	errorChannel chan error
)

// Start starts markov based on instructions pDuration
func Start(sI StartInstructions) {
	instructions = sI

	createFolders()
	loadStats()
	loadChains()
	go tickerLoops()
}

func TempTriggerWrite() {
	writeLoop()
}

func tickerLoops() {
	var writingTicker *time.Ticker
	var zippingTicker *time.Ticker

	writingTicker = writeTicker()

	zippingTicker = time.NewTicker(zipInterval)
	stats.NextZipTime = time.Now().Add(zipInterval)

	for {
		select {
		case <-writingTicker.C:
			//go writeLoop()
		case <-zippingTicker.C:
			//go zipChains()
		}
	}
}
