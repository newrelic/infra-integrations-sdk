package integration

import (
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/persist"
)

func TestDefaultValues(t *testing.T) {
	// Redirecting standard output to a file
	f, err := ioutil.TempFile("", "stdout")
	assert.NoError(t, err)
	back := os.Stdout
	defer func() {
		os.Stdout = back
	}()
	os.Stdout = f

	// Given an integration builder without parameters
	i, err := NewBuilder("integration", "4.0").Build()

	// The Build method does not return any error
	assert.NoError(t, err)

	// And the data is correctly set (including defaults)
	assert.Equal(t, "integration", i.Name)
	assert.Equal(t, "4.0", i.IntegrationVersion)
	assert.Equal(t, "2", i.ProtocolVersion)
	assert.Equal(t, 0, len(i.Data))

	// And when publishing the payload
	assert.NoError(t, i.Publish())
	f.Close()

	// It is redirected to standard output, and non-prettified
	payload, err := ioutil.ReadFile(f.Name())
	assert.NoError(t, err)
	assert.Equal(t, "{\"name\":\"integration\",\"protocol_version\":\"2\",\"integration_version\":\"4.0\",\"data\":[]}",
		string(payload))
}

func TestIntegrationBuilder(t *testing.T) {
	// Redirecting standard output to a file
	output, err := ioutil.TempFile("", "stdout")
	assert.NoError(t, err)

	// Needed for initialising os.Args + flags (emulating).
	os.Args = []string{"cmd", "--pretty"}
	flag.CommandLine = flag.NewFlagSet("name", 0)

	// Given an integration builder with all the parameters set
	var arguments args.DefaultArgumentList
	i, err := NewBuilder("integration", "7.0").
		ParsedArguments(&arguments).
		Writer(output).
		Build()

	// The Build method does not return any error
	assert.NoError(t, err)

	// And the data is correctly set
	assert.Equal(t, "integration", i.Name)
	assert.Equal(t, "7.0", i.IntegrationVersion)
	assert.Equal(t, "2", i.ProtocolVersion)
	assert.Equal(t, 0, len(i.Data))

	// And when publishing the payload
	i.Publish()
	output.Close()

	// The output works as specified
	payload, err := ioutil.ReadFile(output.Name())
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(payload))

	// And also is prettified
	assert.True(t, arguments.Pretty)
	assert.Contains(t, payload, uint8('\n'))
}

func TestWrongArguments(t *testing.T) {
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
		_, err := NewBuilder("integration", "7.0").ParsedArguments(arg).Build()
		assert.Error(t, err)
	}
}

func TestItStoresOnDiskByDefault(t *testing.T) {
	i, err := NewBuilder("cool-integration", "1.0").Writer(ioutil.Discard).Build()
	assert.NoError(t, err)

	i.storer.Set("hello", 12.33)

	assert.NoError(t, i.Publish())

	// assert data has been flushed to disk
	c, err := persist.NewFileStore(persist.DefaultPath("cool-integration"), log.Discard)
	assert.NoError(t, err)

	v, ts, ok := c.Get("hello")
	assert.True(t, ok)
	assert.NotEqual(t, 0, ts)
	assert.InDelta(t, 12.33, v, 0.1)
}

func TestInMemoryStoreDoesNotPersistOnDisk(t *testing.T) {
	i, err := NewBuilder("cool-integration2", "1.0").Writer(ioutil.Discard).InMemoryStore().Build()
	assert.NoError(t, err)

	i.storer.Set("hello", 12.33)

	assert.NoError(t, i.Publish())

	// assert data has not been flushed to disk
	c, err := persist.NewFileStore(persist.DefaultPath("cool-integration2"), log.Discard)
	assert.NoError(t, err)

	_, _, ok := c.Get("hello")
	assert.False(t, ok)
}

func TestCustomStorer(t *testing.T) {
	customStorer := fakeStorer{}
	i, err := NewBuilder("cool-integration", "1.0").Writer(ioutil.Discard).Storer(&customStorer).Build()
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
