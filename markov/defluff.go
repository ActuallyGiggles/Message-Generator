package markov

import (
	"encoding/json"
	"os"
	"time"
)

// determineDefluffTime will determine the next defluff time and whether the ticker should be new or not.
func determineDefluffTime(init bool) (newTicker *time.Ticker) {
	// If a positive amount of time is left (did not miss defluff time), return defluff ticker.
	if timeRemaining := time.Until(stats.NextDefluffTime); timeRemaining > 0 && init {
		stats.NextDefluffTime = time.Now().Add(timeRemaining)
		return time.NewTicker(timeRemaining)
	}

	stats.NextDefluffTime = time.Now().Add(defluffInterval)
	return time.NewTicker(defluffInterval)
}

func defluff() {
	if !instructions.ShouldDefluff {
		return
	}

	busy.Lock()
	defer duration(track("defluffing duration"))

	var totalRemoved int

	for _, chain := range Chains() {
		if w, ok := workerMap[chain]; ok {
			w.ChainMx.Lock()
			totalRemoved += defluffChain(w.Name)
			w.ChainMx.Unlock()
		}
	}

	busy.Unlock()
	// fmt.Println("Total defluffed:", totalRemoved)
	stats.NextDefluffTime = time.Now().Add(defluffInterval)
}

func defluffChain(chain string) (othersRemoved int) {
	defaultPath := "./markov-chains/" + chain + ".json"
	newPath := "./markov-chains/" + chain + "_defluffed.json"

	// Open existing chain file
	f, err := os.Open(defaultPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Start a new decoder
	dec := json.NewDecoder(f)

	// Get beginning token
	_, err = dec.Token()
	if err != nil {
		panic(err)
	}

	// Create a new chain file
	fN, err := os.Create(newPath)
	if err != nil {
		panic(err)
	}
	defer fN.Close()

	// Start the new file encoder
	var enc encode
	if err = StartEncoder(&enc, fN); err != nil {
		panic(err)
	}

	// For every new item in the existing chain
	for dec.More() {
		var existingParent parent
		var existingParentValue int
		var updatedParent parent

		err := dec.Decode(&existingParent)
		if err != nil {
			panic(err)
		}

		// Do for every parent except start key
		if existingParent.Word != instructions.EndKey {
			for _, eChild := range existingParent.Children {
				// If used less than x times in the past day, ignore
				if eChild.Value <= instructions.DefluffTriggerValue {
					othersRemoved++
					continue
				}

				existingParentValue++

				// Add child into new list
				updatedParent.Children = append(updatedParent.Children, child{
					Word:  eChild.Word,
					Value: eChild.Value,
				})
			}
		}

		// Do for every parent except start key
		if existingParent.Word != instructions.StartKey {
			for _, eGrandparent := range existingParent.Grandparents {
				// If used less than x times in the past day, ignore
				if eGrandparent.Value <= instructions.DefluffTriggerValue {
					othersRemoved++
					continue
				}

				existingParentValue++

				// Add grandparent into new list
				updatedParent.Grandparents = append(updatedParent.Grandparents, grandparent{
					Word:  eGrandparent.Word,
					Value: eGrandparent.Value,
				})
			}
		}

		if (len(updatedParent.Children) == 0 && len(updatedParent.Grandparents) == 0) || existingParentValue < instructions.DefluffTriggerValue {
			continue
		}

		updatedParent.Word = existingParent.Word
		err = enc.AddEntry(updatedParent)
		if err != nil {
			panic(err)
		}
	}

	// Close the new file encoder
	if err := enc.CloseEncoder(); err != nil {
		panic(err)
	}

	// Remove the old file and rename the new file with the old file name
	removeAndRename(defaultPath, newPath)

	return othersRemoved
}
