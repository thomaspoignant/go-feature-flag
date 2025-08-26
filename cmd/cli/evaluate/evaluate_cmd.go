package evaluate

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/thomaspoignant/go-feature-flag/cmd/cli/helper"
)

var (
	evalFlagFormat string
	evalConfigFile string
	evalFlag       string
	evalCtx        string
)

func NewEvaluateCmd() *cobra.Command {
	evaluateCmd := &cobra.Command{
		Use:   "evaluate",
		Short: "⚙️ Evaluate feature flags based on configuration and context",
		Long: "⚙️ Evaluate feature flags based on configuration and context," +
			" if no specific flag requested it will evaluate all flags",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEvaluate(cmd, args, evalFlagFormat, evalConfigFile, evalFlag, evalCtx)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	evaluateCmd.Flags().StringVarP(&evalFlagFormat,
		"format", "f", "yaml", "Format of your input file (YAML, JSON or TOML)")
	evaluateCmd.Flags().StringVarP(&evalConfigFile,
		"config", "c", "", "Location of your GO Feature Flag local configuration file")
	evaluateCmd.Flags().StringVar(&evalFlag,
		"flag", "", "Name of the flag to evaluate, if empty we will return the evaluation of all the flags")
	evaluateCmd.Flags().StringVar(&evalCtx,
		"ctx", "{}", "Evaluation context in JSON format")
	_ = evaluateCmd.MarkFlagRequired("config")
	return evaluateCmd
}

func runEvaluate(
	cmd *cobra.Command,
	_ []string,
	flagFormat string,
	configFile string,
	flag string,
	ctx string) error {
	output := helper.Output{}
	e := evaluate{
		config:        configFile,
		fileFormat:    flagFormat,
		flag:          flag,
		evaluationCtx: ctx,
	}
	result, err := e.Evaluate()
	if err != nil {
		return err
	}

	detailed, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	output.Add(string(detailed), helper.DefaultLevel)
	output.PrintLines(cmd)
	return nil
}
