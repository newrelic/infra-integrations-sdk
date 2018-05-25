package metric

import (
	"fmt"
	"testing"
	"time"

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
					tt.testCase, tt.key, sourceType, tt.value, ms.Metrics[tt.key], tt.out)
			}

			v, _, err := storer.Get(key)
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
	storer := persist.NewInMemoryStore()

	ms, err := NewSet("some-event-type", storer)
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
