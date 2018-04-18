package metric_test

import (
	"testing"
	"time"

	"github.com/newrelic/infra-integrations-sdk/metric"
	"github.com/newrelic/infra-integrations-sdk/persist"
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
	storer     interface{}
}{
	{"rateKey1", 10, metric.RATE, 0.0, 10.0},
	{"rateKey1", 100, metric.RATE, 90.0, 100.0},
	{"key1", .22323333, metric.RATE, 0.0, 0.22323333},
	{"key2", 100, metric.RATE, 0.0, 100.0},
	{"key2", 110, metric.RATE, 10.0, 110.0},
	{"key3", 10, metric.DELTA, 0.0, 10.0},
	{"key3", 110, metric.DELTA, 100.0, 110.0},
}

func TestSet_SetMetricGauge(t *testing.T) {
	fd := FakeData{}
	persist.SetNow(fd.Now)

	ms := metric.NewSet("some-event-type")

	ms.SetMetric("key", 10, metric.GAUGE)

	if ms["key"] != 10 {
		t.Errorf("metric stored not valid: %v", ms["key"])
	}
}

func TestSet_SetMetricAttribute(t *testing.T) {
	fd := FakeData{}
	persist.SetNow(fd.Now)

	ms := metric.NewSet("some-event-type")

	ms.SetMetric("key", "some-attribute", metric.ATTRIBUTE)

	if ms["key"] != "some-attribute" {
		t.Errorf("metric stored not valid: %v", ms["key"])
	}
}

func TestSetMetricStorer(t *testing.T) {
	fd := FakeData{}
	persist.SetNow(fd.Now)

	ms := metric.NewSet("eventType")

	for _, tt := range metricTests {
		ms.SetMetric(tt.key, tt.value, tt.metricType)

		if ms[tt.key] != tt.out {
			t.Errorf("SetMetric(\"%s\", %s, %v) => %s, want %s", tt.key, tt.value, tt.metricType, ms[tt.key], tt.out)
		}

		v, _, ok := persist.Get(tt.key)
		if !ok {
			t.Errorf("persist.Get(\"%v\") ==> %v, want %v", true, v, ok)
		} else if tt.storer != v {
			t.Errorf("persist.Get(\"%v\") ==> %v, want %v", tt.key, v, tt.storer)
		}
	}
}
