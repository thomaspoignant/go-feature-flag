package ffclient_test

import (
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/model"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"github.com/thomaspoignant/go-feature-flag/testutils/mock"
	"github.com/thomaspoignant/go-feature-flag/testutils/mock/mockretriever"
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
			fields:     fields{waitingDuration: 2 * time.Second, pollingInterval: 1 * time.Second},
			hasRefresh: true,
		},
		{
			name:       "Should not be refreshed",
			fields:     fields{waitingDuration: 2 * time.Second, pollingInterval: 3 * time.Second},
			hasRefresh: false,
		},
		{
			name:       "Should not crash in offline mode",
			fields:     fields{waitingDuration: 2 * time.Second, pollingInterval: 3 * time.Second},
			hasRefresh: false,
			offline:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gff, _ := ffclient.New(ffclient.Config{
				PollingInterval: tt.fields.pollingInterval,
				Retriever:       &fileretriever.Retriever{Path: "testdata/flag-config.yaml"},
				Offline:         tt.offline,
			})

			date1 := gff.GetCacheRefreshDate()
			time.Sleep(tt.fields.waitingDuration)
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

	// checking that the cache has been refreshed
	gffClient.ForceRefresh()
	assert.NotEqual(t, refreshTime, gffClient.GetCacheRefreshDate())
	gffClient.SetOffline(true)
	gffClient.ForceRefresh()
	assert.Equal(t, time.Time{}, gffClient.GetCacheRefreshDate())
}

func Test_PersistFlagConfigurationOnDisk(t *testing.T) {
	configFile1, err := os.CreateTemp("", "")
	assert.NoError(t, err)

	persistFile, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	defer func() {
		_ = os.Remove(configFile1.Name())
		_ = os.Remove(persistFile.Name())
	}()
	content, err := os.ReadFile("testdata/flag-config.yaml")
	assert.NoError(t, err)
	err = os.WriteFile(configFile1.Name(), content, os.ModePerm)
	assert.NoError(t, err)

	gffClient, err := ffclient.New(ffclient.Config{
		PollingInterval:                 1 * time.Second,
		Retriever:                       &fileretriever.Retriever{Path: configFile1.Name()},
		LeveledLogger:                   slog.Default(),
		Offline:                         false,
		PersistentFlagConfigurationFile: persistFile.Name(),
	})
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond) // Waiting for the go routine to write the persistent file
	// 1. Checking that the persistence happened
	contentP, err := os.ReadFile(persistFile.Name())
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(contentP))

	// 2. Modifying the configuration file
	content2, err := os.ReadFile("testdata/flag-config-2nd-file.yaml")
	assert.NoError(t, err)
	err = os.WriteFile(configFile1.Name(), content2, os.ModePerm)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second) // Waiting for the go routine to write the persistent file
	// 3. Checking that the persistence happened and that the content is different
	contentP2, err := os.ReadFile(persistFile.Name())
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(contentP2))
	assert.NotEqual(t, contentP, contentP2)

	// 4. Stopping GO Feature Flag and restart with a retriever that will fail
	gffClient.Close()
	configFile2, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	err = os.Remove(configFile2.Name())
	assert.NoError(t, err)

	gffClient2, err := ffclient.New(ffclient.Config{
		PollingInterval:                 500 * time.Millisecond,
		Retriever:                       &fileretriever.Retriever{Path: configFile2.Name()},
		LeveledLogger:                   slog.Default(),
		Offline:                         false,
		PersistentFlagConfigurationFile: persistFile.Name(),
	})
	assert.NoError(t, err)
	defer gffClient2.Close()

	time.Sleep(100 * time.Millisecond) // Waiting for the go routine to write the persistent file
	// 5. Checking that the flags have been loaded from the persistent file
	details, _ := gffClient2.BoolVariationDetails(
		"foo-flag",
		ffcontext.NewEvaluationContext("random-key"),
		false,
	)
	assert.NotEqual(t, "ERROR", details.Reason)

	time.Sleep(2 * time.Second) // Waiting to be sure that it continue to check updates
	flags, err := gffClient2.GetFlagsFromCache()
	assert.NoError(t, err)

	// 6. Modifying the failed configuration file
	content3, err := os.ReadFile("testdata/flag-config-3rd-file.yaml")
	assert.NoError(t, err)
	err = os.WriteFile(configFile2.Name(), content3, os.ModePerm)
	assert.NoError(t, err)

	// 7. Checking that the flags have been updated
	time.Sleep(1000 * time.Millisecond) // Waiting to be sure that it continue to check updates
	flags2, err := gffClient2.GetFlagsFromCache()
	assert.NoError(t, err)
	assert.NotEqual(t, len(flags), len(flags2))
	// 8. Checking that the persistence happened and that the file is different from the previous one
	contentP3, err := os.ReadFile(persistFile.Name())
	assert.NoError(t, err)
	assert.NotEqual(t, contentP2, contentP3)
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
