package main

import (
	"github.com/spf13/cobra"
	"github.com/thomaspoignant/go-feature-flag/cmd/cli/evaluate"
)

var (
	evalFlagFormat string
	evalConfigFile string
	evalFlag       string
	evalCtx        string
)

func init() {
	evaluateCmd := &cobra.Command{
		Use:   "evaluate",
		Short: "Evaluate feature flags based on configuration and context",
		Long:  "Evaluate feature flags based on configuration and context, if no specific flag requested it will evaluate all flags",
		RunE: func(cmd *cobra.Command, args []string) error {
			return evaluate.RunEvaluate(cmd, args, evalFlagFormat, evalConfigFile, evalFlag, evalCtx)

		},
	}
	rootCmd.AddCommand(evaluateCmd)
	evaluateCmd.Flags().StringVarP(&evalFlagFormat, "format", "f", "yaml", "Format of your input file (YAML, JSON or TOML)")
	evaluateCmd.Flags().StringVarP(&evalConfigFile, "config", "c", "", "Location of your GO Feature Flag local configuration file")
	evaluateCmd.Flags().StringVar(&evalFlag, "flag", "", "Name of the flag to evaluate, if empty we will return the evaluation of all the flags")
	evaluateCmd.Flags().StringVar(&evalCtx, "ctx", "{}", "Evaluation context in JSON format")
	_ = evaluateCmd.MarkFlagRequired("config")
}
