package histogram

type Pair struct {
	Key   string
	Value uint
}

type Pairlist []Pair

func (pl Pairlist) Len() int { return len(pl) }

func (pl Pairlist) Less(i, j int) bool {
	if pl[i].Value == pl[j].Value {
		return pl[i].Key < pl[j].Key
	}
	return pl[i].Value < pl[j].Value
}

func (pl Pairlist) Swap(i, j int) {
	pl[i], pl[j] = pl[j], pl[i]
}

// NewPairList returns a Pairlist containing pairs (key, value) from the give map
func NewPairList(m map[string]uint) Pairlist {
	p := make(Pairlist, len(m))

	i := 0
	for k, v := range m {
		p[i] = Pair{k, v}
		i++
	}

	return p
}

// TotalValues returns the sum of values across all pairs in the PairList
func (pl *Pairlist) TotalValues() uint {
	totalValue := uint(0)
	for _, p := range *pl {
		totalValue += p.Value
	}

	return totalValue
}
