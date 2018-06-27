package integration

import (
	"io"

	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/persist"
)

// Option sets an option on integration level.
type Option func(*Integration) error

// Writer replaces the output writer.
func Writer(w io.Writer) Option {
	return func(i *Integration) error {
		i.writer = w

		return nil
	}
}

// Logger replaces the logger.
func Logger(l log.Logger) Option {
	return func(i *Integration) error {
		i.logger = l

		return nil
	}
}

// Storer replaces the metrics storage.
func Storer(s persist.Storer) Option {
	return func(i *Integration) error {
		i.storer = s

		return nil
	}
}

// InMemoryStore replaces the metrics storage by an ephemeral in-memory store.
func InMemoryStore() Option {
	return func(i *Integration) error {
		i.storer = persist.NewInMemoryStore()

		return nil
	}
}

// Args sets the destination struct (pointer) where the command-line flags will be parsed to.
func Args(a interface{}) Option {
	return func(i *Integration) error {
		i.args = a

		return nil
	}
}
