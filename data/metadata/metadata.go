package metadata

import "reflect"

// Metadata stores entity Metadata. Serialized as "entity"
// TODO: not a fan of having the fields exposed
type Metadata struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	EntityType  string `json:"type"`
	Tags        TagMap `json:"tags"`
}

func New(name string, entityType string, displayName string) *Metadata {
	return &Metadata{
		Name:        name,
		DisplayName: displayName,
		EntityType:  entityType,
		Tags: map[string]interface{}{},
	}
}

// EqualsTo returns true when both metadata are equal.
func (m *Metadata) EqualsTo(b *Metadata) bool {
	if m.Name != b.Name || m.DisplayName != b.DisplayName || m.EntityType != b.EntityType {
		return false
	}

	return reflect.DeepEqual(m.GetTags(), b.GetTags())
}

// AddTag adds a tag to the entity
func (m *Metadata) AddTag(key string, value interface{}) {
	m.Tags[key] = value
}

func (m *Metadata) GetTag(key string) interface{} {
	return m.Tags[key]
}

func (m *Metadata) GetTags() TagMap {
	return m.Tags
}




