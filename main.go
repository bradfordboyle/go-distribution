package main

import (
	"log"
	"os"
)

func main() {
	s := NewSettings(os.Args[0], os.Args[1:])

	var t Tokenizer
	if s.GraphValues == "vk" {
		t = NewValueKeyTokenizer()
	} else if s.GraphValues == "kv" {
		t = NewKeyValueTokenizer()
	} else if s.NumOnly != "XXX" {
		os.Exit(0)
	} else if s.Tokenize != "" {
		t = NewRegexTokenizer(s.Tokenize, s.MatchRegexp)
	} else {
		t = NewLineTokenizer(s.MatchRegexp)
	}

	pl, err := t.Tokenize(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	h := NewHistogram(s)
	h.WriteHist(os.Stdout, pl)
}
