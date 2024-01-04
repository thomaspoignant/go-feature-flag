package ffclient_test

import (
	"errors"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/testutils/initializableretriever"
	"log"
	"os"
	"testing"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/retriever"

	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/s3retriever"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/testutils/mock"
)

func TestStartWithoutRetriever(t *testing.T) {
	_, err := ffclient.New(ffclient.Config{
		PollingInterval: 60 * time.Second,
		Logger:          log.New(os.Stdout, "", 0),
	})
	assert.Error(t, err)
}

func TestMultipleRetrievers(t *testing.T) {
	client, err := ffclient.New(ffclient.Config{
		PollingInterval: 60 * time.Second,
		Logger:          log.New(os.Stdout, "", 0),
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
		Logger:          log.New(os.Stdout, "", 0),
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

func TestStartWithNegativeInterval(t *testing.T) {
	_, err := ffclient.New(ffclient.Config{
		PollingInterval: -60 * time.Second,
		Retriever:       &fileretriever.Retriever{Path: "testdata/flag-config.yaml"},
		Logger:          log.New(os.Stdout, "", 0),
	})
	assert.Error(t, err)
}

func TestStartWithMinInterval(t *testing.T) {
	_, err := ffclient.New(ffclient.Config{
		PollingInterval: 2,
		Retriever:       &fileretriever.Retriever{Path: "testdata/flag-config.yaml"},
		Logger:          log.New(os.Stdout, "", 0),
	})
	assert.NoError(t, err)
}

func TestValidUseCase(t *testing.T) {
	// Valid use case
	err := ffclient.Init(ffclient.Config{
		PollingInterval: 5 * time.Second,
		Retriever:       &fileretriever.Retriever{Path: "testdata/flag-config.yaml"},
		Logger:          log.New(os.Stdout, "", 0),
		DataExporter: ffclient.DataExporter{
			FlushInterval:    10 * time.Second,
			MaxEventInMemory: 1000,
			Exporter: &mock.Exporter{
				Bulk: true,
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
}

func TestAllFlagsFromCache(t *testing.T) {
	err := ffclient.Init(ffclient.Config{
		Retriever:       &fileretriever.Retriever{Path: "testdata/flag-config.yaml"},
		PollingInterval: 5 * time.Second,
	})
	defer ffclient.Close()

	assert.NoError(t, err)
	flags, err := ffclient.GetFlagsFromCache()

	assert.NoError(t, err)
	assert.Len(t, flags, 2)
}

func TestValidUseCaseToml(t *testing.T) {
	// Valid use case
	gffClient, err := ffclient.New(ffclient.Config{
		PollingInterval: 5 * time.Second,
		Retriever:       &fileretriever.Retriever{Path: "testdata/flag-config.toml"},
		Logger:          log.New(os.Stdout, "", 0),
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
		Logger:          log.New(os.Stdout, "", 0),
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
		Retriever:       &fileretriever.Retriever{Path: "testdata/flag-config-multiline-query.json"},
		Logger:          log.New(os.Stdout, "", 0),
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

func TestS3RetrieverReturnError(t *testing.T) {
	_, err := ffclient.New(ffclient.Config{
		Retriever: &s3retriever.Retriever{
			Bucket:    "unknown-bucket",
			Item:      "unknown-item",
			AwsConfig: aws.Config{},
		},
		Logger: log.New(os.Stdout, "", 0),
	})

	assert.Error(t, err)
}

func Test2GoFeatureFlagInstance(t *testing.T) {
	gffClient1, err := ffclient.New(ffclient.Config{
		PollingInterval: 5 * time.Second,
		Retriever:       &fileretriever.Retriever{Path: "testdata/flag-config.yaml"},
		Logger:          log.New(os.Stdout, "", 0),
	})
	defer gffClient1.Close()

	gffClient2, err2 := ffclient.New(ffclient.Config{
		PollingInterval: 10 * time.Second,
		Retriever:       &fileretriever.Retriever{Path: "testdata/test-instance2.yaml"},
		Logger:          log.New(os.Stdout, "", 0),
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
		Logger:          log.New(os.Stdout, "", 0),
	})
	defer gffClient1.Close()

	flagValue, _ := gffClient1.BoolVariation("test-flag", ffcontext.NewEvaluationContext("random-key"), false)
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

	flagValue, _ = gffClient1.BoolVariation("test-flag", ffcontext.NewEvaluationContext("random-key"), false)
	assert.True(t, flagValue)

	time.Sleep(2 * time.Second)

	flagValue, _ = gffClient1.BoolVariation("test-flag", ffcontext.NewEvaluationContext("random-key"), false)
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
		Logger:          log.New(os.Stdout, "", 0),
	})
	defer gffClient1.Close()

	flagValue, _ := gffClient1.BoolVariation("test-flag", ffcontext.NewEvaluationContext("random-key"), false)
	assert.True(t, flagValue)

	flagValue, _ = gffClient1.BoolVariation("test-flag", ffcontext.NewEvaluationContext("random-key"), false)
	assert.True(t, flagValue)

	// remove file we should still take the last version in consideration
	os.Remove(flagFile.Name())
	time.Sleep(2 * time.Second)

	flagValue, _ = gffClient1.BoolVariation("test-flag", ffcontext.NewEvaluationContext("random-key"), false)
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
		Logger:                  log.New(os.Stdout, "", 0),
		StartWithRetrieverError: true,
	})
	defer gff.Close()

	assert.NoError(t, err, "should not return any error even if we can't retrieve the file")

	flagValue, _ := gff.StringVariation("test-flag", ffcontext.NewEvaluationContext("random-key"), "SDKdefault")
	assert.Equal(t, "SDKdefault", flagValue, "should use the SDK default value")

	_ = os.WriteFile(flagFilePath, []byte(initialFileContent), os.ModePerm)
	time.Sleep(2 * time.Second)

	flagValue, _ = gff.StringVariation("test-flag", ffcontext.NewEvaluationContext("random-key"), "SDKdefault")
	assert.Equal(t, "true", flagValue, "should use the true value")
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

	r := initializableretriever.NewMockInitializableRetriever(f.Name(), retriever.RetrieverReady)
	gff, err := ffclient.New(ffclient.Config{
		PollingInterval: 5 * time.Second,
		Retriever:       &r,
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

	r := initializableretriever.NewMockInitializableRetriever(f.Name(), retriever.RetrieverNotReady)
	gff, err := ffclient.New(ffclient.Config{
		PollingInterval: 5 * time.Second,
		Retriever:       &r,
	})
	defer gff.Close()
	assert.NoError(t, err)
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
