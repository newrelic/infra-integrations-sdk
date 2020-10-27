package integration

import (
	"io/ioutil"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/stretchr/testify/assert"
)

var (
	integrationName     = "TestIntegration"
	integrationVersion  = "1.0"
	integrationProvider = "newRelic"
)

func Test_Integration_CreateIntegrationInitializesCorrectly(t *testing.T) {
	i := newTestIntegration(t)

	assert.Equal(t, "TestIntegration", i.Metadata.Name)
	assert.Equal(t, "1.0", i.Metadata.Version)
	assert.Equal(t, "4", i.ProtocolVersion)
	assert.Len(t, i.Entities, 0)
}

func Test_Integration_DefaultIntegrationWritesToStdout(t *testing.T) {
	// capture stdout to file
	f, err := ioutil.TempFile("", "stdout")
	assert.NoError(t, err)
	back := os.Stdout
	defer func() {
		os.Stdout = back
		_ = os.Remove(f.Name())
	}()
	os.Stdout = f

	i, err := New("integration", "4.0", "newRelic")

	assert.NoError(t, err)
	assert.Equal(t, "integration", i.Metadata.Name)
	assert.Equal(t, "4.0", i.Metadata.Version)
	assert.Equal(t, "4", i.ProtocolVersion)
	assert.Equal(t, 0, len(i.Entities))

	assert.NoError(t, i.Publish())

	// integration published metadata to stdout
	_ = f.Close()
	payload, err := ioutil.ReadFile(f.Name())
	assert.NoError(t, err)
	assert.Equal(t, stripBlanks([]byte(`{"protocol_version":"4","integration":{"name":"integration","version":"4.0"},"data":[]}`)), stripBlanks(payload))
}

func Test_Integration_EntitiesWithSameBasicMetadataAreEqual(t *testing.T) {
	i := newTestIntegration(t)

	e1, err := i.NewEntity("name", "displayName", "type")
	assert.NoError(t, err)

	e2, err := i.NewEntity("name", "displayName", "type")
	assert.NoError(t, err)

	assert.Equal(t, e1, e2)
}

func Test_Integration_EntitiesWithSameMetadataAreEqual(t *testing.T) {
	i := newTestIntegration(t)

	e1, err := i.NewEntity("name", "displayName", "type")
	assert.NoError(t, err)

	e2, err := i.NewEntity("name", "displayName", "type")
	assert.NoError(t, err)

	assert.Equal(t, e1, e2)
}

func Test_Integration_EntitiesWithDifferentMetadataAreNotEqual(t *testing.T) {
	i := newTestIntegration(t)

	e1, err := i.NewEntity("name", "ns", "")
	assert.NoError(t, err)
	e2, err := i.NewEntity("name2", "ns", "")
	assert.NoError(t, err)
	e3, err := i.NewEntity("name", "ns", "")
	assert.NoError(t, err)

	assert.NotEqual(t, e1, e2, "Different names create different entities")
	assert.Equal(t, e1, e3, "Same type & name create same entity")
}

func Test_Integration_EntitiesWithDifferentTagsAreNotEqual(t *testing.T) {
	i := newTestIntegration(t)

	e1, err := i.NewEntity("name", "ns", "")
	assert.NoError(t, err)
	_ = e1.AddTag("k", "v")

	e2, err := i.NewEntity("name", "ns", "")
	assert.NoError(t, err)

	e3, err := i.NewEntity("name", "ns", "")
	assert.NoError(t, err)
	_ = e3.AddTag("k", "v")

	assert.False(t, e1.SameAs(e2), "Different tags create different entities")
	assert.True(t, e1.SameAs(e3), "Same metadata creates/retrieves same entity")
}

func Test_Integration_EntitiesWithSameMetadataAreTheSame(t *testing.T) {
	i := newTestIntegration(t)

	e1, err := i.NewEntity("Entity1", "test", "")
	assert.NoError(t, err)

	e2, err := i.NewEntity("Entity1", "test", "")
	assert.NoError(t, err)

	assert.True(t, e1.SameAs(e2))

	i.AddEntity(e1)
	i.AddEntity(e2)

	assert.Len(t, i.Entities, 2)
}

func Test_Integration_LoggerReturnsDefaultLogger(t *testing.T) {
	i, err := New(integrationName, integrationVersion, integrationProvider)
	assert.NoError(t, err)

	assert.Equal(t, i.logger, i.Logger())
}

func Test_Integration_LoggerReturnsInjectedInstance(t *testing.T) {
	l := log.NewStdErr(false)

	i, err := New(integrationName, integrationVersion, integrationProvider, Logger(l))
	assert.NoError(t, err)

	assert.Equal(t, l, i.Logger())
}

func Test_Integration_PublishThrowsNoError(t *testing.T) {
	w := testWriter{
		func(integrationBytes []byte) {
			expectedOutputRaw := []byte(`
			{
			  "protocol_version": "4",
			  "integration": {
				"name": "TestIntegration",
				"version": "1.0"
			  },
			  "data": [
				{
				  "common": {},
				  "entity": {
					"name": "EntityOne",
					"displayName": "",
					"type": "test",
					"metadata": {
					  "tags.env": "prod"
					}
				  },
				  "metrics": [
					{
					  "timestamp": 10000000,
					  "name": "metric-gauge",
					  "type": "gauge",
					  "attributes": {},
					  "value": 1
					},
					{
					  "timestamp": 10000000,
					  "name": "metric-count",
					  "type": "count",
					  "attributes": {
						"cpu": "amd"
					  },
					  "value": 100
					},
					{
					  "timestamp": 10000000,
					  "name": "metric-summary",
					  "type": "summary",
					  "attributes": {
						"distribution": "debian",
						"os": "linux"
					  },
					  "value": {
						"count": 1,
						"average": 10,
						"sum": 100,
						"min": 1,
						"max": 100
					  }
					}
				  ],
				  "inventory": {
					"custom/example": {
					  "version": "1.2.3"
					}
				  },
				  "events": [
					{
					  "timestamp": 10000000,
					  "summary": "evnt1sum",
					  "category": "evnt1cat",
					  "attributes": {
						"attr1": "attr1Val",
						"attr2": 42
					  }
					},
					{
					  "timestamp": 10000000,
					  "summary": "evnt2sum",
					  "category": "evnt2cat"
					}
				  ]
				},
				{
				  "common": {},
				  "entity": {
					"name": "EntityTwo",
					"displayName": "",
					"type": "test",
					"metadata": {}
				  },
				  "metrics": [
					{
					  "timestamp": 10000000,
					  "name": "metricOne",
					  "type": "gauge",
					  "attributes": {
						"processName": "java"
					  },
					  "value": 2
					}
				  ],
				  "inventory": {},
				  "events": []
				},
				{
				  "common": {},
				  "entity": {
					"name": "EntityThree",
					"displayName": "",
					"type": "test",
					"metadata": {}
				  },
				  "metrics": [
					{
					  "timestamp": 10000000,
					  "name": "metric-summary-with-nan",
					  "type": "summary",
					  "attributes": {},
					  "value": {
						"count": 1,
						"average": null,
						"sum": 100,
						"min": null,
						"max": null
					  }
					}
				  ],
				  "inventory": {},
				  "events": []
				},
				{
				  "common": {},
				  "metrics": [
					{
					  "timestamp": 10000000,
					  "name": "cumulative-count",
					  "type": "cumulative-count",
					  "attributes": {},
					  "value": 120
					},
					{
					  "timestamp": 10000000,
					  "name": "rate",
					  "type": "rate",
					  "attributes": {},
					  "value": 120
					},
					{
					  "timestamp": 10000000,
					  "name": "cumulative-rate",
					  "type": "cumulative-rate",
					  "attributes": {},
					  "value": 120
					},
					{
					  "timestamp": 10000000,
					  "name": "prometheus-histogram",
					  "type": "prometheus-histogram",
					  "attributes": {},
					  "value": {
						"sample_count": 2,
						"sample_sum": 3,
						"buckets": [
						  {
							"cumulative_count": 1,
							"upper_bound": 1
						  },
						  {
							"cumulative_count": 2,
							"upper_bound": 2
						  }
						]
					  }
					},
					{
					  "timestamp": 10000000,
					  "name": "prometheus-summary",
					  "type": "prometheus-summary",
					  "attributes": {},
					  "value": {
						"sample_count": 2,
						"sample_sum": 2,
						"quantiles": [
						  {
							"quantile": 0.5,
							"value": 1
						  },
						  {
							"quantile": 0.9,
							"value": 1
						  }
						]
					  }
					}
				  ],
				  "inventory": {
					"some-inventory": {
					  "some-field": "some-value"
					}
				  },
				  "events": [
					{
					  "timestamp": 10000000,
					  "summary": "evnt2sum",
					  "category": "evnt2cat"
					}
				  ]
				}
			  ]
			}`)

			// awful but cannot compare with json.Unmarshal as it's not supported by Integration
			assert.Equal(t, stripBlanks(expectedOutputRaw), stripBlanks(integrationBytes))
		},
	}

	i, err := New("TestIntegration", "1.0", "newRelic", Logger(log.Discard), Writer(w))
	assert.NoError(t, err)

	e, err := i.NewEntity("EntityOne", "test", "")
	assert.NoError(t, err)
	_ = e.AddTag("env", "prod")

	gauge, _ := Gauge(time.Unix(10000000, 0), "metric-gauge", 1)
	count, _ := Count(time.Unix(10000000, 0), "metric-count", 100)
	_ = count.AddDimension("cpu", "amd")
	summary, _ := Summary(time.Unix(10000000, 0), "metric-summary", 1, 10, 100, 1, 100)
	// attributes should be ordered by key in lexicographic order
	_ = summary.AddDimension("os", "linux")
	_ = summary.AddDimension("distribution", "debian")
	// add metrics to entity 1
	e.AddMetric(gauge)
	e.AddMetric(count)
	e.AddMetric(summary)
	// add 1st event to entity 1
	ev1, err := i.NewEvent(time.Unix(10000000, 0), "evnt1sum", "evnt1cat")
	assert.NoError(t, err)
	_ = ev1.AddAttribute("attr1", "attr1Val")
	_ = ev1.AddAttribute("attr2", 42)
	e.AddEvent(ev1)
	// add 2nd event to entity 1
	ev2, err := i.NewEvent(time.Unix(10000000, 0), "evnt2sum", "evnt2cat")
	assert.NoError(t, err)
	e.AddEvent(ev2)
	// add inventory to entity 1. only one because order is not guaranteed and the test is comparing with a static string
	err = e.AddInventoryItem("custom/example", "version", "1.2.3")
	assert.NoError(t, err)
	// add entity 1 to integration
	i.AddEntity(e)

	// add entity 2
	e2, err := i.NewEntity("EntityTwo", "test", "")
	assert.NoError(t, err)
	// add metric to entity 2
	gauge, _ = Gauge(time.Unix(10000000, 0), "metricOne", 2)
	_ = gauge.AddDimension("processName", "java")
	e2.AddMetric(gauge)
	// add entity 2 to integration
	i.AddEntity(e2)

	// add inventory to the "host" entity
	err = i.HostEntity.AddInventoryItem("some-inventory", "some-field", "some-value")
	assert.NoError(t, err)

	// add event to the "host" entity (will not create one, inventory before already created it)
	ev3, err := i.NewEvent(time.Unix(10000000, 0), "evnt2sum", "evnt2cat")
	assert.NoError(t, err)
	i.HostEntity.AddEvent(ev3)

	// add a cumulative count metric to the host entity
	ccount, _ := CumulativeCount(time.Unix(10000000, 0), "cumulative-count", 120)
	i.HostEntity.AddMetric(ccount)
	// add a rate metric to the host entity
	rate, _ := Rate(time.Unix(10000000, 0), "rate", 120)
	i.HostEntity.AddMetric(rate)

	// add a cumulative rate metric to the host entity
	crate, _ := CumulativeRate(time.Unix(10000000, 0), "cumulative-rate", 120)
	i.HostEntity.AddMetric(crate)

	// add entity 3
	e3, err := i.NewEntity("EntityThree", "test", "")
	assert.NoError(t, err)
	// add metric to entity 2
	summary2, _ := Summary(time.Unix(10000000, 0), "metric-summary-with-nan", 1, math.NaN(), 100, math.NaN(), math.NaN())
	e3.AddMetric(summary2)
	// add entity 3 to integration
	i.AddEntity(e3)

	phisto, _ := PrometheusHistogram(time.Unix(10000000, 0), "prometheus-histogram", 2, 3)
	phisto.AddBucket(1, 1)
	phisto.AddBucket(2, 2)
	i.HostEntity.AddMetric(phisto)

	psum, _ := PrometheusSummary(time.Unix(10000000, 0), "prometheus-summary", 2, 2)
	psum.AddQuantile(0.5, 1)
	psum.AddQuantile(0.9, 1)
	i.HostEntity.AddMetric(psum)

	assert.NoError(t, i.Publish())

	// check integration  was reset
	assert.Empty(t, i.Entities)
}

func Test_Integration_FindEntity(t *testing.T) {
	i := newTestIntegration(t)

	_, found := i.FindEntity("some-entity-name")
	assert.False(t, found)

	e, err := i.NewEntity("some-entity-name", "test", "")
	assert.NoError(t, err)
	assert.NotNil(t, e)
	// not added yet
	_, found = i.FindEntity("some-entity-name")
	assert.False(t, found)

	i.AddEntity(e)
	// after adding
	assert.Len(t, i.Entities, 1)
	e1, found1 := i.FindEntity("some-entity-name")
	assert.True(t, found1)
	assert.True(t, e.SameAs(e1))
}

//--- helpers
func newTestIntegration(t *testing.T) *Integration {
	i, err := New(integrationName, integrationVersion, integrationProvider, Logger(log.Discard), Writer(ioutil.Discard))
	assert.NoError(t, err)

	return i
}

func stripBlanks(b []byte) string {
	return strings.Replace(
		strings.Replace(
			strings.Replace(
				string(b),
				" ", "", -1),
			"\n", "", -1),
		"\t", "", -1)
}

type testWriter struct {
	testFunc func([]byte)
}

func (w testWriter) Write(p []byte) (n int, err error) {
	w.testFunc(p)

	return len(p), err
}
