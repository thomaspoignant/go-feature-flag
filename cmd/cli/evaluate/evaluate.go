package evaluate

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	ffclient "github.com/thomaspoignant/go-feature-flag"
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
	c := ffclient.Config{
		PollingInterval:       10 * time.Minute,
		DisableNotifierOnInit: true,
		Context:               context.Background(),
		Retriever:             e.retriever,
		FileFormat:            e.fileFormat,
	}

	goff, err := ffclient.New(c)
	if err != nil {
		return nil, err
	}

	if e.evaluationCtx == "" {
		return nil, errors.New("invalid evaluation context (missing targeting key)")
	}

	var ctxAsMap map[string]interface{}
	result := map[string]model.RawVarResult{}
	err = json.Unmarshal([]byte(e.evaluationCtx), &ctxAsMap)
	if err != nil {
		return nil, err
	}

	if targetingKey, ok := ctxAsMap["targetingKey"].(string); ok {
		convertedEvaluationCtx := utils.ConvertEvaluationCtxFromRequest(targetingKey, ctxAsMap)
		listFlags := make([]string, 0)
		if e.flag != "" {
			listFlags = append(listFlags, e.flag)
		} else {
			flags, err := goff.GetFlagsFromCache()
			if err != nil {
				return nil, err
			}
			for key := range flags {
				listFlags = append(listFlags, key)
			}
		}

		for _, flag := range listFlags {
			res, err := goff.RawVariation(flag, convertedEvaluationCtx, nil)
			if err != nil {
				return nil, err
			}
			result[flag] = res
		}
		return result, nil
	}
	return nil, errors.New("invalid evaluation context (missing targeting key)")
}
