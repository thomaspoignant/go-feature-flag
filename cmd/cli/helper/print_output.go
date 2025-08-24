package helper

import (
	"fmt"

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

func (o *Output) Add(line string, level Level) Output {
	o.Lines = append(o.Lines, OutputLine{Text: line, Level: level})
	return *o
}

func (o *Output) FormatError(err error) error {
	formattedText := pterm.Error.Sprint(err.Error())
	return fmt.Errorf(formattedText)
}

func (o *Output) PrintLines(cmd *cobra.Command) {
	for _, line := range o.Lines {
		var outputText string
		switch line.Level {
		case InfoLevel:
			outputText = pterm.Info.Sprint(line.Text)
		case WarnLevel:
			outputText = pterm.Warning.Sprint(line.Text)
		case ErrorLevel:
			outputText = pterm.Error.Sprint(line.Text)
		default:
			outputText = pterm.Sprint(line.Text)
		}
		fmt.Fprintln(cmd.OutOrStdout(), outputText)
	}
}
