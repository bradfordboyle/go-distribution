package main

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type Tokenizer interface {
	Tokenize(io.Reader) (Pairlist, error)
}

type preTalliedTokenizer struct {
	extractor *regexp.Regexp
	keyIdx    int
	valueIdx  int
}

const (
	KEY_VALUE_REGEX = `^\s*(.+)\s+(\d+)$`
	VALUE_KEY_REGEX = `^\s*(\d+)\s+(.+)$`
)

func NewKeyValueTokenizer() Tokenizer {
	return preTalliedTokenizer{
		extractor: regexp.MustCompile(KEY_VALUE_REGEX),
		keyIdx:    1,
		valueIdx:  2,
	}
}

func NewValueKeyTokenizer() Tokenizer {
	return preTalliedTokenizer{
		extractor: regexp.MustCompile(VALUE_KEY_REGEX),
		keyIdx:    2,
		valueIdx:  1,
	}
}

func (p preTalliedTokenizer) Tokenize(reader io.Reader) (Pairlist, error) {
	pl := make(Pairlist, 0)

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		res := p.extractor.FindStringSubmatch(line)
		key := res[p.keyIdx]
		value, err := strconv.ParseUint(res[p.valueIdx], 10, 32)
		if err != nil {
			return nil, err
		}
		pl = append(pl, pair{key: key, value: uint(value)})
	}

	return pl, nil
}

type regexTokenizer struct {
	splitter         *regexp.Regexp
	matcher          *regexp.Regexp
	keyPruneInterval int
}

const (
	WHITESPACE_REGEX = `\s+`
	WORD_SPLIT_REGEX = `\W`
	WORD_MATCH_REGEX = `^[A-Z,a-z]+$`
	NUM_MATCH_REGEX  = `^\d+$`
)

func NewRegexTokenizer(splitter string, matcher string) Tokenizer {
	t := &regexTokenizer{}

	switch splitter {
	case "white":
		t.splitter = regexp.MustCompile(WHITESPACE_REGEX)
	case "word":
		t.splitter = regexp.MustCompile(WORD_SPLIT_REGEX)
	default:
		t.splitter = regexp.MustCompile(splitter)
	}

	switch matcher {
	case "word":
		t.matcher = regexp.MustCompile(WORD_MATCH_REGEX)
	case "num":
		t.matcher = regexp.MustCompile(NUM_MATCH_REGEX)
	default:
		t.matcher = regexp.MustCompile(matcher)
	}

	return t
}

func (r *regexTokenizer) Tokenize(reader io.Reader) (Pairlist, error) {
	tokenDict := make(map[string]uint)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\n")
		for _, token := range r.splitter.Split(line, -1) {
			if r.matcher.MatchString(token) {
				tokenDict[token]++
			}
		}
	}

	return NewPairList(tokenDict), nil
}

type lineTokenizer struct {
	matcher          *regexp.Regexp
	keyPruneInterval int
}

func NewLineTokenizer(matcher string) Tokenizer {
	t := &lineTokenizer{}

	switch matcher {
	case "word":
		t.matcher = regexp.MustCompile(WORD_MATCH_REGEX)
	case "num":
		t.matcher = regexp.MustCompile(NUM_MATCH_REGEX)
	default:
		t.matcher = regexp.MustCompile(matcher)
	}

	return t
}

func (l lineTokenizer) Tokenize(reader io.Reader) (Pairlist, error) {
	tokenDict := make(map[string]uint)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\n")
		if l.matcher.MatchString(line) {
			tokenDict[line]++
		}
	}

	return NewPairList(tokenDict), nil
}
