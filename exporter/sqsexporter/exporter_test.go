package sqsexporter

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"log"
	"os"
	"strings"
	"testing"
)

type SQSSendMessageAPIMock struct {
	messages []sqs.SendMessageInput
}

func (s *SQSSendMessageAPIMock) SendMessage(ctx context.Context,
	params *sqs.SendMessageInput,
	optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	if params.QueueUrl != nil && strings.HasSuffix(*params.QueueUrl, "error") {
		return nil, fmt.Errorf("random error")
	}
	s.messages = append(s.messages, *params)
	return nil, nil
}

func TestSQS_IsBulk(t *testing.T) {
	exporter := Exporter{}
	assert.False(t, exporter.IsBulk(), "Exporter exporter is not a bulk exporter")
}

func TestExporter_Export(t *testing.T) {
	type fields struct {
		QueueURL   string
		AwsConfig  *aws.Config
		sqsService SQSSendMessageAPIMock
	}
	tests := []struct {
		name          string
		fields        fields
		featureEvents []exporter.FeatureEvent
		wantErr       bool
	}{
		{
			name: "should return an error if no QueueURL provided",
			fields: fields{
				//QueueURL:   "https://sqs.eu-west-1.amazonaws.com/XXX/test-queue",
				sqsService: SQSSendMessageAPIMock{},
			},
			wantErr: true,
			featureEvents: []exporter.FeatureEvent{
				{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
		},
		{
			name: "should receive an event with a valid feature event",
			fields: fields{
				QueueURL:   "https://sqs.eu-west-1.amazonaws.com/XXX/test-queue",
				sqsService: SQSSendMessageAPIMock{},
			},
			wantErr: false,
			featureEvents: []exporter.FeatureEvent{
				{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
				{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCDEF", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
		},
		{
			name: "should return an error if AWS SQS is returning an error",
			fields: fields{
				QueueURL:   "https://sqs.eu-west-1.amazonaws.com/XXX/error",
				sqsService: SQSSendMessageAPIMock{},
			},
			wantErr: true,
			featureEvents: []exporter.FeatureEvent{
				{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
				{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCDEF", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Exporter{
				QueueURL:   tt.fields.QueueURL,
				AwsConfig:  tt.fields.AwsConfig,
				sqsService: &tt.fields.sqsService,
			}

			logger := log.New(os.Stdout, "", 0)
			err := f.Export(context.TODO(), logger, tt.featureEvents)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			want := make([]sqs.SendMessageInput, len(tt.featureEvents))
			for index, event := range tt.featureEvents {
				messageBody, _ := json.Marshal(event)
				want[index] = sqs.SendMessageInput{
					MessageBody:  aws.String(string(messageBody)),
					QueueUrl:     aws.String(tt.fields.QueueURL),
					DelaySeconds: 0,
					MessageAttributes: map[string]types.MessageAttributeValue{
						"emitter": types.MessageAttributeValue{
							DataType:    aws.String("String"),
							StringValue: aws.String("GO Feature Flag"),
						},
					},
				}
			}
			assert.Equal(t, want, tt.fields.sqsService.messages)
		})
	}
}
