package metric_test

import (
	"testing"
	"time"

	"fmt"

	"github.com/newrelic/infra-integrations-sdk/metric"
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

	ms, err := metric.NewSet("some-event-type", nil)
	assert.NoError(t, err)

	ms.SetMetric("key", 10, metric.GAUGE)

	if ms.Metrics["key"] != 10 {
		t.Errorf("metric stored not valid: %v", ms.Metrics["key"])
	}
}

func TestSet_SetMetricAttribute(t *testing.T) {
	fd := FakeData{}
	persist.SetNow(fd.Now)

	ms, err := metric.NewSet("some-event-type", nil)
	assert.NoError(t, err)

	ms.SetMetric("key", "some-attribute", metric.ATTRIBUTE)

	if ms.Metrics["key"] != "some-attribute" {
		t.Errorf("metric stored not valid: %v", ms.Metrics["key"])
	}
}

func TestSet_SetMetricCachesRateAndDeltas(t *testing.T) {
	storer := persist.NewInMemoryStore()

	fd := FakeData{}
	for _, sourceType := range []metric.SourceType{metric.DELTA, metric.RATE} {
		persist.SetNow(fd.Now)

		ms, err := metric.NewSet("some-event-type", storer)
		assert.NoError(t, err)

		for _, tt := range rateAndDeltaTests {
			// user should not store different types under the same key
			key := fmt.Sprintf("%s:%d", tt.key, sourceType)
			ms.SetMetric(key, tt.value, sourceType)

			if ms.Metrics[key] != tt.out {
				t.Errorf("setting %s %s source-type %d and value %v returned: %v, expected: %v",
					tt.testCase, tt.key, sourceType, tt.value, ms.Metrics[tt.key], tt.out)
			}

			v, _, ok := storer.Get(key)
			if !ok {
				t.Errorf("key %s not in cache for case %s", tt.key, tt.testCase)
			} else if tt.cache != v {
				t.Errorf("cache.Get(\"%v\") ==> %v, want %v", tt.key, v, tt.cache)
			}
		}
	}
}
