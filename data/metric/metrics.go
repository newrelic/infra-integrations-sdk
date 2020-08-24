package metric

import (
	"math"
	"time"

	err "github.com/newrelic/infra-integrations-sdk/data/errors"
)

// Dimensions stores the metric dimensions
type Dimensions map[string]string

// Metrics is the basic structure for storing metrics.
type Metrics []Metric

// Metric is the common interface for all metric types
type Metric interface {
	AddDimension(key string, value string) error
	Dimension(key string) string
	GetDimensions() Dimensions
}

type metricBase struct {
	Timestamp  int64      `json:"timestamp"`
	Name       string     `json:"name"`
	Type       string     `json:"type"`
	Dimensions Dimensions `json:"attributes"`
}

// gauge is a metric of type gauge
type gauge struct {
	metricBase
	Value float64 `json:"value"`
}

// count is a metric of type count
// This indicates to the Infra agent that the value should be interpreted as a count that is reset in each interval
type count struct {
	metricBase
	Value float64 `json:"count"`
}

// summary is a metric of type summary.
type summary struct {
	metricBase
	Count   *float64 `json:"count"`
	Average *float64 `json:"average"`
	Sum     *float64 `json:"sum"`
	Min     *float64 `json:"min"`
	Max     *float64 `json:"max"`
}

// cumulativeCount is a metric of type cumulative count
// This indicates to the Infra agent that the value should be calculated as cumulative count (ever increasing value)
type cumulativeCount count

// rate is a metric of type rate
// This indicates to the Infra agent that the value should be calculated as a rate
type rate gauge

// cumulativeRate is a metric of type cumulative rate
// This indicates to the Infra agent that the value should be calculated as a cumulative rate
type cumulativeRate rate

// NewGauge creates a new metric of type gauge
func NewGauge(timestamp time.Time, name string, value float64) (Metric, error) {
	if len(name) == 0 {
		return nil, err.ParameterCannotBeEmpty("name")
	}

	return &gauge{
		metricBase: metricBase{
			Timestamp:  timestamp.Unix(),
			Name:       name,
			Type:       SourcesTypeToName[GAUGE],
			Dimensions: Dimensions{},
		},
		Value: value,
	}, nil
}

// NewCount creates a new metric of type count
func NewCount(timestamp time.Time, name string, value float64) (Metric, error) {
	if len(name) == 0 {
		return nil, err.ParameterCannotBeEmpty("name")
	}
	if value < 0 {
		return nil, err.ParameterCannotBeNegative("value", value)
	}
	return &count{
		metricBase: metricBase{
			Timestamp:  timestamp.Unix(),
			Name:       name,
			Type:       SourcesTypeToName[COUNT],
			Dimensions: Dimensions{},
		},
		Value: value,
	}, nil
}

// NewSummary creates a new metric of type summary
func NewSummary(timestamp time.Time, name string, count float64, average float64, sum float64,
	min float64, max float64) (Metric, error) {
	if len(name) == 0 {
		return nil, err.ParameterCannotBeEmpty("name")
	}
	if count < 0 {
		return nil, err.ParameterCannotBeNegative("count", count)
	}

	return &summary{
		metricBase: metricBase{
			Timestamp:  timestamp.Unix(),
			Name:       name,
			Type:       SourcesTypeToName[SUMMARY],
			Dimensions: Dimensions{},
		},
		Count:   asFloatPtr(count),
		Average: asFloatPtr(average),
		Sum:     asFloatPtr(sum),
		Min:     asFloatPtr(min),
		Max:     asFloatPtr(max),
	}, nil
}

// NewCumulativeCount creates a new metric of type cumulative count
func NewCumulativeCount(timestamp time.Time, name string, value float64) (Metric, error) {
	if len(name) == 0 {
		return nil, err.ParameterCannotBeEmpty("name")
	}
	if value < 0 {
		return nil, err.ParameterCannotBeNegative("value", value)
	}
	return &cumulativeCount{
		metricBase: metricBase{
			Timestamp:  timestamp.Unix(),
			Name:       name,
			Type:       SourcesTypeToName[CUMULATIVE_COUNT],
			Dimensions: Dimensions{},
		},
		Value: value,
	}, nil
}

// NewRate creates a new metric of type rate
func NewRate(timestamp time.Time, name string, value float64) (Metric, error) {
	if len(name) == 0 {
		return nil, err.ParameterCannotBeEmpty("name")
	}

	return &rate{
		metricBase: metricBase{
			Timestamp:  timestamp.Unix(),
			Name:       name,
			Type:       SourcesTypeToName[RATE],
			Dimensions: Dimensions{},
		},
		Value: value,
	}, nil

}

// NewCumulativeRate creates a new metric of type cumulative rate
func NewCumulativeRate(timestamp time.Time, name string, value float64) (Metric, error) {
	if len(name) == 0 {
		return nil, err.ParameterCannotBeEmpty("name")
	}

	return &cumulativeRate{
		metricBase: metricBase{
			Timestamp:  timestamp.Unix(),
			Name:       name,
			Type:       SourcesTypeToName[CUMULATIVE_RATE],
			Dimensions: Dimensions{},
		},
		Value: value,
	}, nil
}

// AddDimension adds a dimension to the metric instance
func (m *metricBase) AddDimension(key string, value string) error {
	if len(key) == 0 {
		return err.ParameterCannotBeEmpty("name")
	}

	m.Dimensions[key] = value
	return nil
}

// Dimension returns an dimension by key
func (m *metricBase) Dimension(key string) string {
	return m.Dimensions[key]
}

// GetDimensions gets all the dimensions
func (m *metricBase) GetDimensions() Dimensions {
	return m.Dimensions
}

func asFloatPtr(value float64) *float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return nil
	}
	return &value
}
