package evaluate

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thomaspoignant/go-feature-flag/cmd/cli/helper"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	retrieverInit "github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf/init"
)

var (
	evalFlagFormat string
	kind           string
	evalConfigFile string
	path           string
	authToken      string
	githubToken    string
	repositorySlug string
	branch         string
	baseURL        string
	bucket         string
	item           string
	url            string
	method         string
	body           string
	headers        []string
	timeout        int64
	evalFlag       string
	evalCtx        string
	checkMode      bool
)

func NewEvaluateCmd() *cobra.Command {
	evaluateCmd := &cobra.Command{
		Use:   "evaluate",
		Short: "⚙️ Evaluate feature flags based on configuration and context",
		Example: `
# Evaluate a specific flag using deprecated flag --config
evaluate --config ./config.yaml --flag flag1 --ctx '{"targetingKey": "user-123"}'

# Evaluate a specific flag using new flag --path
evaluate --kind file --path ./config.yaml --flag flag1 --ctx '{"targetingKey": "user-123"}'

# Evaluate a specific flag using http retriever
evaluate --kind http --url http://localhost:8080/config.yaml --header 'ContentType: application/json' --header 'X-Auth=Token' --flag flag1 --ctx '{"targetingKey": "user-123"}'
`,
		Long: "⚙️ Evaluate feature flags based on configuration and context," +
			" if no specific flag requested it will evaluate all flags",
		RunE: func(cmd *cobra.Command, args []string) error {
			parsedHeaders := parseHeaders()

			retrieverConf := retrieverconf.RetrieverConf{
				Kind:           retrieverconf.RetrieverKind(kind),
				RepositorySlug: repositorySlug,
				Branch:         branch,
				Path: func() string {
					if path != "" {
						return path
					}
					return evalConfigFile
				}(),
				GithubToken: githubToken,
				AuthToken:   authToken,
				BaseURL:     baseURL,
				Bucket:      bucket,
				Item:        item,
				URL:         url,
				Timeout:     timeout,
				HTTPMethod:  method,
				HTTPBody:    body,
				HTTPHeaders: parsedHeaders,
			}

			err := retrieverConf.IsValid()
			if err != nil {
				return err
			}

			if checkMode {
				return runCheck(cmd, retrieverConf)
			} else {
				return runEvaluate(cmd, args, evalFlagFormat, retrieverConf, evalFlag, evalCtx)
			}
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	evaluateCmd.Flags().StringVarP(&evalFlagFormat,
		"format", "f", "yaml", "Format of your input file (YAML, JSON or TOML)")
	evaluateCmd.Flags().StringVarP(&kind,
		"kind", "k", "file", "Kind of the configuration file (file, http, redis, gitlab, k8s, ...)")
	evaluateCmd.Flags().StringVarP(&evalConfigFile,
		"config", "c", "", "Location of your GO Feature Flag local configuration file")
	evaluateCmd.Flags().StringVarP(&path, "path", "p", "",
		"Path to your GO Feature Flag configuration file (local or remote)")
	evaluateCmd.Flags().StringVar(&authToken,
		"auth-token", "", "Authentication token to access your configuration file")
	evaluateCmd.Flags().StringVar(&githubToken,
		"github-token", "", "Authentication token to access your configuration file on GitHub")
	evaluateCmd.Flags().StringVar(&repositorySlug,
		"repository-slug", "", "Repository slug to access your configuration file on GitHub")
	evaluateCmd.Flags().StringVar(&branch,
		"branch", "", "Branch to access your configuration file on GitHub")
	evaluateCmd.Flags().StringVar(&baseURL,
		"base-url", "", "Base URL of your configuration file on Gitlab")
	evaluateCmd.Flags().StringVar(&bucket,
		"bucket", "", "Bucket of your configuration file on S3")
	evaluateCmd.Flags().StringVar(&item,
		"item", "", "Item of your configuration file on S3")
	evaluateCmd.Flags().StringVar(&url,
		"url", "", "URL of your configuration file on HTTP")
	evaluateCmd.Flags().StringVar(&method,
		"method", "GET", "Method to access your configuration file on HTTP")
	evaluateCmd.Flags().StringVar(&body,
		"body", "", "Body to access your configuration file on HTTP")
	evaluateCmd.Flags().StringArrayVar(&headers,
		"header", nil, "Header to access your configuration file on HTTP (may be repeated). See example of `evaluate` command for usages")
	evaluateCmd.Flags().Int64Var(&timeout,
		"timeout", 0, "Timeout in seconds to access your configuration file on HTTP")
	evaluateCmd.Flags().StringVar(&evalFlag,
		"flag", "", "Name of the flag to evaluate, if empty we will return the evaluation of all the flags")
	evaluateCmd.Flags().StringVar(&evalCtx,
		"ctx", "{}", "Evaluation context in JSON format")
	evaluateCmd.Flags().BoolVar(&checkMode,
		"check-mode", false, "Check only mode - it does not perform any evaluation and returns the configuration of spanned retriever")
	_ = evaluateCmd.Flags().MarkDeprecated("github-token", "Use auth-token instead")
	_ = evaluateCmd.Flags().MarkDeprecated("config", "Use path instead")
	_ = evaluateCmd.Flags()
	return evaluateCmd
}

func runEvaluate(
	cmd *cobra.Command,
	_ []string,
	flagFormat string,
	retrieverConf retrieverconf.RetrieverConf,
	flag string,
	ctx string) error {
	output := helper.Output{}

	r, err := retrieverInit.InitRetriever(&retrieverConf)
	if err != nil {
		return err
	}

	e := evaluate{
		retriever:     r,
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

func runCheck(
	cmd *cobra.Command,
	retrieverConf retrieverconf.RetrieverConf) error {
	output := helper.Output{}

	detailed, err := json.MarshalIndent(retrieverConf, "", "  ")
	if err != nil {
		return err
	}

	output.Add(string(detailed), helper.DefaultLevel)
	output.PrintLines(cmd)
	return nil
}

func parseHeaders() map[string][]string {
	result := make(map[string][]string)
	for _, h := range headers {
		parts := strings.SplitN(h, "=", 2)
		if len(parts) != 2 {
			parts = strings.SplitN(h, ":", 2)
		}
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		result[key] = append(result[key], val)
	}
	return result
}
