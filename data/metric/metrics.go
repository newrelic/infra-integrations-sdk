package metric

import (
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

// Gauge is a metric of type gauge
type Gauge struct {
	metricBase
	Value float64 `json:"value"`
}

// Count is a metric of type count
type Count struct {
	metricBase
	Count uint64 `json:"count"`
}

// Summary is a metric of type summary.
type Summary struct {
	metricBase
	Count   uint64  `json:"count"`
	Average float64 `json:"average"`
	Sum     float64 `json:"sum"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
}

// NewGauge creates a new metric of type gauge
func NewGauge(timestamp time.Time, name string, value float64) (Metric, error) {
	if len(name) == 0 {
		return nil, err.ParameterCannotBeEmpty("name")
	}

	return &Gauge{
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
func NewCount(timestamp time.Time, name string, count uint64) (Metric, error) {
	if len(name) == 0 {
		return nil, err.ParameterCannotBeEmpty("name")
	}

	return &Count{
		metricBase: metricBase{
			Timestamp:  timestamp.Unix(),
			Name:       name,
			Type:       SourcesTypeToName[COUNT],
			Dimensions: Dimensions{},
		},
		Count: count,
	}, nil
}

// NewSummary creates a new metric of type summary
func NewSummary(timestamp time.Time, name string, count uint64, average float64, sum float64,
	min float64, max float64) (Metric, error) {
	if len(name) == 0 {
		return nil, err.ParameterCannotBeEmpty("name")
	}

	return &Summary{
		metricBase: metricBase{
			Timestamp:  timestamp.Unix(),
			Name:       name,
			Type:       SourcesTypeToName[SUMMARY],
			Dimensions: Dimensions{},
		},
		Count:   count,
		Average: average,
		Sum:     sum,
		Min:     min,
		Max:     max,
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

// Dimension returns an attribute by key
func (m *metricBase) Dimension(key string) string {
	return m.Dimensions[key]
}

// GetDimensions gets all the dimensions
func (m *metricBase) GetDimensions() Dimensions {
	return m.Dimensions
}
