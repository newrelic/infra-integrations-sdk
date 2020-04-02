package metric

import (
	"errors"
	"time"
)

// Errors
var (
	ErrNonNumeric = errors.New("non-numeric value for rate/delta")
	//ErrNoStoreToCalcDiff = errors.New("cannot use deltas nor rates without persistent store")
	//ErrTooCloseSamples   = errors.New("samples too close in time, skipping")
	//ErrNegativeDiff      = errors.New("source was reset, skipping")
	//ErrOverrideSetAttrs  = errors.New("cannot overwrite metric-set attributes")
	//ErrDeltaWithNoAttrs  = errors.New("delta/rate metrics should be attached to an attribute identified metric-set")
)

// Dimensions stores the metric dimensions
type Dimensions map[string]string

// Set is the basic structure for storing metrics.
type Set []Metric

// Metric is the common interface for all metric types
type Metric interface {
	AddAttribute(key string, value string)
	Attribute(key string) string
	Attributes() Dimensions
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
	Interval int64 `json:"interval.ms"`
	Count    int64 `json:"count"`
}

// Summary is a metric of type summary.
type Summary struct {
	metricBase
	Interval int64   `json:"interval.ms"`
	Count    int64   `json:"count"`
	Average  float64 `json:"average"`
	Sum      float64 `json:"sum"`
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
}

// NewGauge creates a new metric of type Gauge
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

// NewCount creates a new metric of type Count
func NewCount(timestamp time.Time, interval time.Duration, name string, count int64) Metric {
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

// NewSummary creates a new metric of type Summary
func NewSummary(timestamp time.Time, interval time.Duration, name string, count int64, average float64, sum float64,
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

// AddAttribute adds an attribute (dimension) to the metric instance
func (m *metricBase) AddAttribute(key string, value string) {
	m.Dimensions[key] = value
}

// Attribute returns an attribute by key
func (m *metricBase) Attribute(key string) string {
	return m.Dimensions[key]
}

// Attributes returns all the dimensions of the metric
func (m *metricBase) Attributes() Dimensions {
	// TODO evaluate locking and making a copy instead
	return m.Dimensions
}
