package markov

import (
	"encoding/json"
	"os"
	"strings"
)

func Cleanse(entry string) (totalCleansed int) {
	busy.Lock()
	defer duration(track("cleanse duration"))

	for _, chain := range chains(false, true) {
		if w, ok := workerMap[chain[:len(chain)-5]]; ok {
			w.ChainMx.Unlock()

			if strings.Contains(chain, "_head") {
				totalCleansed += cleanseHead(chain, entry)
			}

			if strings.Contains(chain, "_body") {
				totalCleansed += cleanseBody(chain, entry)
			}

			if strings.Contains(chain, "_tail") {
				totalCleansed += cleanseTail(chain, entry)
			}

			w.ChainMx.Unlock()
		}
	}

	busy.Unlock()
	debugLog("Total cleansed:", totalCleansed)
	return totalCleansed
}

func cleanseHead(chain, entry string) (removed int) {
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
			if strings.Contains(existingChild.Word, entry) {
				removed++
				continue
			}

			// Add child into new list
			enc.AddEntry(child{
				Word:  existingChild.Word,
				Value: existingChild.Value,
			})
		}

		enc.CloseEncoder()
		fN.Close()
	}

	f.Close()

	removeAndRename(defaultPath, newPath)

	return removed
}

func cleanseBody(chain, entry string) (removed int) {
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
				if strings.Contains(eChild.Word, entry) {
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
				if strings.Contains(eGrandparent.Word, entry) {
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
			enc.AddEntry(updatedParent)
		}

		enc.CloseEncoder()
		fN.Close()

	}

	f.Close()

	removeAndRename(defaultPath, newPath)

	return removed
}

func cleanseTail(chain, entry string) (removed int) {
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
			if strings.Contains(existingGrandparent.Word, entry) {
				removed++
				continue
			}

			// Add grandparent into new list
			enc.AddEntry(grandparent{
				Word:  existingGrandparent.Word,
				Value: existingGrandparent.Value,
			})
		}

		enc.CloseEncoder()
		fN.Close()

	}

	f.Close()

	removeAndRename(defaultPath, newPath)
	return removed
}
