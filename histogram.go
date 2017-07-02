package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

type Histogram struct {
	s            *Settings
	height       uint
	width        uint
	keyColor     string
	regularColor string
	ctColor      string
	pctColor     string
	graphColor   string
}

func NewHistogram(s *Settings) *Histogram {
	return &Histogram{
		s:            s,
		height:       s.Height,
		width:        s.Width,
		keyColor:     s.KeyColour,
		regularColor: s.RegularColour,
		ctColor:      s.CtColour,
		pctColor:     s.PctColour,
		graphColor:   s.GraphColour,
	}
}

func (h *Histogram) WriteHist(writer io.Writer, pairlist Pairlist) {
	maxTokenLen := 0
	maxVal := uint(0)

	maxValueWidth := 0
	maxPctWidth := 0

	sort.Sort(sort.Reverse(pairlist))
	totalValue := pairlist.TotalValues()

	for i, p := range pairlist {

		if i == 0 {
			maxValueWidth = len(fmt.Sprintf("%d", p.value))
			maxPctWidth = len(fmt.Sprintf("(%2.2f%%)", float64(p.value)*1.0/float64(totalValue)*100.0))
		}

		tokenLen := len(p.key)
		if tokenLen > maxTokenLen {
			maxTokenLen = tokenLen
		}

		if p.value > maxVal {
			maxVal = p.value
		}

		if uint(i) >= h.height-1 {
			break
		}
	}

	h.s.EndTime = time.Now().UnixNano()
	totalMillis := float64(h.s.EndTime-h.s.StartTime) / 1e6
	if h.s.Verbose {

		os.Stderr.WriteString(fmt.Sprintf("tokens/lines examined: %s\n", humanize.Comma(int64(h.s.TotalObjects))))
		os.Stderr.WriteString(fmt.Sprintf(" tokens/lines matched: %s\n", humanize.Comma(int64(h.s.TotalValues))))
		os.Stderr.WriteString(fmt.Sprintf("       histogram keys: %d\n", pairlist.Len()))
		os.Stderr.WriteString(fmt.Sprintf("              runtime: %sms\n", humanize.Commaf(totalMillis)))
	}

	histWidth := int(h.width) - (maxTokenLen + 1) - (maxValueWidth + 1) - (maxPctWidth + 1) - 1

	os.Stderr.WriteString(Rjust("Key", maxTokenLen))
	os.Stderr.WriteString("|")
	os.Stderr.WriteString(Ljust("Ct", maxValueWidth))
	os.Stderr.WriteString(" ")
	os.Stderr.WriteString(Ljust("(Pct)", maxPctWidth))
	os.Stderr.WriteString("  Histogram")
	os.Stderr.WriteString(h.keyColor)
	os.Stderr.WriteString("\n")

	outputLimit := pairlist.Len()
	if outputLimit > int(h.height) {
		outputLimit = int(h.height)
	}
	for i, p := range (pairlist)[:outputLimit] {
		io.WriteString(writer, Rjust(p.key, maxTokenLen))
		io.WriteString(writer, h.regularColor)
		io.WriteString(writer, "|")
		io.WriteString(writer, h.ctColor)

		outVal := fmt.Sprintf("%d", p.value)
		io.WriteString(writer, Rjust(outVal, maxValueWidth))
		io.WriteString(writer, " ")

		pctStr := fmt.Sprintf("(%2.2f%%)", float64(p.value)*1.0/float64(totalValue)*100.0)
		io.WriteString(writer, h.pctColor)
		io.WriteString(writer, Rjust(pctStr, maxPctWidth))
		io.WriteString(writer, " ")

		io.WriteString(writer, h.graphColor)
		io.WriteString(writer, h.HistogramBar(histWidth, maxVal, p.value))

		if i == outputLimit-1 {
			io.WriteString(writer, h.regularColor)
			break
		} else {
			io.WriteString(writer, h.keyColor)
		}
		io.WriteString(writer, "\n")
	}
}

func (h *Histogram) HistogramBar(histWidth int, maxVal uint, barVal uint) string {
	// given a value and max, return string for histogram bar of the proper
	// number of characters, including unicode partial-width characters

	// first case is partial-width chars
	var zeroChar, oneChar string
	if h.s.CharWidth < 1.0 {
		zeroChar = h.s.GraphChars[len(h.s.GraphChars)-1]
	} else if len(h.s.HistogramChar) > 1 && h.s.UnicodeMode == false {
		zeroChar = string(h.s.HistogramChar[0])
		oneChar = string(h.s.HistogramChar[1])
	} else {
		zeroChar = h.s.HistogramChar
		oneChar = h.s.HistogramChar
	}

	// write out the full-width integer portion of the histogram
	var intWidth int
	var remainderWidth float32
	if h.s.Logarithmic {
		// TODO
	} else {
		width := float32(barVal) * 1.0 / float32(maxVal) * float32(histWidth)
		intWidth = int(width)
		remainderWidth = width - float32(intWidth)
	}

	// write the zeroeth character intWidth times...
	bar := strings.Repeat(zeroChar, intWidth)

	// we always have at least one remaining char for histogram - if
	// we have full-width chars, then just print it, otherwise do a
	// calculation of how much remainder we need to print
	//
	// FIXME: The remainder partial char printed does not take into
	// account logarithmic scale (can humans notice?).
	if h.s.CharWidth == 1 {
		bar += oneChar
	} else if h.s.CharWidth < 1 {
		// this is high-resolution, so figure out what remainder we
		// have to represent
		if remainderWidth > h.s.CharWidth {
			whichChar := int(remainderWidth / h.s.CharWidth)
			bar += h.s.GraphChars[whichChar]
		}
	}

	return bar
}

func Ljust(s string, width int) string {
	return fmt.Sprintf(fmt.Sprintf("%%-%ds", width), s)
}

func Rjust(s string, width int) string {
	return fmt.Sprintf(fmt.Sprintf("%%%ds", width), s)
}
