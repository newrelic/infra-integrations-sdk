package metric

import (
	"fmt"
	"io/ioutil"
	"path"
	"testing"
	"time"

	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/persist"
	"github.com/stretchr/testify/assert"
)

type FakeData struct {
	timestamp time.Time
}

func (fd *FakeData) Now() time.Time {
	if fd.timestamp.IsZero() {
		fd.timestamp = time.Now()
	} else {
		fd.timestamp = fd.timestamp.Add(1 * time.Second)
	}
	return fd.timestamp
}

var rateAndDeltaTests = []struct {
	testCase string
	key      string
	value    interface{}
	out      interface{}
	cache    interface{}
}{
	{"1st data in key", "key1", .22323333, 0.0, 0.22323333},
	{"1st data in key", "key2", 100, 0.0, 100.0},
	{"2nd data in key", "key2", 110, 10.0, 110.0},
}

func TestSet_SetMetricGauge(t *testing.T) {
	fd := FakeData{}
	persist.SetNow(fd.Now)

	ms, err := NewSet("some-event-type", nil)
	assert.NoError(t, err)

	assert.NoError(t, ms.SetMetric("key", 10, GAUGE))

	assert.Equal(t, 10.0, ms.Metrics["key"], "stored gauge should be float")
}

func TestSet_SetMetricAttribute(t *testing.T) {
	fd := FakeData{}
	persist.SetNow(fd.Now)

	ms, err := NewSet("some-event-type", nil)
	assert.NoError(t, err)

	ms.SetMetric("key", "some-attribute", ATTRIBUTE)

	if ms.Metrics["key"] != "some-attribute" {
		t.Errorf("metric stored not valid: %v", ms.Metrics["key"])
	}
}

func TestSet_SetMetricCachesRateAndDeltas(t *testing.T) {
	storer := persist.NewInMemoryStore()

	fd := FakeData{}
	for _, sourceType := range []SourceType{DELTA, RATE} {
		persist.SetNow(fd.Now)

		ms, err := NewSet("some-event-type", storer)
		assert.NoError(t, err)

		for _, tt := range rateAndDeltaTests {
			// user should not store different types under the same key
			key := fmt.Sprintf("%s:%d", tt.key, sourceType)
			ms.SetMetric(key, tt.value, sourceType)

			if ms.Metrics[key] != tt.out {
				t.Errorf("setting %s %s source-type %d and value %v returned: %v, expected: %v",
					tt.testCase, tt.key, sourceType, tt.value, ms.Metrics[key], tt.out)
			}

			var v interface{}
			_, err := storer.Get(key, &v)
			if err == persist.ErrNotFound {
				t.Errorf("key %s not in cache for case %s", tt.key, tt.testCase)
			} else if tt.cache != v {
				t.Errorf("cache.Get(\"%v\") ==> %v, want %v", tt.key, v, tt.cache)
			}
		}
	}
}

func TestSet_SetMetric_NilStorer(t *testing.T) {
	ms, err := NewSet("some-event-type", nil)
	assert.NoError(t, err)

	err = ms.SetMetric("foo", 1, RATE)
	assert.Error(t, err, "integrations built with no-store can't use DELTAs and RATEs")

	err = ms.SetMetric("foo", 1, DELTA)
	assert.Error(t, err, "integrations built with no-store can't use DELTAs and RATEs")
}

func TestSet_SetMetric_IncorrectMetricType(t *testing.T) {
	ms, err := NewSet("some-event-type", persist.NewInMemoryStore())
	assert.NoError(t, err)

	err = ms.SetMetric("foo", "bar", RATE)
	assert.Error(t, err, "non-numeric source type for rate/delta metric foo")

	err = ms.SetMetric("foo", "bar", DELTA)
	assert.Error(t, err, "non-numeric source type for rate/delta metric foo")

	err = ms.SetMetric("foo", "bar", GAUGE)
	assert.Error(t, err, "non-numeric source type for gauge metric foo")

	err = ms.SetMetric("foo", 1, ATTRIBUTE)
	assert.Error(t, err, "non-string source type for attribute foo")

	err = ms.SetMetric("foo", 1, 666)
	assert.Error(t, err, "unknown source type for key foo")

}

func TestSet_MarshalJSON(t *testing.T) {
	ms, err := NewSet("some-event-type", persist.NewInMemoryStore())
	assert.NoError(t, err)

	ms.SetMetric("foo", 1, RATE)
	ms.SetMetric("bar", 1, DELTA)
	ms.SetMetric("baz", 1, GAUGE)
	ms.SetMetric("quux", "bar", ATTRIBUTE)

	marshaled, err := ms.MarshalJSON()

	assert.NoError(t, err)
	assert.Equal(t,
		`{"bar":0,"baz":1,"event_type":"some-event-type","foo":0,"quux":"bar"}`,
		string(marshaled),
	)
}

func TestNewSet_FileStore_StoresBetweenRuns(t *testing.T) {
	fd := FakeData{}
	persist.SetNow(fd.Now)

	storeFile := tempFile()

	s, err := persist.NewFileStore(storeFile, log.Discard, 1*time.Hour)
	assert.NoError(t, err)

	set1, err := NewSet("type", s)
	assert.NoError(t, err)

	assert.NoError(t, set1.SetMetric("foo", 1, DELTA))

	assert.NoError(t, s.Save())

	s2, err := persist.NewFileStore(storeFile, log.Discard, 1*time.Hour)
	assert.NoError(t, err)

	set2, err := NewSet("type", s2)
	assert.NoError(t, err)

	assert.NoError(t, set2.SetMetric("foo", 3, DELTA))

	assert.Equal(t, 2.0, set2.Metrics["foo"])
}

func TestNewSetRelatedTo_AddsAttributes(t *testing.T) {
	fd := FakeData{}
	persist.SetNow(fd.Now)

	storeFile := tempFile()

	// write in same store/integration-run
	storeWrite, err := persist.NewFileStore(storeFile, log.Discard, 1*time.Hour)
	assert.NoError(t, err)

	set, err := NewSet(
		"type",
		storeWrite,
		RelatedTo("pod", "pod-a"),
		RelatedTo("node", "node-a"),
	)
	assert.NoError(t, err)

	assert.Equal(t, "pod-a", set.Metrics["pod"])
	assert.Equal(t, "node-a", set.Metrics["node"])
}

func TestNewSetRelatedTo_SolvesCacheCollision(t *testing.T) {
	fd := FakeData{}
	persist.SetNow(fd.Now)

	storeFile := tempFile()

	// write in same store/integration-run
	storeWrite, err := persist.NewFileStore(storeFile, log.Discard, 1*time.Hour)
	assert.NoError(t, err)

	ms1, err := NewSet("type", storeWrite, RelatedTo("pod", "pod-a"))
	assert.NoError(t, err)
	ms2, err := NewSet("type", storeWrite, RelatedTo("pod", "pod-a"))
	assert.NoError(t, err)
	ms3, err := NewSet("type", storeWrite, RelatedTo("pod", "pod-b"))
	assert.NoError(t, err)

	assert.NoError(t, ms1.SetMetric("field", 1, DELTA))
	assert.NoError(t, ms2.SetMetric("field", 2, DELTA))
	assert.NoError(t, ms3.SetMetric("field", 3, DELTA))

	assert.NoError(t, storeWrite.Save())

	// retrieve from another store/integration-run
	storeRead, err := persist.NewFileStore(storeFile, log.Discard, 1*time.Hour)
	assert.NoError(t, err)

	msRead, err := NewSet("type", storeRead, RelatedTo("pod", "pod-a"))
	assert.NoError(t, err)

	// write is required to make data available for read
	assert.NoError(t, msRead.SetMetric("field", 10, DELTA))

	assert.Equal(t, 8.0, msRead.Metrics["field"], "read metrics: %+v", msRead.Metrics)
}

func tempFile() string {
	dir, err := ioutil.TempDir("", "file_store")
	if err != nil {
		panic(err)
	}

	return path.Join(dir, "test.json")
}
