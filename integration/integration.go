package integration

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/newrelic/infra-integrations-sdk/data/event"
	"github.com/newrelic/infra-integrations-sdk/data/metric"

	"github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/persist"
)

// Custom attribute keys:
const (
	CustomAttrPrefix  = "NRI_"
	CustomAttrCluster = "cluster_name"
	CustomAttrService = "service_name"
)

// NR infrastructure agent protocol version
const (
	protocolVersion = "4"
)

// Integration defines the format of the output JSON that integrations will return for protocol 2.
type Integration struct {
	Name               string    `json:"name"`
	ProtocolVersion    string    `json:"protocol_version"`
	IntegrationVersion string    `json:"integration_version"`
	Entities           []*Entity `json:"data"`
	anonEntity         *Entity   // anonEntity is the equivalent to "local entity", ie, no metadata.
	locker             sync.Locker
	storer             persist.Storer
	prettyOutput       bool
	writer             io.Writer
	logger             log.Logger
	args               interface{}
}

// New creates new integration with sane default values.
func New(name, version string, opts ...Option) (i *Integration, err error) {

	if name == "" {
		err = errors.New("integration name cannot be empty")
		return
	}

	if version == "" {
		err = errors.New("integration version cannot be empty")
		return
	}

	i = &Integration{
		Name:               name,
		ProtocolVersion:    protocolVersion,
		IntegrationVersion: version,
		Entities:           []*Entity{},
		writer:             os.Stdout,
		locker:             &sync.Mutex{},
	}

	for _, opt := range opts {
		if err = opt(i); err != nil {
			err = fmt.Errorf("error applying option to integration. %s", err)
			return
		}
	}

	// arguments
	if err = i.checkArguments(); err != nil {
		return
	}
	if err = args.SetupArgs(i.args); err != nil {
		return
	}
	defaultArgs := args.GetDefaultArgs(i.args)
	i.prettyOutput = defaultArgs.Pretty

	// Setting default values, if not set yet
	if i.logger == nil {
		i.logger = log.NewStdErr(defaultArgs.Verbose)
	}

	if i.storer == nil {
		var err error
		i.storer, err = persist.NewFileStore(persist.DefaultPath(i.Name), i.logger, persist.DefaultTTL)
		if err != nil {
			return nil, fmt.Errorf("can't create store: %s", err)
		}
	}

	i.anonEntity = newAnonymousEntity(i.storer)

	return
}

// NewEntity method creates a new (uniquely named) Entity.
// The `name` of the Entity must be unique for the account otherwise it will cause conflicts
func (i *Integration) NewEntity(name string, entityType string, displayName string) (e *Entity, err error) {
	i.locker.Lock()
	defer i.locker.Unlock()

	e, err = newEntity(name, entityType, displayName, i.storer)
	if err != nil {
		return nil, err
	}

	i.addDefaultAttributes(e)

	return e, nil
}

// AddEntity adds an entity to the list of entities. No check for "duplicates" is performed
func (i *Integration) AddEntity(e *Entity) {
	i.Entities = append(i.Entities, e)
}

// AddInventoryItem adds the item to the anonymous entity (if it exists, otherwise creates one)
// To add Inventory to an entity, use the AddInventoryItem method in an Entity instance
func (i *Integration) AddInventoryItem(key string, field string, value interface{}) error {
	return i.anonEntity.AddInventoryItem(key, field, value)
}

// NewEvent creates a new event
func (i *Integration) NewEvent(timestamp time.Time, summary string, category string) *event.Event {
	return event.New(timestamp, summary, category)
}

// AddEvent adds the specified event to the anonymous entity  (if it exists, otherwise creates one)
func (i *Integration) AddEvent(ev *event.Event) error {
	return i.anonEntity.AddEvent(ev)
}

// Publish runs all necessary tasks before publishing the data. Currently, it
// stores the Storer, prints the JSON representation of the integration using a writer (stdout by default)
// and re-initializes the integration object (allowing re-use it during the
// execution of your code).
func (i *Integration) Publish() error {
	defer i.Clear()
	if i.storer != nil {
		if err := i.storer.Save(); err != nil {
			return err
		}
	}

	// add the anon entity to the list of entity to be serialized, if not empty
	if notEmpty(i.anonEntity) {
		i.Entities = append(i.Entities, i.anonEntity)
	}
	output, err := i.toJSON(i.prettyOutput)
	if err != nil {
		return err
	}
	output = append(output, []byte{'\n'}...)
	_, err = i.writer.Write(output)

	return err
}

// Clear re-initializes the Inventory, Metrics and Events for this integration.
// Used after publishing so the object can be reused.
func (i *Integration) Clear() {
	i.locker.Lock()
	defer i.locker.Unlock()
	i.Entities = []*Entity{} // empty array preferred instead of null on marshaling.
	// reset the anon entity
	i.anonEntity = newAnonymousEntity(i.storer)
}

// MarshalJSON serializes integration to JSON, fulfilling Marshaler interface.
func (i *Integration) MarshalJSON() (output []byte, err error) {
	output, err = json.Marshal(*i)
	if err != nil {
		err = fmt.Errorf("error marshalling to JSON: %s", err)
	}

	return
}

// toJSON serializes integration as JSON. If the pretty attribute is
// set to true, the JSON will be indented for easy reading.
func (i *Integration) toJSON(pretty bool) (output []byte, err error) {
	if !pretty {
		return i.MarshalJSON()
	}

	return json.MarshalIndent(*i, "", "\t")
}

// Logger returns the integration logger instance.
func (i *Integration) Logger() log.Logger {
	return i.logger
}

// Gauge creates a metric of type gauge
func Gauge(timestamp time.Time, metricName string, value float64) metric.Metric {
	return metric.NewGauge(timestamp, metricName, value)
}

// PDelta creates a metric of type pdelta
func PDelta(timestamp time.Time, metricName string, value float64) metric.Metric {
	return metric.NewPDelta(timestamp, metricName, value)
}

// Count creates a metric of type count
func Count(timestamp time.Time, interval time.Duration, metricName string, value uint64) metric.Metric {
	return metric.NewCount(timestamp, interval, metricName, value)
}

// Summary creates a metric of type summary
func Summary(timestamp time.Time, interval time.Duration, metricName string, count uint64,
	average float64, sum float64, min float64, max float64) metric.Metric {
	return metric.NewSummary(timestamp, interval, metricName, count, average, sum, min, max)
}

// -- private
// is entity empty?
func notEmpty(entity *Entity) bool {
	return len(entity.Events) > 0 || len(entity.Metrics) > 0 || len(entity.Inventory.Items()) > 0
}

func (i *Integration) checkArguments() error {
	if i.args == nil {
		i.args = new(struct{})
		return nil
	}
	val := reflect.ValueOf(i.args)

	if val.Kind() == reflect.Ptr && val.Elem().Kind() == reflect.Struct {
		return nil
	}

	return errors.New("arguments must be a pointer to a struct (or nil)")
}

func (i *Integration) addDefaultAttributes(e *Entity) {
	defaultArgs := args.GetDefaultArgs(i.args)

	// get env vars values for "custom" prefixed vars (NRIA_) and add them as attributes to the entity
	if defaultArgs.Metadata {
		for _, element := range os.Environ() {
			variable := strings.Split(element, "=")
			prefix := fmt.Sprintf("%s%s_", CustomAttrPrefix, strings.ToUpper(i.Name))
			if strings.HasPrefix(variable[0], prefix) {
				e.AddTag(strings.TrimPrefix(variable[0], prefix), variable[1])
			}
		}
	}

	// TODO: should these custom "attributes" be added to "common", "metrics" and "events"?
	if defaultArgs.NriCluster != "" {
		e.AddTag(CustomAttrCluster, defaultArgs.NriCluster)
	}
	if defaultArgs.NriService != "" {
		e.AddTag(CustomAttrService, defaultArgs.NriService)
	}
}
