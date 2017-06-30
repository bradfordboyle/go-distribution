package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

type Histogram struct {
}

func (h *Histogram) WriteHist(s *Settings, tokenDict map[string]uint64) {
	maxTokenLen := 0
	maxVal := uint64(0)

	maxValueWidth := 0
	maxPctWidth := 0

	pairlist := NewPairList(tokenDict)
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

		if uint(i) >= s.Height-1 {
			break
		}
	}

	s.EndTime = time.Now().UnixNano()
	totalMillis := float64(s.EndTime-s.StartTime) / 1e6
	if s.Verbose {

		os.Stderr.WriteString(fmt.Sprintf("tokens/lines examined: %s\n", humanize.Comma(int64(s.TotalObjects))))
		os.Stderr.WriteString(fmt.Sprintf(" tokens/lines matched: %s\n", humanize.Comma(int64(s.TotalValues))))
		os.Stderr.WriteString(fmt.Sprintf("       histogram keys: %d\n", len(tokenDict)))
		os.Stderr.WriteString(fmt.Sprintf("              runtime: %sms\n", humanize.Commaf(totalMillis)))
	}

	histWidth := int(s.Width) - (maxTokenLen + 1) - (maxValueWidth + 1) - (maxPctWidth + 1) - 1

	os.Stderr.WriteString(rjust("Key", maxTokenLen))
	os.Stderr.WriteString("|")
	os.Stderr.WriteString(ljust("Ct", maxValueWidth))
	os.Stderr.WriteString(" ")
	os.Stderr.WriteString(ljust("(Pct)", maxPctWidth))
	os.Stderr.WriteString("  Histogram")
	os.Stderr.WriteString(s.KeyColour)
	os.Stderr.WriteString("\n")

	outputLimit := pairlist.Len()
	if outputLimit > int(s.Height) {
		outputLimit = int(s.Height)
	}
	for i, p := range (pairlist)[:outputLimit] {
		os.Stdout.WriteString(rjust(p.key, maxTokenLen))
		os.Stdout.WriteString(s.RegularColour)
		os.Stdout.WriteString("|")
		os.Stdout.WriteString(s.CtColour)

		outVal := fmt.Sprintf("%d", p.value)
		os.Stdout.WriteString(rjust(outVal, maxValueWidth))
		os.Stdout.WriteString(" ")

		pctStr := fmt.Sprintf("(%2.2f%%)", float64(p.value)*1.0/float64(totalValue)*100.0)
		os.Stdout.WriteString(s.PctColour)
		os.Stdout.WriteString(rjust(pctStr, maxPctWidth))
		os.Stdout.WriteString(" ")

		os.Stdout.WriteString(s.GraphColour)
		os.Stdout.WriteString(h.HistogramBar(s, histWidth, maxVal, p.value))

		if i == outputLimit-1 {
			os.Stdout.WriteString(s.RegularColour)
			break
		} else {
			os.Stdout.WriteString(s.KeyColour)
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

	nextStat := time.Now().Add(time.Duration(s.StatInterval))

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

		if s.Verbose && time.Now().After(nextStat) {
			os.Stderr.WriteString(
				fmt.Sprintf(
					"tokens/lines examined: %s ; hash prunes: %s...\r",
					humanize.Comma(int64(s.TotalObjects)),
					humanize.Comma(int64(s.NumPrunes)),
				))
			nextStat = time.Now().Add(time.Duration(s.StatInterval))
		}
	}
}

func (i *InputReader) PruneKeys(s *Settings) {
	prunedTokenCounts := make(map[string]uint64, s.MaxKeys)
	pl := NewPairList(i.TokenDict)
	sort.Sort(sort.Reverse(pl))

	for i, p := range pl {
		prunedTokenCounts[p.key] = p.value
		if uint(i)+1 > s.MaxKeys {
			break
		}
	}
	i.TokenDict = prunedTokenCounts
	s.NumPrunes += 1
}

func (i *InputReader) ReadPretalliedTokens(s *Settings) {
	// the input is already just a series of keys with the frequency of the
	// keys precomputed, as in "du -sb" - vk means the number is first, key
	// second. kv means key first, number second
	vk := regexp.MustCompile(`^\s*(\d+)\s+(.+)$`)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		res := vk.FindStringSubmatch(line)
		key := res[2]
		value, err := strconv.ParseUint(res[1], 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		i.TokenDict[key] = value
		s.TotalValues += value
		s.TotalObjects += 1
	}
}

type Settings struct {
	ScriptName       string
	TotalMillis      uint
	StartTime        int64
	EndTime          int64
	WidthArg         uint
	HeightArg        uint
	Width            uint
	Height           uint
	HistogramChar    string
	ColourisedOutput bool
	Logarithmic      bool
	NumOnly          string
	Verbose          bool
	GraphValues      string
	Size             string
	Tokenize         string
	MatchRegexp      string
	StatInterval     int
	NumPrunes        uint
	ColourPalette    string
	RegularColour    string
	KeyColour        string
	CtColour         string
	PctColour        string
	GraphColour      string
	TotalObjects     uint
	TotalValues      uint64
	KeyPruneInterval uint
	MaxKeys          uint
	UnicodeMode      bool
	CharWidth        float32
	GraphChars       []string
	PartialBlocks    []string
	PartialLines     []string
}

func NewSettings() *Settings {
	// default settings
	s := &Settings{
		ScriptName:       os.Args[0],
		TotalMillis:      0,
		StartTime:        time.Now().UnixNano(),
		EndTime:          0,
		WidthArg:         0,
		HeightArg:        0,
		Width:            80,
		Height:           15,
		HistogramChar:    "-",
		ColourisedOutput: false,
		Logarithmic:      false,
		NumOnly:          "XXX",
		Verbose:          false,
		GraphValues:      "",
		Size:             "",
		Tokenize:         "",
		MatchRegexp:      ".",
		StatInterval:     1e9,
		NumPrunes:        0,
		ColourPalette:    "0,0,32,35,34",
		RegularColour:    "",
		KeyColour:        "",
		CtColour:         "",
		PctColour:        "",
		GraphColour:      "",
		TotalObjects:     0,
		TotalValues:      0,
		KeyPruneInterval: 1500000,
		MaxKeys:          5000,
		UnicodeMode:      false,
		CharWidth:        1.0,
		GraphChars:       []string{},
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
		s.Width, s.Height = callTput()
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
	if s.WidthArg != 0 {
		s.Width = s.WidthArg
	}
	if s.HeightArg != 0 {
		s.Height = s.HeightArg
	}

	// maxKeys should be at least a few thousand greater than height to reduce odds
	// of throwing away high-count values that appear sparingly in the data
	if s.MaxKeys < s.Height+3000 {
		s.MaxKeys = s.Height + 3000
		if s.Verbose {
			os.Stderr.WriteString(fmt.Sprintf("Update MaxKeys to %d (height + 3000)\n", s.MaxKeys))
		}
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
		s.UnicodeMode = true
		s.HistogramChar = "□"
	}

	// sub-full character width graphing systems
	if s.HistogramChar == "pb" {
		s.CharWidth = 0.125
		s.GraphChars = s.PartialBlocks
	} else if s.HistogramChar == "pl" {
		s.CharWidth = 0.3334
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
	os.Stdout.WriteString(fmt.Sprintf("usage: <commandWithOutput> | %s\n", s.ScriptName))
	os.Stdout.WriteString("         [--size={sm|med|lg|full} | --width=<width> --height=<height>]\n")
	os.Stdout.WriteString("         [--color] [--palette=r,k,c,p,g]\n")
	os.Stdout.WriteString("         [--tokenize=<tokenChar>]\n")
	os.Stdout.WriteString("         [--graph[=[kv|vk]] [--numonly[=derivative,diff|abs,absolute,actual]]\n")
	os.Stdout.WriteString("         [--char=<barChars>|<substitutionString>]\n")
	os.Stdout.WriteString("         [--help] [--verbose]\n")
	os.Stdout.WriteString(fmt.Sprintf("  --keys=K       every %d values added, prune hash to K keys (default 5000)\n", s.KeyPruneInterval))
	os.Stdout.WriteString("  --char=C       character(s) to use for histogram character, some substitutions follow:\n")
	os.Stdout.WriteString("        pl       Use 1/3-width unicode partial lines to simulate 3x actual terminal width\n")
	os.Stdout.WriteString("        pb       Use 1/8-width unicode partial blocks to simulate 8x actual terminal width\n")
	os.Stdout.WriteString("        ba       (▬) Bar\n")
	os.Stdout.WriteString("        bl       (Ξ) Building\n")
	os.Stdout.WriteString("        em       (—) Emdash\n")
	os.Stdout.WriteString("        me       (⋯) Mid-Elipses\n")
	os.Stdout.WriteString("        di       (♦) Diamond\n")
	os.Stdout.WriteString("        dt       (•) Dot\n")
	os.Stdout.WriteString("        sq       (□) Square\n")
	os.Stdout.WriteString("  --color        colourise the output\n")
	os.Stdout.WriteString("  --graph[=G]    input is already key/value pairs. vk is default:\n")
	os.Stdout.WriteString("        kv       input is ordered key then value\n")
	os.Stdout.WriteString("        vk       input is ordered value then key\n")
	os.Stdout.WriteString("  --height=N     height of histogram, headers non-inclusive, overrides --size\n")
	os.Stdout.WriteString("  --help         get help\n")
	os.Stdout.WriteString("  --logarithmic  logarithmic graph\n")
	os.Stdout.WriteString("  --match=RE     only match lines (or tokens) that match this regexp, some substitutions follow:\n")
	os.Stdout.WriteString("        word     ^[A-Z,a-z]+\\$ - tokens/lines must be entirely alphabetic\n")
	os.Stdout.WriteString("        num      ^\\d+\\$        - tokens/lines must be entirely numeric\n")
	os.Stdout.WriteString("  --numonly[=N]  input is numerics, simply graph values without labels\n")
	os.Stdout.WriteString("        actual   input is just values (default - abs, absolute are synonymous to actual)\n")
	os.Stdout.WriteString("        diff     input monotonically-increasing, graph differences (of 2nd and later values)\n")
	os.Stdout.WriteString("  --palette=P    comma-separated list of ANSI colour values for portions of the output\n")
	os.Stdout.WriteString("                 in this order: regular, key, count, percent, graph. implies --color.\n")
	os.Stdout.WriteString("  --rcfile=F     use this rcfile instead of ~/.distributionrc - must be first argument!\n")
	os.Stdout.WriteString("  --size=S       size of histogram, can abbreviate to single character, overridden by --width/--height\n")
	os.Stdout.WriteString("        small    40x10\n")
	os.Stdout.WriteString("        medium   80x20\n")
	os.Stdout.WriteString("        large    120x30\n")
	os.Stdout.WriteString("        full     terminal width x terminal height (approximately)\n")
	os.Stdout.WriteString("  --tokenize=RE  split input on regexp RE and make histogram of all resulting tokens\n")
	os.Stdout.WriteString("        word     [^\\w] - split on non-word characters like colons, brackets, commas, etc\n")
	os.Stdout.WriteString("        white    \\s    - split on whitespace\n")
	os.Stdout.WriteString("  --width=N      width of the histogram report, N characters, overrides --size\n")
	os.Stdout.WriteString("  --verbose      be verbose\n")
	os.Stdout.WriteString("\n")
	os.Stdout.WriteString("You can use single-characters options, like so: -h=25 -w=20 -v. You must still include the =\n")
	os.Stdout.WriteString("\n")
	os.Stdout.WriteString("Samples:\n")
	os.Stdout.WriteString(fmt.Sprintf("  du -sb /etc/* | %s --palette=0,37,34,33,32 --graph\n", s.ScriptName))
	os.Stdout.WriteString(fmt.Sprintf("  du -sk /etc/* | awk '{print $2\" \"$1}' | %s --graph=kv\n", s.ScriptName))
	os.Stdout.WriteString(fmt.Sprintf("  zcat /var/log/syslog*gz | %s --char=o --tokenize=white\n", s.ScriptName))
	os.Stdout.WriteString(fmt.Sprintf("  zcat /var/log/syslog*gz | awk '{print \\$5}'  | %s -t=word -m-word -h=15 -c=/\n", s.ScriptName))
	os.Stdout.WriteString(fmt.Sprintf("  zcat /var/log/syslog*gz | cut -c 1-9        | %s -width=60 -height=10 -char=em\n", s.ScriptName))
	os.Stdout.WriteString(fmt.Sprintf("  find /etc -type f       | cut -c 6-         | %s -tokenize=/ -w=90 -h=35 -c=dt\n", s.ScriptName))
	os.Stdout.WriteString(fmt.Sprintf("  cat /usr/share/dict/words | awk '{print length(\\$1)}' | %s -c=* -w=50 -h=10 | sort -n\n", s.ScriptName))
	os.Stdout.WriteString("\n")
}

type pair struct {
	key   string
	value uint64
}

type Pairlist []pair

func (pl Pairlist) Len() int { return len(pl) }

func (pl Pairlist) Less(i, j int) bool {
	if pl[i].value == pl[j].value {
		return pl[i].key < pl[j].key
	}
	return pl[i].value < pl[j].value
}

func (pl Pairlist) Swap(i, j int) {
	pl[i], pl[j] = pl[j], pl[i]
}

// NewPairList returns a Pairlist containing pairs (key, value) from the give map
func NewPairList(m map[string]uint64) Pairlist {
	p := make(Pairlist, len(m))

	i := 0
	for k, v := range m {
		p[i] = pair{k, v}
		i++
	}

	return p
}

// TotalValues returns the sum of values across all pairs in the PairList
func (pl *Pairlist) TotalValues() uint64 {
	totalValue := uint64(0)
	for _, p := range *pl {
		totalValue += p.value
	}

	return totalValue
}

func callTput() (uint, uint) {
	cmdName := "echo"
	cmdArgs := []string{"$(tput cols) $(tput lines)"}

	var (
		cmdOut  []byte
		err     error
		columns uint64
		lines   uint64
	)

	if cmdOut, err = exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		log.Fatal(err)
	}
	columnLinesStr := strings.Split(string(cmdOut), " ")
	if columns, err = strconv.ParseUint(columnLinesStr[0], 10, 16); err != nil {
		log.Fatal(err)
	}
	if lines, err = strconv.ParseUint(columnLinesStr[1], 10, 16); err != nil {
		log.Fatal(err)
	}

	return uint(columns), uint(lines)
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
