package integration

import (
	"errors"
	"sync"

	"github.com/newrelic/infra-integrations-sdk/data/event"
	"github.com/newrelic/infra-integrations-sdk/data/inventory"
	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/persist"
)

// Entity is the producer of the data. Entity could be a host, a container, a pod, or whatever unit of meaning.
type Entity struct {
	Metadata  *EntityMetadata      `json:"entity,omitempty"`
	Metrics   []*metric.Set        `json:"metrics"`
	Inventory *inventory.Inventory `json:"inventory"`
	Events    []*event.Event       `json:"events"`
	storer    persist.Storer
	lock      sync.Locker
}

// EntityMetadata stores entity Metadata
type EntityMetadata struct {
	Name      string `json:"name"`
	Namespace string `json:"type"` // For compatibility reasons we keep the type.
}

// newLocalEntity creates unique default entity without identifier (name & type)
func newLocalEntity(storer persist.Storer) *Entity {
	return &Entity{
		// empty array or object preferred instead of null on marshaling.
		Metrics:   []*metric.Set{},
		Inventory: inventory.New(),
		Events:    []*event.Event{},
		storer:    storer,
		lock:      &sync.Mutex{},
	}
}

// newEntity creates a new remote-entity.
func newEntity(name, namespace string, storer persist.Storer) (*Entity, error) {
	// If one of the attributes is defined, both Name and Namespace are needed.
	if name == "" && namespace != "" || name != "" && namespace == "" {
		return nil, errors.New("entity name and type are required when defining one")
	}

	d := Entity{
		// empty array or object preferred instead of null on marshaling.
		Metrics:   []*metric.Set{},
		Inventory: inventory.New(),
		Events:    []*event.Event{},
		storer:    storer,
		lock:      &sync.Mutex{},
	}

	// Entity data is optional. When not specified, data from the integration is reported for the agent's own entity.
	if name != "" && namespace != "" {
		d.Metadata = &EntityMetadata{
			Name:      name,
			Namespace: namespace,
		}
	}

	return &d, nil
}

// isLocalEntity returns true if entity is the default one (has no identifier: name & type)
func (e *Entity) isLocalEntity() bool {
	return e.Metadata == nil || e.Metadata.Name == ""
}

// NewMetricSet returns a new instance of Set with its sample attached to the integration.
func (e *Entity) NewMetricSet(eventType string, nameSpacingAttributes ...metric.Attribute) *metric.Set {
	s := metric.NewSet(eventType, e.storer, nameSpacingAttributes...)

	e.lock.Lock()
	defer e.lock.Unlock()
	e.Metrics = append(e.Metrics, s)
	return s
}

// AddEvent method adds a new Event.
func (e *Entity) AddEvent(event *event.Event) error {
	if event.Summary == "" {
		return errors.New("summary of the event cannot be empty")
	}

	e.lock.Lock()
	defer e.lock.Unlock()
	e.Events = append(e.Events, event)
	return nil
}

// SetInventoryItem method adds a inventory item.
func (e *Entity) SetInventoryItem(key string, field string, value interface{}) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	return e.Inventory.SetItem(key, field, value)
}
