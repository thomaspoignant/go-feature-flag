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
	ff = newGoFeatureFlag(Config{
		PollInterval: 60,
	})
	err := ff.startUpdater()
	assert.Error(t, err)
	ff = nil
}

func TestStartWithNegativeInterval(t *testing.T) {
	ff = newGoFeatureFlag(Config{
		PollInterval: -60,
		Retriever:    &FileRetriever{Path: "testdata/test.yaml"},
	})
	err := ff.startUpdater()
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
	ff = newGoFeatureFlag(Config{
		Retriever: &S3Retriever{
			Bucket:    "unknown-bucket",
			Item:      "unknown-item",
			AwsConfig: aws.Config{},
		},
	})

	err := ff.startUpdater()
	assert.Error(t, err)
}
