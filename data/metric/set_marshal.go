package metric

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

const (
	// metricNameTag is the struct tag for specifying a metric name on a struct
	metricNameTag = "metric_name"

	// sourceTypeTag is the struct tag for specifying the SourceType of a metric.
	// Tag is case insensitive and should match the type names.
	sourceTypeTag = "source_type"
)

func (ms *Set) MarshalMetrics(data interface{}) error {
	t := reflect.TypeOf(data)
	r := reflect.ValueOf(data)
	value := reflect.Indirect(r)

	if value.Kind() != reflect.Struct {
		return errors.New("metric: can only directly unmarshal structs")
	}
	return marshalStruct(t, value, ms)
}

// marshalValue takes in a struct field and does a kind switch on it to determine further
// marshaling.
func marshalValue(f reflect.StructField, v reflect.Value, ms *Set) error {
	switch v.Kind() {
	case reflect.Struct:
		return marshalStruct(v.Type(), v, ms)
	case reflect.Ptr:
		if v.IsNil() {
			return fmt.Errorf("metric: Marshal(nil %s)", f.Type.String())
		}

		return marshalValue(f, v.Elem(), ms)
	default:
		return marshalField(f, v, ms)
	}
}

// marshalStruct marshals a struct into it's separate fields
func marshalStruct(t reflect.Type, v reflect.Value, ms *Set) error {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fv := v.FieldByName(f.Name)
		if err := marshalValue(f, fv, ms); err != nil {
			return err
		}
	}

	return nil
}

// marshalField marshals a struct field into a metric if both metric tags
// are present.
func marshalField(f reflect.StructField, v reflect.Value, ms *Set) error {
	// Get struct tag values
	metricName, hasMetricName := f.Tag.Lookup(metricNameTag)
	sourceTypeString, hasSourceType := f.Tag.Lookup(sourceTypeTag)

	// Validate that we have all the needed tag information to process the metric
	if !hasMetricName && !hasSourceType {
		return nil
	} else if hasMetricName != hasSourceType {
		return fmt.Errorf("metric: Field '%s' must have both %s and %s struct tags", f.Name, metricNameTag, sourceTypeTag)
	}

	// Convert source_type tag to a value
	sourceType, err := parseSourceType(sourceTypeString)
	if err != nil {
		return err
	}

	// Sets the metric, passing a good deal of additional error handling onto this function as
	// it already handles type checking per sourceType.
	return ms.SetMetric(metricName, v.Interface(), sourceType)
}

// parseSourceType does a case insensitive conversion from a string
// to a SourceType. An error will be returned if no valid SourceType matched.
func parseSourceType(sourceTypeTag string) (SourceType, error) {
	switch strings.ToLower(sourceTypeTag) {
	case "attribute":
		return ATTRIBUTE, nil
	case "rate":
		return RATE, nil
	case "delta":
		return DELTA, nil
	case "gauge":
		return GAUGE, nil
	default:
		return 0, fmt.Errorf("metric: Unknown source_type %s", sourceTypeTag)
	}
}
