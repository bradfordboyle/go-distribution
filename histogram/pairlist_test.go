package histogram

import (
	"testing"
)

func TestNewPairList(t *testing.T) {
	m := map[string]uint{
		"rsc": 3711,
		"r":   2138,
		"gri": 1908,
		"adg": 912,
	}
	pl := NewPairList(m)
	for _, p := range pl {
		i, ok := m[p.Key]
		if !ok {
			t.Errorf("Original map did not contaim %s", p.Key)
		}
		if i != p.Value {
			t.Errorf("PairList had the wrong value for %s; expected %d, actual %d", p.Key, i, p.Value)
		}
	}
}

func TestPairlist_Len(t *testing.T) {
	pl := pairlist([]pair{
		{Key: "a", Value: 1},
		{Key: "a", Value: 2},
		{Key: "b", Value: 1},
	})
	if pl.Len() != 3 {
		t.Errorf("PairList had the wrong value; expected %d, actual %d", 3, pl.Len())
	}
}

func TestPairlist_Less(t *testing.T) {
	pl := pairlist([]pair{
		{Key: "a", Value: 1},
		{Key: "a", Value: 2},
		{Key: "b", Value: 1},
	})

	if !pl.Less(0, 1) {
		t.Errorf("PairList.Less() returned incorrect result; expected %t, actual %t", true, pl.Less(0, 1))
	}

	if !pl.Less(0, 2) {
		t.Errorf("PairList.Less() returned incorrect result; expected %t, actual %t", true, pl.Less(0, 2))
	}
}

func TestPairlist_Swap(t *testing.T) {
	pl := pairlist([]pair{
		{Key: "a", Value: 1},
		{Key: "a", Value: 2},
		{Key: "b", Value: 1},
	})

	pl.Swap(0, 2)

	if pl[0].Value != 1 && pl[0].Key != "b" {
		t.Error("PairList.Swap() did no swap correctly")
	}

	if pl[1].Value != 2 && pl[1].Key != "a" {
		t.Error("PairList.Swap() did no swap correctly")
	}

	if pl[2].Value != 1 && pl[2].Key != "a" {
		t.Error("PairList.Swap() did no swap correctly")
	}
}

func TestPairlist_TotalValues(t *testing.T) {
	pl := pairlist([]pair{
		{Key: "a", Value: 1},
		{Key: "a", Value: 2},
		{Key: "b", Value: 1},
	})

	if pl.TotalValues() != 4 {
		t.Errorf("PairList.TotalValue() returned incorrect result; expected %d, actual %d", 4, pl.TotalValues())
	}
}
