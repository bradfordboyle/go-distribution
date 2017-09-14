package settings

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strconv"
	"strings"
	"time"
)

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

func NewSettings(scriptName string, args []string) *Settings {
	// default settings
	s := &Settings{
		ScriptName:       scriptName,
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
	if len(args) > 0 && strings.HasPrefix(args[0], "--rcfile") {
		rcFile = strings.Split(args[0], "=")[1]
	} else {
		u, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		rcFile = path.Join(u.HomeDir, ".distributionrc")
	}

	// parse opts from the rcFile if it exists
	// don't die or in fact do anything if rcfile doesn't exist
	file, _ := os.Open(rcFile)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimRight(line, " ")
		rcOpt := strings.Split(trimmedLine, "#")[0]
		if rcOpt != "" {
			args = append(args, "")
			copy(args[1:], args[0:])
			args[0] = rcOpt
		}
	}

	// manual argument parsing
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			doUsage(s, os.Stderr)
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
		s.Width, s.Height = TerminalSize()
		s.Height -= 3
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

func doUsage(s *Settings, writer io.Writer) {
	io.WriteString(writer, "")
	io.WriteString(writer, fmt.Sprintf("usage: <commandWithOutput> | %s\n", s.ScriptName))
	io.WriteString(writer, "         [--size={sm|med|lg|full} | --width=<width> --height=<height>]\n")
	io.WriteString(writer, "         [--color] [--palette=r,k,c,p,g]\n")
	io.WriteString(writer, "         [--Tokenize=<tokenChar>]\n")
	io.WriteString(writer, "         [--graph[=[kv|vk]] [--numonly[=derivative,diff|abs,absolute,actual]]\n")
	io.WriteString(writer, "         [--char=<barChars>|<substitutionString>]\n")
	io.WriteString(writer, "         [--help] [--verbose]\n")
	io.WriteString(writer, fmt.Sprintf("  --keys=K       every %d values added, prune hash to K keys (default 5000)\n", s.KeyPruneInterval))
	io.WriteString(writer, "  --char=C       character(s) to use for histogram character, some substitutions follow:\n")
	io.WriteString(writer, "        pl       Use 1/3-width unicode partial lines to simulate 3x actual terminal width\n")
	io.WriteString(writer, "        pb       Use 1/8-width unicode partial blocks to simulate 8x actual terminal width\n")
	io.WriteString(writer, "        ba       (▬) Bar\n")
	io.WriteString(writer, "        bl       (Ξ) Building\n")
	io.WriteString(writer, "        em       (—) Emdash\n")
	io.WriteString(writer, "        me       (⋯) Mid-Elipses\n")
	io.WriteString(writer, "        di       (♦) Diamond\n")
	io.WriteString(writer, "        dt       (•) Dot\n")
	io.WriteString(writer, "        sq       (□) Square\n")
	io.WriteString(writer, "  --color        colourise the output\n")
	io.WriteString(writer, "  --graph[=G]    input is already key/value pairs. vk is default:\n")
	io.WriteString(writer, "        kv       input is ordered key then value\n")
	io.WriteString(writer, "        vk       input is ordered value then key\n")
	io.WriteString(writer, "  --height=N     height of histogram, headers non-inclusive, overrides --size\n")
	io.WriteString(writer, "  --help         get help\n")
	io.WriteString(writer, "  --logarithmic  logarithmic graph\n")
	io.WriteString(writer, "  --match=RE     only match lines (or tokens) that match this regexp, some substitutions follow:\n")
	io.WriteString(writer, "        word     ^[A-Z,a-z]+\\$ - tokens/lines must be entirely alphabetic\n")
	io.WriteString(writer, "        num      ^\\d+\\$        - tokens/lines must be entirely numeric\n")
	io.WriteString(writer, "  --numonly[=N]  input is numerics, simply graph values without labels\n")
	io.WriteString(writer, "        actual   input is just values (default - abs, absolute are synonymous to actual)\n")
	io.WriteString(writer, "        diff     input monotonically-increasing, graph differences (of 2nd and later values)\n")
	io.WriteString(writer, "  --palette=P    comma-separated list of ANSI colour values for portions of the output\n")
	io.WriteString(writer, "                 in this order: regular, key, count, percent, graph. implies --color.\n")
	io.WriteString(writer, "  --rcfile=F     use this rcfile instead of ~/.distributionrc - must be first argument!\n")
	io.WriteString(writer, "  --size=S       size of histogram, can abbreviate to single character, overridden by --width/--height\n")
	io.WriteString(writer, "        small    40x10\n")
	io.WriteString(writer, "        medium   80x20\n")
	io.WriteString(writer, "        large    120x30\n")
	io.WriteString(writer, "        full     terminal width x terminal height (approximately)\n")
	io.WriteString(writer, "  --Tokenize=RE  split input on regexp RE and make histogram of all resulting tokens\n")
	io.WriteString(writer, "        word     [^\\w] - split on non-word characters like colons, brackets, commas, etc\n")
	io.WriteString(writer, "        white    \\s    - split on whitespace\n")
	io.WriteString(writer, "  --width=N      width of the histogram report, N characters, overrides --size\n")
	io.WriteString(writer, "  --verbose      be verbose\n")
	io.WriteString(writer, "\n")
	io.WriteString(writer, "You can use single-characters options, like so: -h=25 -w=20 -v. You must still include the =\n")
	io.WriteString(writer, "\n")
	io.WriteString(writer, "Samples:\n")
	io.WriteString(writer, fmt.Sprintf("  du -sb /etc/* | %s --palette=0,37,34,33,32 --graph\n", s.ScriptName))
	io.WriteString(writer, fmt.Sprintf("  du -sk /etc/* | awk '{print $2\" \"$1}' | %s --graph=kv\n", s.ScriptName))
	io.WriteString(writer, fmt.Sprintf("  zcat /var/log/syslog*gz | %s --char=o --Tokenize=white\n", s.ScriptName))
	io.WriteString(writer, fmt.Sprintf("  zcat /var/log/syslog*gz | awk '{print \\$5}'  | %s -t=word -m-word -h=15 -c=/\n", s.ScriptName))
	io.WriteString(writer, fmt.Sprintf("  zcat /var/log/syslog*gz | cut -c 1-9        | %s -width=60 -height=10 -char=em\n", s.ScriptName))
	io.WriteString(writer, fmt.Sprintf("  find /etc -type f       | cut -c 6-         | %s -Tokenize=/ -w=90 -h=35 -c=dt\n", s.ScriptName))
	io.WriteString(writer, fmt.Sprintf("  cat /usr/share/dict/words | awk '{print length(\\$1)}' | %s -c=* -w=50 -h=10 | sort -n\n", s.ScriptName))
	io.WriteString(writer, "\n")
}

func TerminalSize() (uint, uint) {

	stdErr := new(bytes.Buffer)
	cmd := exec.Command("stty", "size")
	// from https://stackoverflow.com/a/15840610
	// It sounds daft at first, but you will probably find that you can use
	// either stdout or stderr as the input for stty and it will adjust the
	// terminal
	cmd.Stdin = os.Stderr
	cmd.Stderr = stdErr

	cmdOut, err := cmd.Output()
	if err != nil {
		log.Fatalf("Error: `%s`; Command stderr: %#v", err.Error(), stdErr.String())
	}

	columnLinesStr := strings.Split(strings.TrimRight(string(cmdOut), "\n"), " ")

	lines, err := strconv.ParseUint(columnLinesStr[0], 10, 16)
	if err != nil {
		log.Fatal(err)
	}

	columns, err := strconv.ParseUint(columnLinesStr[1], 10, 16)
	if err != nil {
		log.Fatal(err)
	}

	return uint(columns), uint(lines)
}
