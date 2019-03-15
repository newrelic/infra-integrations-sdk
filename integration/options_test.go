package integration

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/persist"
)

func TestWriter(t *testing.T) {
	var w bytes.Buffer

	i, err := New("integration", "7.0", Writer(&w), InMemoryStore())
	assert.NoError(t, err)

	assert.NoError(t, i.Publish())

	assert.Equal(t, `{"name":"integration","protocol_version":"3","integration_version":"7.0","data":[]}`+"\n", w.String())
}

func TestArgs(t *testing.T) {
	// arguments are read from os
	os.Args = []string{"cmd", "--pretty"}
	flag.CommandLine = flag.NewFlagSet("name", 0)
	var arguments args.DefaultArgumentList

	// capture output
	var writer bytes.Buffer

	i, err := New("integration", "7.0", Args(&arguments), Writer(&writer), InMemoryStore())
	assert.NoError(t, err)

	assert.NoError(t, i.Publish())

	assert.Contains(t, writer.String(), "\n", "output should be prettified")
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
	i, err := New(integrationName, integrationVersion)
	assert.NoError(t, err)

	i.storer.Set("hello", 12.33)

	assert.NoError(t, i.Publish())

	// assert data has been flushed to disk
	c, err := persist.NewFileStore(persist.DefaultPath(integrationName), log.NewStdErr(true), persist.DefaultTTL)
	assert.NoError(t, err)

	var v float64
	ts, err := c.Get("hello", &v)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, ts)
	assert.InDelta(t, 12.33, v, 0.1)
}

func TestInMemoryStoreDoesNotPersistOnDisk(t *testing.T) {
	randomName := strconv.Itoa(rand.Int())

	i, err := New(randomName, integrationVersion, InMemoryStore())
	assert.NoError(t, err)

	i.storer.Set("hello", 12.33)

	assert.NoError(t, i.Publish())

	// assert data has not been flushed to disk

	// create folder in case it does not exists to enable store creation
	path := persist.DefaultPath(randomName)
	assert.NoError(t, os.MkdirAll(path, 0755))

	s, err := persist.NewFileStore(path, log.Discard, persist.DefaultTTL)
	assert.NoError(t, err)

	var v float64
	_, err = s.Get("hello", &v)
	assert.Equal(t, persist.ErrNotFound, err)
}

func TestConcurrentModeHasNoDataRace(t *testing.T) {
	in, err := New("TestIntegration", "1.0", Logger(log.Discard), Writer(ioutil.Discard))
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

func (fakeStorer) Set(key string, value interface{}) int64 {
	return 0
}

func (fakeStorer) Get(key string, valuePtr interface{}) (int64, error) {
	return 0, nil
}

func (fakeStorer) Delete(key string) error {
	return nil
}
