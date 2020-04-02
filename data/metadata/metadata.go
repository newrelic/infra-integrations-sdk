package metadata

import "reflect"

// Metadata stores entity Metadata. Serialized as "entity"
type Metadata struct {
	// TODO: not a fan of having the fields exposed
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	EntityType  string `json:"type"`
	Tags        TagMap `json:"tags"`
}

// New creates a new metadata section that "identifies" an entity
func New(name string, entityType string, displayName string) *Metadata {
	return &Metadata{
		Name:        name,
		DisplayName: displayName,
		EntityType:  entityType,
		Tags:        map[string]interface{}{},
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

// GetTag gets a specific tag from the metadata section of an entity
func (m *Metadata) GetTag(key string) interface{} {
	return m.Tags[key]
}

// GetTags gets all the tags added to the metadata section of an entity
func (m *Metadata) GetTags() TagMap {
	return m.Tags
}
