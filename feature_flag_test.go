package ffclient

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/ffuser"
)

func TestStartWithoutRetriever(t *testing.T) {
	_, err := New(Config{
		PollInterval: 60,
	})
	assert.Error(t, err)
	ff = nil
}

func TestStartWithNegativeInterval(t *testing.T) {
	_, err := New(Config{
		PollInterval: -60,
		Retriever:    &FileRetriever{Path: "testdata/test.yaml"},
	})
	assert.Error(t, err)
	ff = nil
}

func TestValidUseCase(t *testing.T) {
	// Valid use case
	err := Init(Config{
		PollInterval: 5,
		Retriever:    &FileRetriever{Path: "testdata/test.yaml"},
		Logger:       log.New(os.Stdout, "", 0),
	})
	defer Close()

	assert.NoError(t, err)
	user := ffuser.NewUser("random-key")
	hasTestFlag, _ := BoolVariation("test-flag", user, false)
	assert.True(t, hasTestFlag, "User should have test flag")
	hasUnknownFlag, _ := BoolVariation("unknown-flag", user, false)
	assert.False(t, hasUnknownFlag, "User should use default value if flag does not exists")
}

func TestS3RetrieverReturnError(t *testing.T) {
	_, err := New(Config{
		Retriever: &S3Retriever{
			Bucket:    "unknown-bucket",
			Item:      "unknown-item",
			AwsConfig: aws.Config{},
		},
	})

	assert.Error(t, err)
}

func Test2GoFeatureFlagInstance(t *testing.T) {
	gffClient1, err := New(Config{
		PollInterval: 5,
		Retriever:    &FileRetriever{Path: "testdata/test.yaml"},
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
