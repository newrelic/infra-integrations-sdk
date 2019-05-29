package metric

import (
	"errors"
	"fmt"
	"reflect"
)

const (
	// metricNameTag is the struct tag for specifying a metric name on a struct
	metricNameTag = "metric_name"

	// sourceTypeTag is the struct tag for specifying the SourceType of a metric.
	// Tag is case insensitive and should match the type names.
	sourceTypeTag = "source_type"
)

// MarshalMetrics creates metrics for primitive values of v.
//
// MarshalMetrics traverses the value of v recursively.
// Once a non-struct or non-pointer value is reached that has
// the accepted tags, SetMetric is then called with the field's value.
//
// Pointers are dereferenced until a base value is found or nil.
// If nil, the field is skipped regardless of if it had the appropriate
// struct field tags.
//
// Needed struct field tags are "metric_name" and "source_type". The value of
// "metric_name" will be the name argument to SetMetric. The value
// of "source_type" is case insensitively matched against values below to a SourceType
// and passed as the sourceType argument to SetMetric.
// If the value does not match one of the values below an error will be returned.
//   - gauge
//   - rate
//   - prate
//   - delta
//   - pdelta
//   - attribute
//
// If one of the required tags is missing an error will be returned.
// If both are missing SetMetric will not be called for the given field.
//
// Examples of struct field tags:
//   type Data struct {
//      Gauge     int     `metric_name:"metric.gauge" source_type:"Gauge"`
//      Attribute string  `metric_name:"metric.attribute" source_type:"attribute"`
//      Rate      float64 `metric_name:"metric.rate" source_type:"RATE"`
//      Delta     float64 `metric_name:"metric.delta" source_type:"delta"`
//      PRate     float64 `metric_name:"metric.prate" source_type:"prate"`
//      PDelta    float64 `metric_name:"metric.pdelta" source_type:"pdelta"`
//   }
//
// Any non-struct/non-pointer value that has the correct struct field tags
// will be passed to SetMetric. If the value causes an error to be returned
// from SetMetric this will be bubbled up and returned by MarshalMetrics.
//
// If a cyclic data structure is passed in this will result in
// infinite recursion.
func (ms *Set) MarshalMetrics(v interface{}) error {
	r := reflect.ValueOf(v)
	value := reflect.Indirect(r)

	if value.Kind() != reflect.Struct {
		return errors.New("metric: can only directly unmarshal structs")
	}
	return marshalStruct(value.Type(), value, ms)
}

// marshalValue takes in a struct field and does a kind switch on it to determine further
// marshaling.
func marshalValue(f reflect.StructField, v reflect.Value, ms *Set) error {
	switch v.Kind() {
	case reflect.Struct:
		return marshalStruct(v.Type(), v, ms)
	case reflect.Interface:
		fallthrough
	case reflect.Ptr:
		// If the pointer is nil we don't process it
		// regardless of if it had the correct struct field tags
		if v.IsNil() {
			return nil
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
	// Get struct field tag values
	metricName, hasMetricName := f.Tag.Lookup(metricNameTag)
	sourceTypeString, hasSourceType := f.Tag.Lookup(sourceTypeTag)

	// Validate that we have all the needed tag information to process the metric
	if !hasMetricName && !hasSourceType {
		return nil
	} else if hasMetricName != hasSourceType {
		return fmt.Errorf("metric: Field '%s' must have both %s and %s struct tags", f.Name, metricNameTag, sourceTypeTag)
	}

	// Convert source_type tag to a value
	sourceType, err := SourceTypeForName(sourceTypeString)
	if err != nil {
		return err
	}

	// Sets the metric, passing a good deal of additional error handling onto this function as
	// it already handles type checking per sourceType.
	return ms.SetMetric(metricName, v.Interface(), sourceType)
}
