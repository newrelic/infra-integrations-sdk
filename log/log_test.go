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
	l := NewWriter(false, &writer)

	l.Errorf("foo")
	l.Errorf("bar")

	logged := writer.String()

	assert.True(t, strings.Contains(logged, "foo"))
	assert.True(t, strings.Contains(logged, "bar"))
	assert.Equal(t, 2, strings.Count(logged, "\n"), "should add carriage return on each error")
}

func TestDefaultLogger_Warnf(t *testing.T) {
	var writer bytes.Buffer
	l := NewWriter(false, &writer)

	l.Warnf("foo")
	l.Warnf("bar")

	logged := writer.String()

	assert.True(t, strings.Contains(logged, "foo"))
	assert.True(t, strings.Contains(logged, "bar"))
	assert.Equal(t, 2, strings.Count(logged, "\n"), "should add carriage return on each error")
}

func TestDefaultLogger_Infof(t *testing.T) {
	var writer bytes.Buffer
	l := NewWriter(false, &writer)

	l.Infof("foo")

	logged := writer.String()

	assert.True(t, strings.Contains(logged, "foo"))
	assert.Equal(t, 1, strings.Count(logged, "\n"))
}

func TestDefaultLogger_Debugf_DoesNotLogWhenNotActive(t *testing.T) {
	var writer bytes.Buffer
	l := NewWriter(false, &writer)

	l.Debugf("foo")

	logged := writer.String()

	assert.False(t, strings.Contains(logged, "foo"))
	assert.Equal(t, 0, strings.Count(logged, "\n"))
}

func TestDefaultLogger_Debugf_LogsWhenActive(t *testing.T) {
	var writer bytes.Buffer
	l := NewWriter(true, &writer)

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
	defer r.Close()

	SetupLogging(true)

	Debug("hello everybody")
	Error("goodbye friend")
	assert.NoError(t, w.Close())

	stdErrBytes := new(bytes.Buffer)
	stdErrBytes.ReadFrom(r)
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
	defer r.Close()

	SetupLogging(false)

	Debug("hello everybody")
	Error("goodbye friend")
	assert.NoError(t, w.Close())

	stdErrBytes := new(bytes.Buffer)
	stdErrBytes.ReadFrom(r)
	assert.NotContains(t, stdErrBytes.String(), "hello everybody")
	assert.Contains(t, stdErrBytes.String(), "goodbye friend")
}

func TestConfigureLoggerVerbose(t *testing.T) {
	// Capturing standard error of global logger
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	back := os.Stderr
	os.Stderr = w
	defer func() {
		os.Stderr = back
	}()
	defer r.Close()

	l := defaultLogger{}
	ConfigureLogger(&l, true)

	l.Debugf("hello everybody")
	l.Errorf("goodbye friend")
	assert.NoError(t, w.Close())

	stdErrBytes := new(bytes.Buffer)
	stdErrBytes.ReadFrom(r)
	assert.Contains(t, stdErrBytes.String(), "hello everybody")
	assert.Contains(t, stdErrBytes.String(), "goodbye friend")
}

func TestConfigureLoggerNotVerbose(t *testing.T) {
	// Capturing standard error of global logger
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	back := os.Stderr
	os.Stderr = w
	defer func() {
		os.Stderr = back
	}()
	defer r.Close()

	l := defaultLogger{}
	ConfigureLogger(&l, false)

	l.Debugf("hello everybody")
	l.Errorf("goodbye friend")
	assert.NoError(t, w.Close())

	stdErrBytes := new(bytes.Buffer)
	stdErrBytes.ReadFrom(r)
	assert.NotContains(t, stdErrBytes.String(), "hello everybody")
	assert.Contains(t, stdErrBytes.String(), "goodbye friend")
}

func TestNewVerbose(t *testing.T) {
	// Capturing standard error of global logger
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	back := os.Stderr
	os.Stderr = w
	defer func() {
		os.Stderr = back
	}()
	defer r.Close()

	l := New(true)

	l.Debugf("hello everybody")
	l.Errorf("goodbye friend")
	assert.NoError(t, w.Close())

	stdErrBytes := new(bytes.Buffer)
	stdErrBytes.ReadFrom(r)
	assert.Contains(t, stdErrBytes.String(), "hello everybody")
	assert.Contains(t, stdErrBytes.String(), "goodbye friend")
}

func TestNewNotVerbose(t *testing.T) {
	// Capturing standard error of global logger
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	back := os.Stderr
	os.Stderr = w
	defer func() {
		os.Stderr = back
	}()
	defer r.Close()

	l := New(false)

	l.Debugf("hello everybody")
	l.Errorf("goodbye friend")
	assert.NoError(t, w.Close())

	stdErrBytes := new(bytes.Buffer)
	stdErrBytes.ReadFrom(r)
	assert.NotContains(t, stdErrBytes.String(), "hello everybody")
	assert.Contains(t, stdErrBytes.String(), "goodbye friend")
}
