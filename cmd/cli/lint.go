package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type LintResult struct {
	Valid    bool     `json:"valid"`
	Warnings []string `json:"warnings"`
	Errors   []string `json:"errors"`
}

func init() {
	lintCmd := &cobra.Command{
		Use:   "lint",
		Short: "Lint feature flag configuration",
		Long:  `Validate feature flag configuration file for common issues and best practices`,
		RunE:  runLint,
	}

	lintCmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to flag configuration file (required)")
	lintCmd.MarkFlagRequired("config")

	rootCmd.AddCommand(lintCmd)
}

func runLint(cmd *cobra.Command, args []string) error {
	config, err := loadConfig(configFile)
	if err != nil {
		return err
	}

	result := lintConfig(config)
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

func lintConfig(config *FlagConfig) *LintResult {
	result := &LintResult{
		Valid:    true,
		Warnings: []string{},
		Errors:   []string{},
	}

	for name, flag := range config.Flags {
		// Validate flag name
		if !isValidFlagName(name) {
			result.Errors = append(result.Errors, 
				fmt.Sprintf("Invalid flag name '%s': should be lowercase with hyphens", name))
			result.Valid = false
		}

		// Validate variations
		if len(flag.Variations) == 0 {
			result.Errors = append(result.Errors, 
				fmt.Sprintf("Flag '%s' has no variations defined", name))
			result.Valid = false
		}

		// Check for default rule
		hasDefaultRule := false
		for _, rule := range flag.Rules {
			if rule.IsDefault() {
				hasDefaultRule = true
				break
			}
		}
		if !hasDefaultRule {
			result.Warnings = append(result.Warnings, 
				fmt.Sprintf("Flag '%s' has no default rule", name))
		}
	}

	return result
}

func isValidFlagName(name string) bool {
	return strings.ToLower(name) == name && !strings.Contains(name, " ")
}