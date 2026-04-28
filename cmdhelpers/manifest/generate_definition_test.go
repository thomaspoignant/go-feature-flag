package manifest_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	helper "github.com/thomaspoignant/go-feature-flag/cmdhelpers/configfile"
	"github.com/thomaspoignant/go-feature-flag/modules/core/dto"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
)

func TestXXX(t *testing.T) {
	flagDTOs, err := helper.LoadConfigFile(
		"../../testdata/flag-config.yaml",
		"yaml",
		helper.ConfigFileDefaultLocations,
	)
	require.NoError(t, err)
	flags := make(map[string]flag.InternalFlag)
	for k, v := range flagDTOs {
		flags[k] = dto.ConvertDtoToInternalFlag(v)
	}

	// fmt.Println(manifest.GenerateDefinition(flags, fflog.FFLogger{}))
}
