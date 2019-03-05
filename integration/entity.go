package integration

import (
	"errors"
	"sort"
	"sync"

	"github.com/newrelic/infra-integrations-sdk/data/event"
	"github.com/newrelic/infra-integrations-sdk/data/inventory"
	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/persist"
)

// Entity is the producer of the data. Entity could be a host, a container, a pod, or whatever unit of meaning.
type Entity struct {
	Metadata    *EntityMetadata      `json:"entity,omitempty"`
	Metrics     []*metric.Set        `json:"metrics"`
	Inventory   *inventory.Inventory `json:"inventory"`
	Events      []*event.Event       `json:"events"`
	AddHostname bool                 `json:"add_hostname,omitempty"` // add hostname to metadata at agent level
	storer      persist.Storer
	lock        sync.Locker
	// CustomAttributes []metric.Attribute `json:"custom_attributes,omitempty"`
	customAttributes []metric.Attribute
}

//SortedIDAttributes this sorted list ensures uniqueness for the entity Key.
type SortedIDAttributes []IDAttribute

// EntityMetadata stores entity Metadata
type EntityMetadata struct {
	Name      string             `json:"name"`
	Namespace string             `json:"type"`          // For compatibility reasons we keep the type.
	IDAttrs   SortedIDAttributes `json:"id_attributes"` // For entity Key uniqueness
}

// newLocalEntity creates unique default entity without identifier (name & type)
func newLocalEntity(storer persist.Storer, addHostnameToMetadata bool) *Entity {
	return &Entity{
		// empty array or object preferred instead of null on marshaling.
		Metrics:     []*metric.Set{},
		Inventory:   inventory.New(),
		Events:      []*event.Event{},
		AddHostname: addHostnameToMetadata,
		storer:      storer,
		lock:        &sync.Mutex{},
	}
}

// newEntity creates a new remote-entity with entity attributes.
func newEntity(
	name,
	namespace string,
	storer persist.Storer,
	addHostnameToMetadata bool,
	idAttrs ...IDAttribute,
) (*Entity, error) {

	// If one of the idAttr is defined, both Name and Namespace are needed.
	if name == "" || namespace == "" {
		return nil, errors.New("entity name and type are required when defining one")
	}

	d := Entity{
		// empty array or object preferred instead of null on marshaling.
		Metrics:     []*metric.Set{},
		Inventory:   inventory.New(),
		Events:      []*event.Event{},
		AddHostname: addHostnameToMetadata,
		storer:      storer,
		lock:        &sync.Mutex{},
	}

	// Entity data is optional. When not specified, data from the integration is reported for the agent's own entity.
	d.Metadata = &EntityMetadata{
		Name:      name,
		Namespace: namespace,
		IDAttrs:   sortIDAttributes(idAttrs...),
	}

	return &d, nil
}

//IDAttributesSorter ...
type IDAttributesSorter struct {
	IDAttributes SortedIDAttributes
}

// Len is part of sort.Interface.
func (s *IDAttributesSorter) Len() int {
	return len(s.IDAttributes)
}

// Swap is part of sort.Interface.
func (s *IDAttributesSorter) Swap(i, j int) {
	s.IDAttributes[i], s.IDAttributes[j] = s.IDAttributes[j], s.IDAttributes[i]
}

// Less is part of sort.Interface.
func (s *IDAttributesSorter) Less(i, j int) bool {
	return s.IDAttributes[i].Key < s.IDAttributes[j].Key
}

func sortIDAttributes(idAttrs ...IDAttribute) SortedIDAttributes {
	sorted := make(SortedIDAttributes, len(idAttrs))
	for i, attr := range idAttrs {
		sorted[i] = attr
	}
	sorter := &IDAttributesSorter{
		IDAttributes: sorted,
	}
	sort.Sort(sorter)

	return sorted
}

// isLocalEntity returns true if entity is the default one (has no identifier: name & type)
func (e *Entity) isLocalEntity() bool {
	return e.Metadata == nil || e.Metadata.Name == ""
}

// NewMetricSet returns a new instance of Set with its sample attached to the integration.
func (e *Entity) NewMetricSet(eventType string, nameSpacingAttributes ...metric.Attribute) *metric.Set {

	s := metric.NewSet(eventType, e.storer, nameSpacingAttributes...)

	if len(e.customAttributes) > 0 {
		metric.AddCustomAttributes(s, e.customAttributes)
	}

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

func (e *Entity) setCustomAttribute(key string, value string) {
	attribute := metric.Attribute{key, value}
	e.customAttributes = append(e.customAttributes, attribute)
}
