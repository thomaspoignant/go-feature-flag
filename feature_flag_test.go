package ffclient_test

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/s3retriever"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/testutils/mock"

	"github.com/thomaspoignant/go-feature-flag/ffuser"
)

func TestStartWithoutRetriever(t *testing.T) {
	_, err := ffclient.New(ffclient.Config{
		PollingInterval: 60 * time.Second,
		Logger:          log.New(os.Stdout, "", 0),
	})
	assert.Error(t, err)
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
	user := ffuser.NewUser("random-key")
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
	user := ffuser.NewUser("random-key")
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
	user := ffuser.NewUser("random-key")
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

	user := ffuser.NewUser("random-key")

	// Client1 is supposed to have the flag at true
	hasTestFlagClient1, _ := gffClient1.BoolVariation("test-flag", user, false)
	assert.True(t, hasTestFlagClient1, "User should have test flag")

	// Client2 is supposed to have the flag at true
	hasTestFlagClient2, _ := gffClient2.BoolVariation("test-flag", user, false)
	assert.False(t, hasTestFlagClient2, "User should have test flag")
}

func TestUpdateFlag(t *testing.T) {
	initialFileContent := `test-flag:
  rule: key eq "random-key"
  percentage: 100
  true: true
  false: false
  default: false`

	flagFile, _ := os.CreateTemp("", "")
	_ = os.WriteFile(flagFile.Name(), []byte(initialFileContent), os.ModePerm)

	gffClient1, _ := ffclient.New(ffclient.Config{
		PollingInterval: 1 * time.Second,
		Retriever:       &fileretriever.Retriever{Path: flagFile.Name()},
		Logger:          log.New(os.Stdout, "", 0),
	})
	defer gffClient1.Close()

	flagValue, _ := gffClient1.BoolVariation("test-flag", ffuser.NewUser("random-key"), false)
	assert.True(t, flagValue)

	updatedFileContent := `test-flag:
  rule: key eq "random-key2"
  percentage: 100
  true: true
  false: false
  default: false`

	_ = os.WriteFile(flagFile.Name(), []byte(updatedFileContent), os.ModePerm)

	flagValue, _ = gffClient1.BoolVariation("test-flag", ffuser.NewUser("random-key"), false)
	assert.True(t, flagValue)

	time.Sleep(2 * time.Second)

	flagValue, _ = gffClient1.BoolVariation("test-flag", ffuser.NewUser("random-key"), false)
	assert.False(t, flagValue)
}

func TestImpossibleToLoadfile(t *testing.T) {
	initialFileContent := `test-flag:
  rule: key eq "random-key"
  percentage: 100
  true: true
  false: false
  default: false`

	flagFile, _ := os.CreateTemp("", "impossible")
	_ = os.WriteFile(flagFile.Name(), []byte(initialFileContent), os.ModePerm)

	gffClient1, _ := ffclient.New(ffclient.Config{
		PollingInterval: 1 * time.Second,
		Retriever:       &fileretriever.Retriever{Path: flagFile.Name()},
		Logger:          log.New(os.Stdout, "", 0),
	})
	defer gffClient1.Close()

	flagValue, _ := gffClient1.BoolVariation("test-flag", ffuser.NewUser("random-key"), false)
	assert.True(t, flagValue)

	flagValue, _ = gffClient1.BoolVariation("test-flag", ffuser.NewUser("random-key"), false)
	assert.True(t, flagValue)

	// remove file we should still take the last version in consideration
	os.Remove(flagFile.Name())
	time.Sleep(2 * time.Second)

	flagValue, _ = gffClient1.BoolVariation("test-flag", ffuser.NewUser("random-key"), false)
	assert.True(t, flagValue)
}

func TestFlagFileUnreachable(t *testing.T) {
	initialFileContent := `test-flag:
  rule: key eq "random-key"
  percentage: 100
  true: "true"
  false: "false"
  default: "false"`

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

	flagValue, _ := gff.StringVariation("test-flag", ffuser.NewUser("random-key"), "SDKdefault")
	assert.Equal(t, "SDKdefault", flagValue, "should use the SDK default value")

	_ = os.WriteFile(flagFilePath, []byte(initialFileContent), os.ModePerm)
	time.Sleep(2 * time.Second)

	flagValue, _ = gff.StringVariation("test-flag", ffuser.NewUser("random-key"), "SDKdefault")
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
	user := ffuser.NewUser("random-key")
	hasTestFlag, _ := gff.BoolVariation("test-flag99", user, false)
	assert.True(t, hasTestFlag, "User should have test flag")
	hasUnknownFlag, _ := gff.BoolVariation("unknown-flag", user, false)
	assert.False(t, hasUnknownFlag, "User should use default value if flag does not exists")
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
