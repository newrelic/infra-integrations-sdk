package inventory

import (
	"encoding/json"
	"fmt"
	"sync"
)

// Limits
const (
	MaxKeyLen = 375
)

// Items ...
type Items map[string]Item

// Item ...
type Item map[string]interface{}

// Inventory is the data type for inventory data produced by an integration data
// source and emitted to the agent's inventory data store.
type Inventory struct {
	items Items
	lock  sync.Mutex
}

// MarshalJSON Marshals the items map into a JSON
func (i Inventory) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.items)
}

// SetItem stores a value into the inventory, key is limited to 375 characters.
func (i Inventory) SetItem(key string, field string, value interface{}) error {
	if len(key) > MaxKeyLen {
		return fmt.Errorf("maximum inventory key length is %d, current key %s has %d characters", MaxKeyLen, key, len(key))
	}

	i.lock.Lock()
	defer i.lock.Unlock()

	if _, ok := i.items[key]; ok {
		i.items[key][field] = value
	} else {
		i.items[key] = Item{field: value}
	}

	return nil
}

// Item returns stored item
func (i Inventory) Item(key string) (item Item, exists bool) {
	item, exists = i.items[key]
	return
}

// Items returns all stored items
func (i Inventory) Items() Items {
	return i.items
}

// New creates new inventory.
func New() *Inventory {
	return &Inventory{
		items: make(Items),
	}
}
