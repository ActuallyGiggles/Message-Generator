package markov

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Out takes output instructions and returns an output and error.
func Out(oi OutputInstructions) (output string, err error) {
	name := oi.Chain
	method := oi.Method
	target := oi.Target

	defer duration(track("output duration"))

	if exists := DoesChainExist(name); !exists {
		return "", errors.New("Chain '" + name + "' is not found in directory.")
	}

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
	}

	if err == nil {
		stats.TotalOutputs++
		stats.SessionOutputs++
	}

	return output, err
}

func likelyBeginning(name string) (output string, err error) {
	var child string

	parentWord, err := getStartWord(name)
	if err != nil {
		return "", err
	}

	output = parentWord

	for true {
		f, err := os.Open("./markov-chains/" + name + "_body.json")
		if err != nil {
			return "", err
		}
		defer f.Close()

		dec := json.NewDecoder(f)
		_, err = dec.Token()
		if err != nil {
			panic(err)
		}

		parentExists := false
		for dec.More() {
			var currentParent parent

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
		}

		if !parentExists {
			return output, errors.New(fmt.Sprintf("parent %s does not exist in chain %s", parentWord, name))
		}
	}

	return output, nil
}

func likelyEnding(name string) (output string, err error) {
	var grandparent string

	parentWord, err := getEndWord(name)
	if err != nil {
		return "", err
	}

	output = parentWord

	for true {
		f, err := os.Open("./markov-chains/" + name + "_body.json")
		if err != nil {
			return "", err
		}
		defer f.Close()

		dec := json.NewDecoder(f)
		_, err = dec.Token()
		if err != nil {
			panic(err)
		}

		parentExists := false
		for dec.More() {
			var currentParent parent

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
		}

		if !parentExists {
			return output, errors.New(fmt.Sprintf("parent %s does not exist in chain %s", parentWord, name))
		}
	}

	return output, nil
}

func targetedBeginning(name, target string) (output string, err error) {
	var parentWord string
	var childChosen string

	var initialList []Choice

	f, err := os.Open("./markov-chains/" + name + "_head.json")
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
		var currentParent child

		err = dec.Decode(&currentParent)
		if err != nil {
			panic(err)
		}

		numberOfTargets := len(strings.Split(target, instructions.SeparationKey))

		if numberOfTargets == 1 {
			potentialParentSplit := strings.Split(currentParent.Word, " ")
			if potentialParentSplit[0] == target {
				initialList = append(initialList, Choice{
					Word:   currentParent.Word,
					Weight: currentParent.Value,
				})
			} else if currentParent.Word == target {
				initialList = append(initialList, Choice{
					Word:   currentParent.Word,
					Weight: currentParent.Value,
				})
			}
		}

		if numberOfTargets == 2 {
			if currentParent.Word == target {
				initialList = append(initialList, Choice{
					Word:   currentParent.Word,
					Weight: currentParent.Value,
				})
			}
		}

		if numberOfTargets > 2 {
			return "", errors.New(fmt.Sprintf("You can only have 1 or 2 targets under the current system. Number of current targets: %d", numberOfTargets))
		}
	}

	if len(initialList) <= 0 {
		return "", errors.New(fmt.Sprintf("%s does not contain parents that match: %s", name, target))
	}

	parentWord, err = weightedRandom(initialList)
	if err != nil {
		return "", err
	}
	output = parentWord

	for true {
		f, err := os.Open("./markov-chains/" + name + "_body.json")
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
		for dec.More() {
			var currentParent parent

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
		}

		if !parentExists {
			return output, errors.New(fmt.Sprintf("parent %s does not exist in chain %s", parentWord, name))
		}
	}

	return output, nil
}

func targetedEnding(name, target string) (output string, err error) {
	var parentWord string
	var grandparentChosen string

	var initialList []Choice

	f, err := os.Open("./markov-chains/" + name + "_tail.json")
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
		var currentParent grandparent

		err = dec.Decode(&currentParent)
		if err != nil {
			panic(err)
		}

		numberOfTargets := len(strings.Split(target, instructions.SeparationKey))

		if numberOfTargets == 1 {
			potentialParentSplit := strings.Split(currentParent.Word, " ")
			if potentialParentSplit[0] == target {
				initialList = append(initialList, Choice{
					Word:   currentParent.Word,
					Weight: currentParent.Value,
				})
			} else if currentParent.Word == target {
				initialList = append(initialList, Choice{
					Word:   currentParent.Word,
					Weight: currentParent.Value,
				})
			}
		}

		if numberOfTargets == 2 {
			if currentParent.Word == target {
				initialList = append(initialList, Choice{
					Word:   currentParent.Word,
					Weight: currentParent.Value,
				})
			}
		}

		if numberOfTargets > 2 {
			return "", errors.New(fmt.Sprintf("You can only have 1 or 2 targets under the current system. Number of current targets: %d", numberOfTargets))
		}
	}

	if len(initialList) <= 0 {
		return "", errors.New(fmt.Sprintf("%s does not contain parents that match: %s", name, target))
	}

	parentWord, err = weightedRandom(initialList)
	if err != nil {
		return "", err
	}
	parentSplit := strings.Split(parentWord, " ")
	if len(parentSplit) == 1 {
		output = parentSplit[0] + " "
	} else {
		output = parentSplit[1] + " "
	}

	for true {
		f, err := os.Open("./markov-chains/" + name + "_body.json")
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
		for dec.More() {
			var currentParent parent

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
		}

		if !parentExists {
			return output, errors.New(fmt.Sprintf("parent %s does not exist in chain %s", parentWord, name))
		}
	}

	return output, nil
}

func targetedMiddle(name, target string) (output string, err error) {
	var parentWord string
	var childChosen string
	var grandparentChosen string

	var initialList []Choice

	f, err := os.Open("./markov-chains/" + name + "_body.json")
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
		var currentParent parent

		err = dec.Decode(&currentParent)
		if err != nil {
			panic(err)
		}

		numberOfTargets := len(strings.Split(target, instructions.SeparationKey))

		if numberOfTargets == 1 {
			potentialParentSplit := strings.Split(currentParent.Word, " ")
			if potentialParentSplit[0] == target {
				goto addParent
			} else if currentParent.Word == target {
				goto addParent
			} else {
				continue
			}
		}

		if numberOfTargets == 2 {
			if currentParent.Word == target {
				goto addParent
			} else {
				continue
			}
		}

		if numberOfTargets > 2 {
			return "", errors.New(fmt.Sprintf("You can only have 1 or 2 targets under the current system. Number of current targets: %d", numberOfTargets))
		}

	addParent:
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

	if len(initialList) <= 0 {
		return "", errors.New(fmt.Sprintf("%s does not contain parents that match: %s", name, target))
	}

	parentWord, err = weightedRandom(initialList)
	if err != nil {
		return "", err
	}
	output = parentWord
	originalParentWord := parentWord

	var forwardComplete bool

goThroughBody:
	f, err = os.Open("./markov-chains/" + name + "_body.json")
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
		var currentParent parent

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
	}

	if !parentExists {
		return output, errors.New(fmt.Sprintf("parent %s does not exist in chain %s", parentWord, name))
	}

	return "", errors.New("Internal error - code should not reach this point - TargetedMiddle")
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

func pickRandomParent(parents []string) (parent string) {
	parent = pickRandomFromSlice(parents)

	return parent
}

func getStartWord(name string) (phrase string, err error) {
	var sum int

	f, err := os.Open("./markov-chains/" + name + "_head.json")
	if err != nil {
		return "", err
	}

	dec := json.NewDecoder(f)
	_, err = dec.Token()
	if err != nil {
		panic(err)
	}

	for dec.More() {
		var child child

		err = dec.Decode(&child)
		if err != nil {
			fmt.Println(name)
			fmt.Println(child)
			panic(err)
		}

		sum += child.Value
	}

	f.Close()

	r, err := randomNumber(0, sum)
	if err != nil {
		return "", err
	}

	f, err = os.Open("./markov-chains/" + name + "_head.json")
	if err != nil {
		return "", err
	}
	defer f.Close()

	dec = json.NewDecoder(f)
	_, err = dec.Token()
	if err != nil {
		panic(err)
	}

	for dec.More() {
		var child child

		err = dec.Decode(&child)
		if err != nil {
			fmt.Println(name)
			fmt.Println(child)
			panic(err)
		}

		r -= child.Value

		if r < 0 {
			return child.Word, nil
		}
	}

	return "", errors.New("Internal error - code should not reach this point")
}

func getEndWord(name string) (phrase string, err error) {
	var sum int

	f, err := os.Open("./markov-chains/" + name + "_tail.json")
	if err != nil {
		return "", err
	}

	dec := json.NewDecoder(f)
	_, err = dec.Token()
	if err != nil {
		panic(err)
	}

	for dec.More() {
		var grandparent grandparent

		err = dec.Decode(&grandparent)
		if err != nil {
			fmt.Println(name)
			fmt.Println(grandparent)
			panic(err)
		}

		sum += grandparent.Value
	}

	f.Close()

	r, err := randomNumber(0, sum)
	if err != nil {
		return "", err
	}

	f, err = os.Open("./markov-chains/" + name + "_tail.json")
	if err != nil {
		return "", err
	}
	defer f.Close()

	dec = json.NewDecoder(f)
	_, err = dec.Token()
	if err != nil {
		panic(err)
	}

	for dec.More() {
		var grandparent grandparent

		err = dec.Decode(&grandparent)
		if err != nil {
			fmt.Println(name)
			fmt.Println(grandparent)
			panic(err)
		}

		r -= grandparent.Value

		if r < 0 {
			return grandparent.Word, nil
		}
	}

	return "", errors.New("Internal error - code should not reach this point")
}
