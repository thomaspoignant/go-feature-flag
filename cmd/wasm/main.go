package main

import (
	"encoding/json"
	"fmt"

	"github.com/thomaspoignant/go-feature-flag/cmd/wasm/helpers"
	"github.com/thomaspoignant/go-feature-flag/modules/core/evaluation"
	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/model"
	"github.com/thomaspoignant/go-feature-flag/modules/core/utils"
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
func evaluate(valuePosition *uint32, length uint32) (result uint64) {
	// The copy into module memory runs outside safeEvaluation's recover; a
	// panic here must not trap either (a trap poisons the instance), so
	// return 0, which hosts treat as "no output produced".
	defer func() {
		if recover() != nil {
			result = 0
		}
	}()
	return helpers.WasmCopyBufferToMemory([]byte(safeEvaluation(valuePosition, length)))
}

// errorResult serializes a structured evaluation error the host can parse.
func errorResult(code flag.ErrorCode, details string) string {
	return model.VariationResult[any]{
		ErrorCode:    code,
		ErrorDetails: details,
	}.ToJsonStr()
}

// evaluationFn is an indirection over localEvaluation so tests can exercise
// the panic-recovery path of safeEvaluation.
var evaluationFn = localEvaluation

// safeEvaluation wraps the evaluation so that any Go panic becomes a
// structured error result instead of a WASM trap. A trap must be avoided:
// it leaves the module's shadow stack pointer unrestored, which permanently
// poisons the instance (every later call faults inside malloc). Note that
// recover cannot catch a stack-overflow trap itself — that is what the
// nesting-depth guards in localEvaluation are for.
func safeEvaluation(valuePosition *uint32, length uint32) (result string) {
	defer func() {
		if r := recover(); r != nil {
			result = errorResult(flag.ErrorCodeGeneral,
				fmt.Sprintf("recovered from panic during evaluation: %v", r))
		}
	}()
	input := helpers.WasmReadBufferFromMemory(valuePosition, length)
	return evaluationFn(string(input))
}

// localEvaluation is the function that will be called from the evaluate function.
// It will unmarshal the input, call the evaluation function and return the result.
func localEvaluation(input string) string {
	if depth := jsonNestingDepth(input); depth > maxInputNestingDepth {
		return errorResult(flag.ErrorCodeParseError, fmt.Sprintf(
			"input JSON exceeds the maximum nesting depth (%d)", maxInputNestingDepth))
	}

	var evaluateInput EvaluateInput
	err := json.Unmarshal([]byte(input), &evaluateInput)
	if err != nil {
		return errorResult(flag.ErrorCodeParseError, err.Error())
	}

	if depth, limit, over := firstQueryOverLimit(&evaluateInput.Flag); over {
		return errorResult(flag.ErrorCodeParseError, fmt.Sprintf(
			"targeting query exceeds the maximum nesting depth (%d > %d)", depth, limit))
	}

	if items, limit, over := firstQueryOverBreadth(&evaluateInput.Flag); over {
		return errorResult(flag.ErrorCodeParseError, fmt.Sprintf(
			"targeting query list exceeds the maximum item count (%d > %d)", items, limit))
	}

	if conditions, limit, over := firstQueryOverConditionCount(&evaluateInput.Flag); over {
		return errorResult(flag.ErrorCodeParseError, fmt.Sprintf(
			"targeting query exceeds the maximum condition count (%d > %d)", conditions, limit))
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
