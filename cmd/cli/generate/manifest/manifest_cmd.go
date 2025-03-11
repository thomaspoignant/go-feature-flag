package manifest

import (
	"github.com/spf13/cobra"
)

var (
	evalFlagFormat          string
	evalConfigFile          string
	flagManifestDestination string
)

func NewManifestCmd() *cobra.Command {
	manifestCmd := &cobra.Command{
		Use:   "manifest",
		Short: "📄 Generate an OpenFeature flag manifest based on your flag configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, _ := NewManifest(evalConfigFile, evalFlagFormat, flagManifestDestination)
			output, err := m.Generate()
			if err != nil {
				cmd.SilenceUsage = true
				return err
			}
			output.PrintLines(cmd)
			return nil
		},
	}
	manifestCmd.Flags().StringVarP(&evalFlagFormat,
		"format", "f", "yaml", "Format of your input file (YAML, JSON or TOML)")
	manifestCmd.Flags().StringVarP(&evalConfigFile,
		"config", "c", "", "Location of your GO Feature Flag local configuration file")
	manifestCmd.Flags().StringVar(&flagManifestDestination,
		"flag_manifest_destination", "", "Destination of your flag manifest file. "+
			"If not provided, the manifest will be printed to the console.")
	_ = manifestCmd.MarkFlagRequired("flag_manifest_destination")
	return manifestCmd
}
