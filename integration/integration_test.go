package integration

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/stretchr/testify/assert"
)

var (
	integrationName    = "TestIntegration"
	integrationVersion = "1.0"
)

func Test_Integration_CreateIntegrationInitializesCorrectly(t *testing.T) {
	i := newTestIntegration(t)

	assert.Equal(t, "TestIntegration", i.Name)
	assert.Equal(t, "1.0", i.IntegrationVersion)
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

	i, err := New("integration", "4.0")

	assert.NoError(t, err)
	assert.Equal(t, "integration", i.Name)
	assert.Equal(t, "4.0", i.IntegrationVersion)
	assert.Equal(t, "4", i.ProtocolVersion)
	assert.Equal(t, 0, len(i.Entities))

	assert.NoError(t, i.Publish())

	// integration published metadata to stdout
	_ = f.Close()
	payload, err := ioutil.ReadFile(f.Name())
	assert.NoError(t, err)
	assert.Equal(t, stripBlanks([]byte(`{"name":"integration","protocol_version":"4","integration_version":"4.0","data":[]}`)), stripBlanks(payload))
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

func Test_Integration_NewEntityReturnsExistingEntity(t *testing.T) {
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
	i, err := New(integrationName, integrationVersion)
	assert.NoError(t, err)

	assert.Equal(t, i.logger, i.Logger())
}

func Test_Integration_LoggerReturnsInjectedInstance(t *testing.T) {
	l := log.NewStdErr(false)

	i, err := New(integrationName, integrationVersion, Logger(l))
	assert.NoError(t, err)

	assert.Equal(t, l, i.Logger())
}

func Test_Integration_PublishThrowsNoError(t *testing.T) {
	w := testWriter{
		func(integrationBytes []byte) {
			expectedOutputRaw := []byte(`
			{
			  "name": "TestIntegration",
			  "protocol_version": "4",
			  "integration_version": "1.0",
			  "data": [
				{
				  "common":{},
				  "entity": {
					"name": "EntityOne",
					"displayName":"",
					"type": "test",
					"tags": {
						"env":"prod"
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
					  	"count": 100
					},
					{	
						"timestamp": 10000000,
						"name": "metric-summary",
						"type": "summary",
						"attributes": {
							"distribution": "debian",							
							"os": "linux"	
						},
					  	"count": 1,
						"average": 10,
						"sum": 100,
						"min": 1,
						"max": 100
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
				  "common":{},
				  "entity": {
					"name": "EntityTwo",
					"displayName":"",
					"type": "test",
					"tags": {}
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
                  "metrics": [
				   {
						"timestamp": 10000000,
						"name": "cumulative-count",
						"type": "cumulative-count",
						"attributes": {},
						"count": 120
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

	i, err := New("TestIntegration", "1.0", Logger(log.Discard), Writer(w))
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

	assert.NoError(t, i.Publish())

	// check integration  was reset
	assert.Empty(t, i.Entities)
}

//--- helpers
func newTestIntegration(t *testing.T) *Integration {
	i, err := New(integrationName, integrationVersion, Logger(log.Discard), Writer(ioutil.Discard))
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
