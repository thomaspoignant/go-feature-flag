package kinesisexporter

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
	assert.False(t, exp.IsBulk(), "DeprecatedExporterV1 is not a bulk exporter")
}

func TestExporter_ExportBasicWithStreamName(t *testing.T) {
	mock := MockKinesisSender{}
	exp := Exporter{
		Format:   "json",
		sender:   &mock,
		Settings: NewSettings(WithStreamName("test-stream")),
	}

	logger := &fflog.FFLogger{LeveledLogger: slog.Default()}

	assert.Nil(t, exp.AwsConfig)
	err := exp.Export(
		context.Background(),
		logger,
		[]exporter.ExportableEvent{
			NewFeatureEvent(),
			NewFeatureEvent(),
			NewFeatureEvent(),
		},
	)

	assert.NoError(t, err)
	assert.Len(t, mock.PutRecordsInputs, 1)
	assert.Len(t, mock.PutRecordsInputs[0].Records, 3)
	for idx := range mock.PutRecordsInputs {
		assert.Equal(t, mock.PutRecordsInputs[idx].StreamName, exp.Settings.StreamName)
		assert.Equal(t, mock.PutRecordsInputs[idx].StreamARN, exp.Settings.StreamArn)
	}

	assert.NotNil(t, exp.AwsConfig)
	assert.NoError(t, err)
}

func TestExporter_ExportBasicWithStreamArn(t *testing.T) {
	mock := MockKinesisSender{}
	exp := Exporter{
		Format:   "json",
		sender:   &mock,
		Settings: NewSettings(WithStreamArn("test-stream")),
	}

	logger := &fflog.FFLogger{LeveledLogger: slog.Default()}

	assert.Nil(t, exp.AwsConfig)
	err := exp.Export(
		context.Background(),
		logger,
		[]exporter.ExportableEvent{
			NewFeatureEvent(),
			NewFeatureEvent(),
			NewFeatureEvent(),
		},
	)

	assert.NoError(t, err)
	assert.Len(t, mock.PutRecordsInputs, 1)
	assert.Len(t, mock.PutRecordsInputs[0].Records, 3)
	for idx := range mock.PutRecordsInputs {
		assert.Equal(t, mock.PutRecordsInputs[idx].StreamName, exp.Settings.StreamName)
		assert.Equal(t, mock.PutRecordsInputs[idx].StreamARN, exp.Settings.StreamArn)
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
		[]exporter.ExportableEvent{NewFeatureEvent()},
	)

	assert.Error(t, err)
}

func TestExporter_ExportAWSConfigurationCustomisation(t *testing.T) {
	mock := MockKinesisSender{}

	exp := Exporter{
		Format: "json",
		sender: &mock,
		Settings: NewSettings(
			WithStreamName("test-stream"),
			WithPartitionKey(func(context.Context, exporter.ExportableEvent) string {
				return "test-key"
			}),
		),
		AwsConfig: &aws.Config{
			Region: "unexistent-region",
		},
	}

	logger := &fflog.FFLogger{LeveledLogger: slog.Default()}

	err := exp.Export(
		context.Background(),
		logger,
		[]exporter.ExportableEvent{
			NewFeatureEvent(),
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, *mock.PutRecordsInputs[0].Records[0].PartitionKey, "test-key")
	assert.Equal(t, exp.AwsConfig.Region, "unexistent-region")
}

func TestExporter_ExportSenderError(t *testing.T) {
	mock := MockKinesisSenderWithError{}

	exp := Exporter{
		Format:   "json",
		sender:   &mock,
		Settings: NewSettings(WithStreamName("test-stream")),
	}

	logger := &fflog.FFLogger{LeveledLogger: slog.Default()}

	err := exp.Export(
		context.Background(),
		logger,
		[]exporter.ExportableEvent{
			NewFeatureEvent(),
		},
	)

	assert.Error(t, err)
}

func TestExporterSettingsCreation(t *testing.T) {
	{
		settings := NewSettings()
		assert.Equal(t, settings.PartitionKey(context.TODO(), NewFeatureEvent()), "default")
		assert.Nil(t, settings.StreamName)
		assert.Nil(t, settings.StreamArn)
		assert.Nil(t, settings.ExplicitHashKey)
	}
	{
		settings := NewSettings(WithStreamArn("test-stream-arn"))
		assert.Equal(t, settings.PartitionKey(context.TODO(), NewFeatureEvent()), "default")
		assert.Nil(t, settings.StreamName)
		assert.Equal(t, *settings.StreamArn, "test-stream-arn")
		assert.Nil(t, settings.ExplicitHashKey)
	}
	{
		settings := NewSettings(WithStreamName("test-stream-name"))
		assert.Equal(t, settings.PartitionKey(context.TODO(), NewFeatureEvent()), "default")
		assert.Equal(t, *settings.StreamName, "test-stream-name")
		assert.Nil(t, settings.StreamArn)
		assert.Nil(t, settings.ExplicitHashKey)
	}
	{
		settings := NewSettings(WithExplicitHashKey("test-explicit-hash-key"))
		assert.Equal(t, settings.PartitionKey(context.TODO(), NewFeatureEvent()), "default")
		assert.Nil(t, settings.StreamName)
		assert.Nil(t, settings.StreamArn)
		assert.Equal(t, *settings.ExplicitHashKey, "test-explicit-hash-key")
	}
	{
		settings := NewSettings(
			WithStreamName("test-stream-name"),
			WithStreamArn("test-stream-arn"),
			WithExplicitHashKey("test-explicit-hash-key"),
			WithPartitionKey(
				func(_ context.Context, _ exporter.ExportableEvent) string { return "non-default" },
			),
		)
		assert.Equal(t, settings.PartitionKey(context.TODO(), NewFeatureEvent()), "non-default")
		assert.Nil(t, settings.StreamName) // overwritten by streamArn
		assert.Equal(t, *settings.StreamArn, "test-stream-arn")
		assert.Equal(t, *settings.ExplicitHashKey, "test-explicit-hash-key")
	}
	{
		settings := NewSettings(
			WithStreamArn("test-stream-arn"),
			WithStreamName("test-stream-name"),
		)
		assert.Nil(t, settings.StreamArn) // overwritten by streamName
		assert.Equal(t, *settings.StreamName, "test-stream-name")
	}
}

func TestHugeMessageExportFlow(t *testing.T) {
	event := exporter.FeatureEvent{
		Kind:         "feature",
		ContextKind:  "anonymousUser",
		UserKey:      "ABCD",
		CreationDate: 1617970547,
		Key:          "random-key",
		Variation:    "Default",
		Value:        "YO",
		Default:      false,
	}
	event.Value = string(make([]byte, Mb))

	mock := MockKinesisSender{}

	exp := Exporter{
		Format:   "json",
		sender:   &mock,
		Settings: NewSettings(WithStreamName("test-stream")),
	}

	logger := &fflog.FFLogger{LeveledLogger: slog.Default()}

	err := exp.Export(
		context.Background(),
		logger,
		[]exporter.ExportableEvent{
			event,
			event,
			event,
			event,
		},
	)

	assert.NoError(t, err)
	assert.Len(t, mock.PutRecordsInputs, 0)
}

func NewFeatureEvent() exporter.ExportableEvent {
	return &exporter.FeatureEvent{
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

func (k *MockKinesisSender) SendMessages(
	_ context.Context,
	msgs *kinesis.PutRecordsInput,
) (*kinesis.PutRecordsOutput, error) {
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

func (k *MockKinesisSenderWithError) SendMessages(
	_ context.Context,
	_ *kinesis.PutRecordsInput,
) (*kinesis.PutRecordsOutput, error) {
	return nil, errors.New("failure to send message: datacenter on fire")
}
