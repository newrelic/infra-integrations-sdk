package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"sync"

	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/persist"
	"github.com/pkg/errors"
	"gopkg.in/newrelic/infra-integrations-sdk.v2/args"
)

// Integration defines the format of the output JSON that integrations will return for protocol 2.
type Integration struct {
	Name               string    `json:"name"`
	ProtocolVersion    string    `json:"protocol_version"`
	IntegrationVersion string    `json:"integration_version"`
	Entities           []*Entity `json:"data"`
	locker             sync.Locker
	storer             persist.Storer
	prettyOutput       bool
	writer             io.Writer
	logger             log.Logger
	args               interface{} // UGLY
}

// Option sets an option on integration level.
type Option func(*Integration) error

// New creates new integration with sane default values.
func New(name, version string, opts ...Option) (*Integration, error) {

	// TODO check for empty name and version

	i := &Integration{
		Name:               name,
		ProtocolVersion:    protocolVersion,
		IntegrationVersion: version,
		Entities:           []*Entity{},
		writer:             os.Stdout, // defaults to stdout
		locker:             disabledLocker{},
		logger:             log.NewStdErr(false),
	}

	for _, o := range opts {
		err := o(i)
		if err != nil {
			return i, fmt.Errorf("error applying option to integration. %s", err)
		}
	}

	if err := i.checkArguments(); err != nil {
		return i, err
	}

	if i.storer == nil {
		var err error
		i.storer, err = persist.NewFileStore(persist.DefaultPath(i.Name), i.logger)
		if err != nil {
			return nil, fmt.Errorf("can't create store: %s", err)
		}
	}

	err := args.SetupArgs(i.args)
	if err != nil {
		return nil, err
	}

	defaultArgs := args.GetDefaultArgs(i.args)
	i.prettyOutput = defaultArgs.Pretty

	return i, nil
}

// DefaultEntity retrieves default entity to monitorize.
func (i *Integration) DefaultEntity() *Entity {
	i.locker.Lock()
	defer i.locker.Unlock()

	for _, e := range i.Entities {
		if e.IsDefaultEntity() {
			return e
		}
	}

	e := newDefaultEntity(i.storer)

	i.Entities = append(i.Entities, e)

	return e
}

// Entity method creates or retrieves an already created entity.
func (i *Integration) Entity(entityName, entityType string) (e *Entity, err error) {
	i.locker.Lock()
	defer i.locker.Unlock()

	// we should change this to map for performance
	for _, e = range i.Entities {
		if e.Metadata.Name == entityName && e.Metadata.Type == entityType {
			return e, nil
		}
	}

	e, err = newEntity(entityName, entityType, i.storer)
	if err != nil {
		return nil, err
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
		err = errors.Wrap(err, "error marshalling to JSON")
	}

	return
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

func Writer(w io.Writer) Option {
	return func(i *Integration) error {
		i.writer = w

		return nil
	}
}

func Logger(l log.Logger) Option {
	return func(i *Integration) error {
		i.logger = l

		return nil
	}
}

func Storer(s persist.Storer) Option {
	return func(i *Integration) error {
		i.storer = s

		return nil
	}
}

func InMemoryStore(i *Integration) error {
	i.storer = persist.NewInMemoryStore()

	return nil
}

func Synchronized(i *Integration) error {
	i.locker = &sync.Mutex{}

	return nil
}

func Args(a interface{}) Option {
	return func(i *Integration) error {
		i.args = a

		return nil
	}
}
