package cache_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/newrelic/infra-integrations-sdk/cache"
	"github.com/stretchr/testify/assert"
)

func TestGlobalCache(t *testing.T) {
	cache.Set("key1234", float64(12345))

	value, ts, exists := cache.Get("key1234")
	assert.InDelta(t, float64(12345), value, 0.1)
	assert.NotEqual(t, 0, ts)
	assert.True(t, exists)
}

func TestGlobalCacheUnexistingKey(t *testing.T) {
	_, _, exists := cache.Get("alsdkfjlaksdjflkasdj")
	assert.False(t, exists)
}

func TestGlobalCacheOverwrite(t *testing.T) {
	cache.Set("key43210", float64(32100))
	cache.Set("key43210", float64(12345))

	value, ts, exists := cache.Get("key43210")
	assert.InDelta(t, float64(12345), value, 0.1)
	assert.NotEqual(t, 0, ts)
	assert.True(t, exists)
}

func TestDefaultPath(t *testing.T) {
	assert.Equal(t, filepath.Join(os.TempDir(), "nr-integrations", "file.json"), cache.DefaultPath("file"))
}
