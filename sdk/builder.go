package sdk

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"sync"

	"github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/persist"
	"github.com/pkg/errors"
)

const protocolVersion = "2"

// IntegrationBuilder provides a fluent interface for creating a configured Integration.
type IntegrationBuilder interface {
	// Build returns the Integration resulting from the applied configuration on this builder.
	// The integration data is empty, ready to add new entities' data.
	Build() (*Integration, error)
	// ParsedArguments sets the destination struct (pointer) where the command-line flags will be parsed to.
	ParsedArguments(interface{}) IntegrationBuilder
	// Synchronized sets the built Integration ready to be managed concurrently from multiple threads.
	// By default, the build integration is not synchronized.
	Synchronized() IntegrationBuilder
	// Writer sets the output stream where the integration resulting payload will be written to.
	// By default, the standard output (os.Stdout).
	Writer(io.Writer) IntegrationBuilder
	// Storer sets the persistence implementation that will be used to persist data between executions of the same integration.
	// By default, it will be a Disk-backed storage named stored in the file returned by the
	// persist.DefaultPath(integrationName) function.
	Storer(persist.Storer) IntegrationBuilder
	// NoStorer disables the storage for this integration.
	NoStorer() IntegrationBuilder
}

type integrationBuilderImpl struct {
	integration *Integration
	hasStore    bool
	arguments   interface{}
}

type disabledLocker struct{}

func (disabledLocker) Lock()   {}
func (disabledLocker) Unlock() {}

// NewIntegration creates a new IntegrationBuilder for the given integration name and version.
func NewIntegration(name string, version string) IntegrationBuilder {
	return &integrationBuilderImpl{
		integration: &Integration{
			Name:               name,
			ProtocolVersion:    protocolVersion,
			IntegrationVersion: version,
			Data:               []*EntityData{},
			writer:             os.Stdout, // defaults to stdout
		},
		hasStore: true,
	}
}

func (b *integrationBuilderImpl) Synchronized() IntegrationBuilder {
	b.integration.locker = &sync.Mutex{}
	return b
}

func (b *integrationBuilderImpl) Writer(writer io.Writer) IntegrationBuilder {
	b.integration.writer = writer
	return b
}

func (b *integrationBuilderImpl) ParsedArguments(dstPointer interface{}) IntegrationBuilder {
	b.arguments = dstPointer
	return b
}

func (b *integrationBuilderImpl) Storer(c persist.Storer) IntegrationBuilder {
	b.integration.Storer = c
	b.hasStore = true
	return b
}

func (b *integrationBuilderImpl) NoStorer() IntegrationBuilder {
	b.integration.Storer = nil
	b.hasStore = false
	return b
}

func (b *integrationBuilderImpl) Build() (*Integration, error) {
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

	persist.SetupLogging(defaultArgs.Verbose)

	if b.integration.Storer == nil && b.hasStore {
		// TODO: set Log(log) function to this builder
		b.integration.Storer, err = persist.NewStorer(persist.DefaultPath(b.integration.Name), persist.GlobalLog)
		if err != nil {
			return nil, fmt.Errorf("can't create store: %s", err.Error())
		}
	}

	b.integration.prettyOutput = defaultArgs.Pretty

	return b.integration, nil
}

// Returns error if the parsed arguments destination is not from an acceptable type. It can be nil or a pointer to a
// struct.
func (b *integrationBuilderImpl) checkArguments() error {
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
