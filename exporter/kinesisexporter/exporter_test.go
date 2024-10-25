package kinesysexporter

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/aws/smithy-go/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

func TestExporter_IsBulk(t *testing.T) {
	exp := Exporter{}
	assert.False(t, exp.IsBulk(), "DeprecatedExporter is not a bulk exporter")
}

func TestExporter_ExportBasicWithStreamName(t *testing.T) {
	mock := MockKinesisSender{}
	exp := Exporter{
		Format:     "json",
		sender:     &mock,
		StreamName: aws.String("test-stream"),
	}

	logger := &fflog.FFLogger{LeveledLogger: slog.Default()}

	assert.Nil(t, exp.AwsConfig)
	err := exp.Export(
		context.Background(),
		logger,
		[]exporter.FeatureEvent{
			NewFeatureEvent(),
			NewFeatureEvent(),
			NewFeatureEvent(),
		},
	)

	assert.NoError(t, err)
	assert.Len(t, mock.PutRecordsInputs, 1)
	assert.Len(t, mock.PutRecordsInputs[0].Records, 3)
	for idx := range mock.PutRecordsInputs {
		assert.Equal(t, mock.PutRecordsInputs[idx].StreamName, exp.StreamName)
		assert.Equal(t, mock.PutRecordsInputs[idx].StreamARN, exp.StreamArn)
	}

	assert.NotNil(t, exp.AwsConfig)
	assert.NoError(t, err)
}

func TestExporter_ExportBasicWithStreamArn(t *testing.T) {
	mock := MockKinesisSender{}
	exp := Exporter{
		Format:    "json",
		sender:    &mock,
		StreamArn: aws.String("test-stream"),
	}

	logger := &fflog.FFLogger{LeveledLogger: slog.Default()}

	assert.Nil(t, exp.AwsConfig)
	err := exp.Export(
		context.Background(),
		logger,
		[]exporter.FeatureEvent{
			NewFeatureEvent(),
			NewFeatureEvent(),
			NewFeatureEvent(),
		},
	)

	assert.NoError(t, err)
	assert.Len(t, mock.PutRecordsInputs, 1)
	assert.Len(t, mock.PutRecordsInputs[0].Records, 3)
	for idx := range mock.PutRecordsInputs {
		assert.Equal(t, mock.PutRecordsInputs[idx].StreamName, exp.StreamName)
		assert.Equal(t, mock.PutRecordsInputs[idx].StreamARN, exp.StreamArn)
	}

	assert.NotNil(t, exp.AwsConfig)

	assert.NoError(t, err)
	assert.NotNil(t, exp.AwsConfig)
}

func TestExporter_ShouldRaiseErrorIfNoStreamIsSpecified(t *testing.T) {
	mock := MockKinesisSender{}
	exp := Exporter{
		Format: "json",
		sender: &mock,
	}

	logger := &fflog.FFLogger{LeveledLogger: slog.Default()}

	assert.Nil(t, exp.AwsConfig)
	err := exp.Export(
		context.Background(),
		logger,
		[]exporter.FeatureEvent{NewFeatureEvent()},
	)

	assert.Error(t, err)
}

func TestExporter_ExportAWSConfigurationCustomisation(t *testing.T) {
	mock := MockKinesisSender{}
	exp := Exporter{
		Format:     "json",
		sender:     &mock,
		StreamName: aws.String("test-stream"),
		AwsConfig: &aws.Config{
			Region: "unexistent-region",
		},
	}

	logger := &fflog.FFLogger{LeveledLogger: slog.Default()}

	err := exp.Export(
		context.Background(),
		logger,
		[]exporter.FeatureEvent{
			NewFeatureEvent(),
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, exp.AwsConfig.Region, "unexistent-region")
}

func TestExporter_ExportSenderError(t *testing.T) {
	mock := MockKinesisSenderWithError{}

	exp := Exporter{
		Format:     "json",
		sender:     &mock,
		StreamName: aws.String("test-stream"),
	}

	logger := &fflog.FFLogger{LeveledLogger: slog.Default()}

	err := exp.Export(
		context.Background(),
		logger,
		[]exporter.FeatureEvent{
			NewFeatureEvent(),
		},
	)

	assert.Error(t, err)
}

func NewFeatureEvent() exporter.FeatureEvent {
	return exporter.FeatureEvent{
		Kind:         "feature",
		ContextKind:  "anonymousUser",
		UserKey:      "ABCD",
		CreationDate: 1617970547,
		Key:          "random-key",
		Variation:    "Default",
		Value:        "YO",
		Default:      false,
	}
}

type MockKinesisSender struct {
	PutRecordsInputs []*kinesis.PutRecordsInput
}

func (k *MockKinesisSender) SendMessages(ctx context.Context, msgs *kinesis.PutRecordsInput) (*kinesis.PutRecordsOutput, error) {
	k.PutRecordsInputs = append(k.PutRecordsInputs, msgs)
	failedRecordCount := int32(0)
	output := kinesis.PutRecordsOutput{
		Records:           []types.PutRecordsResultEntry{},
		FailedRecordCount: &failedRecordCount,
		EncryptionType:    types.EncryptionTypeNone,
		ResultMetadata:    middleware.Metadata{},
	}
	return &output, nil
}

type MockKinesisSenderWithError struct{}

func (k *MockKinesisSenderWithError) SendMessages(ctx context.Context, msgs *kinesis.PutRecordsInput) (*kinesis.PutRecordsOutput, error) {
	return nil, errors.New("failure to send message: datacenter on fire")
}
