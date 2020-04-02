package attribute

const (
	// nsAttributeSeparator is the metric attribute key-value separator applied to generate the metric ns.
	nsAttributeSeparator = "=="
)

// Attributes list of attributes
type Attributes map[string]string

// Len ...
func (a Attributes) Len() int { return len(a) }

