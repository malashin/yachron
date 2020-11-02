package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/atotto/clipboard"
	"github.com/malashin/ffinfo"
	"golang.org/x/crypto/ssh/terminal"
)

var input = "input.txt"

func main() {
	cpbText, err := clipboard.ReadAll()
	if err != nil {
		panic(err)
	}

	fmt.Printf("INPUT: %q\n\n", cpbText)
	lines := strings.Split(strings.Replace(cpbText, "\r\n", "\n", -1), "\n")

	maxLen := 0
	for _, input := range lines {
		n := utf8.RuneCountInString(path.Base(input))
		if n > maxLen {
			maxLen = n
		}
	}

	fmt.Printf("%v\t%v\n", truncPad("FILENAME:", maxLen, 'l'), "DURATION:")

	var total float64
	for _, input := range lines {
		f, err := ffinfo.Probe(input)
		if err != nil {
			panic(err)
		}
		d, err := strconv.ParseFloat(f.Format.Duration, 64)
		if err != nil {
			panic(err)
		}

		fmt.Printf("%v\t%v\n", truncPad(path.Base((input)), maxLen, 'l'), secondsToHHMMSSMS(d))

		total += d
	}
	fmt.Printf("%v\t%v\n", truncPad("TOTAL MINUTES:", maxLen, 'r'), math.Floor(total/60))
	fmt.Printf("%v\t%v\n", truncPad("TOTAL HH:MM:SS.MS:", maxLen, 'r'), secondsToHHMMSSMS(total))

	err = waitForAnyKey()
	if err != nil {
		panic(err)
	}
}

// secondsToHHMMSS converts seconds (SS | SS.MS) to timecode (HH:MM:SS,MS).
func secondsToHHMMSSMS(s float64) string {
	// s, _ := strconv.ParseFloat(seconds, 64)
	hh := math.Floor(s / 3600)
	mm := math.Floor((s - hh*3600) / 60)
	ss := int64(math.Floor(s - hh*3600 - mm*60))
	ms := s - float64(int(s))

	hhString := strconv.FormatInt(int64(hh), 10)
	mmString := strconv.FormatInt(int64(mm), 10)
	ssString := strconv.FormatInt(int64(ss), 10)
	msString := fmt.Sprintf("%03d", int((math.Round(ms*1000)/1000)*1000))

	if msString == "1000" {
		ssString = strconv.FormatInt(int64(ss+1), 10)
		msString = "000"
	}

	if hh < 10 {
		hhString = "0" + hhString
	}
	if mm < 10 {
		mmString = "0" + mmString
	}
	if ss < 10 {
		ssString = "0" + ssString
	}
	return hhString + ":" + mmString + ":" + ssString + "." + msString
}

// truncPad truncs or pads string to needed length.
// If side is 'r' the string is padded and aligned to the right side.
// Otherwise it is aligned to the left side.
func truncPad(s string, n int, side byte) string {
	len := utf8.RuneCountInString(s)
	if len > n {
		return string([]rune(s)[0:n-3]) + "..."
	}
	if side == 'r' {
		return strings.Repeat(" ", n-len) + s
	}
	return s + strings.Repeat(" ", n-len)
}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// waitForAnyKey await for any key press to continue.
func waitForAnyKey() error {
	fd := int(os.Stdin.Fd())
	if !terminal.IsTerminal(fd) {
		return fmt.Errorf("it's not a terminal descriptor")
	}
	state, err := terminal.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("cannot set raw mode")
	}
	defer terminal.Restore(fd, state)

	b := [1]byte{}
	os.Stdin.Read(b[:])
	return nil
}
