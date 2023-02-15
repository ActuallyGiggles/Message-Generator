package markov

import (
	"strings"
)

// In adds an entry into a specific chain.
func In(chainName string, content string) {
	if content == "" || len(content) <= 0 {
		return
	}

	workerMapMx.Lock()
	w, ok := workerMap[chainName]
	workerMapMx.Unlock()

	if !ok {
		w = newWorker(chainName)
	}

	w.ChainMx.Lock()
	w.addInput(content)
	w.ChainMx.Unlock()
}

func (w *worker) addInput(content string) {
	slice := prepareContentForChainProcessing(content)
	w.Chain.extractHead(w.Name, slice)
	w.Chain.extractBody(w.Name, slice)
	w.Chain.extractTail(w.Name, slice)

	w.Intake++
	stats.TotalInputs++
	stats.SessionInputs++
}

func prepareContentForChainProcessing(content string) []string {
	var returnSlice []string
	returnSlice = append(returnSlice, instructions.StartKey)
	slice := strings.Split(content, instructions.SeparationKey)
	for i := 0; i <= len(slice)-1; i += 3 {
		if len(slice)-1 > i+1 {
			returnSlice = append(returnSlice, slice[i]+instructions.SeparationKey+slice[i+1]+instructions.SeparationKey+slice[i+2])
		} else if len(slice)-1 > i {
			returnSlice = append(returnSlice, slice[i]+instructions.SeparationKey+slice[i+1])
		} else {
			returnSlice = append(returnSlice, slice[i])
		}
	}
	returnSlice = append(returnSlice, instructions.EndKey)
	return returnSlice
}

func (c *chain) extractHead(name string, slice []string) {
	start := slice[0]
	next := slice[1]

	parentExists := false
	for i := 0; i < len(c.Parents); i++ {
		parent := &c.Parents[i]
		if parent.Word == start {
			parentExists = true

			childExists := false
			for i := 0; i < len(parent.Children); i++ {
				child := &parent.Children[i]
				if child.Word == next {
					childExists = true
					child.Value += 1
				}
			}

			if !childExists {
				child := child{
					Word:  next,
					Value: 1,
				}
				parent.Children = append(parent.Children, child)
			}
		}
	}

	if !parentExists {
		var children []child
		child := child{
			Word:  next,
			Value: 1,
		}
		children = append(children, child)
		parent := parent{
			Word:     start,
			Children: children,
		}
		c.Parents = append(c.Parents, parent)
	}
}

func (c *chain) extractBody(name string, slice []string) {
	for i := 0; i < len(slice)-2; i++ {
		current := slice[i+1]
		next := slice[i+2]
		previous := slice[i]

		parentExists := false
		for i := 0; i < len(c.Parents); i++ {
			parent := &c.Parents[i]
			if parent.Word == current {
				parentExists = true

				// Deal with child
				childExists := false
				for i := 0; i < len(parent.Children); i++ {
					child := &parent.Children[i]
					if child.Word == next {
						childExists = true
						child.Value += 1
					}
				}

				if !childExists {
					child := child{
						Word:  next,
						Value: 1,
					}
					parent.Children = append(parent.Children, child)
				}

				// Deal with grandparent
				grandparentExists := false
				for i := 0; i < len(parent.Grandparents); i++ {
					grandparent := &parent.Grandparents[i]
					if grandparent.Word == previous {
						grandparentExists = true
						grandparent.Value += 1
					}
				}

				if !grandparentExists {
					grandparent := grandparent{
						Word:  previous,
						Value: 1,
					}
					parent.Grandparents = append(parent.Grandparents, grandparent)
				}
			}
		}

		if !parentExists {
			// Deal with child
			var children []child
			child := child{
				Word:  next,
				Value: 1,
			}
			children = append(children, child)

			// Deal with grandparent
			var grandparents []grandparent
			grandparent := grandparent{
				Word:  previous,
				Value: 1,
			}
			grandparents = append(grandparents, grandparent)

			// Add all to parent
			parent := parent{
				Word:         current,
				Children:     children,
				Grandparents: grandparents,
			}
			c.Parents = append(c.Parents, parent)
		}
	}
}

func (c *chain) extractTail(name string, slice []string) {
	end := slice[len(slice)-1]
	previous := slice[len(slice)-2]

	parentExists := false
	for i := 0; i < len(c.Parents); i++ {
		parent := &c.Parents[i]
		if parent.Word == end {
			parentExists = true

			grandparentExists := false
			for i := 0; i < len(parent.Grandparents); i++ {
				grandparent := &parent.Grandparents[i]
				if grandparent.Word == previous {
					grandparentExists = true
					grandparent.Value += 1
				}
			}

			if !grandparentExists {
				grandparent := grandparent{
					Word:  previous,
					Value: 1,
				}
				parent.Grandparents = append(parent.Grandparents, grandparent)
			}
		}
	}

	if !parentExists {
		var grandparents []grandparent
		grandparent := grandparent{
			Word:  previous,
			Value: 1,
		}
		grandparents = append(grandparents, grandparent)
		parent := parent{
			Word:         end,
			Grandparents: grandparents,
		}
		c.Parents = append(c.Parents, parent)
	}
}
