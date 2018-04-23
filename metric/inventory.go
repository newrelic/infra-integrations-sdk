package metric

type inventoryItem map[string]interface{}

// Inventory is the data type for inventory data produced by an integration data
// source and emitted to the agent's inventory data store.
type Inventory map[string]inventoryItem

// SetItem stores a value into the inventory data structure.
func (i Inventory) SetItem(key string, field string, value interface{}) {
	if _, ok := i[key]; ok {
		i[key][field] = value
	} else {
		i[key] = inventoryItem{field: value}
	}

}

// NewInventory creates new inventory.
func NewInventory() Inventory {
	return make(Inventory)
}
