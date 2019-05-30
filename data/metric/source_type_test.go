package metric

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSourceType_String(t *testing.T) {
	st := RATE
	assert.Equal(t, "rate", st.String())
}

func TestSourceType_Positive(t *testing.T) {

	var testCases = []struct {
		sourceType SourceType
		isPositive bool
	}{
		{GAUGE, false},
		{RATE, false},
		{DELTA, false},
		{PRATE, true},
		{PDELTA, true},
	}

	for _, tc := range testCases {
		t.Run(string(tc.sourceType), func(t *testing.T) {
			assert.Equal(t, tc.isPositive, tc.sourceType.IsPositive())
		})
	}
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
