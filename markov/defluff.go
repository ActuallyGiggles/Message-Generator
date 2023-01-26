package markov

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

func defluff() {
	busy.Lock()
	debugLog("defluff ticker went off")
	defer duration(track("defluffing duration"))

	for _, chain := range chains(false, true) {
		if strings.Contains(chain, "_head") {
			defluffHead(chain)
		}

		if strings.Contains(chain, "_body") {
			defluffBody(chain)
		}

		if strings.Contains(chain, "_tail") {
			defluffTail(chain)
		}
	}

	busy.Unlock()
	debugLog("Done Defluffing at", time.Now().String())
}

func defluffHead(chain string) {
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

	err = os.Remove(defaultPath)
	if err != nil {
		panic(err)
	}

	err = os.Rename(newPath, defaultPath)
	if err != nil {
		panic(err)
	}
}

func defluffBody(chain string) {
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
					continue
				}

				// Add child into new list
				updatedParent.Children = append(updatedParent.Children, child{
					Word:  eChild.Word,
					Value: eChild.Value,
				})
				fmt.Println(eChild.Word, eChild.Value)
			}

			for _, eGrandparent := range existingParent.Grandparents {
				// If used less than x times in the past day, ignore
				if eGrandparent.Value < instructions.DefluffTriggerValue {
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

	err = os.Remove(defaultPath)
	if err != nil {
		panic(err)
	}

	err = os.Rename(newPath, defaultPath)
	if err != nil {
		panic(err)
	}
}

func defluffTail(chain string) {
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
				continue
			}

			// Add child into new list
			enc.AddEntry(child{
				Word:  existingGrandparent.Word,
				Value: existingGrandparent.Value,
			})
		}

		enc.CloseEncoder()
		fN.Close()

	}

	f.Close()

	err = os.Remove(defaultPath)
	if err != nil {
		panic(err)
	}

	err = os.Rename(newPath, defaultPath)
	if err != nil {
		panic(err)
	}
}
