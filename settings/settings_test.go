package settings

import (
	"bytes"
	"testing"
)

const RC_FILE = "--rcfile=/dev/null"

func TestNewSettings(t *testing.T) {
	testCases := []struct {
		arg     string
		checker func(*Settings) bool
	}{
		{"--color", func(s *Settings) bool { return s.ColourisedOutput }},
		{"", func(s *Settings) bool { return s.GraphValues == "" }},
		{"-g", func(s *Settings) bool { return s.GraphValues == "vk" }},
		{"--graph", func(s *Settings) bool { return s.GraphValues == "vk" }},
		{"-l", func(s *Settings) bool { return s.Logarithmic }},
		{"--logarithmic", func(s *Settings) bool { return s.Logarithmic }},
		{"-n", func(s *Settings) bool { return s.NumOnly == "abs" }},
		{"--numonly", func(s *Settings) bool { return s.NumOnly == "abs" }},
		{"-v", func(s *Settings) bool { return s.Verbose }},
		{"--verbose", func(s *Settings) bool { return s.Verbose }},
		{"-w=16", func(s *Settings) bool { return s.Width == 16 }},
		{"--width=16", func(s *Settings) bool { return s.Width == 16 }},
		{"-h=32", func(s *Settings) bool { return s.Height == 32 }},
		{"--height=32", func(s *Settings) bool { return s.Height == 32 }},
		{"-k=65535", func(s *Settings) bool { return s.MaxKeys == 65535 }},
		{"--keys=65535", func(s *Settings) bool { return s.MaxKeys == 65535 }},
		{"-c==>", func(s *Settings) bool { return s.HistogramChar == "=>" }},
		{"--char==>", func(s *Settings) bool { return s.HistogramChar == "=>" }},
		{"-g=kv", func(s *Settings) bool { return s.GraphValues == "kv" }},
		{"--graph=kv", func(s *Settings) bool { return s.GraphValues == "kv" }},
		{"-g=vk", func(s *Settings) bool { return s.GraphValues == "vk" }},
		{"--graph=vk", func(s *Settings) bool { return s.GraphValues == "vk" }},
		{"-n=actual", func(s *Settings) bool { return s.NumOnly == "abs" }},
		{"--numonly=actual", func(s *Settings) bool { return s.NumOnly == "abs" }},
		{"-n=n", func(s *Settings) bool { return s.NumOnly == "abs" }},
		{"--numonly=n", func(s *Settings) bool { return s.NumOnly == "abs" }},
		{"-n=i", func(s *Settings) bool { return s.NumOnly == "mon" }},
		{"--numonly=i", func(s *Settings) bool { return s.NumOnly == "mon" }},
		{"-n=diff", func(s *Settings) bool { return s.NumOnly == "mon" }},
		{"--numonly=m", func(s *Settings) bool { return s.NumOnly == "mon" }},
		{"-n=diff", func(s *Settings) bool { return s.NumOnly == "mon" }},
		{"--numonly=m", func(s *Settings) bool { return s.NumOnly == "mon" }},
		{"-p=30,31,32,33,34", func(s *Settings) bool { return s.ColourPalette == "30,31,32,33,34" && s.ColourisedOutput }},
		{"--palette=30,31,32,33,34", func(s *Settings) bool { return s.ColourPalette == "30,31,32,33,34" && s.ColourisedOutput }},
		// omitting "full" as this calls `TerminalSize()` which does not
		// always work when being tested
		//{"-s=full", func(s *Settings) bool { return s.Size == "full" }},
		//{"--size=full", func(s *Settings) bool { return s.Size == "full" }},
		{"-s=small", func(s *Settings) bool { return s.Size == "small" && s.Width == 60 && s.Height == 10 }},
		{"--size=small", func(s *Settings) bool { return s.Size == "small" && s.Width == 60 && s.Height == 10 }},
		{"-s=sm", func(s *Settings) bool { return s.Size == "sm" && s.Width == 60 && s.Height == 10 }},
		{"--size=sm", func(s *Settings) bool { return s.Size == "sm" && s.Width == 60 && s.Height == 10 }},
		{"-s=s", func(s *Settings) bool { return s.Size == "s" && s.Width == 60 && s.Height == 10 }},
		{"--size=s", func(s *Settings) bool { return s.Size == "s" && s.Width == 60 && s.Height == 10 }},
		{"-s=medium", func(s *Settings) bool { return s.Size == "medium" && s.Width == 100 && s.Height == 20 }},
		{"--size=medium", func(s *Settings) bool { return s.Size == "medium" && s.Width == 100 && s.Height == 20 }},
		{"-s=med", func(s *Settings) bool { return s.Size == "med" && s.Width == 100 && s.Height == 20 }},
		{"--size=med", func(s *Settings) bool { return s.Size == "med" && s.Width == 100 && s.Height == 20 }},
		{"-s=m", func(s *Settings) bool { return s.Size == "m" && s.Width == 100 && s.Height == 20 }},
		{"--size=m", func(s *Settings) bool { return s.Size == "m" && s.Width == 100 && s.Height == 20 }},
		{"-s=large", func(s *Settings) bool { return s.Size == "large" && s.Width == 140 && s.Height == 35 }},
		{"--size=large", func(s *Settings) bool { return s.Size == "large" && s.Width == 140 && s.Height == 35 }},
		{"-s=lg", func(s *Settings) bool { return s.Size == "lg" && s.Width == 140 && s.Height == 35 }},
		{"--size=lg", func(s *Settings) bool { return s.Size == "lg" && s.Width == 140 && s.Height == 35 }},
		{"-s=l", func(s *Settings) bool { return s.Size == "l" && s.Width == 140 && s.Height == 35 }},
		{"--size=l", func(s *Settings) bool { return s.Size == "l" && s.Width == 140 && s.Height == 35 }},
		{"-t=\\w", func(s *Settings) bool { return s.Tokenize == "\\w" }},
		{"--tokenize=\\w", func(s *Settings) bool { return s.Tokenize == "\\w" }},
		{"-m=\\d", func(s *Settings) bool { return s.MatchRegexp == "\\d" }},
		{"--match=\\d", func(s *Settings) bool { return s.MatchRegexp == "\\d" }},
		// the following test special values for certain keys
		{"--keys=10", func(s *Settings) bool { return s.MaxKeys == s.Height+3000 }},
		{"--char=ba", func(s *Settings) bool { return s.UnicodeMode && s.HistogramChar == "\u25ac" }},
		{"--char=bl", func(s *Settings) bool { return s.UnicodeMode && s.HistogramChar == "\u039e" }},
		{"--char=em", func(s *Settings) bool { return s.UnicodeMode && s.HistogramChar == "\u2014" }},
		{"--char=me", func(s *Settings) bool { return s.UnicodeMode && s.HistogramChar == "\u22ef" }},
		{"--char=di", func(s *Settings) bool { return s.UnicodeMode && s.HistogramChar == "\u2666" }},
		{"--char=dt", func(s *Settings) bool { return s.UnicodeMode && s.HistogramChar == "\u2022" }},
		{"--char=sq", func(s *Settings) bool { return s.UnicodeMode && s.HistogramChar == "\u25a1" }},
		{"--char=pb", func(s *Settings) bool {
			return s.CharWidth == 0.125 && len(s.GraphChars) == len(s.PartialBlocks)
		}},
		{"--char=pl", func(s *Settings) bool {
			return s.CharWidth == 0.3334 && len(s.GraphChars) == len(s.PartialLines)
		}},
	}

	for _, tc := range testCases {
		t.Run(tc.arg, func(t *testing.T) {
			s := NewSettings(t.Name(), []string{RC_FILE, tc.arg})
			if !tc.checker(s) {
				t.Errorf("Option '%s' incorrectly handled", tc.arg)
			}
		})
	}
}

func TestDoUsage(t *testing.T) {
	buf := new(bytes.Buffer)
	s := NewSettings(t.Name(), []string{})

	doUsage(s, buf)

	if buf.Len() == 0 {
		t.Error("No usage printed")
	}
}
