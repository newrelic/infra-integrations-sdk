package metric

import (
	"testing"

	"github.com/newrelic/infra-integrations-sdk/persist"
	"github.com/stretchr/testify/assert"
)

func TestSet_MarshalMetricsSimpleStruct(t *testing.T) {
	ms := NewSet("some-event-type", persist.NewInMemoryStore(), Attr("k", "v"))

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

	err := ms.MarshalMetrics(simpleStruct)
	assert.NoError(t, err, "marshal error")

	assert.Equal(t, float64(10), ms.Metrics["metric.gauge"])
	assert.Equal(t, simpleStruct.Attribute, ms.Metrics["metric.attribute"])
	assert.Equal(t, float64(0), ms.Metrics["metric.rate"])
	assert.Equal(t, float64(0), ms.Metrics["metric.delta"])
}

func TestSet_MarshalMetricsComplexStruct(t *testing.T) {
	ms := NewSet("some-event-type", persist.NewInMemoryStore(), Attr("k", "v"))

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

	err := ms.MarshalMetrics(complexStruct)
	assert.NoError(t, err, "marshal error")

	assert.Equal(t, float64(10), ms.Metrics["metric.gauge"])
	assert.Equal(t, complexStruct.Attribute, ms.Metrics["metric.attribute"])
	assert.Equal(t, float64(0), ms.Metrics["metric.delta"])
	assert.Equal(t, float64(40), ms.Metrics["metric.interface"])
}

func TestSet_MarshalMetricsNonStruct(t *testing.T) {
	ms := NewSet("some-event-type", persist.NewInMemoryStore(), Attr("k", "v"))

	err := ms.MarshalMetrics(1)
	assert.Error(t, err, "MarshalMetrics must take in a struct or struct pointer")
}

func TestSet_MarshalMetricsMissingTags(t *testing.T) {
	testCases := []struct {
		name  string
		input interface{}
		ms    *Set
	}{
		{
			"Missing metric_name",
			struct {
				Gauge int `source_type:"gauge"`
			}{
				10,
			},
			NewSet("some-event-type", persist.NewInMemoryStore(), Attr("k", "v")),
		},
		{
			"Missing source_type",
			struct {
				Gauge int `metric_name:"metric.gauge"`
			}{
				10,
			},
			NewSet("some-event-type", persist.NewInMemoryStore(), Attr("k", "v")),
		},
	}

	for _, tc := range testCases {
		err := tc.ms.MarshalMetrics(tc.input)
		assert.Error(t, err, "field must have both metric_name and source_type if it wished to be marshaled")
	}
}

func TestSet_MarshalMetricsBadSourceType(t *testing.T) {
	ms := NewSet("some-event-type", persist.NewInMemoryStore(), Attr("k", "v"))

	simpleStruct := struct {
		Gauge int `metric_name:"metric.gauge" source_type:"nope"`
	}{
		10,
	}

	err := ms.MarshalMetrics(simpleStruct)
	assert.Error(t, err, "source_type must be a valid value")
}
