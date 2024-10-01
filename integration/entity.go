package integration

import (
	"errors"
	"sync"

	"github.com/newrelic/infra-integrations-sdk/v3/data/attribute"
	"github.com/newrelic/infra-integrations-sdk/v3/data/event"
	"github.com/newrelic/infra-integrations-sdk/v3/data/inventory"
	"github.com/newrelic/infra-integrations-sdk/v3/data/metric"
	"github.com/newrelic/infra-integrations-sdk/v3/persist"
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
	customAttributes []attribute.Attribute
}

// EntityMetadata stores entity Metadata
type EntityMetadata struct {
	Name      string       `json:"name"`
	Namespace string       `json:"type"`          // For compatibility reasons we keep the type.
	IDAttrs   IDAttributes `json:"id_attributes"` // For entity Key uniqueness
	lock      sync.Locker
}

// EqualsTo returns true when both metadata are equal.
func (m *EntityMetadata) EqualsTo(b *EntityMetadata) bool {
	// prevent checking on Key() for performance
	if m.Name != b.Name || m.Namespace != b.Namespace {
		return false
	}

	k1, err := m.Key()
	if err != nil {
		return false
	}

	k2, err := b.Key()
	if err != nil {
		return false
	}

	return k1.String() == k2.String()
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
		Metadata: &EntityMetadata{
			Name:      name,
			Namespace: namespace,
			IDAttrs:   idAttributes(idAttrs...),
			lock:      &sync.Mutex{},
		},
	}

	return &d, nil
}

// isLocalEntity returns true if entity is the default one (has no identifier: name & type)
func (e *Entity) isLocalEntity() bool {
	return e.Metadata == nil || e.Metadata.Name == ""
}

// SameAs return true when is same entity
func (e *Entity) SameAs(b *Entity) bool {
	if e.Metadata == nil || b.Metadata == nil {
		return false
	}
	return e.Metadata.EqualsTo(b.Metadata)
}

// NewMetricSet returns a new instance of Set with its sample attached to the integration.
func (e *Entity) NewMetricSet(eventType string, nameSpacingAttributes ...attribute.Attribute) *metric.Set {
	s := metric.NewSet(eventType, e.storer, nameSpacingAttributes...)

	if e.Metadata != nil {
		if key, err := e.Metadata.Key(); err == nil {
			s.AddNamespaceAttributes(attribute.Attr("entityKey", key.String()))
		}
	}

	if len(e.customAttributes) > 0 {
		metric.AddCustomAttributes(s, e.customAttributes)
	}

	e.lock.Lock()
	defer e.lock.Unlock()
	e.Metrics = append(e.Metrics, s)
	return s
}

// AddEvent method adds a new Event.
func (e *Entity) AddEvent(evnt *event.Event) error {
	if evnt.Summary == "" {
		return errors.New("summary of the event cannot be empty")
	}

	if len(e.customAttributes) > 0 {
		event.AddCustomAttributes(evnt, e.customAttributes)
	}

	e.lock.Lock()
	defer e.lock.Unlock()
	e.Events = append(e.Events, evnt)
	return nil
}

// SetInventoryItem method adds a inventory item.
func (e *Entity) SetInventoryItem(key string, field string, value interface{}) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	return e.Inventory.SetItem(key, field, value)
}

// AddAttributes adds attributes to every entity metric-set.
func (e *Entity) AddAttributes(attributes ...attribute.Attribute) {
	for _, a := range attributes {
		e.setCustomAttribute(a.Key, a.Value)
	}
}

func (e *Entity) setCustomAttribute(key string, value string) {
	attribute := attribute.Attribute{
		Key:   key,
		Value: value,
	}
	e.customAttributes = append(e.customAttributes, attribute)
}

// Key unique entity identifier within a New Relic customer account.
func (e *Entity) Key() (EntityKey, error) {
	return e.Metadata.Key()
}
