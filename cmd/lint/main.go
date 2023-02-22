package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jessevdk/go-flags"
)

func main() {
	var opts struct {
		InputFile   string `short:"f" long:"input-file" description:"Location of the flag file you want to lint." required:"true"` //nolint: lll
		InputFormat string `long:"input-format" description:"Format of your input file (YAML, JSON or TOML)" required:"true"`      //nolint: lll
	}
	_, err := flags.Parse(&opts)
	if flags.WroteHelp(err) {
		os.Exit(0)
	}
	if err != nil {
		log.Fatal("impossible to parse command line parameters", err)
	}

	linter := Linter{
		InputFile:   opts.InputFile,
		InputFormat: opts.InputFormat,
	}

	errs := linter.Lint()
	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Fprintf(os.Stderr, "%s\n", err)
		}
		os.Exit(len(errs))
	}
}
