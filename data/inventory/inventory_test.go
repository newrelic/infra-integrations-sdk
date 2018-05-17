package inventory

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInventory_SetItem(t *testing.T) {
	i := New()
	i.SetItem("foo", "bar", "baz")

	assert.Equal(t, i.items["foo"]["bar"], "baz")
	// Test update already existing element
	i.SetItem("foo", "bar", "quux")
	assert.Equal(t, i.items["foo"]["bar"], "quux")

}

func TestInventory_Item(t *testing.T) {
	i := New()
	i.SetItem("foo", "bar", "baz")
	element, exists := i.Item("foo")
	assert.Equal(t, exists, true)
	assert.Equal(t, element["bar"], "baz")
}

func TestInventory_Items(t *testing.T) {
	i := New()
	// Add 4 elements
	i.SetItem("foo", "bar", "baz")
	i.SetItem("qux", "bar", "baz")
	i.SetItem("bar", "bar", "baz")
	i.SetItem("baz", "bar", "baz")

	assert.Equal(t, len(i.Items()), 4)
}

func TestInventory_SetItemForbidsLargeKeys(t *testing.T) {
	i := New()

	randStringWithLen := func(n int) string {
		var chars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

		b := make([]rune, n)
		for i := range b {
			b[i] = chars[rand.Intn(len(chars))]
		}
		return string(b)
	}

	assert.NoError(t, i.SetItem(randStringWithLen(MaxKeyLen), "foo", "bar"))
	assert.Error(t, i.SetItem(randStringWithLen(MaxKeyLen+1), "foo", "bar"))
}
