package metric_test

import (
	"testing"
	"time"

	"github.com/newrelic/infra-integrations-sdk/cache"
	"github.com/newrelic/infra-integrations-sdk/metric"
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
	{"1st data in key", "rateKey1", 10, 0.0, 10.0},
	{"2nd data in key", "rateKey1", 100, 90.0, 100.0},
	{"1st data in key", "key1", .22323333, 0.0, 0.22323333},
	{"1st data in key", "key2", 100, 0.0, 100.0},
	{"2nd data in key", "key2", 110, 10.0, 110.0},
	//{"key3", 10, metric.DELTA, 0.0, 10.0},
	//{"key3", 110, metric.DELTA, 100.0, 110.0},
}

func TestSet_SetMetricGauge(t *testing.T) {
	fd := FakeData{}
	cache.SetNow(fd.Now)

	ms := metric.NewSet("some-event-type")

	ms.SetMetric("key", 10, metric.GAUGE)

	if ms["key"] != 10 {
		t.Errorf("metric stored not valid: %v", ms["key"])
	}
}

func TestSet_SetMetricAttribute(t *testing.T) {
	fd := FakeData{}
	cache.SetNow(fd.Now)

	ms := metric.NewSet("some-event-type")

	ms.SetMetric("key", "some-attribute", metric.ATTRIBUTE)

	if ms["key"] != "some-attribute" {
		t.Errorf("metric stored not valid: %v", ms["key"])
	}
}

func TestSet_SetMetricCachesRateAndDeltas(t *testing.T) {
	for _, sourceType := range []metric.SourceType{metric.DELTA, metric.RATE} {
		fd := FakeData{}
		cache.SetNow(fd.Now)

		ms := metric.NewSet("some-event-type")

		for _, tt := range rateAndDeltaTests {
			ms.SetMetric(tt.key, tt.value, sourceType)

			if ms[tt.key] != tt.out {
				t.Errorf("setting %s %s source-type %d and value %v returned: %v, expected: %v",
					tt.testCase, tt.key, sourceType, tt.value, ms[tt.key], tt.out)
			}

			v, _, ok := cache.Get(tt.key)
			if !ok {
				t.Errorf("key %s not in cache for case %s", tt.key, tt.testCase)
			} else if tt.cache != v {
				t.Errorf("cache.Get(\"%v\") ==> %v, want %v", tt.key, v, tt.cache)
			}
		}
	}
}
