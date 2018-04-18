package persist

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

// SetNow forces a different "current time" for the Storer.
// This function is useful only for unit testing.
func SetNow(newNow func() time.Time) {
	now = newNow
}

const ttl = 1 * time.Minute

// Storer is a key-value structure that is initialized and stored in a persistent device.
// It also saves the timestamp when a key was stored.
type Storer interface {
	// Save persists all the data in the Storer.
	Save() error
	// Get looks for a key in the Storer and returns its value together with the
	// timestamp of when it was last set. The third returned value indicates whether
	// the key has been found or not.
	Get(name string) (float64, int64, bool)
	// Set adds a value into the Storer and with the current timestamp. The data is not persisted until the Save()
	// function is invoked.
	Set(name string, value float64) int64
}

type storer struct {
	path       string
	Data       map[string]interface{}
	Timestamps map[string]int64
}

// NewStorer will create and initialize a disk-backed Storer object.
func NewStorer(storePath string, l log.Logger) (Storer, error) {
	store := &storer{
		Data:       make(map[string]interface{}),
		Timestamps: make(map[string]int64),
	}

	store.path = storePath

	// Create the external directory for user-generated json
	storeDir := filepath.Dir(store.path)
	if _, err := os.Stat(storeDir); err != nil {
		if err = os.MkdirAll(storeDir, 0755); err != nil {
			return nil, fmt.Errorf("store directory in %s could not be created", storeDir)
		}
	}

	stat, err := os.Stat(store.path)
	// Store file doesn't exist yet
	if err != nil {
		return store, nil
	}

	if now().Sub(stat.ModTime()) > ttl {
		l.Infof("store file (%s) is older than %v, skipping loading from disk.", storePath, ttl)
		return store, nil
	}

	file, err := ioutil.ReadFile(store.path)
	if err != nil {
		l.Infof("store file (%s) cannot be open for reading.", storePath)
		return store, nil
	}

	// Ignoring unmarshalling errors, returning a clean store
	json.Unmarshal(file, &store) // nolint: errcheck

	return store, nil
}

// Save persists all the data in the Storer.
func (c *storer) Save() error {
	if c.path == "" {
		return nil
	}

	data, err := json.Marshal(c)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(c.path, data, 0644)
}

// Get looks for a key in the store and returns its value together with the
// timestamp of when it was last set. The third returned value indicates whether
// the key has been found or not.
func (c *storer) Get(name string) (float64, int64, bool) {
	val, ok := c.Data[name]
	if ok {
		ts, ok := c.Timestamps[name]
		if ok {
			return val.(float64), int64(ts), ok
		}
	}
	return 0, 0, false
}

// Set adds a value into the store and it also stores the current timestamp.
func (c *storer) Set(name string, value float64) int64 {
	c.Data[name] = value
	c.Timestamps[name] = now().Unix()
	return c.Timestamps[name]
}
