package evaluate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/internal/utils"
	"github.com/thomaspoignant/go-feature-flag/model"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
)

type evaluate struct {
	config        string
	fileFormat    string
	flag          string
	evaluationCtx string
}

func (e evaluate) Evaluate() (map[string]model.RawVarResult, error) {
	c := ffclient.Config{
		PollingInterval:       10 * time.Minute,
		DisableNotifierOnInit: true,
		Context:               context.Background(),
		Retriever:             &fileretriever.Retriever{Path: e.config},
		FileFormat:            e.fileFormat,
	}

	goff, err := ffclient.New(c)
	if err != nil {
		return nil, err
	}

	var ctxAsMap map[string]interface{}
	err = json.Unmarshal([]byte(e.evaluationCtx), &ctxAsMap)
	if targetingKey, ok := ctxAsMap["targetingKey"].(string); ok {
		convertedEvaluationCtx := utils.ConvertEvaluationCtxFromRequest(targetingKey, ctxAsMap)
		if e.flag != "" {
			return e.evaluateSingleFlag(goff, convertedEvaluationCtx, e.flag)
		} else {
			return e.evaluateBulk(goff, convertedEvaluationCtx)
		}
	}
	return nil, errors.New("invalid evaluation context (missing targeting key)")
}

func (e evaluate) evaluateSingleFlag(goff *ffclient.GoFeatureFlag, evalCtx ffcontext.Context, flag string) (map[string]model.RawVarResult, error) {
	res, _ := goff.RawVariation(flag, evalCtx, nil)
	detailed, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(detailed))
	return nil
}

func (e evaluate) evaluateBulk(goff *ffclient.GoFeatureFlag, evalCtx ffcontext.Context) (map[string]model.RawVarResult, error) {

	flags, err := goff.GetFlagsFromCache()
	if err != nil {
		return err
	}

	for flagName, _ := range flags {
		fmt.Println("Flag:", flagName)
		err := e.evaluateSingleFlag(goff, evalCtx, flagName)
		if err != nil {
			return err
		}
		fmt.Println("--------------")
	}
	return nil
}
