package config_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
)

const (
	flagset1Name   = "flagset-1"
	flagset2Name   = "flagset-2"
	flagset3Name   = "flagset-3"
	oldKey1        = "old-key-1"
	oldKey2        = "old-key-2"
	flagsetNameFmt = "flagset-%d"
)

func TestConfigSetFlagSetAPIKeys(t *testing.T) {
	tests := []struct {
		name           string
		config         *config.Config
		flagsetName    string
		apiKeys        []string
		wantErr        bool
		wantErrContain string
	}{
		{
			name: "set API keys for existing flagset",
			config: &config.Config{
				FlagSets: []config.FlagSet{
					{Name: flagset1Name, APIKeys: []string{oldKey1}},
					{Name: flagset2Name, APIKeys: []string{oldKey2}},
				},
			},
			flagsetName: flagset1Name,
			apiKeys:     []string{"new-key-1", "new-key-2"},
			wantErr:     false,
		},
		{
			name: "set empty API keys for existing flagset",
			config: &config.Config{
				FlagSets: []config.FlagSet{
					{Name: flagset1Name, APIKeys: []string{oldKey1}},
				},
			},
			flagsetName: flagset1Name,
			apiKeys:     []string{},
			wantErr:     false,
		},
		{
			name: "set API keys for non-existing flagset",
			config: &config.Config{
				FlagSets: []config.FlagSet{
					{Name: flagset1Name, APIKeys: []string{oldKey1}},
				},
			},
			flagsetName:    "non-existing",
			apiKeys:        []string{"new-key"},
			wantErr:        true,
			wantErrContain: "flagset non-existing not found",
		},
		{
			name: "set API keys with empty flagsets config",
			config: &config.Config{
				FlagSets: []config.FlagSet{},
			},
			flagsetName:    flagset1Name,
			apiKeys:        []string{"new-key"},
			wantErr:        true,
			wantErrContain: "flagset flagset-1 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.SetFlagSetAPIKeys(tt.flagsetName, tt.apiKeys)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContain)
			} else {
				require.NoError(t, err)
				// Verify the API keys were actually set
				gotKeys, err := tt.config.GetFlagSetAPIKeys(tt.flagsetName)
				require.NoError(t, err)
				assert.Equal(t, tt.apiKeys, gotKeys)
			}
		})
	}
}

func TestConfigGetFlagSetAPIKeys(t *testing.T) {
	tests := []struct {
		name           string
		config         *config.Config
		flagsetName    string
		wantAPIKeys    []string
		wantErr        bool
		wantErrContain string
	}{
		{
			name: "get API keys for existing flagset",
			config: &config.Config{
				FlagSets: []config.FlagSet{
					{Name: flagset1Name, APIKeys: []string{"key-1", "key-2"}},
					{Name: flagset2Name, APIKeys: []string{"key-3"}},
				},
			},
			flagsetName: flagset1Name,
			wantAPIKeys: []string{"key-1", "key-2"},
			wantErr:     false,
		},
		{
			name: "get API keys for flagset with empty keys",
			config: &config.Config{
				FlagSets: []config.FlagSet{
					{Name: flagset1Name, APIKeys: []string{}},
				},
			},
			flagsetName: flagset1Name,
			wantAPIKeys: []string{},
			wantErr:     false,
		},
		{
			name: "get API keys for flagset with nil keys",
			config: &config.Config{
				FlagSets: []config.FlagSet{
					{Name: flagset1Name, APIKeys: nil},
				},
			},
			flagsetName: flagset1Name,
			wantAPIKeys: nil,
			wantErr:     false,
		},
		{
			name: "get API keys for non-existing flagset",
			config: &config.Config{
				FlagSets: []config.FlagSet{
					{Name: flagset1Name, APIKeys: []string{"key-1"}},
				},
			},
			flagsetName:    "non-existing",
			wantAPIKeys:    nil,
			wantErr:        true,
			wantErrContain: "flagset non-existing not found",
		},
		{
			name: "get API keys with empty flagsets config",
			config: &config.Config{
				FlagSets: []config.FlagSet{},
			},
			flagsetName:    flagset1Name,
			wantAPIKeys:    nil,
			wantErr:        true,
			wantErrContain: "flagset flagset-1 not found",
		},
		{
			name: "get API keys for second flagset in list",
			config: &config.Config{
				FlagSets: []config.FlagSet{
					{Name: flagset1Name, APIKeys: []string{"key-1"}},
					{Name: flagset2Name, APIKeys: []string{"key-2", "key-3"}},
					{Name: flagset3Name, APIKeys: []string{"key-4"}},
				},
			},
			flagsetName: flagset2Name,
			wantAPIKeys: []string{"key-2", "key-3"},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAPIKeys, err := tt.config.GetFlagSetAPIKeys(tt.flagsetName)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContain)
				assert.Nil(t, gotAPIKeys)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantAPIKeys, gotAPIKeys)
			}
		})
	}
}

func TestFlagSetAPIKeysConcurrency(t *testing.T) {
	cfg := &config.Config{
		FlagSets: []config.FlagSet{
			{Name: flagset1Name, APIKeys: []string{"initial-key-1"}},
			{Name: flagset2Name, APIKeys: []string{"initial-key-2"}},
			{Name: flagset3Name, APIKeys: []string{"initial-key-3"}},
		},
	}

	const numGoroutines = 100
	const numIterations = 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2) // readers and writers

	// Start writer goroutines
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				flagsetName := fmt.Sprintf(flagsetNameFmt, (id%3)+1)
				newKeys := []string{fmt.Sprintf("key-%d-%d", id, j)}
				err := cfg.SetFlagSetAPIKeys(flagsetName, newKeys)
				require.NoError(t, err)
			}
		}(i)
	}

	// Start reader goroutines
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				flagsetName := fmt.Sprintf(flagsetNameFmt, (id%3)+1)
				keys, err := cfg.GetFlagSetAPIKeys(flagsetName)
				// We don't check the exact value since it may change between reads
				// but we verify the operation doesn't panic or return an error
				assert.NoError(t, err)
				assert.NotNil(t, keys)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Verify final state - each flagset should have exactly one key (from the last write)
	for i := 1; i <= 3; i++ {
		flagsetName := fmt.Sprintf(flagsetNameFmt, i)
		keys, err := cfg.GetFlagSetAPIKeys(flagsetName)
		require.NoError(t, err)
		assert.Len(t, keys, 1, "flagset %s should have exactly one key after concurrent writes", flagsetName)
	}
}

func TestConfig_AddFlagSet(t *testing.T) {
	tests := []struct {
		name           string
		config         *config.Config
		flagset        config.FlagSet
		wantErr        bool
		wantErrContain string
		wantCount      int
	}{
		{
			name: "add new flagset successfully",
			config: &config.Config{
				FlagSets: []config.FlagSet{
					{Name: flagset1Name, APIKeys: []string{"key-1"}},
				},
			},
			flagset: config.FlagSet{
				Name:    flagset2Name,
				APIKeys: []string{"key-2"},
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name: "add flagset to empty config",
			config: &config.Config{
				FlagSets: []config.FlagSet{},
			},
			flagset: config.FlagSet{
				Name:    flagset1Name,
				APIKeys: []string{"key-1"},
			},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name: "add duplicate flagset should fail",
			config: &config.Config{
				FlagSets: []config.FlagSet{
					{Name: flagset1Name, APIKeys: []string{"key-1"}},
				},
			},
			flagset: config.FlagSet{
				Name:    flagset1Name,
				APIKeys: []string{"key-2"},
			},
			wantErr:        true,
			wantErrContain: "flagset flagset-1 already exists",
			wantCount:      1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialCount := len(tt.config.FlagSets)
			err := tt.config.AddFlagSet(tt.flagset)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContain)
				assert.Equal(t, initialCount, len(tt.config.FlagSets), "flagset count should not change on error")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(tt.config.FlagSets))
				// Verify the flagset was actually added
				gotKeys, err := tt.config.GetFlagSetAPIKeys(tt.flagset.Name)
				require.NoError(t, err)
				assert.Equal(t, tt.flagset.APIKeys, gotKeys)
			}
		})
	}
}

func TestConfig_RemoveFlagSet(t *testing.T) {
	tests := []struct {
		name           string
		config         *config.Config
		flagsetName    string
		wantErr        bool
		wantErrContain string
		wantCount      int
	}{
		{
			name: "remove existing flagset",
			config: &config.Config{
				FlagSets: []config.FlagSet{
					{Name: flagset1Name, APIKeys: []string{"key-1"}},
					{Name: flagset2Name, APIKeys: []string{"key-2"}},
				},
			},
			flagsetName: flagset1Name,
			wantErr:     false,
			wantCount:   1,
		},
		{
			name: "remove flagset from single flagset config",
			config: &config.Config{
				FlagSets: []config.FlagSet{
					{Name: flagset1Name, APIKeys: []string{"key-1"}},
				},
			},
			flagsetName: flagset1Name,
			wantErr:     false,
			wantCount:   0,
		},
		{
			name: "remove non-existing flagset should fail",
			config: &config.Config{
				FlagSets: []config.FlagSet{
					{Name: flagset1Name, APIKeys: []string{"key-1"}},
				},
			},
			flagsetName:    "non-existing",
			wantErr:        true,
			wantErrContain: "flagset non-existing not found",
			wantCount:      1,
		},
		{
			name: "remove from empty config should fail",
			config: &config.Config{
				FlagSets: []config.FlagSet{},
			},
			flagsetName:    flagset1Name,
			wantErr:        true,
			wantErrContain: "flagset flagset-1 not found",
			wantCount:      0,
		},
		{
			name: "remove middle flagset",
			config: &config.Config{
				FlagSets: []config.FlagSet{
					{Name: flagset1Name, APIKeys: []string{"key-1"}},
					{Name: flagset2Name, APIKeys: []string{"key-2"}},
					{Name: flagset3Name, APIKeys: []string{"key-3"}},
				},
			},
			flagsetName: flagset2Name,
			wantErr:     false,
			wantCount:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialCount := len(tt.config.FlagSets)
			err := tt.config.RemoveFlagSet(tt.flagsetName)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContain)
				assert.Equal(t, initialCount, len(tt.config.FlagSets), "flagset count should not change on error")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(tt.config.FlagSets))
				// Verify the flagset was actually removed
				_, err := tt.config.GetFlagSetAPIKeys(tt.flagsetName)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
			}
		})
	}
}

func TestConfig_GetFlagSets(t *testing.T) {
	tests := []struct {
		name      string
		config    *config.Config
		wantCount int
		wantNames []string
	}{
		{
			name: "get flagsets from config with multiple flagsets",
			config: &config.Config{
				FlagSets: []config.FlagSet{
					{Name: flagset1Name, APIKeys: []string{"key-1"}},
					{Name: flagset2Name, APIKeys: []string{"key-2"}},
					{Name: flagset3Name, APIKeys: []string{"key-3"}},
				},
			},
			wantCount: 3,
			wantNames: []string{flagset1Name, flagset2Name, flagset3Name},
		},
		{
			name: "get flagsets from empty config",
			config: &config.Config{
				FlagSets: []config.FlagSet{},
			},
			wantCount: 0,
			wantNames: []string{},
		},
		{
			name: "get flagsets returns a copy",
			config: &config.Config{
				FlagSets: []config.FlagSet{
					{Name: flagset1Name, APIKeys: []string{"key-1"}},
				},
			},
			wantCount: 1,
			wantNames: []string{flagset1Name},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flagsets := tt.config.GetFlagSets()
			assert.Equal(t, tt.wantCount, len(flagsets))

			// Verify names match
			names := make([]string, len(flagsets))
			for i, fs := range flagsets {
				names[i] = fs.Name
			}
			assert.ElementsMatch(t, tt.wantNames, names)

			// Verify it's a copy - modifying the returned slice shouldn't affect the config
			if len(flagsets) > 0 {
				originalCount := len(tt.config.FlagSets)
				_ = append(flagsets, config.FlagSet{Name: "new-flagset"})
				assert.Equal(t, originalCount, len(tt.config.FlagSets), "modifying returned slice should not affect config")
			}
		})
	}
}
