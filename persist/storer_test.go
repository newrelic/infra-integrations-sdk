package persist

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/stretchr/testify/assert"
)

// setNow forces a different "current time" for the storage.
func setNow(newNow func() time.Time) {
	now = newNow
}

type storerProvider interface {
	new() (Storer, error)
	prepareToRead(s Storer) (Storer, error)
}

type memoryStorerProvider struct{}

func (m *memoryStorerProvider) new() (Storer, error) {
	return NewInMemoryStore(), nil
}

func (m *memoryStorerProvider) prepareToRead(s Storer) (Storer, error) {
	return s, nil
}

type diskStorerProvider struct {
	filePath string
}

func (j *diskStorerProvider) new() (Storer, error) {
	if j.filePath == "" {
		j.filePath = filePath()
	}

	return NewFileStore(j.filePath, log.NewStdErr(true), DefaultTTL)
}

func filePath() string {
	rootDir, err := ioutil.TempDir("", "disk_storage")
	if err != nil {
		panic(err)
	}

	return path.Join(rootDir, "storage.json")
}

func (j *diskStorerProvider) prepareToRead(s Storer) (Storer, error) {
	s.Save()

	return j.new()
}

func getStorerProviders() []storerProvider {
	return []storerProvider{&memoryStorerProvider{}, &diskStorerProvider{}}
}

func TestStorer_Struct(t *testing.T) {
	nowTime := time.Now()
	setNow(func() time.Time {
		return nowTime
	})

	for _, provider := range getStorerProviders() {
		storer, err := provider.new()
		assert.NoError(t, err)

		t.Run(reflect.TypeOf(storer).Name(), func(t *testing.T) {
			// Given a Storer implementation

			// And a stored struct value
			type testStruct struct {
				FloatVal  float64
				StringVal string
				MapVal    map[string]interface{}
				StructVal struct {
					A float64
					B string
				}
			}
			stored := testStruct{
				1, "2",
				map[string]interface{}{"hello": "how are you", "fine": "and you?"},
				struct {
					A float64
					B string
				}{11, "22"},
			}
			ts := storer.Set("my-storage-test", stored)
			assert.Nil(t, err)
			assert.Equal(t, nowTime.Unix(), ts)

			storer, err = provider.prepareToRead(storer)
			assert.NoError(t, err)

			var read testStruct
			// When reading it from the disk
			ts, err = storer.Get("my-storage-test", &read)

			assert.Equal(t, stored, read)

			// As well as the insertion timestamp
			assert.Equal(t, nowTime.Unix(), ts)
			assert.Nil(t, err)
		})
	}
}

func TestStorer_Map(t *testing.T) {
	nowTime := time.Now()
	setNow(func() time.Time {
		return nowTime
	})

	for _, provider := range getStorerProviders() {
		storer, err := provider.new()
		assert.NoError(t, err)

		t.Run(reflect.TypeOf(storer).Name(), func(t *testing.T) {
			// Given a Storer implementation

			// And a stored map
			stored := map[string]interface{}{
				"1": "2",
				"3": map[string]interface{}{"hello": "how are you", "fine": "and you?"},
				"4": 5.0,
			}
			ts := storer.Set("my-storage-test", stored)
			assert.Nil(t, err)
			assert.Equal(t, nowTime.Unix(), ts)

			storer, err = provider.prepareToRead(storer)
			assert.NoError(t, err)

			// When reading it from the disk
			var read map[string]interface{}
			ts, err = storer.Get("my-storage-test", &read)

			// An map equal to the original has been returned
			assert.Equal(t, stored, read)

			// As well as the insertion timestamp
			assert.Equal(t, nowTime.Unix(), ts)
			assert.Nil(t, err)
		})
	}
}

func TestStorer_Array(t *testing.T) {
	nowTime := time.Now()
	setNow(func() time.Time {
		return nowTime
	})

	for _, provider := range getStorerProviders() {
		storer, err := provider.new()
		assert.NoError(t, err)

		t.Run(reflect.TypeOf(storer).Name(), func(t *testing.T) {
			// Given a Storer implementation

			// And a stored array
			stored := []interface{}{"1", 2.0, "3", map[string]interface{}{"hello": "how are you", "fine": "and you?"}}
			ts := storer.Set("my-storage-test", stored)
			assert.Nil(t, err)
			assert.Equal(t, nowTime.Unix(), ts)

			storer, err = provider.prepareToRead(storer)
			assert.NoError(t, err)

			// When reading it from the disk
			var read []interface{}
			ts, err = storer.Get("my-storage-test", &read)

			// It returns an array equal to the original
			assert.Equal(t, stored, read)

			// As well as the insertion timestamp
			assert.Equal(t, nowTime.Unix(), ts)
			assert.Nil(t, err)
		})
	}
}

func TestStorer_String(t *testing.T) {

	nowTime := time.Now()
	setNow(func() time.Time {
		return nowTime
	})

	for _, provider := range getStorerProviders() {
		storer, err := provider.new()
		assert.NoError(t, err)

		t.Run(reflect.TypeOf(storer).Name(), func(t *testing.T) {
			// Given a Storer implementation

			// And a stored string
			stored := "hello my good friend"
			ts := storer.Set("my-storage-test", stored)
			assert.Nil(t, err)
			assert.Equal(t, nowTime.Unix(), ts)

			storer, err = provider.prepareToRead(storer)
			assert.NoError(t, err)

			// When reading it from the disk
			var read string
			ts, err = storer.Get("my-storage-test", &read)

			// It returns a string equal to the original
			assert.Equal(t, stored, read)

			// As well as the insertion timestamp
			assert.Equal(t, nowTime.Unix(), ts)
			assert.Nil(t, err)
		})
	}
}

func TestStorer_Number(t *testing.T) {

	nowTime := time.Now()
	setNow(func() time.Time {
		return nowTime
	})

	for _, provider := range getStorerProviders() {
		storer, err := provider.new()
		assert.NoError(t, err)

		t.Run(reflect.TypeOf(storer).Name(), func(t *testing.T) {
			// Given a Storer implementation

			// And a stored integer
			stored := int(123456)
			ts := storer.Set("my-storage-test", stored)
			assert.Equal(t, nowTime.Unix(), ts)

			storer, err = provider.prepareToRead(storer)
			assert.NoError(t, err)

			// When reading it from the disk
			var read int
			ts, err = storer.Get("my-storage-test", &read)

			// It returns the copy of the original number
			assert.Equal(t, stored, read)

			// As well as the insertion timestamp
			assert.Equal(t, nowTime.Unix(), ts)
			assert.Nil(t, err)
		})
	}
}

func TestStorer_Overwrite(t *testing.T) {
	for _, provider := range getStorerProviders() {
		storer, err := provider.new()
		assert.NoError(t, err)

		t.Run(reflect.TypeOf(storer).Name(), func(t *testing.T) {
			// Given a Storer implementation

			// And a stored record
			nowTime := time.Unix(1234, 5678)
			setNow(func() time.Time {
				return nowTime
			})
			ts := storer.Set("my-storage-test", "initial Value")
			assert.Equal(t, nowTime.Unix(), ts)

			nowTime = time.Unix(78910, 111213)
			// When this record is overwritten
			ts = storer.Set("my-storage-test", "overwritten value")
			assert.Equal(t, nowTime.Unix(), ts)

			storer, err = provider.prepareToRead(storer)
			assert.NoError(t, err)

			// The read operation returns the last version of the record
			var read string
			storer.Get("my-storage-test", &read)
			assert.Equal(t, "overwritten value", read)
		})
	}
}

func TestStorer_NotFound(t *testing.T) {
	for _, provider := range getStorerProviders() {
		storer, err := provider.new()
		assert.NoError(t, err)

		t.Run(reflect.TypeOf(storer).Name(), func(t *testing.T) {
			// Given a Storer implementation

			storer, err = provider.prepareToRead(storer)
			assert.NoError(t, err)

			// When trying to access an nonexistent record
			var read string
			_, err := storer.Get("my-storage-test", &read)

			// The storage returns an ErrNotFound error
			assert.Equal(t, ErrNotFound, err)
		})
	}
}

func TestStorer_Delete(t *testing.T) {
	for _, provider := range getStorerProviders() {
		storer, err := provider.new()
		assert.NoError(t, err)

		t.Run(reflect.TypeOf(storer).Name(), func(t *testing.T) {
			// Given a Storer implementation

			// And a stored record
			storer.Set("my-storage-test", "initial Value")
			// When removing the stored record
			assert.Nil(t, storer.Delete("my-storage-test"))

			storer, err = provider.prepareToRead(storer)
			assert.NoError(t, err)

			// When trying to access an nonexistent record
			var read string
			_, err = storer.Get("my-storage-test", &read)

			// The storage returns an ErrNotFound error
			assert.Equal(t, ErrNotFound, err)
		})
	}
}

func TestStorer_DeleteUnexistent(t *testing.T) {
	for _, provider := range getStorerProviders() {
		storer, err := provider.new()
		assert.NoError(t, err)

		t.Run(reflect.TypeOf(storer).Name(), func(t *testing.T) {
			// Given a Storer implementation

			// When trying to remove a non-existing record
			err := storer.Delete("my-storage-test")

			// The storage does not return any error
			assert.Nil(t, err)
		})
	}
}

func TestFileStorer_Save(t *testing.T) {
	nowTime := time.Now()
	setNow(func() time.Time {
		return nowTime
	})

	// Given a file storer
	dir, err := ioutil.TempDir("", "filestorer_save")
	assert.NoError(t, err)
	filePath := path.Join(dir, "test.json")
	storer, err := NewFileStore(filePath, log.NewStdErr(true), DefaultTTL)
	assert.NoError(t, err)

	type testStruct struct {
		A float64
		B float64
	}

	// When something is Saved
	storer.Set("stringValue", "hello my friend")
	storer.Set("arrayValue", []float64{0, 1, 2, 3, 4})
	storer.Set("floatValue", 3)
	storer.Set("deletedValue", "this won't be persisted")
	storer.Delete("deletedValue")

	stored := testStruct{555, 444}
	storer.Set("structValue", stored)

	storer.Save()

	// And a new storer opens the file
	storer, err = NewFileStore(filePath, log.NewStdErr(true), DefaultTTL)
	assert.NoError(t, err)

	// The data is persisted as expected
	var stringValue string
	_, err = storer.Get("stringValue", &stringValue)
	assert.NoError(t, err)

	_, err = storer.Get("deletedValue", &stringValue)
	assert.Error(t, err)

	// (int arrays are unmarshalled as float arrays)
	arrayValue := make([]interface{}, 0, 0)
	_, err = storer.Get("arrayValue", &arrayValue)
	assert.NoError(t, err)
	for i, v := range arrayValue {
		assert.InDelta(t, float64(i), v, 0.01)
	}

	var floatValue float64
	_, err = storer.Get("floatValue", &floatValue)
	assert.NoError(t, err)
	assert.InDelta(t, 3, floatValue, 0.01)

	var structValue testStruct
	_, err = storer.Get("structValue", &structValue)
	assert.NoError(t, err)
	assert.Equal(t, testStruct{555, 444}, structValue)

}

func TestInMemoryStore_flushCache(t *testing.T) {
	nowTime := time.Now()
	setNow(func() time.Time {
		return nowTime
	})

	s := NewInMemoryStore().(*inMemoryStore)

	_ = s.Set("k", "v")

	assert.NoError(t, s.flushCache())

	assert.Equal(t, fmt.Sprintf("{\"Timestamp\":%d,\"Value\":\"v\"}", nowTime.Unix()), string(s.Data["k"]))
}

// Exaplained through different deserialization approaches
func TestFileStore_Save(t *testing.T) {
	nowTime := time.Now()
	setNow(func() time.Time {
		return nowTime
	})

	expectedTS := nowTime.Unix()

	storeProvider := diskStorerProvider{}
	s, err := storeProvider.new()
	assert.NoError(t, err)

	_ = s.Set("k", "v")

	// assertion 1: using diskStorerProvider
	s, err = storeProvider.prepareToRead(s)
	assert.NoError(t, err)

	var val string
	ts, err := s.Get("k", &val)

	assert.NoError(t, err)
	assert.Equal(t, "v", val)
	assert.Equal(t, expectedTS, ts)

	// reading file contents
	readStore, err := ioutil.ReadFile(storeProvider.filePath)
	assert.NoError(t, err)

	// assertion 2.1: using a store with value deserialization on demand by Get
	unserializedStore := NewInMemoryStore()
	json.Unmarshal(readStore, &unserializedStore)

	var v string
	ts, err = unserializedStore.Get("k", &v)
	assert.NoError(t, err)
	assert.Equal(t, "v", v)
	assert.Equal(t, expectedTS, ts)

	// assertion 2.2: manual deserialization
	expectedContent := fmt.Sprintf("{\"Data\":{ \"k\": { \"Timestamp\":%d, \"Value\":\"v\" } } }", nowTime.Unix())
	var readJSON map[string]map[string][]byte
	json.Unmarshal(readStore, &readJSON)
	var expectedJSON map[string]map[string][]byte
	json.Unmarshal([]byte(expectedContent), &expectedJSON)

	bytes, ok := readJSON["Data"]["k"]
	assert.True(t, ok)
	var entry jsonEntry
	err = json.Unmarshal(bytes, &entry)
	assert.NoError(t, err)

	assert.Equal(t, "v", entry.Value)
	assert.Equal(t, expectedTS, entry.Timestamp)

}
