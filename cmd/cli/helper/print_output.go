package helper

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

type Level = string

const (
	WarnLevel Level = "WARNING"
	InfoLevel Level = "INFO"
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

func (o *Output) PrintLines(cmd *cobra.Command) {
	for _, line := range o.Lines {
		switch line.Level {
		case WarnLevel:
			pterm.Warning.Println(line.Text)
		default: // InfoLevel
			pterm.Info.Println(line.Text)
		}
	}
}
