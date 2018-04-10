package metric_test

import (
	"testing"
	"time"

	"fmt"
	"github.com/newrelic/infra-integrations-sdk/cache"
	"github.com/newrelic/infra-integrations-sdk/metric"
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

var metricTests = []struct {
	key        string
	value      interface{}
	metricType metric.SourceType
	out        interface{}
	cache      interface{}
}{
	{"gaugeKey", 10, metric.GAUGE, 10, nil},
	{"keyAtribute", "sadad", metric.ATTRIBUTE, "sadad", nil},
	{"rateKey1", 10, metric.RATE, 0.0, 10.0},
	{"rateKey1", 100, metric.RATE, 90.0, 100.0},
	{"key1", .22323333, metric.RATE, 0.0, 0.22323333},
	{"key2", 100, metric.RATE, 0.0, 100.0},
	{"key2", 110, metric.RATE, 10.0, 110.0},
	{"key3", 10, metric.DELTA, 0.0, 10.0},
	{"key3", 110, metric.DELTA, 100.0, 110.0},
}

func TestSetMetric(t *testing.T) {
	fd := FakeData{}
	cache.SetNow(fd.Now)

	ms := metric.NewMetricSet("eventType")

	for _, tt := range metricTests {
		ms.SetMetric(tt.key, tt.value, tt.metricType)

		assert.Equal(t, ms[tt.key], tt.out, fmt.Sprintf("SetMetric(\"%s\", %+v, %d)", tt.key, tt.value, tt.metricType))

		v, _, ok := cache.Get(tt.key)
		if tt.cache == nil {
			continue
		}

		assert.Equal(t, ok, true, "cache.Get(\"%s\")", tt.key)
		assert.Equal(t, tt.cache, v, fmt.Sprintf("cache.Get(\"%s\")", tt.key))
	}
}
