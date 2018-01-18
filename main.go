package main

import (
    "bufio"
    "fmt"
    "os"
    "time"

    "github.com/urfave/cli"
    "golang.org/x/crypto/ssh/terminal"
)

func main()  {
    app := cli.NewApp()
    app.Name = "f-trigger"
    app.Usage = "Trigger events based on a frequency on stdin"
    app.Action = run

    // Hide the help subcommand. We don't use subcommands in this app.
    app.HideHelp = true

    app.Flags = []cli.Flag{
        cli.StringFlag{
            Name: "delimiter, d",
            Value: "\n",
            Usage: "This character determines the start and end of each timed unit (not yet implemented)",
        },
    }

    if err := app.Run(os.Args); err != nil {
		fmt.Println(err.Error())
        os.Exit(1)
    }
}

func run(c *cli.Context) error {
    var err error

    lastTime := time.Now()

    scanner := bufio.NewScanner(os.Stdin)

    var bmpFormat string
    if terminal.IsTerminal(int(os.Stdout.Fd())) {
        bmpFormat = "\r\033[K%.0f bmp"
    } else {
        bmpFormat = "%.0f bmp\n"
    }
    
    for scanner.Scan() {
        fmt.Printf(bmpFormat, 1 / time.Now().Sub(lastTime).Minutes())
        lastTime = time.Now()
    }
    if err = scanner.Err(); err != nil {
        fmt.Fprintln(os.Stderr, "Error reading input: ", err.Error())
    }

    return err
}