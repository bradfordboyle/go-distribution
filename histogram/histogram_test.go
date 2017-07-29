package histogram

import (
	"bytes"
	"distribution/settings"
	"testing"
)

func TestLjust(t *testing.T) {
	s := Ljust("a", 4)
	if s != "a   " {
		t.Errorf("Ljust incorrect: expected %s; actual %s", "a   ", s)
	}
}

func TestRjust(t *testing.T) {
	s := Rjust("a", 4)
	if s != "   a" {
		t.Errorf("Ljust incorrect: expected %s; actual %s", "   a", s)
	}
}

func TestHistogram_HistogramBar(t *testing.T) {
	testCases := []struct {
		args      []string
		histWidth int
		maxVal    uint
		barVal    uint
		expected  string
	}{
		{args: []string{"--char==>"}, histWidth: 10, maxVal: 10, barVal: 2, expected: "==>"},
		{args: []string{"--char=dt"}, histWidth: 10, maxVal: 10, barVal: 2, expected: "•••"},
		{args: []string{"--char=pb"}, histWidth: 10, maxVal: 100, barVal: 25, expected: "██▋"},
	}

	for _, tc := range testCases {
		t.Run("something", func(t *testing.T) {
			s := settings.NewSettings("testing", tc.args)
			h := NewHistogram(s)

			bar := h.HistogramBar(tc.histWidth, tc.maxVal, tc.barVal)
			if bar != tc.expected {
				t.Errorf("HistogramBar is wrong: expected %s; actual %s", tc.expected, bar)
			}
		})
	}
}

// TODO Setting rcfile to "/dev/null" is a bit of a hack
const (
	RC_FILE = "--rcfile=/dev/null"
	KV      = "--graph=kv"
	WIDTH   = "--width=15"
)

func TestHistogram_WriteHist(t *testing.T) {
	testCases := []struct {
		name     string
		args     []string
		counts   map[string]uint
		expected string
	}{
		{
			name:     "Empty PairList",
			args:     []string{RC_FILE, KV, WIDTH},
			counts:   make(map[string]uint),
			expected: "",
		},
		{
			name:     "PairList w/ two tokens",
			args:     []string{RC_FILE, KV, WIDTH},
			counts:   map[string]uint{"a": 1, "b": 2},
			expected: "b|2 (66.67%) --\na|1 (33.33%) -",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := settings.NewSettings("testing", tc.args)
			h := NewHistogram(s)
			buf := new(bytes.Buffer)

			h.WriteHist(buf, tc.counts)

			if buf.String() != tc.expected {
				t.Errorf("WriteHist incorrect: expected %s; actual %s", tc.expected, buf.String())
			}
		})
	}
}
