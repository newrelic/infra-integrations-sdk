package cache_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/newrelic/infra-integrations-sdk/cache"
	"github.com/stretchr/testify/assert"
)

func TestDiskCache(t *testing.T) {
	file, err := ioutil.TempFile("", "cache")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	// Create cache with existing file in env
	_, err = cache.NewCache(file.Name(), cache.GlobalLog)
	assert.NoError(t, err)

	// Create cache with unexisting file in env
	tmpDir, err := ioutil.TempDir("", "cache-test")
	assert.NoError(t, err)
	_, err = cache.NewCache(filepath.Join(tmpDir, "newfile.json"), cache.GlobalLog)
	assert.NoError(t, err)
}

func TestCacheSet(t *testing.T) {
	file, err := ioutil.TempFile("", "cache.json")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	dc, err := cache.NewCache(file.Name(), cache.GlobalLog)

	assert.NoError(t, err)

	dc.Set("key", float64(100))
	v, ts, ok := dc.Get("key")
	assert.True(t, ok)
	assert.InDelta(t, float64(100), v, 0.1)
	assert.NotEqual(t, 0, ts)
}

func TestCacheGet(t *testing.T) {
	file, err := ioutil.TempFile("", "cache.json")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	dc, err := cache.NewCache(file.Name(), cache.GlobalLog)

	assert.NoError(t, err)

	dc.Set("key1", float64(100))

	value, ts, exists := dc.Get("key1")

	if value != float64(100) {
		t.Error()
	}
	if ts == int64(0) {
		t.Error()
	}
	if !exists {
		t.Error()
	}

	value, ts, exists = dc.Get("key2")

	if value != float64(0) {
		t.Error()
	}
	if ts != int64(0) {
		t.Error()
	}
	if exists {
		t.Error()
	}
}

func TestCacheSave(t *testing.T) {

	file, err := ioutil.TempFile("", "cache.json")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	dc, err := cache.NewCache(file.Name(), cache.GlobalLog)

	assert.NoError(t, err)

	dc.Set("key1", float64(100))
	dc.Set("key2", float64(200))

	err = dc.Save()
	assert.NoError(t, err)

	dc, err = cache.NewCache(file.Name(), cache.GlobalLog)
	assert.NoError(t, err)

	value, ts, exists := dc.Get("key1")

	assert.True(t, exists)
	assert.InDelta(t, float64(100), value, 0.1)
	assert.NotEqual(t, 0, ts)

	value, ts, exists = dc.Get("key2")

	assert.True(t, exists)
	assert.InDelta(t, float64(200), value, 0.1)
	assert.NotEqual(t, 0, ts)
}
