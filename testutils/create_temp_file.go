package testutils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func CopyFileToNewTempFile(t *testing.T, src string) *os.File {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		require.Fail(t, "File does not exist: %s", src)
	}
	srcContent, err := os.ReadFile(src)
	require.NoError(t, err)

	dir, err := os.MkdirTemp("", "goff")
	require.NoError(t, err)

	file, err := os.CreateTemp(dir, "")
	require.NoError(t, err)
	defer func() {
		_ = file.Close()
	}()

	err = os.WriteFile(file.Name(), srcContent, 0600)
	require.NoError(t, err)
	syncFile(file.Name())
	return file
}

func CopyContentToNewTempFile(t *testing.T, content string) *os.File {
	dir, err := os.MkdirTemp("", "goff")
	require.NoError(t, err)

	file, err := os.CreateTemp(dir, "")
	require.NoError(t, err)
	defer func() {
		_ = file.Close()
	}()

	err = os.WriteFile(file.Name(), []byte(content), 0600)
	require.NoError(t, err)
	syncFile(file.Name())
	return file
}

func CopyContentToExistingTempFile(t *testing.T, content string, file *os.File) *os.File {
	err := os.WriteFile(file.Name(), []byte(content), 0600)
	require.NoError(t, err)
	syncFile(file.Name())
	return file
}

func CopyFileToExistingTempFile(t *testing.T, src string, file *os.File) *os.File {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		require.Fail(t, "File does not exist: %s", src)
	}
	srcContent, err := os.ReadFile(src)
	require.NoError(t, err)
	return CopyContentToExistingTempFile(t, string(srcContent), file)
}

// syncFile ensures the file is written to disk before returning.
// This is important for file watchers that might detect changes before the file is fully written.
// Without syncing, the file watcher might detect the change before the OS has flushed the write,
// causing the reload to read stale or empty file content.
func syncFile(filePath string) {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err == nil {
		file.Sync()
		file.Close()
	}
}
