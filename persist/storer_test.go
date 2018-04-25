package persist_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/persist"
	"github.com/stretchr/testify/assert"
)

func TestDefaultPath(t *testing.T) {
	assert.Equal(t, filepath.Join(os.TempDir(), "nr-integrations", "file.json"), persist.DefaultPath("file"))
}

func TestDiskStorer(t *testing.T) {
	file, err := ioutil.TempFile("", "cache")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	// Create Storer with existing file in env
	_, err = persist.NewFileStore(file.Name(), log.Discard)
	assert.NoError(t, err)

	// Create Storer with unexisting file in env
	tmpDir, err := ioutil.TempDir("", "cache-test")
	assert.NoError(t, err)
	_, err = persist.NewFileStore(filepath.Join(tmpDir, "newfile.json"), log.Discard)
	assert.NoError(t, err)
}

func TestStorerSet(t *testing.T) {
	file, err := ioutil.TempFile("", "cache.json")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	dc, err := persist.NewFileStore(file.Name(), log.Discard)

	assert.NoError(t, err)

	dc.Set("key", float64(100))
	v, ts, ok := dc.Get("key")
	assert.True(t, ok)
	assert.InDelta(t, float64(100), v, 0.1)
	assert.NotEqual(t, 0, ts)
}

func TestStorerGet(t *testing.T) {
	file, err := ioutil.TempFile("", "cache.json")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	dc, err := persist.NewFileStore(file.Name(), log.Discard)

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

func TestStorerSave(t *testing.T) {

	file, err := ioutil.TempFile("", "cache.json")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	dc, err := persist.NewFileStore(file.Name(), log.Discard)

	assert.NoError(t, err)

	dc.Set("key1", float64(100))
	dc.Set("key2", float64(200))

	err = dc.Save()
	assert.NoError(t, err)

	dc, err = persist.NewFileStore(file.Name(), log.Discard)
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
