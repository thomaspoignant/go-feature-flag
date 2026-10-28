package s3retrieverv2

// Whitebox because the real AWS clients built by Init are only reachable through the unexported
// downloader and lister fields. Not parallel because the subtests use t.Setenv.

import (
	"context"
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
		// nilCtx calls Init with a nil context.
		nilCtx bool
		// env sets extra environment variables for the test.
		env map[string]string
		// wantErr expects Init to fail while loading the default AWS configuration.
		wantErr bool
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
		{
			name:       "nil context",
			r:          func() *Retriever { return &Retriever{Bucket: "b", Item: "i"} },
			nilCtx:     true,
			wantRegion: "us-west-2",
		},
		{
			name:    "default AWS config cannot be loaded",
			r:       func() *Retriever { return &Retriever{Bucket: "b", Item: "i"} },
			env:     map[string]string{"AWS_PROFILE": "profile-that-does-not-exist"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("AWS_REGION", "us-west-2")
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			r := tt.r()
			err := initRetriever(t, r, tt.awsConfig, tt.nilCtx)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, retriever.RetrieverError, r.Status())
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, r.downloader)
			assert.Equal(t, tt.wantLister, r.lister != nil)
			assert.Equal(t, tt.wantRegion, r.AwsConfig.Region)
			assert.Equal(t, retriever.RetrieverReady, r.Status())
			assertListerClient(t, r, tt.wantLister, tt.wantAccelerate)
		})
	}
}

// initRetriever calls Init, optionally with an explicit us-east-1 AwsConfig or a nil context.
func initRetriever(t *testing.T, r *Retriever, awsConfig, nilCtx bool) error {
	t.Helper()
	if awsConfig {
		cfg, err := config.LoadDefaultConfig(t.Context(), config.WithRegion("us-east-1"))
		require.NoError(t, err)
		r.AwsConfig = &cfg
	}
	var ctx context.Context = t.Context()
	if nilCtx {
		ctx = nil
	}
	return r.Init(ctx, nil)
}

// assertListerClient checks that the lister built by Init is a real S3 client with the expected
// UseAccelerate option.
func assertListerClient(t *testing.T, r *Retriever, wantLister, wantAccelerate bool) {
	t.Helper()
	if !wantLister {
		return
	}
	client, ok := r.lister.(*s3.Client)
	require.True(t, ok, "the lister should be a real S3 client")
	assert.Equal(t, wantAccelerate, client.Options().UseAccelerate)
}
