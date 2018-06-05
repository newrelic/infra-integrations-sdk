package metric

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/newrelic/infra-integrations-sdk/persist"
	"github.com/pkg/errors"
)

// SourceType defines the kind of data source. Based on this SourceType, metric
// package performs some calculations with it. Check below the description for
// each one.
type SourceType int

// KV represents a metric key-value pair.
type KV struct {
	Key   string
	Value string
}

const (
	// GAUGE is a value that may increase and decrease. It is stored as-is.
	GAUGE SourceType = iota
	// RATE is an ever-growing value which might be reseted. The package calculates the change rate.
	RATE SourceType = iota
	// DELTA is an ever-growing value which might be reseted. The package calculates the difference between samples.
	DELTA SourceType = iota
	// ATTRIBUTE is any string value
	ATTRIBUTE SourceType = iota
)

// Errors
var (
	ErrNonNumeric        = errors.New("non-numeric value for rate/delta")
	ErrNoStoreToCalcDiff = errors.New("can't use deltas nor rates without persistent store")
	ErrTooCloseSamples   = errors.New("samples too close in time, skipping")
	ErrNegativeDiff      = errors.New("source was reset, skipping")
)

// Set is the basic structure for storing metrics.
type Set struct {
	storer  persist.Storer
	Metrics map[string]interface{}
	id      []KV
}

// NewSet creates new metrics set, optionally identified by a list of key-values.
func NewSet(eventType string, storer persist.Storer, kv ...KV) (*Set, error) {
	ms := Set{
		Metrics: make(map[string]interface{}),
		storer:  storer,
		id:      kv,
	}

	err := ms.SetMetric("event_type", eventType, ATTRIBUTE)

	return &ms, err
}

// NewKV creates a new KV pair.
func NewKV(key string, value string) KV {
	return KV{
		Key:   key,
		Value: value,
	}
}

// SetMetric adds a metric to the Set object or updates the metric value if the metric already exists.
// It calculates elapsed difference for RATE and DELTA types.
func (ms *Set) SetMetric(name string, value interface{}, sourceType SourceType) error {
	var err error
	var newValue = value

	// Only sample metrics of numeric type
	switch sourceType {
	case RATE, DELTA:
		newValue, err = ms.elapsedDifference(name, value, sourceType)
		if err != nil {
			return errors.Wrapf(err, "cannot calculate elapsed difference for metric: %s value %v", name, value)
		}
	case GAUGE:
		newValue, err = castToFloat(value)
		if err != nil {
			return fmt.Errorf("non-numeric value for gauge metric: %s value: %v", name, value)
		}
	case ATTRIBUTE:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("non-string source type for attribute %s", name)
		}
	default:
		return fmt.Errorf("unknown source type for key %s", name)
	}

	ms.Metrics[name] = newValue
	return nil
}

func castToFloat(value interface{}) (float64, error) {
	return strconv.ParseFloat(fmt.Sprintf("%v", value), 64)
}

func (ms *Set) elapsedDifference(name string, absolute interface{}, sourceType SourceType) (elapsed float64, err error) {
	if ms.storer == nil {
		err = ErrNoStoreToCalcDiff
		return
	}

	newValue, err := castToFloat(absolute)
	if err != nil {
		err = ErrNonNumeric
		return
	}

	// Fetch last value & time
	var oldValue float64
	oldTime, err := ms.storer.Get(name, &oldValue)
	if err != nil && err != persist.ErrNotFound {
		return
	}

	// Store new value & time (no IO flush until Save)
	newTime := ms.storer.Set(name, newValue)

	// First value
	if err == persist.ErrNotFound {
		return 0, nil
	}

	// Time constraints
	duration := newTime - oldTime
	if duration == 0 {
		err = ErrTooCloseSamples
		return
	}

	elapsed = newValue - oldValue
	if elapsed < 0 {
		err = ErrNegativeDiff
		return
	}

	if sourceType == RATE {
		elapsed = elapsed / float64(duration)
	}

	return
}

// MarshalJSON adapts the internal structure of the metrics Set to the payload that is compliant with the protocol
func (ms Set) MarshalJSON() ([]byte, error) {
	return json.Marshal(ms.Metrics)
}
