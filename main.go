package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	app := cli.NewApp()
	app.Name = "f-trigger"
	app.Usage = "Trigger events based on a frequency on stdin"
	app.Action = run

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "delimiter, d",
			Value: "\n",
			Usage: "This character determines the start and end of each timed unit (not yet implemented)",
		},
	}

	cli.HelpFlag = cli.BoolFlag{
		Name:  "help, h",
		Usage: "Show this help message",
	}
	cli.VersionFlag = cli.BoolFlag{
		Name:  "version, V",
		Usage: "Print the version",
	}

	cli.AppHelpTemplate = `{{.Name}} - {{.Usage}}

Version {{.Version}}

Usage:
   {{.HelpName}} [options] [triggers]...
   {{if .VisibleFlags}}
Options:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}
`

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func run(c *cli.Context) error {
	var err error

	lastTime := time.Now()

	scanner := bufio.NewScanner(os.Stdin)

	var bpmFormat string
	if terminal.IsTerminal(int(os.Stdout.Fd())) {
		bpmFormat = "\r\033[K%.0f bpm"
	} else {
		bpmFormat = "%.0f bpm\n"
	}

	for scanner.Scan() {
		fmt.Printf(bpmFormat, 1/time.Now().Sub(lastTime).Minutes())
		lastTime = time.Now()
	}
	if err = scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input: ", err.Error())
	}

	return err
}
