package metric

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/newrelic/infra-integrations-sdk/persist"
)

// SourceType defines the kind of data source. Based on this SourceType, metric
// package performs some calculations with it. Check below the description for
// each one.
type SourceType int

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

// Set is the basic structure for storing metrics.
type Set struct {
	// TODO: global refactor, separate JSON structs from entities that manage it
	storer  persist.Storer
	Metrics map[string]interface{}
}

// NewSet returns a new Set instance.
// TODO: this function should not be visible to the SDK user
func NewSet(eventType string, storer persist.Storer) *Set {
	ms := Set{
		Metrics: map[string]interface{}{},
		storer:  storer,
	}
	ms.SetMetric("event_type", eventType, ATTRIBUTE) // nolint: errcheck
	return &ms
}

// SetMetric adds a metric to the Set object or updates the metric value
// if the metric already exists, performing a calculation if the SourceType
// (RATE, DELTA) requires it.
func (ms *Set) SetMetric(name string, value interface{}, sourceType SourceType) error {
	var err error
	var newValue = value

	// Only sample metrics of numeric type
	switch sourceType {
	case RATE, DELTA:
		if ms.storer == nil {
			// This will only happen if the user explicitly builds the integration invoking 'NoCache' function
			return fmt.Errorf("integrations built with no-store can't use DELTAs and RATEs")
		}
		if !isNumeric(value) {
			return fmt.Errorf("non-numeric source type for rate/delta metric %s", name)
		}
		newValue, err = ms.sample(name, value, sourceType)
		if err != nil {
			return err
		}
	case GAUGE:
		if !isNumeric(value) {
			return fmt.Errorf("non-numeric source type for gauge metric %s", name)
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

func isNumeric(value interface{}) bool {
	_, err := strconv.ParseFloat(fmt.Sprintf("%v", value), 64)
	return err == nil
}

func (ms *Set) sample(name string, value interface{}, sourceType SourceType) (float64, error) {
	sampledValue := 0.0

	// Convert the value to a float64 so we can compare it with the stored one
	floatValue, err := strconv.ParseFloat(fmt.Sprintf("%v", value), 64)
	if err != nil {
		return sampledValue, fmt.Errorf("can't sample metric of unknown type %s", name)
	}

	// Retrieve the last value and timestamp from Storer
	oldval, oldTime, ok := ms.storer.Get(name)
	// And replace it with the new value which we want to keep
	newTime := ms.storer.Set(name, floatValue)

	if ok {
		duration := (newTime - oldTime)
		if duration == 0 {
			return sampledValue, fmt.Errorf("samples for %s are too close in time, skipping sampling", name)
		}

		if floatValue-oldval < 0 {
			return sampledValue, fmt.Errorf("source for %s was reseted, skipping sampling", name)
		}
		if sourceType == DELTA {
			sampledValue = floatValue - oldval
		} else {
			sampledValue = (floatValue - oldval) / float64(duration)
		}
	}

	return sampledValue, nil
}

// MarshalJSON adapts the internal structure of the metrics Set to the payload that is compliant with the protocol
func (ms Set) MarshalJSON() ([]byte, error) {
	return json.Marshal(ms.Metrics)
}
