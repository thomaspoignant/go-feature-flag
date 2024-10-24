package kinesys

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
)

type Exporter struct {
	// AwsConfig is the AWS SDK configuration object we will use to
	// upload your exported data files.
	AwsConfig *aws.Config

	// Format is the output format you want in your exported file.
	// Available format are JSON, CSV and Parquet.
	// Default: JSON
	Format string

	// S3ClientOptions is a list of functional options to configure the S3 client.
	// Provide additional functional options to further configure the behavior of the client,
	// such as changing the client's endpoint or adding custom middleware behavior.
	// For more information about the options, please check:
	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3#Options
	KinesisOptions []func(*kinesis.Options)

	StreamName *string
	StreamArn  *string

	init sync.Once

	sender *kinesis.Client
	// dialer will create the producer. This field is added for dependency injection during testing as sarama
	// has the annoying tendency to dial as soon as a producer is created.
	// dialer func(addrs []string, config *sarama.Config) (MessageSender, error)
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

		e.sender = kinesis.NewFromConfig(*e.AwsConfig, e.KinesisOptions...)

	})
	return initErr
}

func (e *Exporter) Export(ctx context.Context, logger *fflog.FFLogger, featureEvents []exporter.FeatureEvent) error {
	if e.sender == nil {
		err := e.initializeProducer(ctx)
		if err != nil {
			return fmt.Errorf("writer: %w", err)
		}
	}

	records := make([]types.PutRecordsRequestEntry, 0, len(featureEvents))

	for _, event := range featureEvents {
		formattedEvent, err := e.formatMessage(event)

		if err != nil {
			return fmt.Errorf("format: %w", err)
		}

		partitionKey := hex.EncodeToString(md5.New().Sum(formattedEvent))

		records = append(records, types.PutRecordsRequestEntry{
			Data:            formattedEvent,
			ExplicitHashKey: &partitionKey,
		})
	}

	output, err := e.sender.PutRecords(
		ctx, &kinesis.PutRecordsInput{
			Records:    records,
			StreamARN:  e.StreamArn,
			StreamName: e.StreamName,
		},
	)

	// Logging
	logger.Info("Blah blah blah {}", output.FailedRecordCount)
	logger.Info("Blah blah blah {}", output.Records[0].ErrorCode)
	logger.Info("Blah blah blah {}", output.Records[0].ErrorMessage)

	if err != nil {
		return fmt.Errorf("send: %w", err)
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
