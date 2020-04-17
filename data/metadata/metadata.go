package metadata

import (
	"fmt"
	"reflect"
	"strings"
)

const TagsPrefix = "tags."

// MetadataMap stores the tags for the entity
type MetadataMap map[string]interface{}

// Metadata stores entity Metadata. Serialized as "entity"
type Metadata struct {
	// TODO: not a fan of having the fields exposed
	Name        string      `json:"name"`
	DisplayName string      `json:"displayName"`
	EntityType  string      `json:"type"`
	Metadata    MetadataMap `json:"metadata"`
}

// New creates a new metadata section that "identifies" an entity
func New(name string, entityType string, displayName string) *Metadata {
	return &Metadata{
		Name:        name,
		DisplayName: displayName,
		EntityType:  entityType,
		Metadata:    MetadataMap{},
	}
}

// EqualsTo returns true when both metadata are equal.
func (m *Metadata) EqualsTo(b *Metadata) bool {
	if m.Name != b.Name || m.DisplayName != b.DisplayName || m.EntityType != b.EntityType {
		return false
	}

	return reflect.DeepEqual(m.GetMetadataMap(), b.GetMetadataMap())
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

// GetMetadataMap gets all the tags added to the metadata section of an entity
func (m *Metadata) GetMetadataMap() MetadataMap {
	return m.Metadata
}

func applyTagsPrefix(s string) string {
	if ok := strings.HasPrefix(s, TagsPrefix); !ok {
		s = fmt.Sprintf("%s%s", TagsPrefix, s)
	}
	return s
}