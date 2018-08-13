package metric

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSourceType_String(t *testing.T) {
	st := RATE
	assert.Equal(t, "rate", st.String())
}

func TestSourceTypeForName(t *testing.T) {
	name, err := SourceTypeForName("delta")
	assert.NoError(t, err)
	assert.Equal(t, DELTA, name)

	// ignore case
	name, err = SourceTypeForName("RATE")
	assert.NoError(t, err)
	assert.Equal(t, RATE, name)

	// error
	_, err = SourceTypeForName("invalid")
	assert.Error(t, err)
}
