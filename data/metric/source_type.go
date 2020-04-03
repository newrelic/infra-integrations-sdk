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
	// COUNT is an ever increasing value
	COUNT SourceType = iota
	// SUMMARY is a composite value with avg, min, max sample count and sum
	SUMMARY SourceType = iota
)

// SourcesTypeToName metric sources list mapping its type to readable name.
var SourcesTypeToName = map[SourceType]string{
	GAUGE:   "gauge",
	COUNT:   "count",
	SUMMARY: "summary",
}

// SourcesNameToType metric sources list mapping its name to type.
var SourcesNameToType = map[string]SourceType{
	"gauge":   GAUGE,
	"count":   COUNT,
	"summary": SUMMARY,
}

// String fulfills stringer interface, returning empty string on invalid source types.
func (t SourceType) String() string {
	if s, ok := SourcesTypeToName[t]; ok {
		return s
	}

	return ""
}

// SourceTypeForName does a case insensitive conversion from a string to a SourceType.
// An error will be returned if no valid SourceType matched.
func SourceTypeForName(sourceTypeTag string) (SourceType, error) {
	if st, ok := SourcesNameToType[strings.ToLower(sourceTypeTag)]; ok {
		return st, nil
	}

	return 0, fmt.Errorf("metric: Unknown source_type %s", sourceTypeTag)
}
