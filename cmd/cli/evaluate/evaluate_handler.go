package evaluate

import "github.com/spf13/cobra"

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
	if err := e.Evaluate(); err != nil {
		return err
	}
	return nil
}
