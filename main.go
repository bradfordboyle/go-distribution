package main

import (
	"distribution/histogram"
	"distribution/settings"
	"distribution/tokenize"
	"log"
	"os"
)

func main() {
	s := settings.NewSettings(os.Args[0], os.Args[1:])

	var t tokenize.Tokenizer
	if s.GraphValues == "vk" {
		t = tokenize.NewValueKeyTokenizer()
	} else if s.GraphValues == "kv" {
		t = tokenize.NewKeyValueTokenizer()
	} else if s.NumOnly != "XXX" {
		os.Exit(0)
	} else if s.Tokenize != "" {
		t = tokenize.NewRegexTokenizer(s.Tokenize, s.MatchRegexp)
	} else {
		t = tokenize.NewLineTokenizer(s.MatchRegexp)
	}

	pl, err := t.Tokenize(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	h := histogram.NewHistogram(s)
	h.WriteHist(os.Stdout, pl)
}
