package main

import (
	"bufio"
	"bytes"
	"cmp"
	"flag"
	"fmt"
	"log/slog"
	"math"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"
)

type bracket struct {
	lowerBound *int64
	upperBound *int64
	command    string
}

func (b *bracket) osCommand() *exec.Cmd {
	return exec.Command("sh", "-c", b.command)
}

func (b *bracket) String() string {
	var lowerBound, upperBound string

	if b.lowerBound != nil {
		lowerBound = strconv.FormatInt(*b.lowerBound, 10)
	}
	if b.upperBound != nil {
		upperBound = strconv.FormatInt(*b.upperBound, 10)
	}

	if b.lowerBound == b.upperBound {
		return fmt.Sprintf("%s %#v", lowerBound, b.command)
	}
	return fmt.Sprintf("%s-%s %#v", lowerBound, upperBound, b.command)
}

func main() {
	var delimiter = flag.String("d", "\n", "delimiter; This character or string determines the start and end of each timed unit")
	flag.Parse()

	plan, err := makePlan(flag.Args())
	if err != nil {
		slog.Error("Error parsing commands", "error", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(scanStringFunc(*delimiter))

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
		// TODO: Calculate bpm as an average over a few seconds
		// TODO: Calculate bpm periodically, say once a second, instead of only when we see input
		bpm := math.Floor(1 / time.Since(lastTime).Minutes())
		fmt.Fprintf(os.Stderr, bpmFormat, bpm)
		lastTime = time.Now()

		go func() {
			i := slices.IndexFunc(plan, func(w bracket) bool {
				var lowerBound, upperBound float64

				if w.lowerBound != nil {
					lowerBound = float64(*w.lowerBound)
				} else {
					lowerBound = math.Inf(-1)
				}
				if w.upperBound != nil {
					upperBound = float64(*w.upperBound)
				} else {
					upperBound = math.Inf(1)
				}

				return bpm >= lowerBound && bpm <= upperBound
			})

			if i < 0 || i > len(plan) {
				return
			}

			bracket := plan[i]
			err := bracket.osCommand().Run()
			if err != nil {
				slog.Error("Command failed", "bracket", bracket.String(), "error", err)
			}
		}()
	}
	if err := scanner.Err(); err != nil {
		slog.Error("Error reading input", "error", err.Error())
		os.Exit(1)
	}
}

func makePlan(args []string) ([]bracket, error) {
	plan := []bracket{}
	unprocessed := args
	for len(unprocessed) > 0 {
		frequencies := unprocessed[0]

		if len(unprocessed) < 2 {
			return nil, fmt.Errorf("Missing command for %#v", frequencies)
		}

		command := unprocessed[1]

		firstFreq, secondFreq, isRange := strings.Cut(frequencies, "-")
		var lowerBound, upperBound *int64

		if firstFreq != "" {
			n, err := strconv.ParseInt(firstFreq, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("Invalid lower frequency in %#v: %w", frequencies, err)
			}

			lowerBound = &n
		}

		if isRange {
			if secondFreq != "" {
				n, err := strconv.ParseInt(secondFreq, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("Invalid upper frequency in %#v: %w", frequencies, err)
				}

				upperBound = &n
			}
		} else {
			upperBound = lowerBound
		}

		if lowerBound != nil && upperBound != nil && *lowerBound > *upperBound {
			return nil, fmt.Errorf("Invalid frequencies in %#v: upper bound %#v must be greater than lower bound %#v", frequencies, *upperBound, *lowerBound)
		}

		plan = append(plan, bracket{
			lowerBound: lowerBound,
			upperBound: upperBound,
			command:    command,
		})

		unprocessed = unprocessed[2:]
	}

	slices.SortFunc(plan, func(a bracket, b bracket) int {
		var aLowerBound, bLowerBound float64

		if a.lowerBound == nil {
			aLowerBound = math.Inf(-1)
		} else {
			aLowerBound = float64(*a.lowerBound)
		}
		if b.lowerBound == nil {
			bLowerBound = math.Inf(-1)
		} else {
			bLowerBound = float64(*b.lowerBound)
		}

		return cmp.Compare(aLowerBound, bLowerBound)
	})

	for i := 0; i < len(plan)-1; i++ {
		a := plan[i]
		b := plan[i+1]

		var aUpperBound, bLowerBound float64

		if a.upperBound == nil {
			aUpperBound = math.Inf(1)
		} else {
			aUpperBound = float64(*a.upperBound)
		}
		if b.lowerBound == nil {
			bLowerBound = math.Inf(-1)
		} else {
			bLowerBound = float64(*b.lowerBound)
		}

		if aUpperBound > bLowerBound {
			return nil, fmt.Errorf("frequencies overlap: %s and %s", a.String(), b.String())
		}
	}

	return plan, nil
}

func scanStringFunc(delimiter string) bufio.SplitFunc {
	needle := []byte(delimiter)
	needleLength := len(needle)

	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if index := bytes.Index(data, needle); index >= 0 {
			return index + needleLength, []byte(""), nil
		}

		if atEOF {
			return 0, nil, nil
		}

		return len(data), nil, nil
	}
}
