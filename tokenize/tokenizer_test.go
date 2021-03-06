package tokenize

import (
	"bytes"
	"fmt"
	"testing"
)

func TestKeyValueTokenizer_Tokenize(t *testing.T) {
	kv := NewKeyValueTokenizer()
	buf := new(bytes.Buffer)

	tc, _ := kv.Tokenize(buf)
	if len(tc) != 0 {
		t.Error("Tokenize on empty reader didn't return an empty PairList")
	}

	buf.WriteString("a 1\n")
	tc, _ = kv.Tokenize(buf)
	if len(tc) != 1 {
		t.Error("Tokenize on single line didn't return single Pair")
	}
	if v, ok := tc["a"]; !ok || v != 1 {
		t.Error("Tokenize did not extract key/value correctly")
	}
}

func TestValueKeyTokenizer_Tokenize(t *testing.T) {
	vk := NewValueKeyTokenizer()
	buf := new(bytes.Buffer)

	tc, _ := vk.Tokenize(buf)
	if len(tc) != 0 {
		t.Error("Tokenize on empty reader didn't return an empty PairList")
	}

	buf.WriteString("1 a\n")
	tc, _ = vk.Tokenize(buf)
	if len(tc) != 1 {
		t.Error("Tokenize on single line didn't return single Pair")
	}
	if v, ok := tc["a"]; !ok || v != 1 {
		t.Error("Tokenize did not extract key/value correctly")
	}
}

func TestNewRegexTokenizer(t *testing.T) {
	testCases := []struct {
		splitter string
		matcher  string
	}{
		{"white", "word"},
		{"white", "num"},
		{"white", `|`},
		{"word", "word"},
		{"word", "num"},
		{"word", `|`},
		{`|`, "word"},
		{`|`, "num"},
		{`|`, `|`},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("spliter: %s; matcher: %s", tc.splitter, tc.matcher), func(t *testing.T) {
			r := NewRegexTokenizer(tc.splitter, tc.matcher)
			if r == nil {
				t.Error("Unable to create regexTokenizer w/ shortcuts")
			}
		})
	}
}

func TestRegexTokenizer_Tokenize(t *testing.T) {
	r := NewRegexTokenizer("white", "word")
	buf := new(bytes.Buffer)

	tc, _ := r.Tokenize(buf)
	if len(tc) != 0 {
		t.Error("Tokenize on empty reader didn't return an empty PairList")
	}

	buf.WriteString("a\n")
	tc, _ = r.Tokenize(buf)
	if len(tc) != 1 {
		t.Error("Tokenize on single line didn't return single Pair")
	}

	buf.WriteString("a a\n")
	tc, _ = r.Tokenize(buf)
	if len(tc) != 1 {
		t.Error("Tokenize on buffer w/ single token didn't return single Pair")
	}
	if v, ok := tc["a"]; !ok || v != 2 {
		t.Error("Tokenize did not extract key/value correctly")
	}

	testCase := `Job ` + "`" + `cron.daily'
Normal exit
(root) CMD
Modem hangup
Connect time
Sent 19243191
<info> (ttyUSB0)
<info> (ttyUSB0)
<info> (ttyUSB2)
<info> (ttyUSB2)
`
	buf = new(bytes.Buffer)
	buf.WriteString(testCase)
	tc, _ = r.Tokenize(buf)
	if len(tc) == 0 {
		t.Error("Tokenize on buffer w/ real data is failing")
	}
}

func TestNewLineTokenizer(t *testing.T) {
	testCases := []struct {
		matcher string
	}{
		{"word"},
		{"num"},
		{`|`},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("matcher: %s", tc.matcher), func(t *testing.T) {
			l := NewLineTokenizer(tc.matcher)
			if l == nil {
				t.Error("Unable to create lineTokenizer w/ shortcuts")
			}
		})
	}
}

func TestLineTokenizer_Tokenize(t *testing.T) {
	l := NewLineTokenizer(".")
	buf := new(bytes.Buffer)

	tc, _ := l.Tokenize(buf)
	if len(tc) != 0 {
		t.Error("Tokenize on empty reader didn't return an empty PairList")
	}

	buf.WriteString("a\n")
	tc, _ = l.Tokenize(buf)
	if len(tc) != 1 {
		t.Error("Tokenize on single line didn't return single Pair")
	}

	buf.WriteString("a a\n")
	tc, _ = l.Tokenize(buf)
	if len(tc) != 1 {
		t.Error("Tokenize on buffer w/ single token didn't return single Pair")
	}
	if v, ok := tc["a a"]; !ok || v != 1 {
		t.Error("Tokenize did not extract key/value correctly")
	}

}
