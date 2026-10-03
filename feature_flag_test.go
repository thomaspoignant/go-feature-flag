package ffclient_test

import (
	"bytes"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/model"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"github.com/thomaspoignant/go-feature-flag/testutils/mock"
	"github.com/thomaspoignant/go-feature-flag/testutils/mock/mockretriever"
	"gopkg.in/yaml.v3"
)

func TestStartWithoutRetriever(t *testing.T) {
	_, err := ffclient.New(ffclient.Config{
		PollingInterval: 60 * time.Second,
		LeveledLogger:   slog.Default(),
	})
	assert.Error(t, err)
}

func TestMultipleRetrievers(t *testing.T) {
	client, err := ffclient.New(ffclient.Config{
		PollingInterval: 60 * time.Second,
		LeveledLogger:   slog.Default(),
		Retrievers: []retriever.Retriever{
			&fileretriever.Retriever{Path: "testdata/flag-config-2nd-file.yaml"},
			&fileretriever.Retriever{Path: "testdata/flag-config.yaml"},
		},
	})
	assert.NoError(t, err)
	defer client.Close()
	user := ffcontext.NewEvaluationContext("random-key")
	flagRes1, err := client.BoolVariationDetails("foo-flag", user, false)
	assert.NoError(t, err)
	assert.True(t, flagRes1.Value)
	assert.NotEqual(t, flag.ErrorCodeFlagNotFound, flagRes1.ErrorCode)

	flagRes2, err := client.BoolVariationDetails("test-flag", user, false)
	assert.NoError(t, err)
	assert.True(t, flagRes2.Value)
	assert.NotEqual(t, flag.ErrorCodeFlagNotFound, flagRes2.ErrorCode)
}
func TestMultipleRetrieversWithOverrideFlag(t *testing.T) {
	client, err := ffclient.New(ffclient.Config{
		PollingInterval: 60 * time.Second,
		LeveledLogger:   slog.Default(),
		Retriever:       &fileretriever.Retriever{Path: "testdata/multiple_files/config-1.yaml"},
		Retrievers: []retriever.Retriever{
			&fileretriever.Retriever{Path: "testdata/multiple_files/config-2.yaml"},
		},
	})
	assert.NoError(t, err)
	defer client.Close()
	user := ffcontext.NewEvaluationContext("random-key")
	flagRes1, err := client.BoolVariationDetails("my-flag", user, false)
	assert.NoError(t, err)
	assert.False(t, flagRes1.Value)
	assert.NotEqual(t, flag.ErrorCodeFlagNotFound, flagRes1.ErrorCode)

	flagRes2, err := client.BoolVariationDetails("my-3rd-flag", user, false)
	assert.NoError(t, err)
	assert.True(t, flagRes2.Value)
	assert.NotEqual(t, flag.ErrorCodeFlagNotFound, flagRes2.ErrorCode)
}

func TestStartWithMinInterval(t *testing.T) {
	_, err := ffclient.New(ffclient.Config{
		PollingInterval: 2,
		Retriever:       &fileretriever.Retriever{Path: "testdata/flag-config.yaml"},
		LeveledLogger:   slog.Default(),
	})
	assert.NoError(t, err)
}

func TestValidUseCase(t *testing.T) {
	cliExport := mock.Exporter{Bulk: false}
	// Valid use case
	err := ffclient.Init(ffclient.Config{
		PollingInterval: 5 * time.Second,
		Retriever:       &fileretriever.Retriever{Path: "testdata/flag-config.yaml"},
		LeveledLogger:   slog.Default(),
		DataExporters: []ffclient.DataExporter{
			{
				FlushInterval:    10 * time.Second,
				MaxEventInMemory: 1000,
				Exporter: &mock.Exporter{
					Bulk: true,
				},
			},
			{
				Exporter:          &cliExport,
				ExporterEventType: ffclient.TrackingEventExporter,
			},
		},
	})
	defer ffclient.Close()

	assert.NoError(t, err)
	user := ffcontext.NewEvaluationContext("random-key")
	hasTestFlag, _ := ffclient.BoolVariation("test-flag", user, false)
	assert.True(t, hasTestFlag, "User should have test flag")
	hasUnknownFlag, _ := ffclient.BoolVariation("unknown-flag", user, false)
	assert.False(t, hasUnknownFlag, "User should use default value if flag does not exists")
	assert.NotEqual(t, time.Time{}, ffclient.GetCacheRefreshDate())

	allFlags := ffclient.AllFlagsState(user)
	assert.Equal(t, 2, len(allFlags.GetFlags()))

	ffclient.SetOffline(true)
	assert.True(t, ffclient.IsOffline())
	assert.False(t, ffclient.ForceRefresh())
	ffclient.SetOffline(false)
	assert.False(t, ffclient.IsOffline())
	assert.True(t, ffclient.ForceRefresh())
	ffclient.Track("toto", user, map[string]any{"key": "value"})
	assert.Equal(t, 1, len(cliExport.ExportedEvents))
}

func TestValidUseCaseToml(t *testing.T) {
	// Valid use case
	gffClient, err := ffclient.New(ffclient.Config{
		PollingInterval: 5 * time.Second,
		Retriever:       &fileretriever.Retriever{Path: "testdata/flag-config.toml"},
		LeveledLogger:   slog.Default(),
		FileFormat:      "toml",
	})
	defer gffClient.Close()

	assert.NoError(t, err)
	user := ffcontext.NewEvaluationContext("random-key")
	hasTestFlag, _ := gffClient.BoolVariation("test-flag", user, false)
	assert.True(t, hasTestFlag, "User should have test flag")
	hasUnknownFlag, _ := gffClient.BoolVariation("unknown-flag", user, false)
	assert.False(t, hasUnknownFlag, "User should use default value if flag does not exists")
}

func TestValidUseCaseJson(t *testing.T) {
	// Valid use case
	gffClient, err := ffclient.New(ffclient.Config{
		PollingInterval: 5 * time.Second,
		Retriever:       &fileretriever.Retriever{Path: "testdata/flag-config.json"},
		LeveledLogger:   slog.Default(),
		FileFormat:      "json",
	})
	defer gffClient.Close()

	assert.NoError(t, err)
	user := ffcontext.NewEvaluationContext("random-key")
	hasTestFlag, _ := gffClient.BoolVariation("test-flag", user, false)
	assert.True(t, hasTestFlag, "User should have test flag")
	hasUnknownFlag, _ := gffClient.BoolVariation("unknown-flag", user, false)
	assert.False(t, hasUnknownFlag, "User should use default value if flag does not exists")
	assert.NotEqual(t, time.Time{}, gffClient.GetCacheRefreshDate())
}

func TestValidUseCaseMultilineQueryJson(t *testing.T) {
	// Valid use case
	gffClient, err := ffclient.New(ffclient.Config{
		PollingInterval: 5 * time.Second,
		Retriever: &fileretriever.Retriever{
			Path: "testdata/flag-config-multiline-query.json",
		},
		LeveledLogger: slog.Default(),
		FileFormat:    "json",
	})
	defer gffClient.Close()

	assert.NoError(t, err)
	user := ffcontext.NewEvaluationContext("random-key")
	hasTestFlag, _ := gffClient.BoolVariation("test-flag", user, false)
	assert.True(t, hasTestFlag, "User should have test flag")
	hasUnknownFlag, _ := gffClient.BoolVariation("unknown-flag", user, false)
	assert.False(t, hasUnknownFlag, "User should use default value if flag does not exists")
	assert.NotEqual(t, time.Time{}, gffClient.GetCacheRefreshDate())
}

func Test2GoFeatureFlagInstance(t *testing.T) {
	gffClient1, err := ffclient.New(ffclient.Config{
		PollingInterval: 5 * time.Second,
		Retriever:       &fileretriever.Retriever{Path: "testdata/flag-config.yaml"},
		LeveledLogger:   slog.Default(),
	})
	defer gffClient1.Close()

	gffClient2, err2 := ffclient.New(ffclient.Config{
		PollingInterval: 10 * time.Second,
		Retriever:       &fileretriever.Retriever{Path: "testdata/test-instance2.yaml"},
		LeveledLogger:   slog.Default(),
	})
	defer gffClient2.Close()

	// Init should be OK for both clients.
	assert.NoError(t, err)
	assert.NoError(t, err2)

	user := ffcontext.NewEvaluationContext("random-key")

	// Client1 is supposed to have the flag at true
	hasTestFlagClient1, _ := gffClient1.BoolVariation("test-flag", user, false)
	assert.True(t, hasTestFlagClient1, "User should have test flag")

	// Client2 is supposed to have the flag at true
	hasTestFlagClient2, _ := gffClient2.BoolVariation("test-flag", user, false)
	assert.False(t, hasTestFlagClient2, "User should have test flag")
}

func TestUpdateFlag(t *testing.T) {
	initialFileContent := `
test-flag:
  variations:
    true_var: true
    false_var: false
  targeting:
    - query: key eq "random-key"
      percentage:
        true_var: 100
        false_var: 0
  defaultRule:
    variation: false_var`

	flagFile, _ := os.CreateTemp("", "")
	_ = os.WriteFile(flagFile.Name(), []byte(initialFileContent), os.ModePerm)

	gffClient1, _ := ffclient.New(ffclient.Config{
		PollingInterval: 1 * time.Second,
		Retriever:       &fileretriever.Retriever{Path: flagFile.Name()},
		LeveledLogger:   slog.Default(),
	})
	defer gffClient1.Close()

	flagValue, _ := gffClient1.BoolVariation(
		"test-flag",
		ffcontext.NewEvaluationContext("random-key"),
		false,
	)
	assert.True(t, flagValue)

	updatedFileContent := `
test-flag:
  variations:
    true_var: true
    false_var: false
  targeting:
    - query: key eq "random-key2"
      percentage:
        true_var: 100
        false_var: 0
  defaultRule:
    variation: false_var`

	_ = os.WriteFile(flagFile.Name(), []byte(updatedFileContent), os.ModePerm)

	flagValue, _ = gffClient1.BoolVariation(
		"test-flag",
		ffcontext.NewEvaluationContext("random-key"),
		false,
	)
	assert.True(t, flagValue)

	time.Sleep(2 * time.Second)

	flagValue, _ = gffClient1.BoolVariation(
		"test-flag",
		ffcontext.NewEvaluationContext("random-key"),
		false,
	)
	assert.False(t, flagValue)
}

func TestImpossibleToLoadfile(t *testing.T) {
	initialFileContent := `
test-flag:
  variations:
    true_var: true
    false_var: false
  targeting:
    - query: key eq "random-key"
      percentage:
        true_var: 100
        false_var: 0
  defaultRule:
    variation: false_var`

	flagFile, _ := os.CreateTemp("", "impossible")
	_ = os.WriteFile(flagFile.Name(), []byte(initialFileContent), os.ModePerm)

	gffClient1, _ := ffclient.New(ffclient.Config{
		PollingInterval: 1 * time.Second,
		Retriever:       &fileretriever.Retriever{Path: flagFile.Name()},
		LeveledLogger:   slog.Default(),
	})
	defer gffClient1.Close()

	flagValue, _ := gffClient1.BoolVariation(
		"test-flag",
		ffcontext.NewEvaluationContext("random-key"),
		false,
	)
	assert.True(t, flagValue)

	flagValue, _ = gffClient1.BoolVariation(
		"test-flag",
		ffcontext.NewEvaluationContext("random-key"),
		false,
	)
	assert.True(t, flagValue)

	// remove file we should still take the last version in consideration
	os.Remove(flagFile.Name())
	time.Sleep(2 * time.Second)

	flagValue, _ = gffClient1.BoolVariation(
		"test-flag",
		ffcontext.NewEvaluationContext("random-key"),
		false,
	)
	assert.True(t, flagValue)
}

func TestFlagFileUnreachable(t *testing.T) {
	initialFileContent := `
test-flag:
  variations:
    true_var: "true"
    false_var: "false"
  targeting:
    - query: key eq "random-key"
      percentage:
        true_var: 100
        false_var: 0
  defaultRule:
    variation: false_var`

	tempDir, _ := os.MkdirTemp("", "")
	defer os.Remove(tempDir)

	flagFilePath := tempDir + "_FlagFileUnreachable.yaml"
	gff, err := ffclient.New(ffclient.Config{
		PollingInterval:         1 * time.Second,
		Retriever:               &fileretriever.Retriever{Path: flagFilePath},
		LeveledLogger:           slog.Default(),
		StartWithRetrieverError: true,
	})
	defer gff.Close()

	assert.NoError(t, err, "should not return any error even if we can't retrieve the file")

	flagValue, _ := gff.StringVariation(
		"test-flag",
		ffcontext.NewEvaluationContext("random-key"),
		"SDKdefault",
	)
	assert.Equal(t, "SDKdefault", flagValue, "should use the SDK default value")

	err = os.WriteFile(flagFilePath, []byte(initialFileContent), os.ModePerm)
	assert.NoError(t, err)
	time.Sleep(2 * time.Second)

	flagValue, _ = gff.StringVariation(
		"test-flag",
		ffcontext.NewEvaluationContext("random-key"),
		"SDKdefault",
	)
	assert.Equal(t, "true", flagValue, "should use the true value")
}

func TestInvalidConf(t *testing.T) {
	gff, err := ffclient.New(ffclient.Config{
		PollingInterval: 1 * time.Second,
		Retriever:       &fileretriever.Retriever{Path: "testdata/invalid-flag-config.json"},
		LeveledLogger:   slog.Default(),
	})
	defer gff.Close()
	assert.Error(t, err)
	assert.Equal(
		t,

		"impossible to initialize the retrievers, please check your configuration: impossible to retrieve the flags, please check your configuration: yaml: line 43: did not find expected ',' or '}'",
		err.Error(),
	)
}

func TestInvalidConfAndRetrieverError(t *testing.T) {
	gff, err := ffclient.New(ffclient.Config{
		PollingInterval: 1 * time.Second,
		Retriever: &fileretriever.Retriever{
			Path: "testdata/invalid-flag-config.json",
		},
		LeveledLogger:           slog.Default(),
		StartWithRetrieverError: true,
	})
	defer gff.Close()
	assert.NoError(t, err)
}

func TestValidUseCaseBigFlagFile(t *testing.T) {
	// Valid use case
	gff, err := ffclient.New(ffclient.Config{
		PollingInterval: 5 * time.Second,
		Retriever:       &fileretriever.Retriever{Path: "testdata/flag-config-big.yaml"},
	})
	defer gff.Close()

	assert.NoError(t, err)
	user := ffcontext.NewEvaluationContext("random-key")
	hasTestFlag, _ := gff.BoolVariation("test-flag99", user, false)
	assert.True(t, hasTestFlag, "User should have test flag")
	hasUnknownFlag, _ := gff.BoolVariation("unknown-flag", user, false)
	assert.False(t, hasUnknownFlag, "User should use default value if flag does not exists")
}

func TestInitializableRetrieverWithRetrieverReady(t *testing.T) {
	f, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	// close the handle before removing: Windows cannot delete a file that is still open
	assert.NoError(t, f.Close())
	// we delete the fileTemp to be sure that the retriever will have to create the file
	err = os.Remove(f.Name())
	assert.NoError(t, err)

	r := mockretriever.NewFileInitializableRetriever(f.Name(), retriever.RetrieverReady)
	gff, err := ffclient.New(ffclient.Config{
		PollingInterval: 5 * time.Second,
		Retriever:       r,
	})
	assert.NoError(t, err)
	user := ffcontext.NewEvaluationContext("random-key")
	hasTestFlag, _ := gff.BoolVariation("flag-xxxx-123", user, false)
	assert.True(t, hasTestFlag, "User should have test flag")

	gff.Close()
	_, err = os.Stat(f.Name())
	assert.True(t, errors.Is(err, os.ErrNotExist))
}
func TestInitializableRetrieverWithRetrieverNotReady(t *testing.T) {
	f, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	// close the handle before removing: Windows cannot delete a file that is still open
	assert.NoError(t, f.Close())
	// we delete the fileTemp to be sure that the retriever will have to create the file
	err = os.Remove(f.Name())
	assert.NoError(t, err)

	r := mockretriever.NewFileInitializableRetriever(f.Name(), retriever.RetrieverNotReady)
	gff, err := ffclient.New(ffclient.Config{
		PollingInterval: 5 * time.Second,
		Retriever:       r,
	})
	assert.NoError(t, err)
	defer gff.Close()
	user := ffcontext.NewEvaluationContext("random-key")
	hasTestFlag, _ := gff.BoolVariation("flag-xxxx-123", user, false)
	assert.False(t, hasTestFlag, "Should resolve to default value if retriever is not ready")
}

func TestGoFeatureFlag_GetCacheRefreshDate(t *testing.T) {
	type fields struct {
		pollingInterval time.Duration
		waitingDuration time.Duration
	}

	tests := []struct {
		name       string
		fields     fields
		hasRefresh bool
		offline    bool
	}{
		{
			name:       "Should be refreshed",
			fields:     fields{waitingDuration: 5 * time.Second, pollingInterval: 1 * time.Second},
			hasRefresh: true,
		},
		{
			// The polling interval is far longer than the waiting duration, so a single tick
			// can never happen during the wait, even if time.Sleep overshoots (which it does
			// more on Windows, where the timer granularity is ~15ms).
			name:       "Should not be refreshed",
			fields:     fields{waitingDuration: 500 * time.Millisecond, pollingInterval: 30 * time.Second},
			hasRefresh: false,
		},
		{
			name:       "Should not crash in offline mode",
			fields:     fields{waitingDuration: 500 * time.Millisecond, pollingInterval: 30 * time.Second},
			hasRefresh: false,
			offline:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gff, err := ffclient.New(ffclient.Config{
				PollingInterval: tt.fields.pollingInterval,
				Retriever:       &fileretriever.Retriever{Path: "testdata/flag-config.yaml"},
				Offline:         tt.offline,
			})
			require.NoError(t, err)
			defer gff.Close()

			date1 := gff.GetCacheRefreshDate()

			if tt.hasRefresh {
				// Waiting exactly one polling interval leaves no margin for the scheduler,
				// so we poll for the refresh instead of sleeping for a fixed duration.
				require.Eventually(t, func() bool { return date1.Before(gff.GetCacheRefreshDate()) },
					tt.fields.waitingDuration, 50*time.Millisecond,
					"the cache was never refreshed by the background updater")
			} else {
				require.Never(t, func() bool { return date1.Before(gff.GetCacheRefreshDate()) },
					tt.fields.waitingDuration, 50*time.Millisecond,
					"the cache was refreshed even though the polling interval had not elapsed")
			}

			date2 := gff.GetCacheRefreshDate()
			if !tt.offline {
				assert.NotEqual(t, time.Time{}, date1)
				assert.NotEqual(t, time.Time{}, date2)
			}
			assert.Equal(t, tt.hasRefresh, date1.Before(date2))
		})
	}
}

func TestGoFeatureFlag_SetOffline(t *testing.T) {
	gffClient, err := ffclient.New(ffclient.Config{
		PollingInterval: 1 * time.Second,
		Retriever:       &fileretriever.Retriever{Path: "testdata/flag-config.yaml"},
		LeveledLogger:   slog.Default(),
		Offline:         false,
	})
	assert.NoError(t, err)
	defer gffClient.Close()

	gffClient.SetOffline(true)
	assert.True(t, gffClient.IsOffline())

	time.Sleep(2 * time.Second)

	gffClient.SetOffline(false)
	assert.False(t, gffClient.IsOffline())
}

func Test_GetPollingInterval(t *testing.T) {
	tests := []struct {
		name            string
		pollingInterval time.Duration
	}{
		{
			name:            "60 seconds",
			pollingInterval: 60 * time.Second,
		},
		{
			name:            "6 hour",
			pollingInterval: 6 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goff, err := ffclient.New(ffclient.Config{
				PollingInterval: tt.pollingInterval,
				Retriever:       &fileretriever.Retriever{Path: "testdata/flag-config.yaml"},
			})
			assert.NoError(t, err)
			assert.Equal(t, tt.pollingInterval.Milliseconds(), goff.GetPollingInterval())
		})
	}
}

func Test_ForceRefreshCache(t *testing.T) {
	tempFile, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	// close the handle before reusing/removing it: Windows cannot write to or
	// delete a file that is still open by another handle
	assert.NoError(t, tempFile.Close())
	defer func() { _ = os.Remove(tempFile.Name()) }()
	content, err := os.ReadFile("testdata/flag-config.yaml")
	assert.NoError(t, err)
	err = os.WriteFile(tempFile.Name(), content, os.ModePerm)
	assert.NoError(t, err)

	gffClient, err := ffclient.New(ffclient.Config{
		PollingInterval: 15 * time.Minute,
		Retriever:       &fileretriever.Retriever{Path: tempFile.Name()},
		LeveledLogger:   slog.Default(),
		Offline:         false,
	})
	assert.NoError(t, err)
	defer gffClient.Close()
	refreshTime := gffClient.GetCacheRefreshDate()

	// modify the file to trigger a refresh
	newContent, err := os.ReadFile("testdata/flag-config-2nd-file.yaml")
	assert.NoError(t, err)
	err = os.WriteFile(tempFile.Name(), newContent, os.ModePerm)
	assert.NoError(t, err)
	// checking that the cache has not been refreshed
	assert.Equal(t, refreshTime, gffClient.GetCacheRefreshDate())

	// Ensure the forced refresh lands on a strictly later timestamp even on
	// platforms with a coarse monotonic clock (Windows timer granularity ~15ms),
	// otherwise the initial load and the forced refresh can share the same instant.
	time.Sleep(50 * time.Millisecond)

	// checking that the cache has been refreshed
	gffClient.ForceRefresh()
	assert.NotEqual(t, refreshTime, gffClient.GetCacheRefreshDate())
	gffClient.SetOffline(true)
	gffClient.ForceRefresh()
	assert.Equal(t, time.Time{}, gffClient.GetCacheRefreshDate())
}

// readPersistedFlags returns the content of the persistent flag configuration file and the
// flags it contains. The library writes this file from a goroutine, so a read can catch a
// partial write: in that case ok is false and the caller (a require.Eventually) just retries.
func readPersistedFlags(path string) (content []byte, flags map[string]any, ok bool) {
	content, err := os.ReadFile(path)
	if err != nil || len(content) == 0 {
		return nil, nil, false
	}
	if err := yaml.Unmarshal(content, &flags); err != nil || len(flags) == 0 {
		return nil, nil, false
	}
	return content, flags, true
}

func Test_PersistFlagConfigurationOnDisk(t *testing.T) {
	// Note: Config.Initialize() raises any polling interval below 1s to 1s (see
	// adjustPollingInterval in config.go), so both clients below really poll once per second.
	// Waiting exactly one polling interval with time.Sleep leaves no margin for the OS
	// scheduler (the Windows timer granularity alone is ~15ms), which is why every "wait
	// until something happened" below is a require.Eventually.
	const (
		pollingInterval   = 1 * time.Second
		eventuallyTimeout = 10 * pollingInterval
		eventuallyTick    = 20 * time.Millisecond
	)

	configFile1, err := os.CreateTemp("", "")
	require.NoError(t, err)
	// close the handles before reusing/removing them: Windows cannot write to or
	// delete a file that is still open by another handle
	require.NoError(t, configFile1.Close())

	persistFile, err := os.CreateTemp("", "")
	require.NoError(t, err)
	require.NoError(t, persistFile.Close())
	// t.Cleanup runs after the deferred Close() calls below, so no handle is still open on
	// these files when Windows deletes them.
	t.Cleanup(func() {
		_ = os.Remove(configFile1.Name())
		_ = os.Remove(persistFile.Name())
	})

	content, err := os.ReadFile("testdata/flag-config.yaml")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(configFile1.Name(), content, os.ModePerm))

	gffClient, err := ffclient.New(ffclient.Config{
		PollingInterval:                 pollingInterval,
		Retriever:                       &fileretriever.Retriever{Path: configFile1.Name()},
		LeveledLogger:                   slog.Default(),
		Offline:                         false,
		PersistentFlagConfigurationFile: persistFile.Name(),
	})
	require.NoError(t, err)

	// 1. Checking that the persistence happened (the file is written by a goroutine).
	var contentP []byte
	require.Eventually(t, func() bool {
		c, flags, ok := readPersistedFlags(persistFile.Name())
		if !ok || len(flags) != 2 { // testdata/flag-config.yaml contains 2 flags
			return false
		}
		contentP = c
		return true
	}, eventuallyTimeout, eventuallyTick,
		"the initial flag configuration was never persisted on disk")

	// 2. Modifying the configuration file
	content2, err := os.ReadFile("testdata/flag-config-2nd-file.yaml")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(configFile1.Name(), content2, os.ModePerm))

	// 3. Checking that the persistence happened again and that the content is different.
	// We wait for a flag of the 2nd file instead of only comparing the raw bytes, so that a
	// partially written file can never satisfy the condition.
	var contentP2 []byte
	require.Eventually(t, func() bool {
		c, flags, ok := readPersistedFlags(persistFile.Name())
		if !ok || len(flags) != 2 {
			return false
		}
		if _, hasFooFlag := flags["foo-flag"]; !hasFooFlag {
			return false
		}
		contentP2 = c
		return true
	}, eventuallyTimeout, eventuallyTick,
		"the updated flag configuration was never persisted on disk")
	assert.NotEqual(t, contentP, contentP2)

	// 4. Stopping GO Feature Flag and restart with a retriever that will fail.
	// Close() stops the polling: without it the background updater of gffClient would keep
	// rewriting persistFile behind the back of gffClient2.
	gffClient.Close()

	configFile2, err := os.CreateTemp("", "")
	require.NoError(t, err)
	// close the handle before removing: Windows cannot delete a file that is still open
	require.NoError(t, configFile2.Close())
	require.NoError(t, os.Remove(configFile2.Name()))
	t.Cleanup(func() { _ = os.Remove(configFile2.Name()) })

	gffClient2, err := ffclient.New(ffclient.Config{
		PollingInterval:                 pollingInterval,
		Retriever:                       &fileretriever.Retriever{Path: configFile2.Name()},
		LeveledLogger:                   slog.Default(),
		Offline:                         false,
		PersistentFlagConfigurationFile: persistFile.Name(),
	})
	require.NoError(t, err)
	defer gffClient2.Close()

	// 5. Checking that the flags have been loaded from the persistent file. The fallback is
	// done synchronously by ffclient.New (retrievePersistentLocalDisk), nothing to wait for.
	details, _ := gffClient2.BoolVariationDetails(
		"foo-flag",
		ffcontext.NewEvaluationContext("random-key"),
		false,
	)
	assert.NotEqual(t, "ERROR", details.Reason)

	// 5b. The retriever keeps failing: checking that the flags coming from the persistent file
	// survive the failing polls. This is the only assertion that needs a fixed duration,
	// because we assert that something does NOT happen.
	require.Never(t, func() bool {
		flags, errCache := gffClient2.GetFlagsFromCache()
		return errCache != nil || len(flags) != 2
	}, 2*pollingInterval, eventuallyTick,
		"the flags loaded from the persistent file were lost on a failing poll")

	// 6. Making the configuration file of the failing retriever available
	content3, err := os.ReadFile("testdata/flag-config-3rd-file.yaml")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(configFile2.Name(), content3, os.ModePerm))

	// 7. Checking that the polling goes on after a retriever error, and that the flags of the
	// 3rd file are now served (testdata/flag-config-3rd-file.yaml contains 7 flags).
	require.Eventually(t, func() bool {
		flags2, errCache := gffClient2.GetFlagsFromCache()
		return errCache == nil && len(flags2) == 7
	}, eventuallyTimeout, eventuallyTick,
		"the retriever did not recover after the configuration file was restored")

	// 8. Checking that the persistence happened and that the file is different from the
	// previous one.
	require.Eventually(t, func() bool {
		c, flags3, ok := readPersistedFlags(persistFile.Name())
		return ok && len(flags3) == 7 && !bytes.Equal(contentP2, c)
	}, eventuallyTimeout, eventuallyTick,
		"the new flag configuration was never persisted on disk")
}

func TestUseCustomBucketingKey(t *testing.T) {
	gffClient, err := ffclient.New(ffclient.Config{
		PollingInterval: 1 * time.Second,
		Retriever: &fileretriever.Retriever{
			Path: "testdata/flag-config-custom-bucketingkey.yaml",
		},
		LeveledLogger: slog.Default(),
		Offline:       false,
	})
	assert.NoError(t, err)

	t.Run("should return the default value if the bucketing key is not found", func(t *testing.T) {
		got, err := gffClient.StringVariationDetails(
			"my-flag",
			ffcontext.NewEvaluationContext("random-key"),
			"default",
		)
		assert.NoError(t, err)
		want := model.VariationResult[string]{
			Value:         "default",
			TrackEvents:   true,
			VariationType: "SdkDefault",
			Failed:        true,
			Reason:        flag.ReasonError,
			ErrorCode:     flag.ErrorCodeTargetingKeyMissing,
			ErrorDetails:  "impossible to find bucketingKey in context: nested key not found: teamId",
		}
		assert.Equal(t, want, got)
	})

	t.Run("should return the variation value if the bucketing key is found", func(t *testing.T) {
		got, err := gffClient.StringVariationDetails(
			"my-flag",
			ffcontext.NewEvaluationContextBuilder("random-key").
				AddCustom("teamId", "team-123").
				Build(),
			"default",
		)
		assert.NoError(t, err)
		want := model.VariationResult[string]{
			Value:         "value_A",
			TrackEvents:   true,
			VariationType: "variation_A",
			Failed:        false,
			Reason:        flag.ReasonStatic,
			Cacheable:     true,
		}
		assert.Equal(t, want, got)
	})
}

func Test_DisableNotifierOnInit(t *testing.T) {
	tests := []struct {
		name                 string
		config               *ffclient.Config
		disableNotification  bool
		expectedNotifyCalled bool
	}{
		{
			name: "DisableNotifierOnInit is true",
			config: &ffclient.Config{
				PollingInterval:       60 * time.Second,
				Retriever:             &fileretriever.Retriever{Path: "testdata/flag-config.yaml"},
				DisableNotifierOnInit: true,
			},
			expectedNotifyCalled: false,
		},
		{
			name: "DisableNotifierOnInit is false",
			config: &ffclient.Config{
				PollingInterval:       60 * time.Second,
				Retriever:             &fileretriever.Retriever{Path: "testdata/flag-config.yaml"},
				DisableNotifierOnInit: false,
			},
			expectedNotifyCalled: true,
		},
		{
			name: "DisableNotifierOnInit is not set",
			config: &ffclient.Config{
				PollingInterval: 60 * time.Second,
				Retriever:       &fileretriever.Retriever{Path: "testdata/flag-config.yaml"},
			},
			expectedNotifyCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockNotifier := &mock.Notifier{}
			tt.config.Notifiers = []notifier.Notifier{mockNotifier}

			gffClient, err := ffclient.New(*tt.config)
			assert.NoError(t, err)
			defer gffClient.Close()

			time.Sleep(2 * time.Second) // wait for the goroutine to call Notify()
			assert.Equal(t, tt.expectedNotifyCalled, mockNotifier.GetNotifyCalls() > 0)
		})
	}
}

func TestStartWithNegativeIntervalToDisablePolling(t *testing.T) {
	content, err := os.ReadFile("testdata/flag-config.yaml")
	assert.NoError(t, err)

	// copy of the file
	tempFile, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(tempFile.Name()) }()
	err = os.WriteFile(tempFile.Name(), content, os.ModePerm)
	assert.NoError(t, err)

	goff, err := ffclient.New(ffclient.Config{
		PollingInterval: -1 * time.Second,
		Retriever:       &fileretriever.Retriever{Path: tempFile.Name()},
		LeveledLogger:   slog.Default(),
	})
	assert.NoError(t, err)

	cacheRefresh := goff.GetCacheRefreshDate()

	// modify the file to trigger a refresh
	newContent, err := os.ReadFile("testdata/flag-config-2nd-file.yaml")
	assert.NoError(t, err)
	err = os.WriteFile(tempFile.Name(), newContent, os.ModePerm)
	assert.NoError(t, err)

	// wait to be sure we give time to the goroutine to refresh the cache
	time.Sleep(2 * time.Second)

	assert.Equal(t, cacheRefresh, goff.GetCacheRefreshDate())

	// we force a refresh to check if the cache is refreshed
	goff.ForceRefresh()
	assert.NotEqual(t, cacheRefresh, goff.GetCacheRefreshDate())
}

func TestGoFeatureFlag_GetEvaluationContextEnrichment(t *testing.T) {
	tests := []struct {
		name       string
		enrichment map[string]any
	}{
		{
			name:       "nil enrichment",
			enrichment: nil,
		},
		{
			name:       "empty enrichment",
			enrichment: map[string]any{},
		},
		{
			name:       "non-empty enrichment",
			enrichment: map[string]any{"foo": "bar", "num": 42},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gff, err := ffclient.New(ffclient.Config{
				Retriever:                   &fileretriever.Retriever{Path: "testdata/flag-config.yaml"},
				EvaluationContextEnrichment: tt.enrichment,
			})
			assert.NoError(t, err)
			got := gff.GetEvaluationContextEnrichment()
			assert.Equal(t, tt.enrichment, got)
		})
	}
}
