package markov

import (
	"encoding/json"
	"os"
	"strings"
	"time"
)

const standardDefluffDuration = 7 * (24 * time.Hour) // 1 week
var durationUntilDefluff time.Duration

// func checkForDefluffDate(setNewDate bool) {
// 	// If no date or specify to set new date, make new date according to standardDefluffDuration.
// 	if stats.DefluffDate.IsZero() || setNewDate {
// 		stats.DefluffDate = time.Now().Add(standardDefluffDuration)
// 		durationUntilDefluff = time.Until(stats.DefluffDate)
// 		saveStats()
// 	} else {
// 		// If date is already set, that means we need to complete the first date with a custom durationUntilDefluff first. Then we can set a ticker for standardDefluffDuration.
// 		durationUntilDefluff = time.Until(stats.DefluffDate)

// 		// Wait for the rest of the duration needed, defluff, and set a new date. If durationUntilDefluff is 0 or negative, it will start immediately.
// 		time.Sleep(durationUntilDefluff)
// 		defluff()
// 		go checkForDefluffDate(true)
// 		return
// 	}

// 	// We are now ready for a ticker to start.
// 	for range time.Tick(durationUntilDefluff) {
// 		go defluff()
// 	}
// }

func defluff() {
	busy.Lock()

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

			// If more than a week has passed since last used, ignore
			if time.Since(existingChild.LastUsed).Hours() > standardDefluffDuration.Hours() {
				continue
			}

			// Add child into new list
			enc.AddEntry(child{
				Word:     existingChild.Word,
				Value:    existingChild.Value,
				LastUsed: existingChild.LastUsed,
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
				// If more than a week has passed since last used, ignore
				if time.Since(eChild.LastUsed).Hours() > standardDefluffDuration.Hours() {
					continue
				}

				// Add child into new list
				updatedParent.Children = append(updatedParent.Children, child{
					Word:     eChild.Word,
					Value:    eChild.Value,
					LastUsed: eChild.LastUsed,
				})
			}

			for _, eGrandparent := range existingParent.Grandparents {
				// If more than a week has passed since last used, ignore
				if time.Since(eGrandparent.LastUsed).Hours() > standardDefluffDuration.Hours() {
					continue
				}

				// Add grandparent into new list
				updatedParent.Grandparents = append(updatedParent.Grandparents, grandparent{
					Word:     eGrandparent.Word,
					Value:    eGrandparent.Value,
					LastUsed: eGrandparent.LastUsed,
				})
			}

			if len(updatedParent.Children) != 0 && len(updatedParent.Grandparents) != 0 {
				enc.AddEntry(updatedParent)
			}
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
			var existingChild child

			err := dec.Decode(&existingChild)
			if err != nil {
				panic(err)
			}

			// If more than a week has passed since last used, ignore
			if time.Since(existingChild.LastUsed).Hours() > standardDefluffDuration.Hours() {
				continue
			}

			// Add child into new list
			enc.AddEntry(child{
				Word:     existingChild.Word,
				Value:    existingChild.Value,
				LastUsed: existingChild.LastUsed,
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
