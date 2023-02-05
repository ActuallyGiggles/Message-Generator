package markov

import (
	"encoding/json"
	"os"
	"strings"
	"time"
)

// determineDefluffTime will determine the next defluff time and whether the ticker should be new or not.
func determineDefluffTime(init bool) (newTicker *time.Ticker) {
	// If a positive amount of time is left (did not miss defluff time), return defluff ticker.
	if timeRemaining := time.Until(stats.NextDefluffTime); timeRemaining > 0 && init {
		return time.NewTicker(timeRemaining)
	}

	return time.NewTicker(defluffInterval)
}

func defluff() {
	busy.Lock()
	defer duration(track("defluffing duration"))

	var totalRemoved int

	for _, chain := range chains(false, true) {
		if w, ok := workerMap[chain[:len(chain)-5]]; ok {
			w.ChainMx.Lock()

			if strings.Contains(chain, "_head") {
				totalRemoved += defluffHead(chain)
			}

			if strings.Contains(chain, "_body") {
				totalRemoved += defluffBody(chain)
			}

			if strings.Contains(chain, "_tail") {
				totalRemoved += defluffTail(chain)
			}

			w.ChainMx.Unlock()
		}
	}

	busy.Unlock()
	debugLog("Total defluffed:", totalRemoved)
	stats.NextDefluffTime = time.Now().Add(defluffInterval)
}

func defluffHead(chain string) (removed int) {
	defaultPath := "./markov-chains/" + chain + ".json"
	newPath := "./markov-chains/" + chain + "_defluffed.json"

	// Open existing chain file
	f, err := os.OpenFile(defaultPath, os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	} else {
		// Start a new decoder
		dec := json.NewDecoder(f)

		// Get beginning token
		_, err = dec.Token()
		if err != nil {
			panic(err)
		}

		fN, err := os.OpenFile(newPath, os.O_CREATE, 0666)
		if err != nil {
			panic(err)
		}

		var enc encode
		if err = StartEncoder(&enc, fN); err != nil {
			panic(err)
		}

		// For everything in old file
		for dec.More() {
			var existingChild child

			err := dec.Decode(&existingChild)
			if err != nil {
				panic(err)
			}

			// If used less than x times in the past day, ignore
			if existingChild.Value < instructions.DefluffTriggerValue {
				removed++
				continue
			}

			// Add child into new list
			err = enc.AddEntry(child{
				Word:  existingChild.Word,
				Value: existingChild.Value,
			})
			if err != nil {
				panic(err)
			}
		}

		err = enc.CloseEncoder()
		if err != nil {
			panic(err)
		}

		fN.Close()
	}

	f.Close()

	removeAndRename(defaultPath, newPath)

	return removed
}

func defluffBody(chain string) (removed int) {
	defaultPath := "./markov-chains/" + chain + ".json"
	newPath := "./markov-chains/" + chain + "_defluffed.json"

	// Open existing chain file
	f, err := os.OpenFile(defaultPath, os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	} else {
		// Start a new decoder
		dec := json.NewDecoder(f)

		// Get beginning token
		_, err = dec.Token()
		if err != nil {
			panic(err)
		}

		fN, err := os.OpenFile(newPath, os.O_CREATE, 0666)
		if err != nil {
			panic(err)
		}

		var enc encode
		if err = StartEncoder(&enc, fN); err != nil {
			panic(err)
		}

		// For everything in old file
		for dec.More() {
			var existingParent parent
			var updatedParent parent

			err := dec.Decode(&existingParent)
			if err != nil {
				panic(err)
			}

			for _, eChild := range existingParent.Children {
				// If used less than x times in the past day, ignore
				if eChild.Value < instructions.DefluffTriggerValue {
					removed++
					continue
				}

				// Add child into new list
				updatedParent.Children = append(updatedParent.Children, child{
					Word:  eChild.Word,
					Value: eChild.Value,
				})
			}

			for _, eGrandparent := range existingParent.Grandparents {
				// If used less than x times in the past day, ignore
				if eGrandparent.Value < instructions.DefluffTriggerValue {
					removed++
					continue
				}

				// Add grandparent into new list
				updatedParent.Grandparents = append(updatedParent.Grandparents, grandparent{
					Word:  eGrandparent.Word,
					Value: eGrandparent.Value,
				})
			}

			if len(updatedParent.Children) == 0 && len(updatedParent.Grandparents) == 0 {
				continue
			}

			updatedParent.Word = existingParent.Word
			err = enc.AddEntry(updatedParent)
			if err != nil {
				panic(err)
			}
		}

		err = enc.CloseEncoder()
		if err != nil {
			panic(err)
		}

		fN.Close()

	}

	f.Close()

	removeAndRename(defaultPath, newPath)

	return removed
}

func defluffTail(chain string) (removed int) {
	defaultPath := "./markov-chains/" + chain + ".json"
	newPath := "./markov-chains/" + chain + "_defluffed.json"

	// Open existing chain file
	f, err := os.OpenFile(defaultPath, os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	} else {
		// Start a new decoder
		dec := json.NewDecoder(f)

		// Get beginning token
		_, err = dec.Token()
		if err != nil {
			panic(err)
		}

		fN, err := os.OpenFile(newPath, os.O_CREATE, 0666)
		if err != nil {
			panic(err)
		}

		var enc encode
		if err = StartEncoder(&enc, fN); err != nil {
			panic(err)
		}

		// For everything in old file
		for dec.More() {
			var existingGrandparent grandparent

			err := dec.Decode(&existingGrandparent)
			if err != nil {
				panic(err)
			}

			// If used less than x times in the past day, ignore
			if existingGrandparent.Value < instructions.DefluffTriggerValue {
				removed++
				continue
			}

			// Add grandparent into new list
			err = enc.AddEntry(grandparent{
				Word:  existingGrandparent.Word,
				Value: existingGrandparent.Value,
			})
			if err != nil {
				panic(err)
			}
		}

		err = enc.CloseEncoder()
		if err != nil {
			panic(err)
		}

		fN.Close()

	}

	f.Close()

	removeAndRename(defaultPath, newPath)
	return removed
}
