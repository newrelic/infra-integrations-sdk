package metadata

import "reflect"

// Metadata stores entity Metadata. Serialized as "entity"
type Metadata struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	EntityType  string `json:"type"`
	Tags        Tags   `json:"tags"`
}

// EqualsTo returns true when both metadata are equal.
func (m *Metadata) EqualsTo(b *Metadata) bool {
	if m.Name != b.Name || m.DisplayName != b.DisplayName || m.EntityType != b.EntityType {
		return false
	}

	return reflect.DeepEqual(m.Tags, b.Tags)
}

// AddTag adds a tag to the entity
func (m *Metadata) AddTag(key string, value string) {
	// maybe lock?
	m.Tags = append(m.Tags, Tag{Key: key, Value: value})
}
