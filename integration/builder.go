package integration

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"sync"

	"github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/persist"
	"github.com/pkg/errors"
)

const protocolVersion = "2"

// Builder OOP builder-pattern to create a new Integration instance.
type Builder struct {
	integration *Integration
	arguments   interface{}
}

type disabledLocker struct{}

func (disabledLocker) Lock()   {}
func (disabledLocker) Unlock() {}

// New creates new integration with sane default values.
func New(name, version string) (*Integration, error) {
	return NewBuilder(name, version).Build()
}

// NewBuilder creates a new Builder for the given integration name and version.
func NewBuilder(name, version string) *Builder {
	return &Builder{
		integration: &Integration{
			Name:               name,
			ProtocolVersion:    protocolVersion,
			IntegrationVersion: version,
			Data:               []*EntityData{},
			writer:             os.Stdout, // defaults to stdout
		},
	}
}

// Synchronized locks data on r/w to enable concurrency.
func (b *Builder) Synchronized() *Builder {
	b.integration.locker = &sync.Mutex{}
	return b
}

// Writer sets the integration output.
func (b *Builder) Writer(writer io.Writer) *Builder {
	b.integration.writer = writer
	return b
}

// ParsedArguments sets the destination struct (pointer) where the command-line flags will be parsed to.
func (b *Builder) ParsedArguments(dstPointer interface{}) *Builder {
	b.arguments = dstPointer
	return b
}

// Storer sets the persistence store.
func (b *Builder) Storer(c persist.Storer) *Builder {
	b.integration.Storer = c
	return b
}

// InMemoryStore sets the persistence store to ephemeral in-memory.
func (b *Builder) InMemoryStore() *Builder {
	b.integration.Storer = persist.NewInMemoryStore()
	return b
}

// Build builds a proper integration.
func (b *Builder) Build() (*Integration, error) {
	// Checking errors
	if b.integration.writer == nil {
		return nil, errors.New("integration writer can't be nil")
	}

	// Setting default values
	if b.integration.locker == nil {
		b.integration.locker = disabledLocker{}
	}

	// Checking arguments
	err := b.checkArguments()
	if err != nil {
		return nil, err
	}
	err = args.SetupArgs(b.arguments)
	if err != nil {
		return nil, err
	}
	defaultArgs := args.GetDefaultArgs(b.arguments)

	if b.integration.Storer == nil {
		// TODO: remove this log and add Log(log) function to this Builder
		b.integration.Storer, err = persist.NewFileStore(persist.DefaultPath(b.integration.Name), log.NewStdErr(false))
		if err != nil {
			return nil, fmt.Errorf("can't create store: %s", err.Error())
		}
	}

	b.integration.prettyOutput = defaultArgs.Pretty

	return b.integration, nil
}

// Returns error if the parsed arguments destination is not from an acceptable type. It can be nil or a pointer to a
// struct.
func (b *Builder) checkArguments() error {
	if b.arguments == nil {
		b.arguments = new(struct{})
		return nil
	}
	val := reflect.ValueOf(b.arguments)

	if val.Kind() == reflect.Ptr && val.Elem().Kind() == reflect.Struct {
		return nil
	}
	return errors.New("arguments must be a pointer to a struct (or nil)")
}
