package persist_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"time"

	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/persist"
	"github.com/stretchr/testify/assert"
)

type FakeData struct {
	timestamp time.Time
}

func (fd *FakeData) DummyTime() time.Time {
	fd.timestamp = time.Unix(1, 1)
	return fd.timestamp
}

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
	defer os.Remove(tmpDir)

	_, err = persist.NewFileStore(filepath.Join(tmpDir, "newfile.json"), log.Discard)
	assert.NoError(t, err)
}

func TestNewFileStoreReturnsErrorOnForbiddenDirectory(t *testing.T) {
	_, err := persist.NewFileStore("/forbidden-directory", log.Discard)
	assert.Error(t, err)
}

func TestNewFileStoreReturnsErrorOnForbiddenFilePath(t *testing.T) {
	_, err := persist.NewFileStore("/forbidden-directory/file.json", log.Discard)
	assert.Error(t, err)
}

func TestStorerSet(t *testing.T) {
	file, err := ioutil.TempFile("", "cache.json")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	dc, err := persist.NewFileStore(file.Name(), log.Discard)
	assert.NoError(t, err)

	dc.Set("key", float64(100))
	v, ts, err := dc.Get("key")
	assert.NoError(t, err)
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

	value, ts, err := dc.Get("key1")
	assert.NoError(t, err)

	if value != float64(100) {
		t.Error()
	}
	if ts == int64(0) {
		t.Error()
	}

	value, ts, err = dc.Get("key2")
	assert.Equal(t, persist.ErrNotFound, err)

	if value != float64(0) {
		t.Error()
	}
	if ts != int64(0) {
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

	value, ts, err := dc.Get("key1")
	assert.NoError(t, err)
	assert.InDelta(t, float64(100), value, 0.1)
	assert.NotEqual(t, 0, ts)

	value, ts, err = dc.Get("key2")
	assert.NoError(t, err)
	assert.InDelta(t, float64(200), value, 0.1)
	assert.NotEqual(t, 0, ts)
}

func TestNewFileStoreIsNotPopulatedWhenModTimeGreaterThanTTL(t *testing.T) {
	ft := time.Now()

	x := func() time.Time {
		if ft.IsZero() {
			return time.Now()
		}
		return ft.Add(24 * time.Hour)
	}

	persist.SetNow(x)

	file, err := ioutil.TempFile("", "cache.json")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	dc, err := persist.NewFileStore(file.Name(), log.Discard)
	assert.NoError(t, err)

	dc.Set("key1", float64(100))

	assert.NoError(t, dc.Save())

	dc, err = persist.NewFileStore(file.Name(), log.Discard)
	assert.NoError(t, err)

	_, _, err = dc.Get("key1")
	assert.Equal(t, persist.ErrNotFound, err)
}

func TestSetNow(t *testing.T) {
	fd := FakeData{timestamp: time.Unix(1, 1)}
	persist.SetNow(fd.DummyTime)

	assert.Equal(t, fd.timestamp, time.Unix(1, 1))
}

func TestInMemoryStore_SaveDoesNothing(t *testing.T) {
	s := persist.NewInMemoryStore()
	assert.NoError(t, s.Save())
}

func TestInMemoryStore_Delete(t *testing.T) {
	s := persist.NewInMemoryStore()

	_ = s.Set("key", 1)

	_, _, err := s.Get("key")
	assert.NoError(t, err)

	s.Delete("key")

	_, _, err = s.Get("key")
	assert.Equal(t, persist.ErrNotFound, err)
}
