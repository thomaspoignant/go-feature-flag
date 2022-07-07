package gcstorageretriever

import (
	"context"
	"crypto/md5" //nolint: gosec
	"io"
	"io/ioutil"
	"testing"

	"cloud.google.com/go/storage"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/testutils"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

func TestGCStorageRetriever_Retrieve(t *testing.T) {
	type fields struct {
		Options []option.ClientOption
		Bucket  string
		Object  string
		rC      io.ReadCloser
	}
	type args struct {
		ctx context.Context
	}
	mockCtrl := gomock.NewController(t)

	tests := []struct {
		name                  string
		fields                fields
		args                  args
		want                  string
		wantErr               bool
		wantReadDataFromCache bool
	}{
		{
			name: "File on Object Not in Cache",
			fields: fields{
				Options: []option.ClientOption{option.WithCredentials(&google.Credentials{})},
				Bucket:  "bucket",
				Object:  "Object",
				rC: &testutils.GCStorageReaderMock{
					ShouldFail: false,
					FileToRead: "./testdata/flag-config-updated.yaml",
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr:               false,
			wantReadDataFromCache: false,
		},
		{
			name: "File on Object in Cache",
			fields: fields{
				Options: []option.ClientOption{option.WithCredentials(&google.Credentials{})},
				Bucket:  "bucket",
				Object:  "Object",
				rC: &testutils.GCStorageReaderMock{
					ShouldFail: true,
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr:               false,
			wantReadDataFromCache: true,
		},
		{
			name: "File not On Object",
			fields: fields{
				Options: []option.ClientOption{option.WithCredentials(&google.Credentials{})},
				Bucket:  "bucket",
				Object:  "Object",
				rC: &testutils.GCStorageReaderMock{
					ShouldFail: true,
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr:               true,
			wantReadDataFromCache: false,
		},
		{
			name: "Option Without Auth",
			fields: fields{
				Options: []option.ClientOption{option.WithoutAuthentication()},
				Bucket:  "bucket",
				Object:  "Object",
				rC: &testutils.GCStorageReaderMock{
					ShouldFail: false,
					FileToRead: "./testdata/flag-config-updated.yaml",
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr:               false,
			wantReadDataFromCache: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Retriever{
				Options: tt.fields.Options,
				Bucket:  tt.fields.Bucket,
				Object:  tt.fields.Object,
				rC:      tt.fields.rC,
				md5:     make([]byte, 16),
			}

			// Read default file.
			want, err := ioutil.ReadFile("./testdata/flag-config.yaml")
			assert.NoError(t, err)
			r.cache = want

			// Compute Hash of this file data.
			md5Hash := md5.Sum(want) //nolint: gosec
			wantedMd5 := md5Hash[:]
			copy(r.md5, wantedMd5)

			obj := testutils.NewMockobject(mockCtrl)
			if tt.wantReadDataFromCache {
				// If expect the data to be in cache, mock the
				// remote hash of the local data.
				obj.EXPECT().Attrs(context.Background()).Return(&storage.ObjectAttrs{MD5: r.md5}, nil).Times(1)
			} else {
				// If expect data not to be in cache, mock the
				// remote hash to a different one that the local hash.

				want, err = ioutil.ReadFile("./testdata/flag-config-updated.yaml")
				assert.NoError(t, err)

				md5Hash = md5.Sum(want) //nolint: gosec
				wantedMd5 = md5Hash[:]
				obj.EXPECT().Attrs(context.Background()).Return(&storage.ObjectAttrs{MD5: wantedMd5}, nil).Times(1)
			}

			r.obj = obj

			got, err := r.Retrieve(tt.args.ctx)

			assert.Equal(t, tt.wantErr, err != nil, "Retrieve() error = %v, wantErr %v", err, tt.wantErr)

			if err == nil {
				assert.Equal(t, want, got, "Retrieve() got = %v, want %v", got, tt.want)
				assert.Equal(t, want, r.cache, "Retrieve() got = Retriever{cache: %v} want = Retriever{cache: %v}", r.cache, want)
				assert.Equal(t, wantedMd5, r.md5, "Retrieve() got = Retriever{md5: %v} want = Retriever{md5: %v}", r.md5, wantedMd5)
			}
		})
	}
}
