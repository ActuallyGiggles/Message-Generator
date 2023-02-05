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

	// Body needs to be last because it will write all existing parents into body file
	var wg2 sync.WaitGroup
	wg2.Add(2)
	go w.writeHead(errCh, &wg2)
	go w.writeTail(errCh, &wg2)
	wg2.Wait()
	w.writeBody(errCh)

	w.Chain.Parents = nil
	w.Intake = 0
}

func (w *worker) writeHead(errCh chan error, wg2 *sync.WaitGroup) {
	defer wg2.Done()
	defaultPath := "./markov-chains/" + w.Name + "_head.json"
	newPath := "./markov-chains/" + w.Name + "_head_new.json"

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
			chainData, _ := json.MarshalIndent(w.Chain.Parents[0].Children, "", "    ")
			w.Chain.removeParent(0)
			f.Write(chainData)
			f.Close()
			return
		} else {
			fN, err := os.Create(newPath)
			if err != nil {
				panic(err)
			}

			var enc encode

			if err = StartEncoder(&enc, fN); err != nil {
				panic(err)
			}

			for i, parent := range *&w.Chain.Parents {
				if parent.Word == instructions.StartKey {

					for dec.More() {
						var existingChild child

						err := dec.Decode(&existingChild)
						if err != nil {
							panic(err)
						}

						childMatch := false

						for j, newChild := range *&parent.Children {

							if newChild.Word == existingChild.Word {
								childMatch = true

								if err := enc.AddEntry(child{
									Word:  newChild.Word,
									Value: newChild.Value + existingChild.Value,
								}); err != nil {
									panic(err)
								}

								parent.removeChild(j)
								continue
							}
						}

						if !childMatch {
							if err := enc.AddEntry(existingChild); err != nil {
								panic(err)
							}
						}
					}

					for _, c := range *&parent.Children {
						if err := enc.AddEntry(c); err != nil {
							panic(err)
						}
					}

					w.Chain.removeParent(i)
				}
			}

			if err := enc.CloseEncoder(); err != nil {
				panic(err)
			}

			err = compareSizes(f, fN)
			if err != nil {
				errCh <- err
			}

			err = fN.Close()
			if err != nil {
				panic(err)
			}
		}
	}

	err = f.Close()
	if err != nil {
		panic(err)
	}

	removeAndRename(defaultPath, newPath)
}

func (w *worker) writeTail(errCh chan error, wg2 *sync.WaitGroup) {
	defer wg2.Done()
	defaultPath := "./markov-chains/" + w.Name + "_tail.json"
	newPath := "./markov-chains/" + w.Name + "_tail_new.json"

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
			chainData, _ := json.MarshalIndent(w.Chain.Parents[0].Grandparents, "", "    ")
			w.Chain.removeParent(0)
			f.Write(chainData)
			f.Close()
			return
		} else {
			fN, err := os.Create(newPath)
			if err != nil {
				panic(err)
			}

			var enc encode

			if err = StartEncoder(&enc, fN); err != nil {
				panic(err)
			}

			// For every parent in parents
			for i, parent := range *&w.Chain.Parents {

				// If the parent is an end key
				if parent.Word == instructions.EndKey {

					// Look through all of its grandparents
					for dec.More() {
						var existingGrandparent grandparent

						err := dec.Decode(&existingGrandparent)
						if err != nil {
							panic(err)
						}

						grandparentMatch := false

						// For every new grandparent in new grandparents
						for j, newGrandparent := range *&parent.Grandparents {

							// If this new grandparent matches the existing/old grandparent
							if newGrandparent.Word == existingGrandparent.Word {
								grandparentMatch = true

								// Create a new grandparent with combined values of the previous two and write to a new file
								if err := enc.AddEntry(grandparent{
									Word:  newGrandparent.Word,
									Value: newGrandparent.Value + existingGrandparent.Value,
								}); err != nil {
									panic(err)
								}

								// Remove that old grandparent
								parent.removeGrandparent(j)
								continue
							}
						}

						// If there is no new grandparent that matches the old grandparent, just add the old one to the new file
						if !grandparentMatch {
							if err := enc.AddEntry(existingGrandparent); err != nil {
								panic(err)
							}
						}
					}

					// Now, for every new grandparent that wasn't matched, also add it to the new file
					for _, g := range *&parent.Grandparents {
						if err := enc.AddEntry(g); err != nil {
							panic(err)
						}
					}

					// Finally, remove the whole new parent that was being worked on from the new chain
					w.Chain.removeParent(i)
				}
			}

			if err := enc.CloseEncoder(); err != nil {
				panic(err)
			}

			err = compareSizes(f, fN)
			if err != nil {
				errCh <- err
			}

			err = fN.Close()
			if err != nil {
				panic(err)
			}
		}
	}

	err = f.Close()
	if err != nil {
		panic(err)
	}

	removeAndRename(defaultPath, newPath)
}

func (w *worker) writeBody(errCh chan error) {
	defaultPath := "./markov-chains/" + w.Name + "_body.json"
	newPath := "./markov-chains/" + w.Name + "_body_new.json"

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
			chainData, _ := json.MarshalIndent(w.Chain.Parents, "", "    ")
			f.Write(chainData)
			f.Close()
			return
		} else {
			fN, err := os.Create(newPath)
			if err != nil {
				panic(err)
			}

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
				for nPIndex, newParent := range *&w.Chain.Parents {

					if newParent.Word == instructions.StartKey || newParent.Word == instructions.EndKey {
						continue
					}

					if newParent.Word == existingParent.Word {
						parentMatch = true

						uParent := parent{
							Word: newParent.Word,
						}

						// Do for child
						// combine values and set into updatedChain
						for _, existingChild := range *&existingParent.Children {
							childMatch := false

							for nCIndex, newChild := range *&newParent.Children {

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

						for _, newChild := range newParent.Children {
							uParent.Children = append(uParent.Children, newChild)
						}

						// Do for grandparent
						// combine values and set into updatedChain
						for _, existingGrandparent := range *&existingParent.Grandparents {
							grandparentMatch := false

							for nPIndex, newGrandparent := range *&newParent.Grandparents {

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

						for _, newGrandparent := range newParent.Grandparents {
							uParent.Grandparents = append(uParent.Grandparents, newGrandparent)
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

			for _, nParent := range w.Chain.Parents {
				if err := enc.AddEntry(nParent); err != nil {
					panic(err)
				}
			}

			if err := enc.CloseEncoder(); err != nil {
				panic(err)
			}

			err = compareSizes(f, fN)
			if err != nil {
				errCh <- err
			}

			fN.Close()
		}
	}

	// Close the chain file
	err = f.Close()
	if err != nil {
		panic(err)
	}

	removeAndRename(defaultPath, newPath)
}
