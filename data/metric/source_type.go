package metric

import (
	"fmt"
	"strings"
)

// SourceType defines the kind of data source. Based on this SourceType, metric
// package performs some calculations with it. Check below the description for
// each one.
type SourceType int

// Source types
// If any more SourceTypes are added update maps: SourcesTypeToName & SourcesNameToType.
const (
	// GAUGE is a value that may increase and decrease. It is stored as-is.
	GAUGE SourceType = iota
	// RATE is an ever-growing value which might be reset. The package calculates the change rate.
	RATE SourceType = iota
	// DELTA is an ever-growing value which might be reset. The package calculates the difference between samples.
	DELTA SourceType = iota
	// ATTRIBUTE is any string value
	ATTRIBUTE SourceType = iota
	// PRATE is a version of RATE that only allows positive values.
	PRATE SourceType = iota
	// PDELTA is a version of DELTA that only allows positive values.
	PDELTA SourceType = iota
)

// SourcesTypeToName metric sources list mapping its type to readable name.
var SourcesTypeToName = map[SourceType]string{
	GAUGE:     "gauge",
	RATE:      "rate",
	DELTA:     "delta",
	ATTRIBUTE: "attribute",
}

// SourcesNameToType metric sources list mapping its name to type.
var SourcesNameToType = map[string]SourceType{
	"gauge":     GAUGE,
	"rate":      RATE,
	"delta":     DELTA,
	"prate":     PRATE,
	"pdelta":    PDELTA,
	"attribute": ATTRIBUTE,
}

// String fulfills stringer interface, returning empty string on invalid source types.
func (t SourceType) String() string {
	if s, ok := SourcesTypeToName[t]; ok {
		return s
	}

	return ""
}

// IsPositive checks that the `SourceType` belongs to the positive only
// list of `SourceType`s
func (t SourceType) IsPositive() bool {
	return t == PRATE || t == PDELTA
}

// SourceTypeForName does a case insensitive conversion from a string to a SourceType.
// An error will be returned if no valid SourceType matched.
func SourceTypeForName(sourceTypeTag string) (SourceType, error) {
	if st, ok := SourcesNameToType[strings.ToLower(sourceTypeTag)]; ok {
		return st, nil
	}

	return 0, fmt.Errorf("metric: Unknown source_type %s", sourceTypeTag)
}
