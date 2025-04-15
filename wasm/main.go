package main

import (
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/internal/evaluation"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/wasm/helpers"
	"strings"
)

func main() {
	// We keep this main empty because it is required by the tinygo when building wasm.
}

//export evaluate
func evaluate(valuePosition *uint32, length uint32) uint64 {
	input := helpers.WasmReadBufferFromMemory(valuePosition, length)
	inputAsString := strings.SplitAfter(string(input), "\n")

	var f flag.InternalFlag
	var flagkey string
	var evaluationCtx ffcontext.EvaluationContext
	var flagCtx flag.Context
	var sdkDefaultValue interface{}

	c, err := evaluation.Evaluate[interface{}](
		&f, flagkey, evaluationCtx, flagCtx, "interface{}", sdkDefaultValue)
	fmt.Println(c, err)
}
