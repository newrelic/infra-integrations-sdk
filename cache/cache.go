package cache

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/newrelic/infra-integrations-sdk/log"
)

const pathEnvVar = "NRIA_CACHE_PATH"

var now = time.Now

// SetNow forces a different "current time" for the cache.
// This function is useful only for unit testing.
func SetNow(newNow func() time.Time) {
	now = newNow
}

const cacheTTL = 1 * time.Minute

// Cache is a key-value structure that is initialized and stored in a persistent device.
// It also saves the timestamp when a key was stored.
type Cache interface {
	// Save persists all the data in the cache.
	Save() error
	// Get looks for a key in the cache and returns its value together with the
	// timestamp of when it was last set. The third returned value indicates whether
	// the key has been found or not.
	Get(name string) (float64, int64, bool)
	// Set adds a value into the cache and with the current timestamp. The data is not persisted until the Save()
	// function is invoked.
	Set(name string, value float64) int64
}

type cacheImpl struct {
	path       string
	Data       map[string]interface{}
	Timestamps map[string]int64
}

// NewCache will create and initialize a disk-backed cache object.
func NewCache(cachePath string, l log.Logger) (Cache, error) {
	cache := &cacheImpl{
		Data:       make(map[string]interface{}),
		Timestamps: make(map[string]int64),
	}

	cache.path = cachePath

	// Create the external directory for user-generated json
	cacheDir := filepath.Dir(cache.path)
	if _, err := os.Stat(cacheDir); err != nil {
		if err = os.MkdirAll(cacheDir, 0755); err != nil {
			return nil, fmt.Errorf("cache directory in %s could not be created", cacheDir)
		}
	}

	stat, err := os.Stat(cache.path)
	// Cache file doesn't exist yet
	if err != nil {
		return cache, nil
	}

	if now().Sub(stat.ModTime()) > cacheTTL {
		l.Infof("cache file (%s) is older than %v, skipping loading from disk.", cachePath, cacheTTL)
		return cache, nil
	}

	file, err := ioutil.ReadFile(cache.path)
	if err != nil {
		l.Infof("cache file (%s) cannot be open for reading.", cachePath)
		return cache, nil
	}

	// Ignoring unmarshalling errors, returning a clean cache
	json.Unmarshal(file, &cache) // nolint: errcheck

	return cache, nil
}

// Save persists all the data in the cache.
func (cache *cacheImpl) Save() error {
	if cache.path == "" {
		return nil
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(cache.path, data, 0644)
}

// Get looks for a key in the cache and returns its value together with the
// timestamp of when it was last set. The third returned value indicates whether
// the key has been found or not.
func (cache *cacheImpl) Get(name string) (float64, int64, bool) {
	val, ok := cache.Data[name]
	if ok {
		ts, ok := cache.Timestamps[name]
		if ok {
			return val.(float64), int64(ts), ok
		}
	}
	return 0, 0, false
}

// Set adds a value into the cache and it also stores the current timestamp.
func (cache *cacheImpl) Set(name string, value float64) int64 {
	cache.Data[name] = value
	cache.Timestamps[name] = now().Unix()
	return cache.Timestamps[name]
}
