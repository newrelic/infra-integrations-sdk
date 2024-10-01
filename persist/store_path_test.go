package persist

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/stretchr/testify/assert"
)

var tmpDir string

func setupTestCase(t *testing.T) func(t *testing.T) {
	t.Log("setup test case")

	assert.NoError(t, os.RemoveAll(filepath.Join(os.TempDir(), integrationsDir)))
	tmpDir = tmpIntegrationDir("")

	files := []struct {
		name    string
		lastMod time.Duration
	}{
		{
			name:    "com.newrelic.fake-a.json",
			lastMod: 1 * time.Second,
		},
		{
			name:    "com.newrelic.fake-b.json",
			lastMod: 80 * time.Second,
		},
		{
			name:    "com.newrelic.fake-c.json",
			lastMod: 80 * time.Second,
		},
		{
			name:    "com.newrelic.flex-b.json",
			lastMod: 80 * time.Second,
		},
	}

	for _, file := range files {
		f, err := os.Create(filepath.Join(tmpDir, file.name))
		assert.NoError(t, err)

		lastChanged := time.Now().Local().Add(-file.lastMod)
		err = os.Chtimes(f.Name(), lastChanged, lastChanged)
		assert.NoError(t, err)
	}

	return func(t *testing.T) {
		t.Log("teardown test case")
		assert.NoError(t, os.RemoveAll(tmpDir))
	}
}

func TestStorePath_CleanOldFiles(t *testing.T) {

	// GIVEN a tmp directory with multiple files
	tearDownFn := setupTestCase(t)
	defer tearDownFn(t)

	// WHEN new store file is generated
	newPath, err := NewStorePath("com.newrelic.fake", "c", "", log.Discard, 1*time.Minute)
	assert.NoError(t, err)

	// THEN only old files with different integration ID are removed
	newPath.CleanOldFiles()

	files, err := filepath.Glob(filepath.Join(tmpDir, "*"))
	assert.NoError(t, err)

	expected := []string{
		filepath.Join(tmpDir, "com.newrelic.fake-a.json"),
		filepath.Join(tmpDir, "com.newrelic.fake-c.json"),
		filepath.Join(tmpDir, "com.newrelic.flex-b.json"),
	}
	assert.Equal(t, expected, files)
}

func TestStorePath_GetFilePath(t *testing.T) {
	cases := []struct {
		tempDir  string
		expected string
	}{
		{
			tempDir:  "",
			expected: filepath.Join(os.TempDir(), integrationsDir, "com.newrelic.fake-c.json"),
		},
		{
			tempDir:  "custom-tmp",
			expected: filepath.Join("custom-tmp", "com.newrelic.fake-c.json"),
		},
	}

	for _, tt := range cases {
		storeFile, err := NewStorePath("com.newrelic.fake", "c", tt.tempDir, log.Discard, 1*time.Minute)
		assert.NoError(t, err)

		assert.Equal(t, tt.expected, storeFile.GetFilePath())
	}
}

func TestStorePath_glob(t *testing.T) {
	storeFile, err := NewStorePath("com.newrelic.fake", "c", "", log.Discard, 1*time.Minute)
	assert.NoError(t, err)

	tmp, ok := storeFile.(*storePath)
	assert.True(t, ok)

	expected := filepath.Join(tmpIntegrationDir(""), "com.newrelic.fake-*.json")
	assert.Equal(t, expected, tmp.glob())
}
