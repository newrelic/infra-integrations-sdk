package metric

import (
	"time"
)

// Dimensions stores the metric dimensions
type Dimensions map[string]string

// Set is the basic structure for storing metrics.
type Set []Metric

// Metric is the common interface for all metric types
type Metric interface {
	AddDimension(key string, value string)
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

// PDelta is a metric of type pdelta (positive delta)
type PDelta struct {
	metricBase
	Value float64 `json:"value"`
}

// Count is a metric of type count
type Count struct {
	metricBase
	Interval int64  `json:"interval.ms"`
	Count    uint64 `json:"count"`
}

// Summary is a metric of type summary.
type Summary struct {
	metricBase
	Interval int64   `json:"interval.ms"`
	Count    uint64  `json:"count"`
	Average  float64 `json:"average"`
	Sum      float64 `json:"sum"`
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
}

// NewGauge creates a new metric of type gauge
func NewGauge(timestamp time.Time, name string, value float64) Metric {
	return &Gauge{
		metricBase: metricBase{
			Timestamp:  timestamp.Unix(),
			Name:       name,
			Type:       SourcesTypeToName[GAUGE],
			Dimensions: Dimensions{},
		},
		Value: value,
	}
}

// NewPDelta creates a new metric of type pdelta
func NewPDelta(timestamp time.Time, name string, value float64) Metric {
	return &PDelta{
		metricBase: metricBase{
			Timestamp:  timestamp.Unix(),
			Name:       name,
			Type:       SourcesTypeToName[PDELTA],
			Dimensions: Dimensions{},
		},
		Value: value,
	}
}

// NewCount creates a new metric of type count
func NewCount(timestamp time.Time, interval time.Duration, name string, count uint64) Metric {
	return &Count{
		metricBase: metricBase{
			Timestamp:  timestamp.Unix(),
			Name:       name,
			Type:       SourcesTypeToName[COUNT],
			Dimensions: Dimensions{},
		},
		Interval: interval.Milliseconds(),
		Count:    count,
	}
}

// NewSummary creates a new metric of type summary
func NewSummary(timestamp time.Time, interval time.Duration, name string, count uint64, average float64, sum float64,
	min float64, max float64) Metric {
	return &Summary{
		metricBase: metricBase{
			Timestamp:  timestamp.Unix(),
			Name:       name,
			Type:       SourcesTypeToName[SUMMARY],
			Dimensions: Dimensions{},
		},
		Interval: interval.Milliseconds(),
		Count:    count,
		Average:  average,
		Sum:      sum,
		Min:      min,
		Max:      max,
	}
}

// AddDimension adds a dimension to the metric instance
func (m *metricBase) AddDimension(key string, value string) {
	m.Dimensions[key] = value
}

// Dimension returns an attribute by key
func (m *metricBase) Dimension(key string) string {
	return m.Dimensions[key]
}

// GetDimensions gets all the dimensions
func (m *metricBase) GetDimensions() Dimensions {
	return m.Dimensions
}
