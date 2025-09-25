package evaluate

import (
	"encoding/json"

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
			}

			err := retrieverConf.IsValid()
			if err != nil {
				return err
			}

			return runEvaluate(cmd, args, evalFlagFormat, retrieverConf, evalFlag, evalCtx)
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
		"auth-token", "", "Authentication token to access your private configuration file")
	evaluateCmd.Flags().StringVar(&githubToken,
		"github-token", "", "Authentication token to access your private configuration file on GitHub")
	evaluateCmd.Flags().StringVar(&repositorySlug,
		"repository-slug", "", "Repository slug to access your private configuration file on GitHub")
	evaluateCmd.Flags().StringVar(&branch,
		"branch", "", "Branch to access your private configuration file on GitHub")
	evaluateCmd.Flags().StringVar(&evalFlag,
		"flag", "", "Name of the flag to evaluate, if empty we will return the evaluation of all the flags")
	evaluateCmd.Flags().StringVar(&evalCtx,
		"ctx", "{}", "Evaluation context in JSON format")
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
