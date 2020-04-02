package metric

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/newrelic/infra-integrations-sdk/data/attribute"
	"github.com/pkg/errors"
)

const (
	// nsSeparator is the metric namespace separator
	nsSeparator = "::"
)

// Errors
var (
	ErrNonNumeric        = errors.New("non-numeric value for rate/delta")
	ErrNoStoreToCalcDiff = errors.New("cannot use deltas nor rates without persistent store")
	ErrTooCloseSamples   = errors.New("samples too close in time, skipping")
	ErrNegativeDiff      = errors.New("source was reset, skipping")
	ErrOverrideSetAttrs  = errors.New("cannot overwrite metric-set attributes")
	ErrDeltaWithNoAttrs  = errors.New("delta/rate metrics should be attached to an attribute identified metric-set")
)

// Set is the basic structure for storing metrics.
type Set []Metric

type Metric interface {
	AddAttribute(key string, value string)
}

type MetricBase struct {
	Timestamp  int64                `json:"timestamp"`
	Name       string               `json:"name"`
	Type       string               `json:"type"`
	Dimensions attribute.Attributes `json:"attributes"`
}

type Gauge struct {
	MetricBase
	Value float64 `json:"value"`
}

type Count struct {
	MetricBase
	Interval int64 `json:"interval.ms"`
	Count    int64 `json:"count"`
}

type Summary struct {
	MetricBase
	Interval int64   `json:"interval.ms"`
	Count    int64   `json:"count"`
	Average  float64 `json:"average"`
	Sum      float64 `json:"sum"`
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
}

func NewGauge(timestamp time.Time, name string, value float64) Metric {
	return &Gauge{
		MetricBase: MetricBase{
			Timestamp:  timestamp.Unix(),
			Name:       name,
			Type:       SourcesTypeToName[GAUGE],
			Dimensions: attribute.Attributes{},
		},
		Value: value,
	}
}

func NewCount(timestamp time.Time, interval time.Duration, name string, count int64) Metric {
	return &Count{
		MetricBase: MetricBase{
			Timestamp:  timestamp.Unix(),
			Name:       name,
			Type:       SourcesTypeToName[COUNT],
			Dimensions: attribute.Attributes{},
		},
		Interval: interval.Milliseconds(),
		Count:    count,
	}
}

func NewSummary(timestamp time.Time, interval time.Duration, name string, count int64, average float64, sum float64,
	min float64, max float64) Metric {
	return &Summary{
		MetricBase: MetricBase{
			Timestamp:  timestamp.Unix(),
			Name:       name,
			Type:       SourcesTypeToName[SUMMARY],
			Dimensions: attribute.Attributes{},
		},
		Interval: interval.Milliseconds(),
		Count:    count,
		Average:  average,
		Sum:      sum,
		Min:      min,
		Max:      max,
	}
}

func (m *MetricBase) AddAttribute(key string, value string) {
	m.Dimensions[key] = value
}

//--- private
func castToFloat(value interface{}) (float64, error) {
	if b, ok := value.(bool); ok {
		if b {
			return 1, nil
		}
		return 0, nil
	}

	parsedValue, err := strconv.ParseFloat(fmt.Sprintf("%v", value), 64)
	if err != nil {
		return 0, err
	}

	if isNaNOrInf(parsedValue) {
		return 0, ErrNonNumeric
	}

	return parsedValue, nil
}

func isNaNOrInf(f float64) bool {
	return math.IsNaN(f) || math.IsInf(f, 0) || math.IsInf(f, -1)
}

//
//func (ms *Set) elapsedDifference(name string, absolute interface{}, sourceType SourceType) (elapsed float64, err error) {
//	if ms.storer == nil {
//		err = ErrNoStoreToCalcDiff
//		return
//	}
//
//	newValue, err := castToFloat(absolute)
//	if err != nil {
//		err = ErrNonNumeric
//		return
//	}
//
//	// Fetch last value & time
//	var oldValue float64
//	oldTime, err := ms.storer.Get(ms.namespace(name), &oldValue)
//	if err != nil && err != persist.ErrNotFound {
//		return
//	}
//
//	// Store new value & time (no IO flush until Save)
//	newTime := ms.storer.Set(ms.namespace(name), newValue)
//
//	// First value
//	if err == persist.ErrNotFound {
//		return 0, nil
//	}
//
//	// Time constraints
//	duration := newTime - oldTime
//	if duration == 0 {
//		err = ErrTooCloseSamples
//		return
//	}
//
//	elapsed = newValue - oldValue
//
//	if elapsed < 0 && sourceType.IsPositive() {
//		err = ErrNegativeDiff
//		return
//	}
//
//	if sourceType == RATE {
//		elapsed = elapsed / float64(duration)
//	}
//
//	return
//}

//// MarshalJSON adapts the internal structure of the metrics Set to the payload that is compliant with the protocol
//func (ms *Set) MarshalJSON() ([]byte, error) {
//	return json.Marshal(ms.Metrics)
//}
//
//// UnmarshalJSON unserializes protocol compliant JSON metrics into the metric set.
//func (ms *Set) UnmarshalJSON(data []byte) error {
//	return json.Unmarshal(data, &ms.Metrics)
//}
