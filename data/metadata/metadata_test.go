package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Metadata_AddMetadataAndTags(t *testing.T) {
	// given a new entity with no metadata
	m := New("foobar", "test", "test nmetadata")

	// when we add metadata and a tag
	m.AddMetadata("m1", 55)
	m.AddTag("t1", "hello there")

	// then we check that the metadata hasn't changed
	_, ok := m.Metadata["m1"]
	assert.True(t, ok)

	// then we check that the tag has been prefixed correctly
	_, ok = m.Metadata[tagsPrefix+"t1"]
	assert.True(t, ok)
}
