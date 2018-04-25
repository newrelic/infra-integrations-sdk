package metric

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestInventory_SetItem(t *testing.T) {
	i := NewInventory()
	i.SetItem("foo", "bar", "baz")

	assert.Equal(t, i.items["foo"]["bar"], "baz")
}
