package inventory

import "sync"

// InventoryItems ...
type InventoryItems map[string]InventoryItem

// InventoryItem ...
type InventoryItem map[string]interface{}

// Inventory is the data type for inventory data produced by an integration data
// source and emitted to the agent's inventory data store.
type Inventory struct {
	items InventoryItems
	lock  sync.Mutex
}

// SetItem stores a value into the inventory data structure.
func (i Inventory) SetItem(key string, field string, value interface{}) {
	i.lock.Lock()
	defer i.lock.Unlock()

	if _, ok := i.items[key]; ok {
		i.items[key][field] = value
	} else {
		i.items[key] = InventoryItem{field: value}
	}
}

// Item returns stored item
func (i Inventory) Item(key string) (item InventoryItem, exists bool) {
	item, exists = i.items[key]
	return
}

// Items returns all stored items
func (i Inventory) Items() InventoryItems {
	return i.items
}

// NewInventory creates new inventory.
func NewInventory() *Inventory {
	return &Inventory{
		items: make(InventoryItems),
	}
}
