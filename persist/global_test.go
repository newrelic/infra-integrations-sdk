package persist_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/newrelic/infra-integrations-sdk/persist"
	"github.com/stretchr/testify/assert"
)

func TestGlobalStorer(t *testing.T) {
	persist.Set("key1234", float64(12345))

	value, ts, exists := persist.Get("key1234")
	assert.InDelta(t, float64(12345), value, 0.1)
	assert.NotEqual(t, 0, ts)
	assert.True(t, exists)
}

func TestGlobalStorerUnexistingKey(t *testing.T) {
	_, _, exists := persist.Get("alsdkfjlaksdjflkasdj")
	assert.False(t, exists)
}

func TestGlobalStorerOverwrite(t *testing.T) {
	persist.Set("key43210", float64(32100))
	persist.Set("key43210", float64(12345))

	value, ts, exists := persist.Get("key43210")
	assert.InDelta(t, float64(12345), value, 0.1)
	assert.NotEqual(t, 0, ts)
	assert.True(t, exists)
}

func TestDefaultPath(t *testing.T) {
	assert.Equal(t, filepath.Join(os.TempDir(), "nr-integrations", "file.json"), persist.DefaultPath("file"))
}
