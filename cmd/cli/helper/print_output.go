package helper

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

type Level = string

const (
	WarnLevel  Level = "WARNING"
	InfoLevel  Level = "INFO"
	ErrorLevel Level = "ERROR"
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

func (o *Output) PrintLines(_ *cobra.Command) {
	for _, line := range o.Lines {
		switch line.Level {
		case InfoLevel:
			pterm.Info.Println(line.Text)
		case WarnLevel:
			pterm.Warning.Println(line.Text)
		case ErrorLevel:
			pterm.Error.Println(line.Text)
		default:
			pterm.Println(line.Text)
		}
	}
}
