package main

import (
	"log"

	"github.com/jessevdk/go-flags"
)

func main() {
	var opts struct {
		InputFile   string `short:"f" long:"input-file" description:"Location of the flag file you want to convert." required:"true"` //nolint: lll
		InputFormat string `long:"input-format" description:"Format of your input file (YAML, JSON or TOML)" required:"false"`        //nolint: lll

		OutputFile   string `short:"o" long:"output-file" description:"Location of the flag file you want to convert." required:"false"` //nolint: lll
		OutputFormat string `long:"output-format" description:"Format of your output file (YAML, JSON or TOML)" required:"false"`        //nolint: lll
	}
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal("impossible to parse command line parameters", err)
	}

	c := FlagConverter{
		InputFile:    opts.InputFile,
		InputFormat:  opts.InputFormat,
		OutputFile:   opts.OutputFile,
		OutputFormat: opts.OutputFormat,
	}

	err = c.Migrate()
	if err != nil {
		log.Fatal(err)
	}
}
