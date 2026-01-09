package testutils

import (
	"os"
	"strings"
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
	syncFile(t, file.Name())
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
	syncFile(t, file.Name())
	return file
}

func CopyContentToExistingTempFile(t *testing.T, content string, file *os.File) *os.File {
	err := os.WriteFile(file.Name(), []byte(content), 0600)
	require.NoError(t, err)
	syncFile(t, file.Name())
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

func ReplaceInFile(t *testing.T, file *os.File, old, new string) {
	content, err := os.ReadFile(file.Name())
	require.NoError(t, err)
	content = []byte(strings.Replace(string(content), old, new, 1))
	err = os.WriteFile(file.Name(), []byte(content), 0600)
	require.NoError(t, err)
	syncFile(t, file.Name())
}

func ReplaceAndCopyFileToExistingFile(t *testing.T, src string, dstFile *os.File, old, new string) *os.File {
	tempConfigFile := CopyFileToNewTempFile(t, src)
	defer func() {
		_ = os.Remove(tempConfigFile.Name())
	}()
	ReplaceInFile(t, tempConfigFile, old, new)
	return CopyFileToExistingTempFile(t, tempConfigFile.Name(), dstFile)
}

// syncFile ensures the file is written to disk before returning.
// This is important for file watchers that might detect changes before the file is fully written.
// Without syncing, the file watcher might detect the change before the OS has flushed the write,
// causing the reload to read stale or empty file content.
func syncFile(t *testing.T, filePath string) {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err == nil {
		err := file.Sync()
		require.NoError(t, err)
		err = file.Close()
		require.NoError(t, err)
	}
}
