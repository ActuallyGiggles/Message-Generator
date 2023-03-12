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

func TempTriggerWrite(name string) {
	exists, w := doesWorkerExist(name)
	if !exists {
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)
	w.writeChainHeader(&wg)
	wg.Wait()
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
			go writeLoop()
		case <-zippingTicker.C:
			//go zipChains()
		}
	}
}
