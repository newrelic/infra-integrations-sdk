package v2

import (
	"io"
	"os"
	"reflect"
	"sync"

	"github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/cache"
	"github.com/newrelic/infra-integrations-sdk/log"
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
}

type integrationBuilderImpl struct {
	integration *Integration
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

	log.SetupLogging(defaultArgs.Verbose)

	// Avoid working with an uninitialized or in error state cache
	if err = cache.Status(); err != nil { // Todo: cache should not be a singleton
		return nil, err
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
