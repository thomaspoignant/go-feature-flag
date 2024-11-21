package evaluate

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func RunEvaluate(
	_ *cobra.Command,
	_ []string,
	flagFormat string,
	configFile string,
	flag string,
	ctx string) error {
	e := evaluate{
		config:        configFile,
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

	fmt.Println(string(detailed))
	return nil
}
