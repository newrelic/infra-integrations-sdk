package metric

import (
	"testing"

	"github.com/newrelic/infra-integrations-sdk/persist"
	"github.com/stretchr/testify/assert"
)

func TestSet_MarshalMetricsSimpleStruct(t *testing.T) {
	simpleStruct := struct {
		Gauge     int     `metric_name:"metric.gauge" source_type:"gauge"`
		Attribute string  `metric_name:"metric.attribute" source_type:"attribute"`
		Rate      float64 `metric_name:"metric.rate" source_type:"rate"`
		Delta     float64 `metric_name:"metric.delta" source_type:"delta"`
	}{
		10,
		"some-attribute",
		float64(20),
		float64(30),
	}

	ms := newTestSet()
	err := ms.MarshalMetrics(simpleStruct)

	assert.NoError(t, err, "marshal error")
	assert.Equal(t, float64(10), ms.Metrics["metric.gauge"])
	assert.Equal(t, simpleStruct.Attribute, ms.Metrics["metric.attribute"])
	assert.Equal(t, float64(0), ms.Metrics["metric.rate"])
	assert.Equal(t, float64(0), ms.Metrics["metric.delta"])
}

func TestSet_MarshalMetricsComplexStruct(t *testing.T) {
	type NestedStruct struct {
		Rate  *float64 `metric_name:"metric.rate" source_type:"rate"`
		Delta float64  `metric_name:"metric.delta" source_type:"delta"`
		Map   map[string]bool
	}

	type InterfaceStruct struct {
		Metric int `metric_name:"metric.interface" source_type:"gauge"`
	}

	complexStruct := &struct {
		Gauge           int    `metric_name:"metric.gauge" source_type:"gauge"`
		Attribute       string `metric_name:"metric.attribute" source_type:"attribute"`
		Nested          *NestedStruct
		Slice           []string
		NestedInterface interface{}
	}{
		10,
		"some-attribute",
		&NestedStruct{
			nil,
			float64(10),
			map[string]bool{"one": true},
		},
		[]string{"one", "two", "three"},
		&InterfaceStruct{
			40,
		},
	}

	expectedMarshall := map[string]interface{}{
		"event_type":       "some-event-type", // added by newTestSet()
		"k":                "v",               // added by newTestSet()
		"metric.gauge":     10.,
		"metric.attribute": "some-attribute",
		"metric.delta":     0.,
		"metric.interface": 40.,
		//"metric.rate" is nil
		// Map has no tags
		// Slice has no tags
	}

	ms := newTestSet()
	assert.NoError(t, ms.MarshalMetrics(complexStruct), "marshal error")

	assert.Len(t, ms.Metrics, len(expectedMarshall))
	for expectedName, expectedValue := range expectedMarshall {
		v, ok := ms.Metrics[expectedName]
		assert.True(t, ok, "lacking metric: %s", expectedName)
		assert.Equal(t, expectedValue, v, "unexpected metric value %v", expectedValue)
	}
}

func TestSet_MarshalMetricsNonStruct(t *testing.T) {
	err := newTestSet().MarshalMetrics(1)

	assert.Error(t, err, "MarshalMetrics must take in a struct or struct pointer")
}

func TestSet_MarshalMetricsMissingOrInvalidTags(t *testing.T) {
	testCases := []struct {
		name  string
		input interface{}
	}{
		{
			"Missing metric_name",
			struct {
				Gauge int `source_type:"gauge"`
			}{
				10,
			},
		},
		{
			"Missing source_type",
			struct {
				Gauge int `metric_name:"metric.gauge"`
			}{
				10,
			},
		},
		{
			"Invalid source_type",
			struct {
				Gauge int `metric_name:"metric.gauge" source_type:"INVALID"`
			}{
				10,
			},
		},
	}

	for _, tc := range testCases {
		err := newTestSet().MarshalMetrics(tc.input)
		assert.Error(t, err, tc.name)
	}
}

func newTestSet() *Set {
	return NewSet("some-event-type", persist.NewInMemoryStore(), Attr("k", "v"))
}
