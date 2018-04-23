package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEntity(t *testing.T) {
	e, err := NewEntity("name", "type")

	assert.NoError(t, err)
	assert.Equal(t, "name", e.Metadata.Name)
	assert.Equal(t, "type", e.Metadata.Type)
}
