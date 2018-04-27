package integration

import (
	"fmt"
	"sync"

	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/persist"
	"github.com/pkg/errors"
)

// Entity is the producer of the data. Entity could be a host, a container, a pod, or whatever unit of meaning.
type Entity struct {
	lock      sync.Mutex
	Metadata  *EntityMetadata   `json:"entity,omitempty"`
	Metrics   []*metric.Set     `json:"metrics"`
	Inventory *metric.Inventory `json:"inventory"`
	Events    []*metric.Event   `json:"events"`
	storer    persist.Storer
}

// EntityMetadata stores entity Metadata
type EntityMetadata struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// EntityID entity identifier
type EntityID string

// newDefaultEntity creates unique default entity without identifier (name & type)
func newDefaultEntity(storer persist.Storer) *Entity {
	return &Entity{
		// empty array or object preferred instead of null on marshaling.
		Metrics:   []*metric.Set{},
		Inventory: metric.NewInventory(),
		Events:    []*metric.Event{},
		storer:    storer,
	}
}

// newEntity creates a new remote-entity.
func newEntity(entityName, entityType string, storer persist.Storer) (*Entity, error) {
	// If one of the attributes is defined, both Name and Type are needed.
	if entityName == "" && entityType != "" || entityName != "" && entityType == "" {
		return nil, errors.New("entity name and type are required when defining one")
	}

	d := Entity{
		// empty array or object preferred instead of null on marshaling.
		Metrics:   []*metric.Set{},
		Inventory: metric.NewInventory(),
		Events:    []*metric.Event{},
		storer:    storer,
	}

	// Entity data is optional. When not specified, data from the integration is reported for the agent's own entity.
	if entityName != "" && entityType != "" {
		d.Metadata = &EntityMetadata{
			Name: entityName,
			Type: entityType,
		}
	}

	return &d, nil
}

// isDefaultEntity returns true if entity is the default one (has no identifier: name & type)
func (e *Entity) isDefaultEntity() bool {
	return e.Metadata == nil || (e.Metadata.Name == "" && e.Metadata.Type == "")
}

// NewMetricSet returns a new instance of Set with its sample attached to the integration.
func (e *Entity) NewMetricSet(eventType string) (s *metric.Set, err error) {
	s, err = metric.NewSet(eventType, e.storer)
	if err != nil {
		return
	}

	e.lock.Lock()
	defer e.lock.Unlock()
	e.Metrics = append(e.Metrics, s)
	return
}

// AddEvent method adds a new Event.
func (e *Entity) AddEvent(event *metric.Event) error {
	if event.Summary == "" {
		return errors.New("summary of the event cannot be empty")
	}

	e.lock.Lock()
	defer e.lock.Unlock()
	e.Events = append(e.Events, event)
	return nil
}

// SetInventoryItem method adds a inventory item.
func (e *Entity) SetInventoryItem(key string, field string, value interface{}) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.Inventory.SetItem(key, field, value)
}

// ID provides the entity id in string format
func (e *Entity) ID() EntityID {
	return EntityID(fmt.Sprintf("%s:%s", e.Metadata.Type, e.Metadata.Name))
}
