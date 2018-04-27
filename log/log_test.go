package log

import (
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
