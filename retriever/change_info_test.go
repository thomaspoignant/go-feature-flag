package retriever_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/retriever"
)

type plainRetriever struct{}

func (plainRetriever) Retrieve(context.Context) ([]byte, error) { return nil, nil }

type providerRetriever struct {
	plainRetriever
	info *retriever.ChangeInfo
	err  error
}

func (p providerRetriever) LastChange(context.Context) (*retriever.ChangeInfo, error) {
	return p.info, p.err
}

func TestNewChangeInfoResolver(t *testing.T) {
	old := &retriever.ChangeInfo{Author: "old", ChangedAt: time.Unix(100, 0)}
	recent := &retriever.ChangeInfo{Author: "recent", ChangedAt: time.Unix(200, 0)}
	tests := []struct {
		name       string
		retrievers []retriever.Retriever
		want       *retriever.ChangeInfo
		wantErr    bool
		wantNil    bool
	}{
		{
			name:       "no retriever able to give the author",
			retrievers: []retriever.Retriever{plainRetriever{}},
			wantNil:    true,
		},
		{
			name: "most recent change wins",
			retrievers: []retriever.Retriever{
				providerRetriever{info: old}, plainRetriever{}, providerRetriever{info: recent},
			},
			want: recent,
		},
		{
			name: "a failing retriever is ignored if another answers",
			retrievers: []retriever.Retriever{
				providerRetriever{err: errors.New("rate limited")}, providerRetriever{info: old},
			},
			want: old,
		},
		{
			name:       "error when nobody answers",
			retrievers: []retriever.Retriever{providerRetriever{err: errors.New("rate limited")}},
			wantErr:    true,
		},
		{
			name:       "provider without information",
			retrievers: []retriever.Retriever{providerRetriever{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolve := retriever.NewChangeInfoResolver(tt.retrievers)
			if tt.wantNil {
				assert.Nil(t, resolve)
				return
			}
			require.NotNil(t, resolve)
			got, err := resolve(context.Background())
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, got)
		})
	}
}
