package main

import (
	"fmt"
	"log"
	"os"

	"github.com/thomaspoignant/go-feature-flag/cmd/migrationcli/converter"

	"github.com/jessevdk/go-flags"
)

func main() {
	var opts struct {
		InputFile    string `short:"f" long:"input-file" description:"Location of the flag file you want to convert." required:"true"` //nolint: lll
		InputFormat  string `long:"input-format" description:"Format of your input file (YAML, JSON or TOML)" required:"true"`         //nolint: lll
		OutputFile   string `short:"o" long:"output-file" description:"Location of the converted flag file." required:"false"`         //nolint: lll
		OutputFormat string `long:"output-format" description:"Format of your output file (YAML, JSON or TOML)" required:"false"`      //nolint: lll
	}
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal("impossible to parse command line parameters", err)
	}

	c := converter.FlagConverter{
		InputFile:    opts.InputFile,
		InputFormat:  opts.InputFormat,
		OutputFormat: opts.OutputFormat,
	}

	content, err := c.Migrate()
	if err != nil {
		log.Fatal(err)
	}

	err = outputResult(content, opts.OutputFile)
	if err != nil {
		log.Fatal(err)
	}
}

// outputResult is the function that display the result of the command line.
// it can put the result directly in the console or in a file depending on the flags.
func outputResult(content []byte, outputFile string) error {
	if outputFile == "" {
		fmt.Println(string(content))
		return nil
	}
	return os.WriteFile(outputFile, content, os.ModePerm)
}
