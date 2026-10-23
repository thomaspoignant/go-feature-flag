package s3retrieverv2

// Whitebox because the real AWS clients built by Init are only reachable through the unexported
// downloader and lister fields. Not parallel because the subtests use t.Setenv.

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/retriever"
)

func TestRetriever_Init_RealClients(t *testing.T) {
	useAccelerate := []func(*s3.Options){func(o *s3.Options) { o.UseAccelerate = true }}
	tests := []struct {
		name       string
		r          func() *Retriever
		awsConfig  bool // set an explicit AwsConfig in us-east-1
		wantRegion string
		wantLister bool
		// wantAccelerate is the expected UseAccelerate option of the built S3 client.
		wantAccelerate bool
	}{
		{
			name:       "item mode with no AwsConfig",
			r:          func() *Retriever { return &Retriever{Bucket: "b", Item: "i"} },
			wantRegion: "us-west-2",
		},
		{
			name:       "prefix mode builds a lister",
			r:          func() *Retriever { return &Retriever{Bucket: "b", Prefix: "f/"} },
			wantRegion: "us-west-2",
			wantLister: true,
		},
		{
			name:       "with AwsConfig",
			r:          func() *Retriever { return &Retriever{Bucket: "b", Item: "i"} },
			awsConfig:  true,
			wantRegion: "us-east-1",
		},
		{
			name:           "with S3 client options",
			r:              func() *Retriever { return &Retriever{Bucket: "b", Prefix: "f/", S3ClientOptions: useAccelerate} },
			wantRegion:     "us-west-2",
			wantLister:     true,
			wantAccelerate: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("AWS_REGION", "us-west-2")
			r := tt.r()
			if tt.awsConfig {
				awsConfig, err := config.LoadDefaultConfig(t.Context(), config.WithRegion("us-east-1"))
				require.NoError(t, err)
				r.AwsConfig = &awsConfig
			}
			require.NoError(t, r.Init(t.Context(), nil))
			assert.NotNil(t, r.downloader)
			assert.Equal(t, tt.wantLister, r.lister != nil)
			assert.Equal(t, tt.wantRegion, r.AwsConfig.Region)
			assert.Equal(t, retriever.RetrieverReady, r.Status())
			if tt.wantLister {
				client, ok := r.lister.(*s3.Client)
				require.True(t, ok, "the lister should be a real S3 client")
				assert.Equal(t, tt.wantAccelerate, client.Options().UseAccelerate)
			}
		})
	}
}
