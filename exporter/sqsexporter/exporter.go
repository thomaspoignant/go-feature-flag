package sqsexporter

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

var _ exporter.Exporter = &Exporter{}

type Exporter struct {
	// QueueURL is the URL of your SQS queue
	// (mandatory)
	QueueURL string

	// AwsConfig is the AWS SDK configuration object we will use to
	// upload your exported data files.
	AwsConfig *aws.Config

	init       sync.Once
	sqsService SQSSendMessageAPI
}

// Export is sending SQS event for each events received.
func (f *Exporter) Export(
	ctx context.Context,
	_ *fflog.FFLogger,
	events []exporter.ExportableEvent,
) error {
	if f.AwsConfig == nil {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return fmt.Errorf("impossible to init SQS exporter: %v", err)
		}
		f.AwsConfig = &cfg
	}

	if f.QueueURL == "" {
		return fmt.Errorf("impossible to init SQS exporter: QueueURL is a mandatory parameter")
	}

	if f.sqsService == nil {
		f.init.Do(func() {
			f.sqsService = sqs.NewFromConfig(*f.AwsConfig)
		})
	}

	for _, event := range events {
		messageBody, err := json.Marshal(event)
		if err != nil {
			return err
		}
		_, err = f.sqsService.SendMessage(ctx, &sqs.SendMessageInput{
			MessageBody: aws.String(string(messageBody)),
			QueueUrl:    aws.String(f.QueueURL),
			MessageAttributes: map[string]types.MessageAttributeValue{
				"emitter": {
					DataType:    aws.String("String"),
					StringValue: aws.String("GO Feature Flag"),
				},
			},
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (f *Exporter) IsBulk() bool {
	return false
}

// SQSSendMessageAPI defines the interface for the GetQueueUrl and SendMessage functions.
// We use this interface to test the functions using a mocked service.
type SQSSendMessageAPI interface {
	SendMessage(ctx context.Context,
		params *sqs.SendMessageInput,
		optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
}
