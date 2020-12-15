package ffclient_test

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
)

func TestInit(t *testing.T) {
	type args struct {
		config ffclient.Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid use case",
			args: args{
				config: ffclient.Config{
					PollInterval: 0,
					LocalFile:    "testdata/test.yaml",
				},
			},
			wantErr: false,
		},
		{
			name: "no retriever",
			args: args{
				config: ffclient.Config{
					PollInterval: 3,
				},
			},
			wantErr: true,
		},
		{
			name: "S3 retriever return error",
			args: args{
				config: ffclient.Config{
					PollInterval: 3,
					S3Retriever: &ffclient.S3Retriever{
						Bucket:    "unkown-bucket",
						Item:      "unknown-item",
						AwsConfig: aws.Config{},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ffclient.Init(tt.args.config)
			defer ffclient.Close()
			assert.Equal(t, tt.wantErr, err != nil)

			if err == nil {
				user := ffuser.NewUser("random-key")
				hasTestFlag, _ := ffclient.BoolVariation("test-flag", user, false)
				assert.True(t, hasTestFlag, "User should have test flag")
				hasUnknownFlag, _ := ffclient.BoolVariation("unknown-flag", user, false)
				assert.False(t, hasUnknownFlag, "User should use default value if flag does not exists")
			}
		})
	}
}
