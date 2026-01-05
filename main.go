package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"time"

	"golang.org/x/term"
)

func main() {
	var delimiter = flag.String("d", "\n", "delimiter; This character or string determines the start and end of each timed unit")
	flag.Parse()

	needle := []byte(*delimiter)
	needleLength := len(needle)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if index := bytes.Index(data, needle); index >= 0 {
			return index + needleLength, []byte(""), nil
		}

		if atEOF {
			return 0, nil, nil
		}

		return len(data), nil, nil
	})

	var bpmFormat string
	if term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stderr.Fd())) {
		bpmFormat = "\033[F\033[K%.0f bpm"
	} else if term.IsTerminal(int(os.Stdout.Fd())) {
		bpmFormat = "\r\033[K%.0f bpm"
	} else {
		bpmFormat = "%.0f bpm\n"
	}

	lastTime := time.Now()

	for scanner.Scan() {
		fmt.Fprintf(os.Stderr, bpmFormat, 1/time.Since(lastTime).Minutes())
		lastTime = time.Now()
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input: ", err.Error())
	}
}
