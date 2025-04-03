package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCmdExist(t *testing.T) {
	t.Run("lint command should exist", func(t *testing.T) {
		cmd := initRootCmd()
		assert.NotNil(t, cmd)
		cmd.SetArgs([]string{"lint", "linter/testdata/valid.yaml", "--format", "yaml"})
		err := cmd.Execute()
		require.NoError(t, err)
	})
	t.Run("evaluate command should exist", func(t *testing.T) {
		cmd := initRootCmd()
		assert.NotNil(t, cmd)
		cmd.SetArgs(
			[]string{
				"evaluate",
				"--config",
				"evaluate/testdata/flag.goff.yaml",
				"--ctx",
				`{"targetingKey": "user-123"}`,
				"--format",
				"yaml",
			},
		)
		err := cmd.Execute()
		require.NoError(t, err)
	})
}
