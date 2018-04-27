package integration

import (
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"bytes"
	"fmt"

	"github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/persist"
)

func TestWriter(t *testing.T) {
	var w bytes.Buffer

	i, err := New("integration", "7.0", Writer(&w))
	assert.NoError(t, err)

	assert.NoError(t, i.Publish())

	assert.Equal(t, `{"name":"integration","protocol_version":"2","integration_version":"7.0","data":[]}`, w.String())
}

func TestArgs(t *testing.T) {
	// arguments are read from os
	os.Args = []string{"cmd", "--pretty"}
	flag.CommandLine = flag.NewFlagSet("name", 0)
	var arguments args.DefaultArgumentList

	// capture output
	var writer bytes.Buffer

	i, err := New("integration", "7.0", Args(&arguments), Writer(&writer))
	assert.NoError(t, err)

	assert.NoError(t, i.Publish())

	payload := writer.String()

	// output is prettified
	assert.Contains(t, payload, "\n")
}

func TestWrongArgumentsCausesError(t *testing.T) {
	var d interface{} = struct{}{}

	arguments := []struct {
		arg interface{}
	}{
		{struct{ thing string }{"abcd"}},
		{1234},
		{"hello"},
		{[]struct{ x string }{{"hello"}, {"my friend"}}},
		{d},
	}
	for _, arg := range arguments {
		_, err := New("integration", "7.0", Args(arg))
		assert.Error(t, err)
	}
}

func TestItStoresOnDiskByDefault(t *testing.T) {
	i := newNoLoggerNoWriter(t)

	i.storer.Set("hello", 12.33)

	assert.NoError(t, i.Publish())

	// assert data has been flushed to disk
	c, err := persist.NewFileStore(persist.DefaultPath(integrationName), log.Discard)
	assert.NoError(t, err)

	v, ts, ok := c.Get("hello")
	assert.True(t, ok)
	assert.NotEqual(t, 0, ts)
	assert.InDelta(t, 12.33, v, 0.1)
}

func TestInMemoryStoreDoesNotPersistOnDisk(t *testing.T) {
	i, err := New("cool-integration2", "1.0", Writer(ioutil.Discard), InMemoryStore())
	assert.NoError(t, err)

	i.storer.Set("hello", 12.33)

	assert.NoError(t, i.Publish())

	// assert data has not been flushed to disk
	c, err := persist.NewFileStore(persist.DefaultPath("cool-integration2"), log.Discard)
	assert.NoError(t, err)

	_, _, ok := c.Get("hello")
	assert.False(t, ok)
}

func TestConcurrentModeHasNoDataRace(t *testing.T) {
	in, err := New("TestIntegration", "1.0", Logger(log.Discard), Writer(ioutil.Discard), Synchronized())
	assert.NoError(t, err)

	for i := 0; i < 10; i++ {
		go func(i int) {
			in.Entity(fmt.Sprintf("entity%v", i), "test")
		}(i)
	}
}

func TestStorer(t *testing.T) {
	customStorer := fakeStorer{}
	i, err := New("cool-integration", "1.0", Writer(ioutil.Discard), Storer(&customStorer))
	assert.NoError(t, err)

	assert.NoError(t, i.Publish())

	assert.True(t, customStorer.saved, "data has not been saved")
}

type fakeStorer struct {
	saved bool
}

func (m *fakeStorer) Save() error {
	m.saved = true
	return nil
}

func (fakeStorer) Get(name string) (float64, int64, bool) {
	return 0, 0, false
}

func (fakeStorer) Set(name string, value float64) int64 {
	return 0
}
