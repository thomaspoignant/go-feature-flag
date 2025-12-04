package helper

import (
	"fmt"
	"os"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

type Level = string

const (
	WarnLevel    Level = "WARNING"
	InfoLevel    Level = "INFO"
	ErrorLevel   Level = "ERROR"
	DefaultLevel Level = "DEFAULT"
)

type OutputLine struct {
	Text  string
	Level Level
}
type Output struct {
	Lines []OutputLine
}

var exitFunc = os.Exit

func (o *Output) Add(line string, level Level) Output {
	o.Lines = append(o.Lines, OutputLine{Text: line, Level: level})
	return *o
}

func (o *Output) PrintLines(cmd *cobra.Command) {
	for _, line := range o.Lines {
		var outputText string
		writer := cmd.OutOrStdout()

		switch line.Level {
		case InfoLevel:
			outputText = pterm.Info.Sprint(line.Text)
		case WarnLevel:
			outputText = pterm.Warning.Sprint(line.Text)
		case ErrorLevel:
			outputText = pterm.Error.Sprint(line.Text)
			writer = cmd.ErrOrStderr()
		default:
			outputText = pterm.Sprint(line.Text)
		}
		_, err := fmt.Fprintln(writer, outputText)
		if err != nil {
			PrintFatalAndExit(err)
		}
	}
}

func PrintFatalAndExit(err error) {
	pterm.Error.Printf("error executing command: %v\n", err)
	exitFunc(1)
}
