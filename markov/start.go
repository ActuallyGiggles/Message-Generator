package markov

import (
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

	errorChannel chan error
)

// Start starts markov based on instructions pDuration
func Start(sI StartInstructions) {
	instructions = sI

	createFolders()
	loadStats()
	loadChains()
	//go tickerLoops()
}

func TempTriggerWrite(streamer string) {
	w, exist := workerMap[streamer]
	if !exist {
		return
	}

	if len(w.Chain.Parents) == 0 {
		return
	}

	w.ChainMx.Lock()
	defer w.ChainMx.Unlock()

	// Find new peak intake chain
	if w.Intake > stats.PeakChainIntake.Amount {
		stats.PeakChainIntake.Chain = w.Name
		stats.PeakChainIntake.Amount = w.Intake
		stats.PeakChainIntake.Time = time.Now()
	}

	w.writeBody()

	w.Chain.Parents = nil
	w.Intake = 0
}

func tickerLoops() {
	var writingTicker *time.Ticker
	var zippingTicker *time.Ticker
	var defluffTicker *time.Ticker

	writingTicker = writeTicker()

	zippingTicker = time.NewTicker(zipInterval)
	stats.NextZipTime = time.Now().Add(zipInterval)

	defluffTicker = determineDefluffTime(true)

	for {
		select {
		case <-writingTicker.C:
			go writeLoop()
		case <-zippingTicker.C:
			go zipChains()
		case <-defluffTicker.C:
			defluffTicker = determineDefluffTime(false)
			go defluff()
		}
	}
}
