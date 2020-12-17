package integration

import (
	"errors"
	"sync"

	"github.com/newrelic/infra-integrations-sdk/v4/data/event"
	"github.com/newrelic/infra-integrations-sdk/v4/data/inventory"
	"github.com/newrelic/infra-integrations-sdk/v4/data/metadata"
	"github.com/newrelic/infra-integrations-sdk/v4/data/metric"
)

// Entity is the producer of the data. Entity could be a host, a container, a pod, or whatever unit of meaning.
type Entity struct {
	Common    *Common              `json:"common"`
	Metadata  *metadata.Metadata   `json:"entity,omitempty"`
	Metrics   metric.Metrics       `json:"metrics"`
	Inventory *inventory.Inventory `json:"inventory"`
	Events    event.Events         `json:"events"`
	lock      sync.Locker
}

// Common is a common set of attributes
type Common struct{}

// SameAs return true when is same entity
func (e *Entity) SameAs(b *Entity) bool {
	if e.Metadata == nil || b.Metadata == nil {
		return false
	}

	return e.Metadata.EqualsTo(b.Metadata)
}

// AddMetric adds a new metric to the entity metrics list
func (e *Entity) AddMetric(metric metric.Metric) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.Metrics = append(e.Metrics, metric)
}

// AddEvent method adds a new Event.
func (e *Entity) AddEvent(event *event.Event) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.Events = append(e.Events, event)
}

// AddInventoryItem method sets the inventory item (only one allowed).
func (e *Entity) AddInventoryItem(key string, field string, value interface{}) error {
	if len(key) == 0 || len(field) == 0 {
		return errors.New("key or field cannot be empty")
	}
	e.lock.Lock()
	defer e.lock.Unlock()
	return e.Inventory.SetItem(key, field, value)
}

// GetMetadata returns all the Entity's metadata
func (e *Entity) GetMetadata() metadata.Map {
	return e.Metadata.Metadata
}

// AddTag adds a new tag to the entity
func (e *Entity) AddTag(key string, value interface{}) error {
	if len(key) == 0 {
		return errors.New("key cannot be empty")
	}
	e.Metadata.AddTag(key, value)
	return nil
}

// AddMetadata adds a new metadata to the entity
func (e *Entity) AddMetadata(key string, value interface{}) error {
	if len(key) == 0 {
		return errors.New("key cannot be empty")
	}
	e.Metadata.AddMetadata(key, value)
	return nil
}

// Name is the unique entity identifier within a New Relic customer account.
func (e *Entity) Name() string {
	return e.Metadata.Name
}

//--- private

// newHostEntity creates a entity without metadata.
func newHostEntity() *Entity {
	return &Entity{
		Common:   &Common{},
		Metadata: nil,
		// empty array or object preferred instead of null on marshaling.
		Metrics:   metric.Metrics{},
		Inventory: inventory.New(),
		Events:    event.Events{},
		lock:      &sync.Mutex{},
	}
}

// isHostEntity returns true if entity has no metadata
func (e *Entity) isHostEntity() bool {
	return e.Metadata == nil || e.Metadata.Name == ""
}

// newEntity creates a new entity with with metadata.
func newEntity(
	name,
	entityType string,
	displayName string) (*Entity, error) {

	if name == "" || entityType == "" {
		return nil, errors.New("entity name and type cannot be empty")
	}

	e := &Entity{
		// empty array or object preferred instead of null on marshaling.
		Common:    &Common{},
		Metadata:  metadata.New(name, entityType, displayName),
		Metrics:   metric.Metrics{},
		Inventory: inventory.New(),
		Events:    []*event.Event{},
		lock:      &sync.Mutex{},
	}
	return e, nil
}
