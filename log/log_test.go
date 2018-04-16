package log

import (
	"testing"

	"bytes"
	"strings"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	var writer bytes.Buffer

	l := New(false, &writer)


	l.Errorf("foo")
	l.Errorf("bar")

	logged := writer.String()

	assert.True(t, strings.Contains(logged, "foo"))
	assert.True(t, strings.Contains(logged, "bar"))
	assert.Equal(t, 2, strings.Count(logged, "\n"), "should add carriage return on each error")

	// show log on verbose
	t.Log(logged)
}
