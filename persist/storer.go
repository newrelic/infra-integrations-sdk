package persist

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"time"

	"github.com/newrelic/infra-integrations-sdk/log"
)

const (
	// DefaultTTL specifies the "Time To Live" of the disk storage.
	DefaultTTL      = 1 * time.Minute
	filePerm        = 0644
	dirFilePerm     = 0755
	integrationsDir = "nr-integrations"
)

var (
	// ErrNotFound defines an error that will be returned when trying to access a storage entry that can't be found
	ErrNotFound = errors.New("key not found")
)

var now = time.Now

// Storer defines the interface of a Key-Value storage system, which is able to store the timestamp
// where the key was stored.
type Storer interface {
	// Set associates a value with a given key. Implementors must save also the time when it has been stored and return it.
	// The value can be any type.
	Set(key string, value interface{}) int64
	// Get gets the value associated to a given key and stores in the value referenced by the pointer passed as argument.
	// It returns the Unix timestamp when the value was stored (in seconds), or an error if the Get operation failed.
	// It may return any type of value.
	Get(key string, valuePtr interface{}) (int64, error)
	// Delete removes the cached data for the given key. If the data does not exist, the system does not return
	// any error.
	Delete(key string) error
	// Save persists all the data in the storer.
	Save() error
}

// In-memory implementation of the storer
type inMemoryStore struct {
	cachedData map[string]jsonEntry
	Data       map[string][]byte
}

// Holder for any entry in the JSON storage
type jsonEntry struct {
	Timestamp int64
	Value     interface{}
}

// fileStore is a Storer implementation that uses the file system as persistence backend, storing
// the objects as JSON.
// This requires that any object that has to be stored is Marshallable and Unmarshallable.
type fileStore struct {
	inMemoryStore
	path string
	ilog log.Logger
}

// SetNow forces a different "current time" for the Storer.
// Deprecated: This function is useful only for unit testing outside the persist package.
func SetNow(newNow func() time.Time) {
	now = newNow
}

// DefaultPath returns a default folder/filename path to a Storer for an integration from the given name. The name of
// the file will be the name of the integration with the .json extension.
func DefaultPath(integrationName string) string {
	dir := filepath.Join(os.TempDir(), integrationsDir)
	baseDir := path.Join(dir, integrationName+".json")
	// Create integrations Storer directory
	if os.MkdirAll(dir, dirFilePerm) != nil {
		baseDir = os.TempDir()
	}
	return baseDir
}

// NewInMemoryStore will create and initialize an in-memory Storer (not persistent).
func NewInMemoryStore() Storer {
	return &inMemoryStore{
		cachedData: make(map[string]jsonEntry),
		Data:       make(map[string][]byte),
	}
}

// NewFileStore returns a disk-backed Storer using the provided file path
func NewFileStore(storagePath string, ilog log.Logger, ttl time.Duration) (Storer, error) {
	ms := NewInMemoryStore().(*inMemoryStore)

	store := fileStore{
		path:          storagePath,
		ilog:          ilog,
		inMemoryStore: *ms,
	}

	if stat, err := os.Stat(storagePath); err == nil {
		if now().Sub(stat.ModTime()) > ttl {
			ilog.Debugf("store file (%s) is older than %v, skipping loading from disk.", storagePath, ttl)
			return &store, nil
		}

		err := store.loadFromDisk()
		if err != nil {
			ilog.Debugf(err.Error())
		}
	} else if os.IsNotExist(err) {
		folder := path.Dir(storagePath)
		err := os.MkdirAll(folder, dirFilePerm)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	return &store, nil
}

func (j *fileStore) Save() error {
	// An in-memory implementation does nothing
	err := j.flushCache()
	if err != nil {
		return err
	}

	bytes, err := json.Marshal(j)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(j.path, bytes, filePerm)
}

func (j *inMemoryStore) Save() error {
	// An in-memory implementation does nothing
	return nil
}

// Set stores a value for a given key. Implementors must save also the time when it was stored.
// This implementation adds a restriction to the key name: it must be a valid file name (without extension).
func (j inMemoryStore) Set(key string, value interface{}) int64 {
	ts := now().Unix()
	j.cachedData[key] = jsonEntry{
		Timestamp: ts,
		Value:     value,
	}
	return ts
}

// Get gets the value associated to a given key and stores it in the value referenced by the pointer passed as
// second argument
// This implementation adds a restriction to the key name: it must be a valid file name (without extension).
func (j inMemoryStore) Get(key string, valuePtr interface{}) (int64, error) {
	rv := reflect.ValueOf(valuePtr)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return 0, errors.New("destination argument must be a pointer")
	}

	entry, ok := j.cachedData[key]

	// If the entry is not cached, it may be stored as a JSON (as loaded from disk), and we unmarshall it
	if !ok {
		bytes, ok := j.Data[key]
		if !ok {
			return 0, ErrNotFound
		}

		entry = jsonEntry{}
		entry.Value = valuePtr
		err := json.Unmarshal(bytes, &entry)
		if err != nil {
			return 0, err
		}
		j.cachedData[key] = entry
	}

	// Using reflection to indirectly set the value passed as reference
	reflect.Indirect(rv).Set(reflect.Indirect(reflect.ValueOf(entry.Value)))

	return entry.Timestamp, nil
}

// flushCache marshalls all the cached data into JSON, ready to be stored into disk
func (j *inMemoryStore) flushCache() error {
	for k, v := range j.cachedData {
		bytes, err := json.Marshal(v)
		if err != nil {
			return err
		}
		j.Data[k] = bytes
	}
	j.cachedData = make(map[string]jsonEntry)
	return nil
}

func (j *fileStore) loadFromDisk() error {
	bytes, err := ioutil.ReadFile(j.path)
	if err != nil {
		return fmt.Errorf("can't read %q: %s. Ignoring", j.path, err.Error())
	}
	err = json.Unmarshal(bytes, j)
	if err != nil {
		return fmt.Errorf("can't unmarshall %q: %s. Ignoring", j.path, err.Error())
	}
	return nil
}

// Delete removes the cached data for the given key. If the data does not exist, the system does not return
// any error.
func (j inMemoryStore) Delete(key string) error {
	delete(j.cachedData, key)
	delete(j.Data, key)
	return nil
}
