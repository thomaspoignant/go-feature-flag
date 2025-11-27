package evaluate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thomaspoignant/go-feature-flag/cmd/cli/helper"
	rerr "github.com/thomaspoignant/go-feature-flag/cmdhelpers/err"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	retrieverInit "github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf/init"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/utils"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
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
	object         string
	namespace      string
	configMap      string
	key            string
	uri            string
	database       string
	collection     string
	container      string
	accountName    string
	accountKey     string
	table          string
	columns        []string
	evalFlag       string
	evalCtx        string
	checkMode      bool
)

// nolint:funlen
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
evaluate --kind http --url http://localhost:8080/config.yaml --header 'ContentType: application/json' --header 
'X-Auth=Token' --flag flag1 --ctx '{"targetingKey": "user-123"}'

# Evaluate a specific flag using github retriever
evaluate --kind github --repository-slug thomaspoignant/go-feature-flag --branch master --flag flag1
--ctx '{"targetingKey": "user-123"}'

# Evaluate a specific flag using gitlab retriever
evaluate --kind gitlab --base-url https://gitlab.com --repository-slug thomaspoignant/go-feature-flag
--branch master --flag flag1 --ctx '{"targetingKey": "user-123"}'

# Evaluate a specific flag using bitbucket retriever
evaluate --kind bitbucket --repository-slug thomaspoignant/go-feature-flag --branch master --flag flag1
--ctx '{"targetingKey": "user-123"}'

# Evaluate a specific flag using s3 retriever
evaluate --kind s3 --bucket my-bucket --item my-item.yaml --flag flag1 --ctx '{"targetingKey": "user-123"}'

# Evaluate a specific flag using gcs retriever
evaluate --kind googleStorage --bucket my-bucket --object my-item.yaml --flag flag1 --ctx
'{"targetingKey": "user-123"}'

# Evaluate a specific flag using configmap retriever
evaluate --kind configmap --namespace default --config-map my-configmap --key my-key.yaml --flag flag1
--ctx '{"targetingKey": "user-123"}'

# Evaluate a specific flag using mongodb retriever
evaluate --kind mongodb --uri mongodb://localhost:27017 --database my-database --collection my-collection --flag flag1
--ctx '{"targetingKey": "user-123"}'

# Evaluate a specific flag using azureblob retriever
evaluate --kind azureblob --container my-container --account-name my-account-name --account-key my-account-key
--object my-object --flag flag1 --ctx '{"targetingKey": "user-123"}'

# Evaluate a specific flag using postgres retriever
evaluate --kind postgres --table my-table --column my-column:my-column-type --flag flag1
--ctx '{"targetingKey": "user-123"}'
`,
		Long: "⚙️ Evaluate feature flags based on configuration and context," +
			" if no specific flag requested it will evaluate all flags",
		RunE: func(cmd *cobra.Command, args []string) error {
			parsedHeaders := parseHTTPHeaders()
			parsedColumns := parsePostgresColumns()

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
				Object:      object,
				Namespace:   namespace,
				ConfigMap:   configMap,
				Key:         key,
				URI:         uri,
				Database:    database,
				Collection:  collection,
				Container:   container,
				AccountName: accountName,
				AccountKey:  accountKey,
				Table:       table,
				Columns:     parsedColumns,
			}

			err := retrieverConf.IsValid()
			if err != nil {
				if rcErr, ok := err.(*rerr.RetrieverConfError); ok {
					return errors.New(rcErr.CliErrorMessage())
				}
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
		"method", "GET", "HTTP method to access your configuration file on HTTP")
	evaluateCmd.Flags().StringVar(&body,
		"body", "", "Http body to access your configuration file on HTTP")
	evaluateCmd.Flags().StringArrayVar(&headers,
		"header", nil,
		"HTTP header to access your configuration file on HTTP (may be repeated). "+
			"See example of `evaluate` command for usages")
	evaluateCmd.Flags().Int64Var(&timeout,
		"timeout", 0, "Timeout in seconds to access your configuration file on HTTP")
	evaluateCmd.Flags().StringVar(&evalFlag,
		"flag", "",
		"Name of the flag to evaluate, if empty we will return the evaluation of all the flags")
	evaluateCmd.Flags().StringVar(&evalCtx,
		"ctx", "{}", "Evaluation context in JSON format")
	evaluateCmd.Flags().BoolVar(&checkMode,
		"check-mode", false,
		"Check only mode - when set, the command will not perform any evaluation and returns "+
			"the configuration of spanned retriever")
	evaluateCmd.Flags().StringVar(&object,
		"object", "", "Object of your configuration file on GCS")
	evaluateCmd.Flags().StringVar(&namespace,
		"namespace", "default", "Namespace of your configuration file on K8s")
	evaluateCmd.Flags().StringVar(&configMap,
		"config-map", "", "Config map of your configuration file on K8s")
	evaluateCmd.Flags().StringVar(&key,
		"key", "", "Key of your configuration file on K8s")
	evaluateCmd.Flags().StringVar(&uri,
		"uri", "", "URI of your configuration file")
	evaluateCmd.Flags().StringVar(&database,
		"database", "", "Database of your configuration file on MongoDB")
	evaluateCmd.Flags().StringVar(&collection,
		"collection", "", "Collection of your configuration file on MongoDB")
	evaluateCmd.Flags().StringVar(&container,
		"container", "", "Container of your configuration file on Azure Blob Storage")
	evaluateCmd.Flags().StringVar(&accountName,
		"account-name", "", "Account name of your configuration file on Azure Blob Storage")
	evaluateCmd.Flags().StringVar(&accountKey,
		"account-key", "", "Account key of your configuration file on Azure Blob Storage")
	evaluateCmd.Flags().StringVar(&table,
		"table", "", "Postgres table of your configuration file on Postgres")
	evaluateCmd.Flags().StringArrayVar(&columns,
		"column", nil,
		"Postgres column mapping of your configuration file on Postgres (may be repeated)")
	_ = evaluateCmd.Flags().MarkDeprecated("github-token", "Use auth-token instead")
	_ = evaluateCmd.Flags().MarkDeprecated("config", "Use path instead")
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

	logger := &fflog.FFLogger{LeveledLogger: slog.Default()}

	if err := tryInitializeStandard(context.Background(), r, logger); err != nil {
		return err
	}

	if err := tryInitializeWithFlagset(context.Background(), r, logger, utils.DefaultFlagSetName); err != nil {
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

	// nolint:musttag
	detailed, err := json.MarshalIndent(retrieverConf, "", "  ")
	if err != nil {
		return err
	}

	output.Add(string(detailed), helper.DefaultLevel)
	output.PrintLines(cmd)
	return nil
}

func tryInitializeStandard(ctx context.Context, r retriever.Retriever, logger *fflog.FFLogger) error {
	if r, ok := r.(retriever.InitializableRetriever); ok {
		if err := r.Init(ctx, logger); err != nil {
			return fmt.Errorf("impossible to init retriever: %v", err)
		}
	}

	return nil
}

func tryInitializeWithFlagset(
	ctx context.Context, r retriever.Retriever, logger *fflog.FFLogger, flagset string) error {
	if r, ok := r.(retriever.InitializableRetrieverWithFlagset); ok {
		if err := r.Init(ctx, logger, &flagset); err != nil {
			return fmt.Errorf("impossible to init flagset retriever: %v", err)
		}
	}

	return nil
}

func parseHTTPHeaders() map[string][]string {
	result := make(map[string][]string)
	for _, h := range headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) != 2 {
			parts = strings.SplitN(h, "=", 2)
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

func parsePostgresColumns() map[string]string {
	result := make(map[string]string)
	for _, c := range columns {
		parts := strings.SplitN(c, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		result[key] = val
	}
	return result
}
