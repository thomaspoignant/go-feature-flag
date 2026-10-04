package internal

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHTTPClient(t *testing.T) {
	transport := http.DefaultTransport.(*http.Transport).Clone()

	client := NewHTTPClient(2*time.Second, transport)

	httpClient, ok := client.(*http.Client)
	require.True(t, ok)
	assert.Equal(t, 2*time.Second, httpClient.Timeout)
	assert.Same(t, transport, httpClient.Transport)
}

func TestNewHTTPClient_WithDefaultTimeout(t *testing.T) {
	client := NewHTTPClient(0, nil)

	httpClient, ok := client.(*http.Client)
	require.True(t, ok)
	assert.Equal(t, 10*time.Second, httpClient.Timeout)
	assert.Nil(t, httpClient.Transport)
}
