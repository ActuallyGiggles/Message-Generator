package markov

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	previousChainSums = make(map[string]int64)
)

// Out takes output instructions and returns an output and error.
// If a chain has less than 50 parent values, it will act as if the chain is not found in the directory.
func Out(oi OutputInstructions) (output string, err error) {
	name := oi.Chain
	method := oi.Method
	target := oi.Target

	if !DoesChainFileExist(name) {
		return "", errors.New("chain [" + name + "] is not found in directory")
	}

	exists, w := doesWorkerExist(name)
	if exists {
		if !w.ChainMx.TryLock() {
			return
		}
		defer w.ChainMx.Unlock()
	}

	defer duration(track("output duration"))

	switch method {
	case "LikelyBeginning":
		output, err = likelyBeginning(name)
	case "LikelyEnding":
		output, err = likelyEnding(name)
	case "TargetedBeginning":
		output, err = targetedBeginning(name, target)
	case "TargetedEnding":
		output, err = targetedEnding(name, target)
	case "TargetedMiddle":
		output, err = targetedMiddle(name, target)
	case "RandomMiddle":
		output, err = randomMiddle(name)
	default:
		return "", errors.New("no correct method provided")
	}

	if err == nil {
		stats.TotalOutputs++
		stats.SessionOutputs++
	}

	return output, err
}

func likelyBeginning(name string) (output string, err error) {
	var child string
	var path = "./markov-chains/" + name + ".json"

	parentWord, err := getStartWord(name)
	if err != nil {
		return "", err
	}

	output = parentWord

	for {
		f, err := os.Open(path)
		if err != nil {
			return "", err
		}
		defer f.Close()

		dec := json.NewDecoder(f)
		_, err = dec.Token()
		if err != nil {
			return "", errors.New("EOF (via likelyBeginning) detected in " + path)
		}

		parentExists := false
		var currentParent parent
		for dec.More() {
			err = dec.Decode(&currentParent)
			if err != nil {
				fmt.Println(name)
				fmt.Println(currentParent)
				panic(err)
			}

			if currentParent.Word == parentWord {
				parentExists = true

				child = getNextWord(currentParent)

				if child == instructions.EndKey {
					return output, nil
				} else {
					output = output + instructions.SeparationKey + child

					parentWord = child
					continue
				}
			}

			currentParent = parent{}
		}

		if !parentExists {
			return output, fmt.Errorf("parent %s does not exist in chain %s", parentWord, name)
		}
	}
}

func likelyEnding(name string) (output string, err error) {
	var grandparent string
	var path = "./markov-chains/" + name + ".json"

	parentWord, err := getEndWord(name)
	if err != nil {
		return "", err
	}

	output = parentWord

	for {
		f, err := os.Open(path)
		if err != nil {
			return "", err
		}
		defer f.Close()

		dec := json.NewDecoder(f)
		_, err = dec.Token()
		if err != nil {
			return "", errors.New("EOF (via likelyEnding) detected in " + path)
		}

		parentExists := false
		var currentParent parent
		for dec.More() {
			err = dec.Decode(&currentParent)
			if err != nil {
				fmt.Println(name)
				fmt.Println(currentParent)
				panic(err)
			}

			if currentParent.Word == parentWord {
				parentExists = true

				grandparent = getPreviousWord(currentParent)

				if grandparent == instructions.StartKey {
					return output, nil
				} else {
					output = grandparent + instructions.SeparationKey + output

					parentWord = grandparent
					continue
				}
			}

			currentParent = parent{}
		}

		if !parentExists {
			return output, fmt.Errorf("parent %s does not exist in chain %s", parentWord, name)
		}
	}
}

func targetedBeginning(name, target string) (output string, err error) {
	var path = "./markov-chains/" + name + ".json"

	if target == "" {
		return "", errors.New("target is empty for TargetedBeginning")
	}

	if len(strings.Split(target, instructions.SeparationKey)) > 1 {
		return "", fmt.Errorf("you can only have 1 target")
	}

	var parentWord string
	var childChosen string
	var initialList []Choice

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	_, err = dec.Token()
	if err != nil {
		return "", errors.New("EOF (via targetedBeginning) detected in " + path)
	}

	var currentChild child
	for dec.More() {

		err = dec.Decode(&currentChild)
		if err != nil {
			panic(err)
		}

		if match, _ := regexp.MatchString("\\b"+target+"\\b", currentChild.Word); match {
			initialList = append(initialList, Choice{
				Word:   currentChild.Word,
				Weight: currentChild.Value,
			})
		}

		currentChild = child{}
	}

	if len(initialList) <= 0 {
		return "", fmt.Errorf("%s does not contain parents that match: %s", name, target)
	}

	parentWord, err = weightedRandom(initialList)
	if err != nil {
		return "", err
	}
	output = parentWord

	for {
		f, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		dec := json.NewDecoder(f)
		_, err = dec.Token()
		if err != nil {
			panic(err)
		}

		var currentParent parent
		parentExists := false
		for dec.More() {

			err = dec.Decode(&currentParent)
			if err != nil {
				panic(err)
			}

			if currentParent.Word == parentWord {
				parentExists = true
				childChosen = getNextWord(currentParent)

				if childChosen == instructions.EndKey {
					return output, nil
				} else {
					output = output + instructions.SeparationKey + childChosen
					parentWord = childChosen
					continue
				}
			}

			currentParent = parent{}
		}

		if !parentExists {
			return output, fmt.Errorf("parent %s does not exist in chain %s", parentWord, name)
		}
	}
}

func targetedEnding(name, target string) (output string, err error) {
	var path = "./markov-chains/" + name + ".json"

	if target == "" {
		return "", errors.New("target is empty for TargetedEnding")
	}

	if len(strings.Split(target, instructions.SeparationKey)) > 1 {
		return "", fmt.Errorf("you can only have 1 target")
	}

	var parentWord string
	var grandparentChosen string
	var initialList []Choice

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	_, err = dec.Token()
	if err != nil {
		return "", errors.New("EOF (via targetedEnding) detected in " + path)
	}

	var currentGrandparent grandparent

	for dec.More() {
		err = dec.Decode(&currentGrandparent)
		if err != nil {
			panic(err)
		}

		if match, _ := regexp.MatchString("\\b"+target+"\\b", currentGrandparent.Word); match {
			initialList = append(initialList, Choice{
				Word:   currentGrandparent.Word,
				Weight: currentGrandparent.Value,
			})
		}

		currentGrandparent = grandparent{}
	}

	if len(initialList) <= 0 {
		return "", fmt.Errorf("%s does not contain parents that match: %s", name, target)
	}

	parentWord, err = weightedRandom(initialList)
	if err != nil {
		return "", err
	}
	output = parentWord

	for {
		f, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		dec := json.NewDecoder(f)
		_, err = dec.Token()
		if err != nil {
			panic(err)
		}

		parentExists := false
		var currentParent parent
		for dec.More() {
			err = dec.Decode(&currentParent)
			if err != nil {
				panic(err)
			}

			if currentParent.Word == parentWord {
				parentExists = true
				grandparentChosen = getPreviousWord(currentParent)

				if grandparentChosen == instructions.StartKey {
					return output, nil
				} else {
					output = grandparentChosen + instructions.SeparationKey + output
					parentWord = grandparentChosen
					continue
				}
			}

			currentParent = parent{}
		}

		if !parentExists {
			return output, fmt.Errorf("parent %s does not exist in chain %s", parentWord, name)
		}
	}
}

func targetedMiddle(name, target string) (output string, err error) {
	var path = "./markov-chains/" + name + ".json"

	if target == "" {
		return "", errors.New("target is empty for TargetedMiddle")
	}

	if len(strings.Split(target, instructions.SeparationKey)) > 1 {
		return "", fmt.Errorf("you can only have 1 target")
	}

	var parentWord string
	var childChosen string
	var grandparentChosen string

	var initialList []Choice

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	_, err = dec.Token()
	if err != nil {
		return "", errors.New("EOF (via targetedMiddle) detected in " + path)
	}

	var currentParent parent
	for dec.More() {
		err = dec.Decode(&currentParent)
		if err != nil {
			panic(err)
		}

		if strings.Contains(currentParent.Word, instructions.SeparationKey+target+instructions.SeparationKey) {
			var totalWeight int

			for _, child := range currentParent.Children {
				totalWeight += child.Value
			}

			for _, grandparent := range currentParent.Grandparents {
				totalWeight += grandparent.Value
			}

			initialList = append(initialList, Choice{
				Word:   currentParent.Word,
				Weight: totalWeight,
			})
		}

		currentParent = parent{}
	}

	if len(initialList) <= 0 {
		return "", fmt.Errorf("%s does not contain parents that match: %s", name, target)
	}

	parentWord, err = weightedRandom(initialList)
	if err != nil {
		return "", err
	}
	output = parentWord
	originalParentWord := parentWord

	var forwardComplete bool

goThroughBody:
	f, err = os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	dec = json.NewDecoder(f)
	_, err = dec.Token()
	if err != nil {
		panic(err)
	}

	parentExists := false
	for dec.More() {
		err = dec.Decode(&currentParent)
		if err != nil {
			panic(err)
		}

		if currentParent.Word == parentWord {

			if !forwardComplete {
				parentExists = true
				childChosen = getNextWord(currentParent)

				if childChosen == instructions.EndKey {
					forwardComplete = true
					parentWord = originalParentWord
					goto goThroughBody
				} else {
					output = output + instructions.SeparationKey + childChosen

					parentWord = childChosen
					goto goThroughBody
				}
			}

			if forwardComplete {
				parentExists = true
				grandparentChosen = getPreviousWord(currentParent)

				if grandparentChosen == instructions.StartKey {
					return output, nil
				} else {
					output = grandparentChosen + instructions.SeparationKey + output

					parentWord = grandparentChosen
					goto goThroughBody
				}
			}
		}

		currentParent = parent{}
	}

	if !parentExists {
		return output, fmt.Errorf("parent %s does not exist in chain %s", parentWord, name)
	}

	return "", errors.New("internal error - code should not reach this point, most likely due to chain being defluffed or being empty  - TargetedMiddle - " + path)
}

func getStartWord(name string) (phrase string, err error) {
	var path = "./markov-chains/" + name + ".json"
	var sum int64
	if s, exists := previousChainSums[name]; !exists {
		calculatedSum, err := getSum(name)
		if err != nil {
			panic(err)
		}
		sum = calculatedSum
	} else {
		sum = s
		go getSum(name)
	}

	r, err := randomNumber(0, sum)
	if err != nil {
		return "", err
	}

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	_, err = dec.Token()
	if err != nil {
		panic(err)
	}
	for dec.More() {
		var parent parent

		err = dec.Decode(&parent)
		if err != nil {
			panic(err)
		}

		if parent.Word == instructions.StartKey {
			for _, child := range parent.Children {
				r -= int64(child.Value)

				if r < 0 {
					return child.Word, nil
				}
			}
		}
	}

	return "", errors.New("internal error - code should not reach this point, most likely due to chain being defluffed or being empty  - getStartWord - " + path)
}

func getEndWord(name string) (phrase string, err error) {
	var path = "./markov-chains/" + name + ".json"
	var sum int64

	if s, exists := previousChainSums[name]; !exists {
		calculatedSum, err := getSum(name)
		if err != nil {
			panic(err)
		}
		sum = calculatedSum
	} else {
		sum = s
		go getSum(name)
	}

	r, err := randomNumber(0, sum)
	if err != nil {
		return "", err
	}

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	_, err = dec.Token()
	if err != nil {
		panic(err)
	}

	for dec.More() {
		var parent parent

		err = dec.Decode(&parent)
		if err != nil {
			panic(err)
		}

		if parent.Word == instructions.EndKey {
			for _, grandparent := range parent.Grandparents {
				r -= int64(grandparent.Value)

				if r < 0 {
					return grandparent.Word, nil
				}
			}
		}
	}

	return "", errors.New("internal error - code should not reach this point, most likely due to chain being defluffed or being empty  - getEndWord - " + path)
}

func getNextWord(parent parent) (child string) {
	var wrS []Choice
	for _, word := range parent.Children {
		w := word.Word
		v := word.Value
		item := Choice{
			Word:   w,
			Weight: v,
		}
		wrS = append(wrS, item)
	}
	child, _ = weightedRandom(wrS)

	return child
}

func getPreviousWord(parent parent) (grandparent string) {
	var wrS []Choice
	for _, word := range parent.Grandparents {
		w := word.Word
		v := word.Value
		item := Choice{
			Word:   w,
			Weight: v,
		}
		wrS = append(wrS, item)
	}
	grandparent, _ = weightedRandom(wrS)

	return grandparent
}

func getSum(name string) (sum int64, err error) {
	var path = "./markov-chains/" + name + ".json"
	f, err := os.Open(path)
	if err != nil {
		return sum, err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	_, err = dec.Token()
	if err != nil {
		return sum, errors.New("EOF (via getRandomParent) detected in " + path)
	}

	var Parent parent
	for dec.More() {
		err = dec.Decode(&Parent)
		if err != nil {
			panic(err)
		}

		for _, child := range Parent.Children {
			sum += int64(child.Value)
		}

		Parent = parent{}
	}

	previousChainSums[name] = sum

	return sum, err
}

func getRandomParent(name string) (parentToReturn string, err error) {
	var path = "./markov-chains/" + name + ".json"
	var sum int64

	if s, exists := previousChainSums[name]; !exists {
		calculatedSum, err := getSum(name)
		if err != nil {
			panic(err)
		}
		sum = calculatedSum
	} else {
		sum = s
		go getSum(name)
	}

	r, err := randomNumber(0, sum)
	if err != nil {
		return
	}

	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	_, err = dec.Token()
	if err != nil {
		panic(err)
	}
	var Parent parent
	for dec.More() {
		err = dec.Decode(&Parent)
		if err != nil {
			panic(err)
		}

		for _, child := range Parent.Children {
			r -= int64(child.Value)

			if r < 0 {
				return Parent.Word, nil
			}
		}

		Parent = parent{}
	}

	return parentToReturn, errors.New("internal error - code should not reach this point, most likely due to chain being defluffed or being empty - getRandomParent - " + path)
}

func randomMiddle(name string) (output string, err error) {
	// Get a random parent
	originalParentWord, err := getRandomParent(name)
	if err != nil {
		return
	}

	output = originalParentWord
	parentWord := originalParentWord

	var path = "./markov-chains/" + name + ".json"
	var forwardComplete bool
	var childChosen string
	var grandparentChosen string

goThroughBody:
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	_, err = dec.Token()
	if err != nil {
		panic(err)
	}

	var currentParent parent
	parentExists := false
	for dec.More() {
		err = dec.Decode(&currentParent)
		if err != nil {
			panic(err)
		}

		if currentParent.Word == parentWord {
			if !forwardComplete {
				parentExists = true
				childChosen = getNextWord(currentParent)

				if childChosen == instructions.EndKey {
					forwardComplete = true
					parentWord = originalParentWord
					goto goThroughBody
				} else {
					output = output + instructions.SeparationKey + childChosen

					parentWord = childChosen
					goto goThroughBody
				}
			}

			if forwardComplete {
				parentExists = true
				grandparentChosen = getPreviousWord(currentParent)

				if grandparentChosen == instructions.StartKey {
					return output, nil
				} else {
					output = grandparentChosen + instructions.SeparationKey + output

					parentWord = grandparentChosen
					goto goThroughBody
				}
			}
		}

		currentParent = parent{}
	}

	if !parentExists {
		return output, fmt.Errorf("parent %s does not exist in chain %s", parentWord, name)
	}

	return "", errors.New("internal error - code should not reach this point, most likely due to chain being defluffed or being empty  - randomMiddle - " + path)
}
