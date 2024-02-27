package tail_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/getsavvyinc/savvy-cli/tail"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testDataDir = "testdata"
)

var (
	testDataFile  = filepath.Join(testDataDir, "data.txt")
	testEmptyFile = filepath.Join(testDataDir, "empty.txt")
)

func TestTail(t *testing.T) {
	t.Run("FailsOnNonExistentFile", func(t *testing.T) {
		_, err := tail.Tail("nonexistentfile", 10)
		require.Error(t, err)
		require.ErrorIs(t, err, os.ErrNotExist)
	})
	t.Run("FailsOnDirectory", func(t *testing.T) {
		_, err := tail.Tail(testDataDir, 10)
		require.Error(t, err)
	})
	t.Run("FailsOnEmptyFile", func(t *testing.T) {
		_, err := tail.Tail(testEmptyFile, 10)
		assert.Error(t, err)
		assert.ErrorIs(t, err, tail.ErrEmptyFile)
	})
	t.Run("FailsOnNegativeLines", func(t *testing.T) {
		_, err := tail.Tail(testDataFile, -1)
		assert.Error(t, err)
		assert.ErrorIs(t, err, tail.ErrInvalidN)
	})
	t.Run("ReturnsEntireFileWhenNIsGreaterThanFileLength", func(t *testing.T) {
		f, err := tail.Tail(testDataFile, 1000)
		require.NoError(t, err)
		defer f.Close()

		got, err := io.ReadAll(f)
		assert.NoError(t, err)

		expected, err := os.ReadFile(testDataFile)
		require.NoError(t, err)
		assert.Equal(t, string(expected), string(got))
	})
	t.Run("ReturnsLastNLines", func(t *testing.T) {
		f, err := tail.Tail(testDataFile, 3)
		require.NoError(t, err)
		defer f.Close()

		expected := `right from their terminal.
Savvy's ClI also allows developers to create runbooks from their shell history
and share them with their team.
`

		got, err := io.ReadAll(f)
		assert.NoError(t, err)
		assert.Equal(t, expected, string(got))
	})
}
