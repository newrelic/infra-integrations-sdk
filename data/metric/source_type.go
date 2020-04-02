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
	GAUGE   SourceType = iota
	COUNT   SourceType = iota
	SUMMARY SourceType = iota
	PDELTA  SourceType = iota
)

// SourcesTypeToName metric sources list mapping its type to readable name.
var SourcesTypeToName = map[SourceType]string{
	GAUGE:   "gauge",
	COUNT:   "count",
	SUMMARY: "summary",
	PDELTA:  "pdelta",
}

// SourcesNameToType metric sources list mapping its name to type.
var SourcesNameToType = map[string]SourceType{
	"gauge":   GAUGE,
	"count":   COUNT,
	"summary": SUMMARY,
	"pdelta":  PDELTA,
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
	return t == COUNT || t == PDELTA
}

// SourceTypeForName does a case insensitive conversion from a string to a SourceType.
// An error will be returned if no valid SourceType matched.
func SourceTypeForName(sourceTypeTag string) (SourceType, error) {
	if st, ok := SourcesNameToType[strings.ToLower(sourceTypeTag)]; ok {
		return st, nil
	}

	return 0, fmt.Errorf("metric: Unknown source_type %s", sourceTypeTag)
}
