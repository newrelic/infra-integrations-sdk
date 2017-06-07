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

var metricTests = []struct {
	key        string
	value      interface{}
	metricType metric.SourceType
	out        interface{}
	cache      interface{}
}{
	{"gaugeKey", 10, metric.GAUGE, 10, nil},
	{"keyAtribute", "sadad", metric.ATTRIBUTE, "sadad", nil},
	{"counterKey1", 10, metric.COUNTER, 0.0, 10.0},
	{"counterKey1", 100, metric.COUNTER, 90.0, 100.0},
	{"key1", .22323333, metric.COUNTER, 0.0, 0.22323333},
	{"key2", 100, metric.COUNTER, 0.0, 100.0},
	{"key2", 110, metric.COUNTER, 10.0, 110.0},
	{"key3", 10, metric.DELTA, 0.0, 10.0},
	{"key3", 110, metric.DELTA, 100.0, 110.0},
}

func TestSetMetric(t *testing.T) {
	fd := FakeData{}
	cache.SetNow(fd.Now)

	ms := metric.NewMetricSet("eventType", "provider")

	for _, tt := range metricTests {
		ms.SetMetric(tt.key, tt.value, tt.metricType)

		if ms[tt.key] != tt.out {
			t.Errorf("SetMetric(\"%s\", %s, %s) => %s, want %s", tt.key, tt.value, tt.metricType, ms[tt.key], tt.out)
		}

		v, _, ok := cache.Get(tt.key)
		if tt.cache != nil {
			if !ok {
				t.Errorf("cache.Get(\"%s\") ==> %s, want %s", true, v, ok)
			} else if tt.cache != v {
				t.Errorf("cache.Get(\"%s\") ==> %s, want %s", tt.key, v, tt.cache)
			}
		}
	}
}
