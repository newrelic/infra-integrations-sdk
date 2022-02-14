package integration

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newrelic/infra-integrations-sdk/v4/args"
	"github.com/newrelic/infra-integrations-sdk/v4/log"
)

func Test_PublishWritesUsingSelectedWriter(t *testing.T) {
	var w bytes.Buffer

	i, err := New("integration", "7.0", Writer(&w))
	assert.NoError(t, err)

	assert.NoError(t, i.Publish())

	assert.Equal(t, `{"protocol_version":"4","integration":{"name":"integration","version":"7.0"},"data":[]}`+"\n", w.String())
}

func Test_PrettyPrintWritesPrettifiedResult(t *testing.T) {
	// arguments are read from os
	os.Args = []string{"cmd", "--pretty"}
	flag.CommandLine = flag.NewFlagSet("name", 0)
	var arguments args.DefaultArgumentList

	// capture output
	var writer bytes.Buffer

	i, err := New("integration", "7.0", Args(&arguments), Writer(&writer))
	assert.NoError(t, err)

	assert.NoError(t, i.Publish())

	assert.Contains(t, writer.String(), "\n", "output should be prettified")
}

func Test_WrongArgumentsCausesError(t *testing.T) {
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

func Test_ConcurrentModeHasNoDataRace(t *testing.T) {
	in, err := New("TestIntegration", "1.0", Logger(log.Discard), Writer(ioutil.Discard))
	assert.NoError(t, err)

	for i := 0; i < 10; i++ {
		go func(i int) {
			_, _ = in.NewEntity(fmt.Sprintf("entity%v", i), "", "test", false)
		}(i)
	}
}

func Test_VerboseLogPrintsDebugMessages(t *testing.T) {
	type argumentList struct {
		args.DefaultArgumentList
	}

	// Given an integration set in verbose mode
	os.Args = []string{"cmd", "--verbose"}
	flag.CommandLine = flag.NewFlagSet("name", 0)

	// Whose log messages are written in the standard error
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	back := os.Stderr
	os.Stderr = w
	defer func() {
		os.Stderr = back
	}()
	defer func() { _ = r.Close() }()

	var al argumentList
	i, err := New("TestIntegration", "1.0", Args(&al))
	assert.NoError(t, err)

	// When logging a debug message
	i.logger.Debugf("hello everybody")
	assert.NoError(t, w.Close())

	// The message is correctly submitted to the standard error
	stdErrBytes := new(bytes.Buffer)
	_, err = stdErrBytes.ReadFrom(r)
	assert.NoError(t, err)
	assert.Contains(t, stdErrBytes.String(), "hello everybody")
}

func Test_CustomArgumentsAreAddedToArgumentList(t *testing.T) {
	type argumentList struct {
		args.DefaultArgumentList
	}

	os.Args = []string{"cmd", "--pretty", "--verbose"}
	flag.CommandLine = flag.NewFlagSet("name", 0)

	var al argumentList
	_, err := New("TestIntegration", "1.0", Logger(log.Discard), Writer(ioutil.Discard), Args(&al))
	assert.NoError(t, err)

	if !al.All() {
		t.Error()
	}
	if !al.Pretty {
		t.Error()
	}
	if !al.Verbose {
		t.Error()
	}
}

func Test_DefaultArguments(t *testing.T) {
	t.Skip("This is failing to due to flag redefinition. We'll take a look later")
	al := args.DefaultArgumentList{}

	i, err := New("TestIntegration", "1.0", Logger(log.Discard), Writer(ioutil.Discard), Args(&al))
	assert.NoError(t, err)

	assert.Equal(t, "TestIntegration", i.Metadata.Name)
	assert.Equal(t, "1.0", i.Metadata.Version)
	assert.Equal(t, "4", i.ProtocolVersion)
	assert.Len(t, i.Entities, 0)
	assert.True(t, al.All())
	assert.False(t, al.Pretty)
	assert.False(t, al.Verbose)
}

func Test_DefaultArgsSetNonVerboseLogging(t *testing.T) {
	type argumentList struct {
		args.DefaultArgumentList
	}

	// Given an integration set in non-verbose mode
	os.Args = []string{"cmd"}
	flag.CommandLine = flag.NewFlagSet("name", 0)

	// Whose log messages are written in the standard error
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	back := os.Stderr
	os.Stderr = w
	defer func() {
		os.Stderr = back
	}()
	defer func() { _ = r.Close() }()

	var al argumentList
	i, err := New("TestIntegration", "1.0", Args(&al))
	assert.NoError(t, err)

	// When logging info, error and debug messages
	i.logger.Debugf("this is a debug")
	i.logger.Infof("this is an info")
	i.logger.Errorf("this is an error")
	assert.NoError(t, w.Close())

	// The standard error shows all the message levels but debug
	stdErrBytes := new(bytes.Buffer)
	_, err = stdErrBytes.ReadFrom(r)
	assert.NoError(t, err)
	assert.Contains(t, stdErrBytes.String(), "this is an error")
	assert.Contains(t, stdErrBytes.String(), "this is an info")
	assert.NotContains(t, stdErrBytes.String(), "this is a debug")
}

func Test_ClusterAndServiceArgumentsAreAddedToMetadata(t *testing.T) {
	al := args.DefaultArgumentList{}

	_ = os.Setenv("NRI_CLUSTER", "foo")
	_ = os.Setenv("NRI_SERVICE", "bar")

	// os.ClearEnv breaks tests in Windows that use the fileStorer because clears the user env vars
	defer func() {
		_ = os.Unsetenv("NRI_CLUSTER")
		_ = os.Unsetenv("NRI_SERVICE")
	}()

	os.Args = []string{"cmd"}
	flag.CommandLine = flag.NewFlagSet("cmd", flag.ContinueOnError)

	i, err := New("TestIntegration", "1.0", Logger(log.Discard), Writer(ioutil.Discard), Args(&al))
	assert.NoError(t, err)

	e, err := i.NewEntity("name", "ns", "", false)
	assert.NoError(t, err)

	assert.Len(t, e.GetMetadata(), 0)
}
