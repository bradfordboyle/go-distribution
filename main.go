package main

import (
	"fmt"
	"bufio"
	"os"
	"strings"
	"regexp"
	"strconv"
	"sort"
)

type Histogram struct {

}

func (h *Histogram) WriteHist(s *Settings, tokenDict map[string]uint64) {
	// FIXME: pull this out somewhere
	regularColor := "\u001b[0m"
	keyColor := "\u001b[32m"
	ctColor := "\u001b[34m"
	pctColor := "\u001b[35m"
	graphColor := "\u001b[37m"
	maxTokenLen := 0
	maxVal := uint64(0)

	maxValueWidth := 0
	maxPctWidth := 0

	pairlist := NewPairList(tokenDict)
	sort.Sort(sort.Reverse(pairlist))
	totalValue := pairlist.TotalValues()

	for i, p := range *pairlist {

		if i == 0 {
			maxValueWidth = len(fmt.Sprintf("%d", p.value))
			maxPctWidth = len(fmt.Sprintf("(%2.2f%%)", float64(p.value) * 1.0 / float64(totalValue) * 100.0))
		}

		tokenLen := len(p.key)
		if tokenLen > maxTokenLen {
			maxTokenLen = tokenLen
		}

		if p.value > maxVal {
			maxVal = p.value
		}

		// FIXME: this should be in settings
		if i > 33 {
			break
		}
	}

	os.Stderr.WriteString("tokens/lines examined: 279\n")
	os.Stderr.WriteString(" tokens/lines matched: 17,444,532\n")
	os.Stderr.WriteString("       histogram keys: 279\n")
	os.Stderr.WriteString("              runtime: 2.00ms\n")

	histWidth := 120 - (maxTokenLen + 1) - (maxValueWidth + 1) - (maxPctWidth + 1) - 1

	os.Stderr.WriteString(rjust("Key", maxTokenLen))
	os.Stderr.WriteString("|")
	os.Stderr.WriteString(ljust("Ct", maxValueWidth))
	os.Stderr.WriteString(" ")
	os.Stderr.WriteString(ljust("(Pct)", maxPctWidth))
	os.Stderr.WriteString("  Histogram")
	os.Stderr.WriteString(keyColor)
	os.Stderr.WriteString("\n")

	for i, p := range *pairlist {
		os.Stdout.WriteString(rjust(p.key, maxTokenLen))
		os.Stdout.WriteString(regularColor)
		os.Stdout.WriteString("|")
		os.Stdout.WriteString(ctColor)

		outVal := fmt.Sprintf("%d", p.value)
		os.Stdout.WriteString(rjust(outVal, maxValueWidth))
		os.Stdout.WriteString(" ")

		pctStr := fmt.Sprintf("(%2.2f%%)", float64(p.value) * 1.0 / float64(totalValue) * 100.0)
		os.Stdout.WriteString(pctColor)
		os.Stdout.WriteString(rjust(pctStr, maxPctWidth))
		os.Stdout.WriteString(" ")

		os.Stdout.WriteString(graphColor)
		os.Stdout.WriteString(h.HistogramBar(s, histWidth, maxVal, p.value))

		if i == 34 {
			os.Stdout.WriteString(regularColor)
		} else {
			os.Stdout.WriteString(keyColor)
		}
		os.Stdout.WriteString("\n")


		// FIXME: this should be in settings
		if i == 34 {
			break
		}
	}
}

func ljust(s string, width int) string {
	return fmt.Sprintf(fmt.Sprintf("%%-%ds", width), s)
}

func rjust(s string, width int) string {
	return fmt.Sprintf(fmt.Sprintf("%%%ds", width), s)
}

func (h *Histogram) HistogramBar(s *Settings, histWidth int, maxVal uint64, barVal uint64) string {
	// FIXME: assuming --char=dt
	zeroChar := "•"
	oneChar := "•"
	width := float64(barVal) * 1.0 / float64(maxVal) * float64(histWidth)
	intWidth := int(width)
	//remainderWidth := width - intWidth

	bar := strings.Repeat(zeroChar, intWidth)
	return bar + oneChar
}

type InputReader struct {
	TokenDict map[string]uint64
}

func NewInputReader() *InputReader {
	i := new(InputReader)
	i.TokenDict = make(map[string]uint64)

	return i
}

func (i *InputReader) TokenizeInput(s *Settings) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\n")
		i.TokenDict[line] += 1
	}
}

func (i * InputReader) ReadPretalliedTokens(s *Settings) {
	vk := regexp.MustCompile(`^\s*(\d+)\s+(.+)$`)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner. Scan() {
		line := scanner.Text()
		res := vk.FindStringSubmatch(line)
		key := res[2]
		// TODO: handle the error here
		value, _ := strconv.ParseUint(res[1], 10, 64)
		i.TokenDict[key] = value
	}
}

type Settings struct {
	GraphValues string
}


type pair struct {
	key string
	value uint64
}

type pairlist []pair

func (pl pairlist) Len() int { return len(pl) }

func (pl pairlist) Less(i, j int) bool {
	if pl[i].value == pl[j].value {
		return pl[i].key < pl[j].key
	}
	return pl[i].value < pl[j].value
}

func (pl pairlist) Swap(i, j int) {
	pl[i], pl[j] = pl[j], pl[i]
}

func NewPairList(m map[string]uint64) *pairlist {
	p := make(pairlist, len(m))

	i := 0
	for k, v := range m {
		p[i] = pair{k, v}
		i += 1
	}

	return &p
}

func (pl *pairlist) TotalValues() uint64 {
	totalValue := uint64(0)
	for _, p := range *pl {
		totalValue += p.value
	}

	return totalValue
}

func main() {
	s := &Settings{
		GraphValues: "vk",
	}
	i := NewInputReader()
	h := &Histogram{}

	if s.GraphValues == "vk" || s.GraphValues == "kv" {
		i.ReadPretalliedTokens(s)
	} else {
		i.TokenizeInput(s)
	}


	i.TokenizeInput(s)
	h.WriteHist(s, i.TokenDict)
}