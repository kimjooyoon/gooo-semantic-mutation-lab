package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/gooo-semantic-mutation-lab/internal/lab"
)

func main() {
	if len(os.Args) < 2 {
		fatal("command is required: run")
	}
	switch os.Args[1] {
	case "run":
		run(os.Args[2:])
	default:
		fatal(fmt.Sprintf("unknown command %q", os.Args[1]))
	}
}

func run(args []string) {
	set := flag.NewFlagSet("run", flag.ExitOnError)
	source := set.String("source", "", "path to declared .gooo source")
	contract := set.String("contract", "", "fixed denominator contract")
	out := set.String("out", "", "caller-owned temporary output directory")
	set.Parse(args)
	if *source == "" || *contract == "" || *out == "" {
		fatal("run requires --source, --contract, and --out")
	}
	report, err := lab.Run(*source, *contract, *out)
	if err != nil {
		fatal(err.Error())
	}
	data, err := json.Marshal(report.Summary)
	if err != nil {
		fatal(err.Error())
	}
	fmt.Println(string(data))
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
