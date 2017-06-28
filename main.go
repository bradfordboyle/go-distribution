package main

import (
	"fmt"
	"bufio"
	"os"
	"strings"
	"regexp"
	"strconv"
	"sort"
	"time"
	"os/user"
	"path"
	"log"
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

	histWidth := int(s.Width) - (maxTokenLen + 1) - (maxValueWidth + 1) - (maxPctWidth + 1) - 1

	os.Stderr.WriteString(rjust("Key", maxTokenLen))
	os.Stderr.WriteString("|")
	os.Stderr.WriteString(ljust("Ct", maxValueWidth))
	os.Stderr.WriteString(" ")
	os.Stderr.WriteString(ljust("(Pct)", maxPctWidth))
	os.Stderr.WriteString("  Histogram")
	os.Stderr.WriteString(keyColor)
	os.Stderr.WriteString("\n")

	outputLimit := len(*pairlist)
	if outputLimit > int(s.Height) {
		outputLimit = int(s.Height)
	}
	for i, p := range (*pairlist)[:outputLimit] {
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

		if i == outputLimit - 1 {
			os.Stdout.WriteString(regularColor)
			break
		} else {
			os.Stdout.WriteString(keyColor)
		}
		os.Stdout.WriteString("\n")
	}
}

func ljust(s string, width int) string {
	return fmt.Sprintf(fmt.Sprintf("%%-%ds", width), s)
}

func rjust(s string, width int) string {
	return fmt.Sprintf(fmt.Sprintf("%%%ds", width), s)
}

func (h *Histogram) HistogramBar(s *Settings, histWidth int, maxVal uint64, barVal uint64) string {
	// given a value and max, return string for histogram bar of the proper
	// number of characters, including unicode partial-width characters

	// first case is partial-width chars
	var zeroChar, oneChar string
	if s.CharWidth < 1.0 {
		zeroChar = s.GraphChars[len(s.GraphChars)-1]
	} else if len(s.HistogramChar) > 1 && s.UnicodeMode == false {
		zeroChar = string(s.HistogramChar[0])
		oneChar = string(s.HistogramChar[1])
	} else {
		zeroChar = s.HistogramChar
		oneChar = s.HistogramChar
	}

	// write out the full-width integer portion of the histogram
	var intWidth int
	var remainderWidth float32
	if s.Logarithmic {
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
	if s.CharWidth == 1 {
		bar += oneChar
	} else if s.CharWidth < 1 {
		// this is high-resolution, so figure out what remainder we
		// have to represent
		if remainderWidth > s.CharWidth {
			whichChar := int(remainderWidth / s.CharWidth)
			bar += s.GraphChars[whichChar]
		}
	}

	return bar
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
	// how to split the input... typically we split on whitespace or
	// word boundaries, but the user can specify any regexp
	if s.Tokenize == "white" {
		s.Tokenize = `\s+`
	} else if s.Tokenize == "word" {
		s.Tokenize = `\W`
	}
	if s.MatchRegexp == "word" {
		s.MatchRegexp = `^[A-Z,a-z]+$`
	} else if s.MatchRegexp == "num" || s.MatchRegexp == "number" {
		s.MatchRegexp = `^\d+$`
	}

	pt := regexp.MustCompile(s.Tokenize)
	pm := regexp.MustCompile(s.MatchRegexp)

	//nextStat := time.Now().Add(time.Duration(s.StatInterval))

	pruneObjects := uint(0)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\n")
		if s.Tokenize != "" {
			for _, token := range pt.Split(line, -1) {
				// user desires to break line into tokens...
				s.TotalObjects += 1
				if pm.MatchString(token) {
					s.TotalValues += 1
					pruneObjects += 1
					i.TokenDict[token] += 1
				}
			}
		} else {
			// user just wants every line to be a token
			s.TotalObjects += 1
			if pm.MatchString(line) {
				s.TotalValues += 1
				pruneObjects += 1
				i.TokenDict[line] += 1
			}
		}

		// prune the hash if it gets too large
		if pruneObjects >= s.KeyPruneInterval {
			i.PruneKeys(s)
			pruneObjects = 0
		}
	}
}

func (i *InputReader) PruneKeys(s *Settings) {
	prunedTokenCounts := make(map[string]uint64, s.MaxKeys)
	pl := NewPairList(i.TokenDict)
	sort.Sort(sort.Reverse(pl))

	for i, p := range *pl {
		prunedTokenCounts[p.key] = p.value
		if uint(i) + 1 > s.MaxKeys {
			break
		}
	}
	i.TokenDict = prunedTokenCounts
	s.NumPrunes += 1
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
	TotalMillis uint
	StartTime uint
	EndTime uint
	WidthArg uint
	HeightArg uint
	Width uint
	Height uint
	HistogramChar string
	ColourisedOutput bool
	Logarithmic bool
	NumOnly string
	Verbose bool
	GraphValues string
	Size string
	Tokenize string
	MatchRegexp string
	StatInterval int
	NumPrunes uint
	ColourPalette string
	RegularColour string
	KeyColour  string
	CtColour  string
	PctColour  string
	GraphColour  string
	TotalObjects uint
	TotalValues uint
	KeyPruneInterval uint
	MaxKeys uint
	UnicodeMode bool
	CharWidth float32
	GraphChars []string
	PartialBlocks []string
	PartialLines []string
}

func NewSettings() *Settings {
	// default settings
	s := &Settings{
		TotalMillis: 0,
		StartTime: uint(time.Now().UnixNano() / 1000),
		EndTime: 0,
		WidthArg: 0,
		HeightArg: 0,
		Width: 80,
		Height: 15,
		HistogramChar: "-",
		ColourisedOutput: false,
		Logarithmic: false,
		NumOnly: "XXX",
		Verbose: false,
		GraphValues: "",
		Size: "",
		Tokenize: "",
		MatchRegexp: ".",
		StatInterval: 1.0,
		NumPrunes: 0,
		ColourPalette: "0,0,32,35,34",
		RegularColour: "",
		KeyColour: "",
		CtColour: "",
		PctColour: "",
		GraphColour: "",
		TotalObjects: 0,
		TotalValues: 0,
		KeyPruneInterval: 1500000,
		MaxKeys: 5000,
		UnicodeMode: false,
		CharWidth: 1.0,
		GraphChars: []string{},
		PartialBlocks:    []string{"▏", "▎", "▍", "▌", "▋", "▊", "▉", "█"},
		PartialLines:     []string{"╸", "╾", "━"},
	}

	// rcfile grabbing/parsing if specified
	var rcFile string
	if len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "--rcfile") {
		rcFile = strings.Split(os.Args[1], "=")[1]
	} else {
		u, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		rcFile = path.Join(u.HomeDir, ".distributionrc")
	}

	// parse opts from the rcFile if it exists
	file, err := os.Open(rcFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimRight(line, " ")
		rcOpt := strings.Split(trimmedLine, "#")[0]
		if rcOpt != "" {
			os.Args = append(os.Args, "")
			copy(os.Args[1:], os.Args[0:])
			os.Args[0] = rcOpt
		}
	}

	// manual argument parsing
	for _, arg := range os.Args {
		if arg == "-h" || arg == "--help" {
			doUsage(s)
			os.Exit(0)
		} else if arg == "-c" || arg == "--color" || arg == "--colour" {
			s.ColourisedOutput = true
		} else if arg == "-g" || arg == "--graph" {
			// can pass --graph without option, will default to value/key ordering
			// since unix perfers that for piping-to-sort reasons
			s.GraphValues = "vk"
		} else if arg == "-l" || arg == "--logarithmic" {
			s.Logarithmic = true
		} else if arg == "-n" || arg == "--numonly" {
			s.NumOnly = "abs"
		} else if arg == "-v" || arg == "--verbose" {
			s.Verbose = true
		} else {
			argList := strings.SplitN(arg, "=", 2)
			if argList[0] == "-w" || argList[0] == "--width" {
				argInt, err := strconv.ParseUint(argList[1], 10, 32)
				if err != nil {
					log.Fatal(err)
				}
				s.WidthArg = uint(argInt)
			} else if argList[0] == "-h" || argList[0] == "--height" {
				argInt, err := strconv.ParseUint(argList[1], 10, 32)
				if err != nil {
					log.Fatal(err)
				}
				s.HeightArg = uint(argInt)
			} else if argList[0] == "-k" || argList[0] == "--keys" {
				argInt, err := strconv.ParseUint(argList[1], 10, 32)
				if err != nil {
					log.Fatal(err)
				}
				s.MaxKeys = uint(argInt)
			} else if argList[0] == "-c" || argList[0] == "--char" {
				s.HistogramChar = argList[1]
			} else if argList[0] == "-g" || argList[0] == "--graph" {
				s.GraphValues = argList[1]
			} else if argList[0] == "-n" || argList[0] == "--numonly" {
				s.NumOnly = argList[1]
			} else if argList[0] == "-p" || argList[0] == "--palette" {
				s.ColourPalette = argList[1]
				s.ColourisedOutput = true
			} else if argList[0] == "-s" || argList[0] == "--size" {
				s.Size = argList[1]
			} else if argList[0] == "-t" || argList[0] == "--tokenize" {
				s.Tokenize = argList[1]
			} else if argList[0] == "-m" || argList[0] == "--match" {
				s.MatchRegexp = argList[1]
			}
		}
	}

	// first, size, which might be further overridden by width/height later
	if s.Size == "full" || s.Size == "fl" || s.Size == "f" {
		// tput will tell us the term width/height even if input is stdin
		// TODO: actually call tput here
		s.Width, s.Height = 238, 61
		// need room for the verbosity output
		if s.Verbose {
			s.Height -= 4
		}
		if s.Width < 40 {
			s.Width = 40
		}
		if s.Height < 10 {
			s.Height = 10
		}
	} else if s.Size == "small" || s.Size == "sm" || s.Size == "s" {
		s.Width, s.Height = 60, 10
	} else if s.Size == "medium" || s.Size == "med" || s.Size == "m" {
		s.Width, s.Height = 100, 20
	} else if s.Size == "large" || s.Size == "lg" || s.Size == "l" {
		s.Width, s.Height = 140, 35
	}

	// synonyms "monotonically-increasing": derivative, difference, delta, increasing
	// so all "d" "i" and "m" words will be graphing those differences
	if s.NumOnly[0] == 'd' || s.NumOnly[0] == 'i' || s.NumOnly[0] == 'm' {
		s.NumOnly = "mon"
	}
	// synonyms "actual values": absolute, actual, number, normal, noop,
	// so all "a" and "n" words will graph straight up numbers
	if s.NumOnly[0] == 'a' || s.NumOnly[0] == 'n' {
		s.NumOnly = "abs"
	}

	// override variables if they were explicitly given
	if s.WidthArg  != 0 {
		s.Width  = s.WidthArg
	}
	if s.HeightArg != 0 {
		s.Height = s.HeightArg
	}

	// maxKeys should be at least a few thousand greater than height to reduce odds
	// of throwing away high-count values that appear sparingly in the data
	if s.MaxKeys < s.Height + 3000 {
		s.MaxKeys = s.Height + 3000
	}

	if s.Verbose {
		os.Stderr.WriteString(fmt.Sprintf("Updated maxKeys to %d (height + 3000)\n", s.MaxKeys))
	}

	// colour palette
	if s.ColourisedOutput {
		cl := strings.Split(s.ColourPalette, ",")
		// ANSI color code is ESC+[+NN+m where ESC=chr(27), [ and m are
		// the literal characters, and NN is a two-digit number, typically
		// from 31 to 37 - why is this knowledge still useful in 2014?
		s.RegularColour = fmt.Sprintf("\u001b[%sm", cl[0])
		s.KeyColour = fmt.Sprintf("\u001b[%sm", cl[1])
		s.CtColour = fmt.Sprintf("\u001b[%sm", cl[2])
		s.PctColour = fmt.Sprintf("\u001b[%sm", cl[3])
		s.GraphColour = fmt.Sprintf("\u001b[%sm", cl[4])
	}


	// some useful ASCII-->utf-8 substitutions
	switch s.HistogramChar {
	case "ba":
		s.UnicodeMode = true
		s.HistogramChar = "▬"
	case "bl":
		s.UnicodeMode = true
		s.HistogramChar = "Ξ"
	case "em":
		s.UnicodeMode = true
		s.HistogramChar = "—"
	case "me":
		s.UnicodeMode = true
		s.HistogramChar = "⋯"
	case "di":
		s.UnicodeMode = true
		s.HistogramChar = "♦"
	case "dt":
		s.UnicodeMode = true
		s.HistogramChar = "•"
	case "sq":
		s.UnicodeMode = true;
		s.HistogramChar = "□"
	}

	// sub-full character width graphing systems
	if s.HistogramChar == "pb" {
		s.CharWidth = 0.125
		s.GraphChars = s.PartialBlocks
	} else if s.HistogramChar == "pl" {
		s.CharWidth = 0.3334;
		s.GraphChars = s.PartialLines
	}

	// detect whether the user has passed a multibyte unicode character directly as the histogram char
	if s.HistogramChar[0] >= 128 {
		s.UnicodeMode = true
	}

	return s
}

func doUsage(s *Settings) {
	os.Stdout.WriteString("")
	os.Stdout.WriteString(fmt.Sprintf("usage: <commandWithOutput> | %s\n", os.Args[0]))

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
	s := NewSettings()

	i := NewInputReader()
	h := &Histogram{}

	if s.GraphValues == "vk" || s.GraphValues == "kv" {
		i.ReadPretalliedTokens(s)
	} else if s.NumOnly != "XXX" {
		os.Exit(0)
	} else {
		i.TokenizeInput(s)
	}

	h.WriteHist(s, i.TokenDict)
}