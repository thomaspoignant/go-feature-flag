package ffclient

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"github.com/thomaspoignant/go-feature-flag/testutils"
)

func TestStartWithoutRetriever(t *testing.T) {
	_, err := New(Config{
		PollInterval: 60,
		Logger:       log.New(os.Stdout, "", 0),
	})
	assert.Error(t, err)
	ff = nil
}

func TestStartWithNegativeInterval(t *testing.T) {
	_, err := New(Config{
		PollInterval: -60,
		Retriever:    &FileRetriever{Path: "testdata/flag-config.yaml"},
		Logger:       log.New(os.Stdout, "", 0),
	})
	assert.Error(t, err)
	ff = nil
}

func TestValidUseCase(t *testing.T) {
	// Valid use case
	err := Init(Config{
		PollInterval: 5,
		Retriever:    &FileRetriever{Path: "testdata/flag-config.yaml"},
		Logger:       log.New(os.Stdout, "", 0),
		DataExporter: DataExporter{
			FlushInterval:    10 * time.Second,
			MaxEventInMemory: 1000,
			Exporter: &testutils.MockExporter{
				Mutex: sync.Mutex{},
			},
		},
	})
	defer Close()

	assert.NoError(t, err)
	user := ffuser.NewUser("random-key")
	hasTestFlag, _ := BoolVariation("test-flag", user, false)
	assert.True(t, hasTestFlag, "User should have test flag")
	hasUnknownFlag, _ := BoolVariation("unknown-flag", user, false)
	assert.False(t, hasUnknownFlag, "User should use default value if flag does not exists")
}

func TestValidUseCaseToml(t *testing.T) {
	// Valid use case
	gffClient, err := New(Config{
		PollInterval: 5,
		Retriever:    &FileRetriever{Path: "testdata/flag-config.toml"},
		Logger:       log.New(os.Stdout, "", 0),
		FileFormat:   "toml",
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
	gffClient, err := New(Config{
		PollInterval: 5,
		Retriever:    &FileRetriever{Path: "testdata/flag-config.json"},
		Logger:       log.New(os.Stdout, "", 0),
		FileFormat:   "json",
	})
	defer gffClient.Close()

	assert.NoError(t, err)
	user := ffuser.NewUser("random-key")
	hasTestFlag, _ := gffClient.BoolVariation("test-flag", user, false)
	assert.True(t, hasTestFlag, "User should have test flag")
	hasUnknownFlag, _ := gffClient.BoolVariation("unknown-flag", user, false)
	assert.False(t, hasUnknownFlag, "User should use default value if flag does not exists")
}

func TestS3RetrieverReturnError(t *testing.T) {
	_, err := New(Config{
		Retriever: &S3Retriever{
			Bucket:    "unknown-bucket",
			Item:      "unknown-item",
			AwsConfig: aws.Config{},
		},
		Logger: log.New(os.Stdout, "", 0),
	})

	assert.Error(t, err)
}

func Test2GoFeatureFlagInstance(t *testing.T) {
	gffClient1, err := New(Config{
		PollInterval: 5,
		Retriever:    &FileRetriever{Path: "testdata/flag-config.yaml"},
		Logger:       log.New(os.Stdout, "", 0),
	})
	defer gffClient1.Close()

	gffClient2, err2 := New(Config{
		PollInterval: 10,
		Retriever:    &FileRetriever{Path: "testdata/test-instance2.yaml"},
		Logger:       log.New(os.Stdout, "", 0),
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

	flagFile, _ := ioutil.TempFile("", "")
	_ = ioutil.WriteFile(flagFile.Name(), []byte(initialFileContent), 0600)

	gffClient1, _ := New(Config{
		PollInterval: 1,
		Retriever:    &FileRetriever{Path: flagFile.Name()},
		Logger:       log.New(os.Stdout, "", 0),
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

	_ = ioutil.WriteFile(flagFile.Name(), []byte(updatedFileContent), 0600)

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

	flagFile, _ := ioutil.TempFile("", "impossible")
	_ = ioutil.WriteFile(flagFile.Name(), []byte(initialFileContent), 0600)

	gffClient1, _ := New(Config{
		PollInterval: 1,
		Retriever:    &FileRetriever{Path: flagFile.Name()},
		Logger:       log.New(os.Stdout, "", 0),
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

func TestWrongWebhookConfig(t *testing.T) {
	_, err := New(Config{
		PollInterval: 5,
		Retriever:    &FileRetriever{Path: "testdata/flag-config.yaml"},
		Webhooks: []WebhookConfig{
			{
				PayloadURL: " https://example.com/hook",
				Secret:     "Secret",
				Meta: map[string]string{
					"my-app": "go-ff-test",
				},
			},
		},
	})

	assert.Errorf(t, err, "wrong url should return an error")
	assert.Equal(t, err.Error(), "wrong configuration in your webhook: parse \" https://example.com/hook\": "+
		"first path segment in URL cannot contain colon")
}
