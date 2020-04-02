package inventory

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Inventory_SetItemAddsInventoryItem(t *testing.T) {
	i := New()
	err := i.SetItem("foo", "bar", "baz")
	assert.NoError(t, err)

	err = i.SetItem("foo1", "bar1", "baz1")
	assert.NoError(t, err)

	assert.Len(t, i.items, 2)
	assert.Equal(t, i.items["foo"]["bar"], "baz")
	assert.Equal(t, i.items["foo1"]["bar1"], "baz1")
}

func Test_Inventory_SetItemWithSameKeyUpdatesExisting(t *testing.T) {
	i := New()
	err := i.SetItem("foo", "bar", "baz")
	assert.NoError(t, err)
	assert.Equal(t, i.items["foo"]["bar"], "baz")

	// updates already existing element
	err = i.SetItem("foo", "bar", "quux")
	assert.NoError(t, err)
	assert.Equal(t, i.items["foo"]["bar"], "quux")

}

func Test_Inventory_GetItemByKey(t *testing.T) {
	i := New()
	_ = i.SetItem("foo", "bar", "baz")
	element, exists := i.Item("foo")
	assert.Equal(t, exists, true)
	assert.Equal(t, element["bar"], "baz")
}

func Test_Inventory_GetAllItems(t *testing.T) {
	i := New()
	// Add 4 elements
	_ = i.SetItem("foo", "bar", "baz")
	_ = i.SetItem("qux", "bar", "baz")
	_ = i.SetItem("bar", "bar", "baz")
	_ = i.SetItem("baz", "bar", "baz")

	assert.Equal(t, len(i.Items()), 4)
}

func Test_Inventory_SetItemForbidsLargeKeys(t *testing.T) {
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
