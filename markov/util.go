package markov

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"
)

func debugLog(v ...any) {
	if instructions.Debug {
		log.Println(v...)
	}
}

// chains gets a list of current chains found in the directory.
func chains(isInit bool, includeFullName bool) (chains []string) {
	files, err := os.ReadDir("./markov-chains/")
	if err != nil {
		return chains
	}

	for _, file := range files {
		if file.Name() == "stats" {
			continue
		}

		if strings.Contains(file.Name(), "_new") && isInit {
			err = os.Remove("./markov-chains/" + file.Name())
			if err != nil {
				panic(err)
			}
			continue
		}

		if includeFullName {
			chains = append(chains, file.Name()[:len(file.Name())-5])
			continue
		}

		chains = append(chains, file.Name()[:len(file.Name())-10])
	}

	return chains
}

// DoesChainExist returns if a chain exists or not.
func DoesChainExist(name string) (w *worker, exists bool) {
	for _, c := range chains(false, false) {
		if c == name {
			for _, w := range workerMap {
				if w.Name == c {
					return w, true
				}
			}
		}
	}
	return nil, false
}

func PrettyPrint(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}
}

// CurrentWorkers returns the names of all workers that have been made.
func CurrentWorkers() []string {
	workerMapMx.Lock()
	var s []string
	for chain := range workerMap {
		s = append(s, chain)
	}
	workerMapMx.Unlock()
	return s
}

// NextWriteTime returns what time the next write cycle will happen.
func NextWriteTime() time.Time {
	return stats.NextWriteTime
}

// PeakIntake returns the highest intake across all workers per session and at what time it happened.
func PeakIntake() PeakIntakeStruct {
	return stats.PeakChainIntake
}

// weightedRandom used weighted random selection to return one of the supplied
// choices.  Weights of 0 are never selected.  All other weight values are
// relative.  E.g. if you have two choices both weighted 3, they will be
// returned equally often; and each will be returned 3 times as often as a
// choice weighted 1.
func weightedRandom(choices []Choice) (string, error) {
	// Based on this algorithm:
	//     http://eli.thegreenplace.net/2010/01/22/weighted-random-generation-in-python/
	sum := 0
	for _, c := range choices {
		sum += c.Weight
	}
	r, err := randomNumber(0, sum)
	if err != nil {
		return "", err
	}
	for _, c := range choices {
		r -= c.Weight
		if r < 0 {
			return c.Word, nil
		}
	}
	err = errors.New("internal error - code should not reach this point")
	return "", err
}

func createFolders() {
	_, err := os.Stat("./markov-chains")
	if os.IsNotExist(err) {
		err := os.MkdirAll("./markov-chains", 0755)
		if err != nil {
			panic(err)
		}
	}

	_, err = os.Stat("./markov-chains/stats")
	if os.IsNotExist(err) {
		err := os.MkdirAll("./markov-chains/stats", 0755)
		if err != nil {
			panic(err)
		}
	}
}

func (p *parent) removeGrandparent(i int) {
	p.Grandparents[i] = p.Grandparents[len(p.Grandparents)-1]
	p.Grandparents = p.Grandparents[:len(p.Grandparents)-1]
}

func (p *parent) removeChild(i int) {
	p.Children[i] = p.Children[len(p.Children)-1]
	p.Children = p.Children[:len(p.Children)-1]
}

func (c *chain) removeParent(i int) {
	c.Parents[i] = c.Parents[len(c.Parents)-1]
	c.Parents = c.Parents[:len(c.Parents)-1]
}

// randomNumber returns a random integer in the range from min to max.
func randomNumber(min, max int) (int, error) {
	var result int
	switch {
	case min > max:
		// Fail with error
		return result, errors.New("min cannot be greater than max")
	case max == min:
		result = max
	case max > min:
		maxRand := max - min
		b, err := rand.Int(rand.Reader, big.NewInt(int64(maxRand)))
		if err != nil {
			return result, err
		}
		result = min + int(b.Int64())
	}
	return result, nil
}

func StartEncoder(enc *encode, file *os.File) (err error) {
	if _, err = file.Write([]byte{'['}); err != nil {
		return err
	}

	encoder := json.NewEncoder(file)

	enc.Encoder = encoder
	enc.File = file

	return nil
}

func (enc *encode) AddEntry(entry interface{}) (err error) {
	if enc.ContinuedEntry {
		if _, err = enc.File.Write([]byte{','}); err != nil {
			return err
		}
	}

	if err := enc.Encoder.Encode(entry); err != nil {
		return err
	}

	enc.ContinuedEntry = true

	return nil
}

func (enc *encode) CloseEncoder() (err error) {
	if _, err = enc.File.Write([]byte{']'}); err != nil {
		panic(err)
	}

	return nil
}

// IsBusy returns false if not writing, zipping, or defluffing. Returns true otherwise.
func IsBusy() bool {
	if !busy.TryLock() {
		return true
	}
	busy.Unlock()

	return false
}

func removeAndRename(defaultPath, newPath string) {
	err := os.Remove(defaultPath)
	if err != nil {
		panic(err)
	}

	err = os.Rename(newPath, defaultPath)
	if err != nil {
		panic(err)
	}
}

// compareSizes will compare the size of an old write file and a new write file and return an error if old is bigger than new. lol
func compareSizes(old, new *os.File) error {
	oldStats, err := old.Stat()
	if err != nil {
		panic(err)
	}

	newStats, err := new.Stat()
	if err != nil {
		panic(err)
	}

	oldSize := oldStats.Size()
	newSize := newStats.Size()

	if newSize < oldSize {
		return errors.New("Old file size is bigger than the new file size!\n" + old.Name() + ": " + ByteCountSI(oldSize) + "\n" + new.Name() + ": " + ByteCountSI(newSize))
	}

	return nil
}

func ByteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}
