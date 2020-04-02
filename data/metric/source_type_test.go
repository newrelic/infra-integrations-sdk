package metric

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSourceType_String(t *testing.T) {
	st := COUNT
	assert.Equal(t, "count", st.String())
}

func TestSourceType_Positive(t *testing.T) {

	var testCases = []struct {
		sourceType SourceType
		isPositive bool
	}{
		{GAUGE, false},
		{COUNT, true},
		{SUMMARY, false},
		{PDELTA, true},
	}

	for _, tc := range testCases {
		t.Run(string(tc.sourceType), func(t *testing.T) {
			assert.Equal(t, tc.isPositive, tc.sourceType.IsPositive())
		})
	}
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
