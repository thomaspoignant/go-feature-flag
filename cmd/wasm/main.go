package main

import (
	"encoding/json"

	"github.com/thomaspoignant/go-feature-flag/cmd/wasm/helpers"
	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/model"
	"github.com/thomaspoignant/go-feature-flag/modules/core/utils"
	"github.com/thomaspoignant/go-feature-flag/modules/evaluation"
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
		return model.VariationResult[any]{
			ErrorCode:    flag.ErrorCodeParseError,
			ErrorDetails: err.Error(),
		}.ToJsonStr()
	}

	evalCtx := convertEvaluationCtx(evaluateInput.EvaluationCtx)

	// we don't care about the error here because the errorCode and errorDetails
	// contains information about the type of the error directly, no need to check the Go error.
	c, _ := evaluation.Evaluate[any](
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
// Note: Empty targeting keys are now allowed - the core evaluation logic will determine
// if a targeting key is required based on whether the flag needs bucketing.
func convertEvaluationCtx(ctx map[string]any) ffcontext.Context {
	// Allow empty or missing targeting keys - core evaluation logic will handle requirements
	targetingKey := ""
	if key, ok := ctx["targetingKey"].(string); ok {
		targetingKey = key
	}

	// Create evaluation context (empty targeting key is allowed)
	evalCtx := utils.ConvertEvaluationCtxFromRequest(targetingKey, ctx)
	return evalCtx
}
