package metric

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"

	"github.com/newrelic/infra-integrations-sdk/data/attribute"
	"github.com/newrelic/infra-integrations-sdk/persist"
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
type Set struct {
	storer       persist.Storer
	Metrics      map[string]interface{}
	nsAttributes []attribute.Attribute
}

// NewSet creates new metrics set, optionally related to a list of attributes. These attributes makes the metric-set unique.
// If related attributes are used, then new attributes are added.
func NewSet(eventType string, storer persist.Storer, attributes ...attribute.Attribute) (s *Set) {
	s = &Set{
		Metrics:      make(map[string]interface{}),
		storer:       storer,
		nsAttributes: attributes,
	}

	s.setSetAttribute("event_type", eventType)

	for _, attr := range attributes {
		s.setSetAttribute(attr.Key, attr.Value)
	}

	return
}

// AddCustomAttributes add customAttributes to MetricSet
func AddCustomAttributes(metricSet *Set, customAttributes []attribute.Attribute) {
	for _, attr := range customAttributes {
		metricSet.setSetAttribute(attr.Key, attr.Value)
	}
}

// SetMetric adds a metric to the Set object or updates the metric value if the metric already exists.
// It calculates elapsed difference for RATE and DELTA types.
func (ms *Set) SetMetric(name string, value interface{}, sourceType SourceType) (err error) {
	var errElapsed error
	var newValue = value

	// Only sample metrics of numeric type
	switch sourceType {
	case RATE, DELTA, PRATE, PDELTA:
		if len(ms.nsAttributes) == 0 {
			err = ErrDeltaWithNoAttrs
			return
		}
		newValue, errElapsed = ms.elapsedDifference(name, value, sourceType)
		if errElapsed != nil {
			return errors.Wrapf(errElapsed, "cannot calculate elapsed difference for metric: %s value %v", name, value)
		}
	case GAUGE:
		newValue, err = castToFloat(value)
		if err != nil {
			return fmt.Errorf("non-numeric value for gauge metric: %s value: %v", name, value)
		}
	case ATTRIBUTE:
		strVal, ok := value.(string)
		if !ok {
			return fmt.Errorf("non-string source type for attribute %s", name)
		}
		for _, attr := range ms.nsAttributes {
			if name == attr.Key && strVal == attr.Value {
				return ErrOverrideSetAttrs
			}
		}
	default:
		return fmt.Errorf("unknown source type for key %s", name)
	}

	ms.Metrics[name] = newValue

	return
}

func (ms *Set) setSetAttribute(name string, value string) {
	ms.Metrics[name] = value
}

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
	oldTime, err := ms.storer.Get(ms.namespace(name), &oldValue)
	if err != nil && err != persist.ErrNotFound {
		return
	}

	// Store new value & time (no IO flush until Save)
	newTime := ms.storer.Set(ms.namespace(name), newValue)

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

	if elapsed < 0 && sourceType.IsPositive() {
		err = ErrNegativeDiff
		return
	}

	if sourceType == RATE || sourceType == PRATE {
		elapsed = elapsed / float64(duration)
	}

	return
}

// prefix a metric name with a namespace based on the alphabetical order of the set related attributes.
func (ms *Set) namespace(metricName string) string {
	ns := ""
	separator := ""

	attrs := ms.nsAttributes
	sort.Sort(attribute.Attributes(attrs))

	for _, attr := range attrs {
		ns = fmt.Sprintf("%s%s%s", ns, separator, attr.Namespace())
		separator = nsSeparator
	}

	return fmt.Sprintf("%s%s%s", ns, separator, metricName)
}

// MarshalJSON adapts the internal structure of the metrics Set to the payload that is compliant with the protocol
func (ms *Set) MarshalJSON() ([]byte, error) {
	return json.Marshal(ms.Metrics)
}

// UnmarshalJSON unserializes protocol compliant JSON metrics into the metric set.
func (ms *Set) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &ms.Metrics)
}
