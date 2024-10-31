package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/variation"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type EvaluationResult struct {
	TrackEvents   bool        `json:"trackEvents"`
	VariationType string      `json:"variationType"`
	Failed        bool        `json:"failed"`
	Version       string      `json:"version"`
	Reason        string      `json:"reason"`
	ErrorCode     string      `json:"errorCode"`
	Value         interface{} `json:"value"`
	Cacheable     bool       `json:"cacheable"`
}

type FlagConfig struct {
	Flags map[string]*variation.Flag `yaml:"flags"`
}

var (
	configFile        string
	flagName         string
	evaluationContext string
)

func init() {
	evaluateCmd := &cobra.Command{
		Use:   "evaluate",
		Short: "Evaluate feature flags",
		Long:  `Evaluate feature flags based on configuration and context`,
		RunE:  runEvaluate,
	}

	evaluateCmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to flag configuration file (required)")
	evaluateCmd.Flags().StringVarP(&flagName, "flag", "f", "", "Name of the flag to evaluate")
	evaluateCmd.Flags().StringVarP(&evaluationContext, "evaluation-context", "e", "{}", "Evaluation context in JSON format")

	evaluateCmd.MarkFlagRequired("config")

	rootCmd.AddCommand(evaluateCmd)
}

func runEvaluate(cmd *cobra.Command, args []string) error {
	config, err := loadConfig(configFile)
	if err != nil {
		return err
	}

	context, err := parseContext(evaluationContext)
	if err != nil {
		return err
	}

	if flagName == "" {
		return evaluateAllFlags(config, context)
	}

	return evaluateFlag(flagName, config, context)
}

func loadConfig(path string) (*FlagConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config FlagConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return &config, nil
}

func parseContext(contextJSON string) (*ffuser.User, error) {
	var context map[string]interface{}
	if err := json.Unmarshal([]byte(contextJSON), &context); err != nil {
		return nil, fmt.Errorf("error parsing evaluation context: %w", err)
	}

	targetingKey, ok := context["targetingKey"].(string)
	if !ok {
		return nil, fmt.Errorf("targetingKey must be a string")
	}

	user := ffuser.NewUser(targetingKey)
	for k, v := range context {
		if k != "targetingKey" {
			user.Custom[k] = v
		}
	}

	return user, nil
}

func evaluateFlag(name string, config *FlagConfig, user *ffuser.User) error {
	flag, exists := config.Flags[name]
	if !exists {
		return fmt.Errorf("flag %s not found in configuration", name)
	}

	result, err := evaluateSingleFlag(flag, user)
	if err != nil {
		return err
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

func evaluateSingleFlag(flag *variation.Flag, user *ffuser.User) (*EvaluationResult, error) {
	eval := flag.Evaluate(user)
	
	return &EvaluationResult{
		TrackEvents:   eval.TrackEvents,
		VariationType: eval.VariationType,
		Failed:        eval.Failed,
		Version:       eval.Version,
		Reason:        string(eval.Reason),
		ErrorCode:     eval.ErrorCode,
		Value:         eval.Value,
		Cacheable:     eval.Cacheable,
	}, nil
}

func evaluateAllFlags(config *FlagConfig, user *ffuser.User) error {
	results := make(map[string]*EvaluationResult)

	for flagName, flag := range config.Flags {
		result, err := evaluateSingleFlag(flag, user)
		if err != nil {
			return fmt.Errorf("error evaluating flag %s: %w", flagName, err)
		}
		results[flagName] = result
	}

	output, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	fmt.Println(string(output))
	return nil
}