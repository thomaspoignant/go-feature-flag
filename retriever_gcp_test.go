package ffclient

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/testutils"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"io"
	"io/ioutil"
	"testing"
)

func TestGCPRetriever_Retrieve(t *testing.T) {
	type fields struct {
		Option option.ClientOption
		Bucket string
		Object string
		rC     io.ReadCloser
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "File on Object",
			fields: fields{
				Option: option.WithCredentials(&google.Credentials{}),
				Bucket: "bucket",
				Object: "object",
				rC: &testutils.GCPStorageReaderMock{
					ShouldFail: false,
				},
			},
			args: args{
				ctx: context.Background(),
			},
			want:    "./testdata/flag-config.yaml",
			wantErr: false,
		},
		{
			name: "File not On Object",
			fields: fields{
				Option: option.WithCredentials(&google.Credentials{}),
				Bucket: "bucket",
				Object: "object",
				rC: &testutils.GCPStorageReaderMock{
					ShouldFail: true,
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
		{
			name: "Option Without Auth",
			fields: fields{
				Option: option.WithoutAuthentication(),
				Bucket: "bucket",
				Object: "object",
				rC: &testutils.GCPStorageReaderMock{
					ShouldFail: false,
				},
			},
			args: args{
				ctx: context.Background(),
			},
			want:    "./testdata/flag-config.yaml",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &GCPRetriever{
				Option: tt.fields.Option,
				Bucket: tt.fields.Bucket,
				Object: tt.fields.Object,
				rC:     tt.fields.rC,
			}
			got, err := r.Retrieve(tt.args.ctx)
			assert.Equal(t, tt.wantErr, err != nil, "Retrieve() error = %v, wantErr %v", err, tt.wantErr)
			if err == nil {
				want, err := ioutil.ReadFile(tt.want)
				assert.NoError(t, err)
				assert.Equal(t, want, got, "Retrieve() got = %v, want %v", got, tt.want)
			}
		})
	}
}
