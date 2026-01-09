package testutils

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

// GetFreePort returns a free port on the local machine.
func GetFreePort(t *testing.T) int {
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

// GetFreePortAsString returns a free port on the local machine as a string.
func GetFreePortAsString(t *testing.T) string {
	port := GetFreePort(t)
	return fmt.Sprintf("%d", port)
}
