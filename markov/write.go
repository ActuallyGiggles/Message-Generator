package markov

import (
	"encoding/json"
	"fmt"
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

func writeLoop() {
	if !busy.TryLock() {
		return
	}

	defer duration(track("writing duration"))

	var wg sync.WaitGroup
	wg.Add(len(workerMap))

	for _, w := range workerMap {
		go w.writeAllPerChain(&wg)
	}

	wg.Wait()

	busy.Unlock()
	saveStats()
	stats.NextWriteTime = time.Now().Add(writeInterval)
}

func (w *worker) writeAllPerChain(wg *sync.WaitGroup) {
	w.ChainMx.Lock()
	defer w.ChainMx.Unlock()
	defer wg.Done()

	if len(w.Chain.Parents) == 0 {
		return
	}

	// Find new peak intake chain
	if w.Intake > stats.PeakChainIntake.Amount {
		stats.PeakChainIntake.Chain = w.Name
		stats.PeakChainIntake.Amount = w.Intake
		stats.PeakChainIntake.Time = time.Now()
	}

	// As of 1/26/23, the head would be first, taking away the start key in the chain.
	// Then, the body would take everything else with it, leaving no end key for the tail.
	// Currently patched in body with an if word == start or end key: continue.
	w.writeHead()
	w.writeBody()
	w.writeTail()

	w.Chain.Parents = nil
	w.Intake = 0
}

func (w *worker) writeHead() {
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

			fN.Close()
		}
	}

	var triedToClose int
tryClose:
	err = f.Close()
	if err != nil {
		time.Sleep(5 * time.Second)
		if triedToClose < 25 {
			fmt.Println("attempting to close:", defaultPath, ", attempt #: ", triedToClose)
			triedToClose++
			goto tryClose
		}
	}

	err = os.Remove(defaultPath)
	if err != nil {
		panic(err)
	}

	err = os.Rename(newPath, defaultPath)
	if err != nil {
		panic(err)
	}
}

func (w *worker) writeBody() {
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

			StartEncoder(&enc, fN)

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
			fN.Close()
		}
	}

	// Close the chain file
	var triedToClose int
tryClose:
	err = f.Close()
	if err != nil {
		time.Sleep(5 * time.Second)
		if triedToClose < 25 {
			fmt.Println("attempting to close:", defaultPath, ", attempt #: ", triedToClose)
			triedToClose++
			goto tryClose
		}
	}

	err = os.Remove(defaultPath)
	if err != nil {
		panic(err)
	}

	err = os.Rename(newPath, defaultPath)
	if err != nil {
		panic(err)
	}
}

func (w *worker) writeTail() {
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

			for i, parent := range *&w.Chain.Parents {
				if parent.Word == instructions.EndKey {

					for dec.More() {
						var existingGrandparent grandparent

						err := dec.Decode(&existingGrandparent)
						if err != nil {
							panic(err)
						}

						grandparentMatch := false

						for j, newGrandparent := range *&parent.Grandparents {

							if newGrandparent.Word == existingGrandparent.Word {
								grandparentMatch = true

								if err := enc.AddEntry(grandparent{
									Word:  newGrandparent.Word,
									Value: newGrandparent.Value + existingGrandparent.Value,
								}); err != nil {
									panic(err)
								}

								parent.removeGrandparent(j)
								continue
							}
						}

						if !grandparentMatch {
							if err := enc.AddEntry(existingGrandparent); err != nil {
								panic(err)
							}
						}
					}

					for _, g := range *&parent.Grandparents {
						if err := enc.AddEntry(g); err != nil {
							panic(err)
						}
					}

					w.Chain.removeParent(i)
				}
			}

			if err := enc.CloseEncoder(); err != nil {
				panic(err)
			}
			fN.Close()
		}
	}

	var triedToClose int
tryClose:
	err = f.Close()
	if err != nil {
		time.Sleep(5 * time.Second)
		if triedToClose < 25 {
			fmt.Println("attempting to close:", defaultPath, ", attempt #: ", triedToClose)
			triedToClose++
			goto tryClose
		}
	}

	err = os.Remove(defaultPath)
	if err != nil {
		panic(err)
	}

	err = os.Rename(newPath, defaultPath)
	if err != nil {
		panic(err)
	}
}
