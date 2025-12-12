package evaluate

import (
	"context"
	"encoding/json"
	"time"

	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/model"
	"github.com/thomaspoignant/go-feature-flag/modules/core/utils"
	"github.com/thomaspoignant/go-feature-flag/retriever"
)

type evaluate struct {
	retriever     retriever.Retriever
	fileFormat    string
	flag          string
	evaluationCtx string
}

// Evaluate evaluates the feature flags based on the configuration and context
func (e evaluate) Evaluate() (map[string]model.RawVarResult, error) {
	goff, err := e.initGOFF()
	if err != nil {
		return nil, err
	}

	convertedEvaluationCtx, err := e.parseEvaluationContext()
	if err != nil {
		return nil, err
	}

	listFlags, err := e.getFlagList(goff)
	if err != nil {
		return nil, err
	}

	return e.evaluateFlags(goff, listFlags, convertedEvaluationCtx)
}

// initGOFF initializes the GO Feature Flag client
func (e evaluate) initGOFF() (*ffclient.GoFeatureFlag, error) {
	c := ffclient.Config{
		PollingInterval:       10 * time.Minute,
		DisableNotifierOnInit: true,
		Context:               context.Background(),
		Retriever:             e.retriever,
		FileFormat:            e.fileFormat,
	}
	return ffclient.New(c)
}

// parseEvaluationContext parses the evaluation context from the command line arguments
func (e evaluate) parseEvaluationContext() (ffcontext.Context, error) {
	if e.evaluationCtx == "" {
		return ffcontext.NewEvaluationContextBuilder("").Build(), nil
	}

	var ctxAsMap map[string]interface{}
	err := json.Unmarshal([]byte(e.evaluationCtx), &ctxAsMap)
	if err != nil {
		return nil, err
	}

	targetingKey, ok := ctxAsMap["targetingKey"].(string)
	if !ok {
		targetingKey = ""
	}

	convertedEvaluationCtx := utils.ConvertEvaluationCtxFromRequest(targetingKey, ctxAsMap)
	return convertedEvaluationCtx, nil
}

// getFlagList gets the list of flags to evaluate
func (e evaluate) getFlagList(goff *ffclient.GoFeatureFlag) ([]string, error) {
	if e.flag != "" {
		return []string{e.flag}, nil
	}

	flags, err := goff.GetFlagsFromCache()
	if err != nil {
		return nil, err
	}

	listFlags := make([]string, 0, len(flags))
	for key := range flags {
		listFlags = append(listFlags, key)
	}
	return listFlags, nil
}

// evaluateFlags evaluates the flags
func (e evaluate) evaluateFlags(
	goff *ffclient.GoFeatureFlag,
	listFlags []string,
	convertedEvaluationCtx ffcontext.Context) (map[string]model.RawVarResult, error) {
	result := make(map[string]model.RawVarResult, len(listFlags))
	for _, flag := range listFlags {
		res, err := goff.RawVariation(flag, convertedEvaluationCtx, nil)
		if err != nil {
			return nil, err
		}
		result[flag] = res
	}
	return result, nil
}
