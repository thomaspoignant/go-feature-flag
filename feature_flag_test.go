package ffclient_test

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
)

func TestNoRetriever(t *testing.T) {
	err := ffclient.Init(ffclient.Config{
		PollInterval: 3,
	})
	assert.Error(t, err)
	ffclient.Close()
}

func TestValidUseCase(t *testing.T) {
	// Valid use case
	err := ffclient.Init(ffclient.Config{
		PollInterval: 0,
		Retriever:    &ffclient.FileRetriever{Path: "testdata/test.yaml"},
	})

	assert.NoError(t, err)
	user := ffuser.NewUser("random-key")
	hasTestFlag, _ := ffclient.BoolVariation("test-flag", user, false)
	assert.True(t, hasTestFlag, "User should have test flag")
	hasUnknownFlag, _ := ffclient.BoolVariation("unknown-flag", user, false)
	assert.False(t, hasUnknownFlag, "User should use default value if flag does not exists")
	ffclient.Close()
}

func TestS3RetrieverReturnError(t *testing.T) {
	// Valid use case
	err := ffclient.Init(ffclient.Config{
		PollInterval: 0,
		Retriever: &ffclient.S3Retriever{
			Bucket:    "unknown-bucket",
			Item:      "unknown-item",
			AwsConfig: aws.Config{},
		},
	})
	assert.Error(t, err)
}
