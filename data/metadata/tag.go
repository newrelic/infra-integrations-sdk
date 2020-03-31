package metadata

//Tags list of tags that will can be added as dimensions.
type Tags []Tag

// Tag is key-value pair.
type Tag struct {
	Key   string
	Value string
}

// Len is part of sort.Interface.
func (a Tags) Len() int {
	return len(a)
}

// Swap is part of sort.Interface.
func (a Tags) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// Less is part of sort.Interface.
func (a Tags) Less(i, j int) bool {
	return a[i].Key < a[j].Key
}

// NewTag creates new identifier attribute.
func NewTag(key string, value string) Tag {
	return Tag{
		Key:   key,
		Value: value,
	}
}
