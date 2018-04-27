package inventory

import "sync"

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

// SetItem stores a value into the inventory data structure.
func (i Inventory) SetItem(key string, field string, value interface{}) {
	i.lock.Lock()
	defer i.lock.Unlock()

	if _, ok := i.items[key]; ok {
		i.items[key][field] = value
	} else {
		i.items[key] = Item{field: value}
	}
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

// NewInventory creates new inventory.
func NewInventory() *Inventory {
	return &Inventory{
		items: make(Items),
	}
}
