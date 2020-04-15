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

func Test_SourceType_TypeAndNameAreSynced(t *testing.T) {

	assert.Equal(t, len(SourcesTypeToName), len(SourcesNameToType))

	for source, name := range SourcesTypeToName {
		st, err := SourceTypeForName(name)
		assert.NoError(t, err)
		assert.Equal(t, source, st)
	}
}
