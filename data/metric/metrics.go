package metric

import (
	"math"
	"time"

	err "github.com/newrelic/infra-integrations-sdk/v4/data/errors"
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
	Value float64 `json:"value"`
}

// summary is a metric of type summary.
type summary struct {
	metricBase
	Value summaryValue `json:"value"`
}

type summaryValue struct {
	Count   *float64 `json:"count"`
	Average *float64 `json:"average"`
	Sum     *float64 `json:"sum"`
	Min     *float64 `json:"min"`
	Max     *float64 `json:"max"`
}

type bucket struct {
	CumulativeCount *uint64  `json:"cumulative_count,omitempty"`
	UpperBound      *float64 `json:"upper_bound,omitempty"`
}

// PrometheusHistogram represents a Prometheus histogram
type PrometheusHistogram struct {
	metricBase
	Value PrometheusHistogramValue `json:"value,omitempty"`
}

// PrometheusHistogramValue represents the Value type for a Prometheus histogram.
type PrometheusHistogramValue struct {
	SampleCount *uint64  `json:"sample_count,omitempty"`
	SampleSum   *float64 `json:"sample_sum,omitempty"`
	// Buckets defines the buckets into which observations are counted. Each
	// element in the slice is the upper inclusive bound of a bucket. The
	// values must are sorted in strictly increasing order.
	Buckets []*bucket `json:"buckets,omitempty"`
}

type quantile struct {
	Quantile *float64 `json:"quantile,omitempty"`
	Value    *float64 `json:"value,omitempty"`
}

// PrometheusSummary represents a Prometheus summary
type PrometheusSummary struct {
	metricBase
	Value PrometheusSummaryValue `json:"value,omitempty"`
}

// PrometheusSummaryValue represents the Value type for a Prometheus summary.
type PrometheusSummaryValue struct {
	SampleCount *uint64     `json:"sample_count,omitempty"`
	SampleSum   *float64    `json:"sample_sum,omitempty"`
	Quantiles   []*quantile `json:"quantiles,omitempty"`
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
		Value: summaryValue{
			Count:   asFloatPtr(count),
			Average: asFloatPtr(average),
			Sum:     asFloatPtr(sum),
			Min:     asFloatPtr(min),
			Max:     asFloatPtr(max),
		},
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

// NewPrometheusHistogram creates a new metric structurally similar to a Prometheus histogram
func NewPrometheusHistogram(timestamp time.Time, name string, sampleCount uint64, sampleSum float64) (*PrometheusHistogram, error) {
	return &PrometheusHistogram{
		metricBase: metricBase{
			Timestamp:  timestamp.Unix(),
			Name:       name,
			Type:       SourcesTypeToName[PROMETHEUS_HISTOGRAM],
			Dimensions: Dimensions{},
		},
		Value: PrometheusHistogramValue{
			SampleCount: &sampleCount,
			SampleSum:   asFloatPtr(sampleSum),
		},
	}, nil
}

// AddBucket adds a new bucket to the histogram.
// Note that no attempt is made to keep buckets ordered, it's on the caller to guarantee the buckets are added
// in the correct order.
func (ph *PrometheusHistogram) AddBucket(cumulativeCount uint64, upperBound float64) {
	// ignore +Inf buckets
	if math.IsNaN(upperBound) || math.IsInf(upperBound, 0) {
		return
	}
	ph.Value.Buckets = append(ph.Value.Buckets, &bucket{
		CumulativeCount: &cumulativeCount,
		UpperBound:      asFloatPtr(upperBound),
	})
}

// NewPrometheusSummary creates a new metric structurally similar to a Prometheus summary
func NewPrometheusSummary(timestamp time.Time, name string, sampleCount uint64, sampleSum float64) (*PrometheusSummary, error) {
	return &PrometheusSummary{
		metricBase: metricBase{
			Timestamp:  timestamp.Unix(),
			Name:       name,
			Type:       SourcesTypeToName[PROMETHEUS_SUMMARY],
			Dimensions: Dimensions{},
		},
		Value: PrometheusSummaryValue{
			SampleCount: &sampleCount,
			SampleSum:   asFloatPtr(sampleSum),
		},
	}, nil
}

// AddQuantile adds a new quantile to the summary.
func (ps *PrometheusSummary) AddQuantile(quant float64, value float64) {
	// ignore invalid quantiles
	if math.IsNaN(quant) || math.IsNaN(value) {
		return
	}
	ps.Value.Quantiles = append(ps.Value.Quantiles, &quantile{
		Quantile: asFloatPtr(quant),
		Value:    asFloatPtr(value),
	})
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
