package histogram

type pair struct {
	Key   string
	Value uint
}

type pairlist []pair

func (pl pairlist) Len() int { return len(pl) }

func (pl pairlist) Less(i, j int) bool {
	if pl[i].Value == pl[j].Value {
		return pl[i].Key < pl[j].Key
	}
	return pl[i].Value < pl[j].Value
}

func (pl pairlist) Swap(i, j int) {
	pl[i], pl[j] = pl[j], pl[i]
}

// NewPairList returns a pairlist containing pairs (key, value) from the give map
func NewPairList(m map[string]uint) pairlist {
	p := make(pairlist, len(m))

	i := 0
	for k, v := range m {
		p[i] = pair{k, v}
		i++
	}

	return p
}

// TotalValues returns the sum of values across all pairs in the PairList
func (pl *pairlist) TotalValues() uint {
	totalValue := uint(0)
	for _, p := range *pl {
		totalValue += p.Value
	}

	return totalValue
}
