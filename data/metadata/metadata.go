package metadata

import (
	"fmt"
	"reflect"
	"strings"
)

const tagsPrefix = "tags."

// Map stores the tags for the entity
type Map map[string]interface{}

// Metadata stores entity Metadata. Serialized as "entity"
type Metadata struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	EntityType  string `json:"type"`
	Metadata    Map    `json:"metadata"`
}

// New creates a new metadata section that "identifies" an entity
func New(name string, entityType string, displayName string) *Metadata {
	return &Metadata{
		Name:        name,
		DisplayName: displayName,
		EntityType:  entityType,
		Metadata:    Map{},
	}
}

// EqualsTo returns true when both metadata are equal.
func (m *Metadata) EqualsTo(b *Metadata) bool {
	if m.Name != b.Name || m.DisplayName != b.DisplayName || m.EntityType != b.EntityType {
		return false
	}

	return reflect.DeepEqual(m.Metadata, b.Metadata)
}

// AddMetadata adds a generic metadata to the entity
func (m *Metadata) AddMetadata(key string, value interface{}) {
	m.Metadata[key] = value
}

// GetMetadata gets a specific metadata from the metadata section of an entity
func (m *Metadata) GetMetadata(key string) interface{} {
	return m.Metadata[key]
}

// AddTag adds a tag to the entity
func (m *Metadata) AddTag(key string, value interface{}) {
	m.Metadata[applyTagsPrefix(key)] = value
}

// GetTag gets a specific tag from the metadata section of an entity
func (m *Metadata) GetTag(key string) interface{} {
	return m.Metadata[applyTagsPrefix(key)]
}

func applyTagsPrefix(s string) string {
	if ok := strings.HasPrefix(s, tagsPrefix); !ok {
		s = fmt.Sprintf("%s%s", tagsPrefix, s)
	}
	return s
}
