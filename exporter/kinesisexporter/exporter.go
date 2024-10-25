package kinesysexporter

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

const (
	formatJSON = "json"
	Mb         = 1024 * 1024
)

type MessageSender interface {
	SendMessages(ctx context.Context, msgs *kinesis.PutRecordsInput) (*kinesis.PutRecordsOutput, error)
}

type DefaultKinesisSender struct {
	*kinesis.Client
}

func (k *DefaultKinesisSender) SendMessages(ctx context.Context, msgs *kinesis.PutRecordsInput) (*kinesis.PutRecordsOutput, error) {
	return k.PutRecords(ctx, msgs)
}

type Exporter struct {
	// AwsConfig is the AWS SDK configuration object we will use to
	// upload your exported data files.
	AwsConfig *aws.Config

	// Format is the output format you want in your exported file.
	// Available format are JSON, CSV and Parquet.
	// Default: JSON
	Format string

	// kinesis.Options is a list of functional options to configure the Kinesis client.
	// Provide additional functional options to further configure the behavior of the client,
	// such as changing the client's endpoint or adding custom middleware behavior.
	// For more information about the options, please check:
	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/kinesis#Options
	KinesisOptions []func(*kinesis.Options)

	StreamName *string
	StreamArn  *string

	init sync.Once

	sender MessageSender
}

func (e *Exporter) initializeProducer(ctx context.Context) error {
	var initErr error
	e.init.Do(func() {
		if e.AwsConfig == nil {
			cfg, err := config.LoadDefaultConfig(ctx)
			if err != nil {
				initErr = fmt.Errorf("impossible to init Kinesis exporter: %v", err)
				return
			}
			e.AwsConfig = &cfg
		}
		if e.sender == nil {
			e.sender = &DefaultKinesisSender{
				kinesis.NewFromConfig(*e.AwsConfig, e.KinesisOptions...),
			}
		}
	})
	return initErr
}

func (e *Exporter) Export(ctx context.Context, logger *fflog.FFLogger, featureEvents []exporter.FeatureEvent) error {
	err := e.initializeProducer(ctx)
	if err != nil {
		return fmt.Errorf("writer: %w", err)
	}

	records := make([]types.PutRecordsRequestEntry, 0, len(featureEvents))

	for _, event := range featureEvents {
		formattedEvent, err := e.formatMessage(event)

		if err != nil {
			return fmt.Errorf("format: %w", err)
		}

		if len(formattedEvent) >= Mb {
			logger.Error("format: Event is too large, skipping", err)
			continue
		}

		partitionKey := hex.EncodeToString(md5.New().Sum(formattedEvent))

		records = append(records, types.PutRecordsRequestEntry{
			Data:            formattedEvent,
			ExplicitHashKey: &partitionKey,
		})
	}

	input := &kinesis.PutRecordsInput{
		Records: records,
	}

	if e.StreamArn != nil {
		input.StreamARN = e.StreamArn
	} else if e.StreamName != nil {
		input.StreamName = e.StreamName
	} else {
		return fmt.Errorf("send: no StreamName or StreamArn provided")
	}

	output, err := e.sender.SendMessages(ctx, input)

	if err != nil {
		return fmt.Errorf("send: %w", err)
	}

	if *output.FailedRecordCount > 0 {
		logger.Error("send: couldn't send %d records to Kinesis", output.FailedRecordCount)
	}

	for _, record := range output.Records {
		if record.ErrorCode != nil || record.ErrorMessage != nil {
			logger.Error(
				"send: couldn't send event to Kinesis: ErrorCode: %s, ErrorMessage: %s",
				record.ErrorCode,
				record.ErrorMessage,
			)
		}
	}

	return nil
}

// formatMessage returns the event encoded in the selected format. Will always use JSON for now.
func (e *Exporter) formatMessage(event exporter.FeatureEvent) ([]byte, error) {
	switch e.Format {
	case formatJSON:
		fallthrough
	default:
		return json.Marshal(event)
	}
}

// IsBulk reports if the producer can handle bulk messages. Will always return false for this exporter.
func (f *Exporter) IsBulk() bool {
	return false
}
