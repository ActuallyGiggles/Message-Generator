package markov

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

var (
	unit time.Duration
)

func writeTicker() *time.Ticker {
	if instructions.WriteInterval == 0 && instructions.IntervalUnit == "" {
		stats.NextWriteTime = time.Now().Add(writeInterval)
		return time.NewTicker(writeInterval)
	}

	switch instructions.IntervalUnit {
	default:
		unit = time.Minute
	case "seconds":
		unit = time.Second
	case "minutes":
		unit = time.Minute
	case "hours":
		unit = time.Hour
	}

	writeInterval = time.Duration(instructions.WriteInterval) * unit
	stats.NextWriteTime = time.Now().Add(writeInterval)
	return time.NewTicker(writeInterval)
}

func writeLoop(errCh chan error) {
	if !busy.TryLock() {
		return
	}

	defer duration(track("writing duration"))

	var wg sync.WaitGroup

	for _, w := range workerMap {
		wg.Add(1)
		go w.writeAllPerChain(errCh, &wg)
	}

	wg.Wait()

	saveStats()
	stats.NextWriteTime = time.Now().Add(writeInterval)
	busy.Unlock()
}

func (w *worker) writeAllPerChain(errCh chan error, wg *sync.WaitGroup) {
	defer wg.Done()

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

	w.writeBody(errCh)

	w.Chain.Parents = nil
	w.Intake = 0
}

func (w *worker) writeBody(errCh chan error) {
	defaultPath := "./markov-chains/" + w.Name + ".json"
	newPath := "./markov-chains/" + w.Name + "_new.json"

	// Open existing chain file
	f, err := os.Open(defaultPath)

	// If chain file does not exist, create it and write the data to it.
	if err != nil {
		f, err = os.Create(defaultPath)
		if err != nil {
			panic(err)
		}
		chainData, _ := json.Marshal(w.Chain.Parents)
		_, err = f.Write(chainData)
		if err != nil {
			panic(err)
		}
		f.Close()
		return
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

	// For every new item in the existing chain
	for dec.More() {
		var existingParent parent

		err = dec.Decode(&existingParent)
		if err != nil {
			panic(err)
		}

		parentMatch := false
		// Find newParent in existingParents
		for nPIndex, newParent := range w.Chain.Parents {

			if newParent.Word == existingParent.Word {
				parentMatch = true

				uParent := parent{
					Word: newParent.Word,
				}

				// Do for every parent except end key
				if newParent.Word != instructions.EndKey {
					// Do for child
					// combine values and set into updatedChain
					for _, existingChild := range existingParent.Children {
						childMatch := false

						for nCIndex, newChild := range newParent.Children {

							if newChild.Word == existingChild.Word {
								childMatch = true

								uParent.Children = append(uParent.Children, child{
									Word:  newChild.Word,
									Value: newChild.Value + existingChild.Value,
								})

								newParent.removeChild(nCIndex)
								break
							}
						}

						if !childMatch {
							uParent.Children = append(uParent.Children, existingChild)
						}
					}

					uParent.Children = append(uParent.Children, newParent.Children...)
				}

				// Do for every parent except start key
				if newParent.Word != instructions.StartKey {
					// Do for grandparent
					// combine values and set into updatedChain
					for _, existingGrandparent := range existingParent.Grandparents {
						grandparentMatch := false

						for nPIndex, newGrandparent := range newParent.Grandparents {

							if newGrandparent.Word == existingGrandparent.Word {
								grandparentMatch = true

								uParent.Grandparents = append(uParent.Grandparents, grandparent{
									Word:  newGrandparent.Word,
									Value: newGrandparent.Value + existingGrandparent.Value,
								})

								newParent.removeGrandparent(nPIndex)
								break
							}
						}

						if !grandparentMatch {
							uParent.Grandparents = append(uParent.Grandparents, existingGrandparent)
						}
					}

					uParent.Grandparents = append(uParent.Grandparents, newParent.Grandparents...)
				}

				if err := enc.AddEntry(uParent); err != nil {
					panic(err)
				}

				w.Chain.removeParent(nPIndex)
				break
			}
		}

		if !parentMatch {
			if err := enc.AddEntry(existingParent); err != nil {
				panic(err)
			}
		}
	}

	// Add every new parent that is left over
	for _, nParent := range w.Chain.Parents {
		if err := enc.AddEntry(nParent); err != nil {
			panic(err)
		}
	}

	// Close the new file encoder
	if err := enc.CloseEncoder(); err != nil {
		panic(err)
	}

	// Verify if new file size is larger than old file size
	err = compareSizes(f, fN)
	if err != nil && errCh != nil {
		errCh <- err
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
}
