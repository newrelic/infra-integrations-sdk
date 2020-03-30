package log

import (
	"os"
	"testing"

	"bytes"
	"strings"

	"github.com/stretchr/testify/assert"
)

func TestDefaultLogger_Errorf(t *testing.T) {
	var writer bytes.Buffer
	l := New(false, &writer)

	l.Errorf("foo")
	l.Errorf("bar")

	logged := writer.String()

	assert.True(t, strings.Contains(logged, "foo"))
	assert.True(t, strings.Contains(logged, "bar"))
	assert.Equal(t, 2, strings.Count(logged, "\n"), "should add carriage return on each error")
}

func TestDefaultLogger_Warnf(t *testing.T) {
	var writer bytes.Buffer
	l := New(false, &writer)

	l.Warnf("foo")
	l.Warnf("bar")

	logged := writer.String()

	assert.True(t, strings.Contains(logged, "foo"))
	assert.True(t, strings.Contains(logged, "bar"))
	assert.Equal(t, 2, strings.Count(logged, "\n"), "should add carriage return on each error")
}

func TestDefaultLogger_Infof(t *testing.T) {
	var writer bytes.Buffer
	l := New(false, &writer)

	l.Infof("foo")

	logged := writer.String()

	assert.True(t, strings.Contains(logged, "foo"))
	assert.Equal(t, 1, strings.Count(logged, "\n"))
}

func TestDefaultLogger_Debugf_DoesNotLogWhenNotActive(t *testing.T) {
	var writer bytes.Buffer
	l := New(false, &writer)

	l.Debugf("foo")

	logged := writer.String()

	assert.False(t, strings.Contains(logged, "foo"))
	assert.Equal(t, 0, strings.Count(logged, "\n"))
}

func TestDefaultLogger_Debugf_LogsWhenActive(t *testing.T) {
	var writer bytes.Buffer
	l := New(true, &writer)

	l.Debugf("foo")

	logged := writer.String()

	assert.True(t, strings.Contains(logged, "foo"))
	assert.Equal(t, 1, strings.Count(logged, "\n"))
}

// Tests for deprecated, old global logger
func TestSetupLoggingVerbose(t *testing.T) {
	// Capturing standard error of global logger
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	back := os.Stderr
	os.Stderr = w
	defer func() {
		os.Stderr = back
	}()
	defer func() { _ = r.Close() }()

	SetupLogging(true)

	Debug("hello everybody")
	Error("goodbye friend")
	assert.NoError(t, w.Close())

	stdErrBytes := new(bytes.Buffer)
	_, err = stdErrBytes.ReadFrom(r)
	assert.NoError(t, err)
	assert.Contains(t, stdErrBytes.String(), "hello everybody")
	assert.Contains(t, stdErrBytes.String(), "goodbye friend")
}

func TestSetupLoggingNotVerbose(t *testing.T) {
	// Capturing standard error of global logger
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	back := os.Stderr
	os.Stderr = w
	defer func() {
		os.Stderr = back
	}()
	defer func() { _ = r.Close() }()

	SetupLogging(false)

	Debug("hello everybody")
	Error("goodbye friend")
	assert.NoError(t, w.Close())

	stdErrBytes := new(bytes.Buffer)
	_, err = stdErrBytes.ReadFrom(r)
	assert.NoError(t, err)
	assert.NotContains(t, stdErrBytes.String(), "hello everybody")
	assert.Contains(t, stdErrBytes.String(), "goodbye friend")
}
