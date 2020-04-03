package metric

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSourceType_String(t *testing.T) {
	st := COUNT
	assert.Equal(t, "count", st.String())
}

func TestSourceTypeForName(t *testing.T) {
	name, err := SourceTypeForName("gauge")
	assert.NoError(t, err)
	assert.Equal(t, GAUGE, name)

	// ignore case
	name, err = SourceTypeForName("count")
	assert.NoError(t, err)
	assert.Equal(t, COUNT, name)

	// error
	_, err = SourceTypeForName("invalid")
	assert.Error(t, err)
}
