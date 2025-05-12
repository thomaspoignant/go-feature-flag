package main

import (
	"encoding/json"
	"fmt"

	"github.com/thomaspoignant/go-feature-flag/evaluation"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/internal/utils"
	"github.com/thomaspoignant/go-feature-flag/model"
	"github.com/thomaspoignant/go-feature-flag/wasm/helpers"
)

// main is the entry point for the wasm module.
// we should keep it to be make sure that the module
// is a valid wasm module for tinygo.
func main() {
	// We keep this main empty because it is required by the tinygo when building wasm.
}

// nolint: unused
// evaluate is the entry point for the wasm module.
// what it does is:
// 1. read the input from the memory
// 2. call the localEvaluation function
// 3. write the result to the memory
//
//export evaluate
func evaluate(valuePosition *uint32, length uint32) uint64 {
	input := helpers.WasmReadBufferFromMemory(valuePosition, length)
	c := localEvaluation(string(input))
	return helpers.WasmCopyBufferToMemory([]byte(c))
}

// localEvaluation is the function that will be called from the evaluate function.
// It will unmarshal the input, call the evaluation function and return the result.
func localEvaluation(input string) string {
	var evaluateInput EvaluateInput
	err := json.Unmarshal([]byte(input), &evaluateInput)
	if err != nil {
		return model.VariationResult[interface{}]{
			ErrorCode:    flag.ErrorCodeParseError,
			ErrorDetails: err.Error(),
		}.ToJsonStr()
	}

	evalCtx, err := convertEvaluationCtx(evaluateInput.EvaluationCtx)
	if err != nil {
		return model.VariationResult[interface{}]{
			ErrorCode:    flag.ErrorCodeTargetingKeyMissing,
			ErrorDetails: err.Error(),
		}.ToJsonStr()
	}

	c, _ := evaluation.Evaluate[interface{}](
		&evaluateInput.Flag,
		evaluateInput.FlagKey,
		evalCtx,
		evaluateInput.FlagContext,
		"interface{}",
		evaluateInput.FlagContext.DefaultSdkValue,
	)
	return c.ToJsonStr()
}

// convertEvaluationCtx converts the evaluation context from the input to a ffcontext.Context.
func convertEvaluationCtx(ctx map[string]any) (ffcontext.Context, error) {
	if targetingKey, ok := ctx["targetingKey"].(string); ok {
		evalCtx := utils.ConvertEvaluationCtxFromRequest(targetingKey, ctx)
		return evalCtx, nil
	}
	return ffcontext.NewEvaluationContextBuilder("").Build(),
		fmt.Errorf("targetingKey not found in context")
}
