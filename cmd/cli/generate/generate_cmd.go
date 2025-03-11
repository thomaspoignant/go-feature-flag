package generate

import (
	"github.com/spf13/cobra"
	"github.com/thomaspoignant/go-feature-flag/cmd/cli/generate/manifest"
)

func NewGenerateCmd() *cobra.Command {
	g := &cobra.Command{
		Use:   "generate",
		Short: "🏗️ Generate GO Feature Flag related files",
		Long:  `🏗️ Generate GO Feature Flag relates files (examples: flag manifest, ...)`,
	}
	g.AddCommand(manifest.NewManifestCmd())
	return g
}
