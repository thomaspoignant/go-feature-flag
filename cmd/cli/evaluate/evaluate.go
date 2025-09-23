package evaluate

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	retrieverInit "github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf/init"
	"github.com/thomaspoignant/go-feature-flag/internal/utils"
	"github.com/thomaspoignant/go-feature-flag/model"
)

type evaluate struct {
	retrieverConf retrieverconf.RetrieverConf
	fileFormat    string
	flag          string
	evaluationCtx string
}

func (e evaluate) Evaluate() (map[string]model.RawVarResult, error) {
	r, err := retrieverInit.InitRetriever(&e.retrieverConf)
	if err != nil {
		return nil, err
	}
	c := ffclient.Config{
		PollingInterval:       10 * time.Minute,
		DisableNotifierOnInit: true,
		Context:               context.Background(),
		Retriever:             r,
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
			flags, _ := goff.GetFlagsFromCache()
			for key := range flags {
				listFlags = append(listFlags, key)
			}
		}

		for _, flag := range listFlags {
			res, _ := goff.RawVariation(flag, convertedEvaluationCtx, nil)
			result[flag] = res
		}
		return result, nil
	}
	return nil, errors.New("invalid evaluation context (missing targeting key)")
}
