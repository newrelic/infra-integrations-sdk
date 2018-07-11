package metric

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/newrelic/infra-integrations-sdk/persist"
	"github.com/pkg/errors"
)

// SourceType defines the kind of data source. Based on this SourceType, metric
// package performs some calculations with it. Check below the description for
// each one.
type SourceType int

// Attribute represents an attribute metric in key-value pair format.
type Attribute struct {
	Key   string
	Value string
}

const (
	// GAUGE is a value that may increase and decrease. It is stored as-is.
	GAUGE SourceType = iota
	// RATE is an ever-growing value which might be reset. The package calculates the change rate.
	RATE SourceType = iota
	// DELTA is an ever-growing value which might be reset. The package calculates the difference between samples.
	DELTA SourceType = iota
	// ATTRIBUTE is any string value
	ATTRIBUTE SourceType = iota
)

const (
	// nsSeparator is the metric namespace separator
	nsSeparator = "::"
	// nsAttributeSeparator is the metric attribute key-value separator applied to generate the metric ns.
	nsAttributeSeparator = "=="
)

// Errors
var (
	ErrNonNumeric        = errors.New("non-numeric value for rate/delta")
	ErrNoStoreToCalcDiff = errors.New("cannot use deltas nor rates without persistent store")
	ErrTooCloseSamples   = errors.New("samples too close in time, skipping")
	ErrOverrideSetAttrs  = errors.New("cannot overwrite metric-set attributes")
	ErrDeltaWithNoAttrs  = errors.New("delta/rate metrics should be attached to an attribute identified metric-set")
)

// Set is the basic structure for storing metrics.
type Set struct {
	storer       persist.Storer
	Metrics      map[string]interface{}
	nsAttributes []Attribute
}

// NewSet creates new metrics set, optionally related to a list of attributes. These attributes makes the metric-set unique.
// If related attributes are used, then new attributes are added.
func NewSet(eventType string, storer persist.Storer, attributes ...Attribute) (s *Set) {
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

// Attr creates an attribute aimed to namespace a metric-set.
func Attr(key string, value string) Attribute {
	return Attribute{
		Key:   key,
		Value: value,
	}
}

// SetMetric adds a metric to the Set object or updates the metric value if the metric already exists.
// It calculates elapsed difference for RATE and DELTA types.
func (ms *Set) SetMetric(name string, value interface{}, sourceType SourceType) (err error) {
	var errElapsed error
	var newValue = value

	// Only sample metrics of numeric type
	switch sourceType {
	case RATE, DELTA:
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
	if sourceType == RATE {
		elapsed = elapsed / float64(duration)
	}

	return
}

// prefix a metric name with a namespace based on the alphabetical order of the set related attributes.
func (ms *Set) namespace(metricName string) string {
	ns := ""
	separator := ""

	attrs := ms.nsAttributes
	sort.Sort(Attributes(attrs))

	for _, attr := range attrs {
		ns = fmt.Sprintf("%s%s%s", ns, separator, attr.Namespace())
		separator = nsSeparator
	}

	return fmt.Sprintf("%s%s%s", ns, separator, metricName)
}

// Namespace generates the string value of an attribute used to namespace a metric.
func (a *Attribute) Namespace() string {
	return fmt.Sprintf("%s%s%s", a.Key, nsAttributeSeparator, a.Value)
}

// MarshalJSON adapts the internal structure of the metrics Set to the payload that is compliant with the protocol
func (ms Set) MarshalJSON() ([]byte, error) {
	return json.Marshal(ms.Metrics)
}

// Required for Go < v.18, as these do not include sort.Slice

// Attributes list of attributes
type Attributes []Attribute

// Len ...
func (a Attributes) Len() int { return len(a) }

// Swap ...
func (a Attributes) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// Less ...
func (a Attributes) Less(i, j int) bool {
	if a[i].Key == a[j].Key {
		return a[i].Value < a[j].Value
	}
	return a[i].Key < a[j].Key
}
