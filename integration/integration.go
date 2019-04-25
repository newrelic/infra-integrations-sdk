package integration

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"sync"

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

// Standard attributes
const (
	AttrReportingEntity   = "reportingEntityKey"
	AttrReportingEndpoint = "reportingEndpoint"
)

// NR infrastructure agent protocol version
const (
	protocolVersion = "3"
)

// Integration defines the format of the output JSON that integrations will return for protocol 2.
type Integration struct {
	Name               string    `json:"name"`
	ProtocolVersion    string    `json:"protocol_version"`
	IntegrationVersion string    `json:"integration_version"`
	Entities           []*Entity `json:"data"`
	addHostnameToMeta  bool
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
	i.addHostnameToMeta = defaultArgs.NriAddHostname

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

	return
}

// LocalEntity retrieves default (local) entity to monitorize.
func (i *Integration) LocalEntity() *Entity {
	i.locker.Lock()
	defer i.locker.Unlock()

	for _, e := range i.Entities {
		if e.isLocalEntity() {
			return e
		}
	}

	e := newLocalEntity(i.storer, i.addHostnameToMeta)

	i.Entities = append(i.Entities, e)

	return e
}

// EntityReportedBy entity being reported from another entity that is not producing the actual entity data.
func (i *Integration) EntityReportedBy(reportingEntity EntityKey, reportedEntityName, reportedEntityNamespace string, idAttributes ...IDAttribute) (e *Entity, err error) {
	e, err = i.Entity(reportedEntityName, reportedEntityNamespace, idAttributes...)
	if err != nil {
		return
	}

	e.setCustomAttribute(AttrReportingEntity, reportingEntity.String())
	return
}

// EntityReportedVia entity being reported from a known endpoint.
func (i *Integration) EntityReportedVia(endpoint, reportedEntityName, reportedEntityNamespace string, idAttributes ...IDAttribute) (e *Entity, err error) {
	e, err = i.Entity(reportedEntityName, reportedEntityNamespace, idAttributes...)
	if err != nil {
		return
	}

	e.setCustomAttribute(AttrReportingEndpoint, endpoint)
	return
}

// Entity method creates or retrieves an already created entity.
func (i *Integration) Entity(name, namespace string, idAttributes ...IDAttribute) (e *Entity, err error) {
	i.locker.Lock()
	defer i.locker.Unlock()

	e, err = newEntity(name, namespace, i.storer, i.addHostnameToMeta, idAttributes...)
	if err != nil {
		return nil, err
	}

	for _, eIt := range i.Entities {
		if e.SameAs(eIt) {
			return eIt, nil
		}
	}

	defaultArgs := args.GetDefaultArgs(i.args)

	if defaultArgs.Metadata {
		for _, element := range os.Environ() {
			variable := strings.Split(element, "=")
			prefix := fmt.Sprintf("%s%s_", CustomAttrPrefix, strings.ToUpper(i.Name))
			if strings.HasPrefix(variable[0], prefix) {
				e.setCustomAttribute(strings.TrimPrefix(variable[0], prefix), variable[1])
			}
		}
	}

	if defaultArgs.NriCluster != "" {
		e.setCustomAttribute(CustomAttrCluster, defaultArgs.NriCluster)
	}
	if defaultArgs.NriService != "" {
		e.setCustomAttribute(CustomAttrService, defaultArgs.NriService)
	}

	i.Entities = append(i.Entities, e)

	return e, nil
}

// Publish runs all necessary tasks before publishing the data. Currently, it
// stores the Storer, prints the JSON representation of the integration using a writer (stdout by default)
// and re-initializes the integration object (allowing re-use it during the
// execution of your code).
func (i *Integration) Publish() error {
	if i.storer != nil {
		if err := i.storer.Save(); err != nil {
			return err
		}
	}

	output, err := i.toJSON(i.prettyOutput)
	if err != nil {
		return err
	}
	output = append(output, []byte{'\n'}...)
	_, err = i.writer.Write(output)
	defer i.Clear()

	return err
}

// Clear re-initializes the Inventory, Metrics and Events for this integration.
// Used after publishing so the object can be reused.
func (i *Integration) Clear() {
	i.locker.Lock()
	defer i.locker.Unlock()
	i.Entities = []*Entity{} // empty array preferred instead of null on marshaling.
}

// MarshalJSON serializes integration to JSON, fulfilling Marshaler interface.
func (i *Integration) MarshalJSON() (output []byte, err error) {
	output, err = json.Marshal(*i)
	if err != nil {
		err = fmt.Errorf("error marshalling to JSON: %s", err)
	}

	return
}

// Logger returns the integration logger instance.
func (i *Integration) Logger() log.Logger {
	return i.logger
}

// toJSON serializes integration as JSON. If the pretty attribute is
// set to true, the JSON will be indented for easy reading.
func (i *Integration) toJSON(pretty bool) (output []byte, err error) {
	output, err = i.MarshalJSON()
	if !pretty {
		return
	}

	var buf bytes.Buffer
	err = json.Indent(&buf, output, "", "\t")
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
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
