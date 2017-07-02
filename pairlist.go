package main

type pair struct {
	key   string
	value uint
}

type Pairlist []pair

func (pl Pairlist) Len() int { return len(pl) }

func (pl Pairlist) Less(i, j int) bool {
	if pl[i].value == pl[j].value {
		return pl[i].key < pl[j].key
	}
	return pl[i].value < pl[j].value
}

func (pl Pairlist) Swap(i, j int) {
	pl[i], pl[j] = pl[j], pl[i]
}

// NewPairList returns a Pairlist containing pairs (key, value) from the give map
func NewPairList(m map[string]uint) Pairlist {
	p := make(Pairlist, len(m))

	i := 0
	for k, v := range m {
		p[i] = pair{k, v}
		i++
	}

	return p
}

// TotalValues returns the sum of values across all pairs in the PairList
func (pl *Pairlist) TotalValues() uint {
	totalValue := uint(0)
	for _, p := range *pl {
		totalValue += p.value
	}

	return totalValue
}
