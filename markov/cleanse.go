package markov

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)

// Cleanse will go through every chain and remove any mention of the entry.
func Cleanse(entry string) (totalCleansed int) {
	busy.Lock()
	defer duration(track("cleanse duration"))

	for _, chain := range chains(false) {
		if w, ok := workerMap[chain]; ok {
			w.ChainMx.Lock()
			totalCleansed += cleanseBody(chain, entry)
			w.ChainMx.Unlock()
		}
	}

	busy.Unlock()
	debugLog("Total cleansed:", totalCleansed)
	return totalCleansed
}

func cleanseBody(chain, entry string) (removed int) {
	defaultPath := "./markov-chains/" + chain + ".json"
	newPath := "./markov-chains/" + chain + "_defluffed.json"

	// Open existing chain file
	f, err := os.Open(defaultPath)
	if err != nil {
		panic(err)
	}

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

	// Start the new file encoder
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

		// Do for every parent except start key
		if existingParent.Word != instructions.EndKey {
			for _, eChild := range existingParent.Children {
				if matches, _ := regexp.MatchString("\\b"+entry+"\\b", eChild.Word); matches {
					removed++
					continue
				}

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
				if matches, _ := regexp.MatchString("\\b"+entry+"\\b", eGrandparent.Word); matches {
					removed++
					continue
				}

				// Add grandparent into new list
				updatedParent.Grandparents = append(updatedParent.Grandparents, grandparent{
					Word:  eGrandparent.Word,
					Value: eGrandparent.Value,
				})
			}
		}

		if matches, _ := regexp.MatchString("\\b"+entry+"\\b", existingParent.Word); matches {
			continue
		}

		updatedParent.Word = existingParent.Word
		err = enc.AddEntry(updatedParent)
		if err != nil {
			panic(err)
		}

		fmt.Println(existingParent)
		fmt.Println(updatedParent)
	}

	// Close the new file encoder
	if err := enc.CloseEncoder(); err != nil {
		panic(err)
	}

	// Close new file
	err = fN.Close()
	if err != nil {
		panic(err)
	}

	// Close old file
	err = f.Close()
	if err != nil {
		panic(err)
	}

	// Remove the old file and rename the new file with the old file name
	removeAndRename(defaultPath, newPath)

	return removed
}
