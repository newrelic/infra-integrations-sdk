package attribute

import "fmt"

const (
	// nsAttributeSeparator is the metric attribute key-value separator applied to generate the metric ns.
	nsAttributeSeparator = "=="
)

// Attribute represents a metric attribute key-value pair.
type Attribute struct {
	Key   string
	Value string
}

// Attributes list of attributes
type Attributes []Attribute

// Required for Go < v.18, as these do not include sort.Slice

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

// Namespace generates the string value of an attribute used to namespace a metric.
func (a *Attribute) Namespace() string {
	return fmt.Sprintf("%s%s%s", a.Key, nsAttributeSeparator, a.Value)
}

// Attr creates an attribute aimed to namespace a metric-set.
func Attr(key string, value string) Attribute {
	return Attribute{
		Key:   key,
		Value: value,
	}
}
