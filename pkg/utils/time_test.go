package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDatePtr_Nil(t *testing.T) {
	assert.Nil(t, ParseDatePtr(nil))
}

func TestParseDatePtr_Empty(t *testing.T) {
	empty := ""
	assert.Nil(t, ParseDatePtr(&empty))
}

func TestParseDatePtr_Valid(t *testing.T) {
	dateStr := "2025-01-15T10:00:00Z"
	result := ParseDatePtr(&dateStr)
	assert.NotNil(t, result)
	assert.Equal(t, 2025, result.Year())
	assert.Equal(t, time.January, result.Month())
	assert.Equal(t, 15, result.Day())
}

func TestParseDatePtr_Invalid(t *testing.T) {
	invalid := "not-a-date"
	assert.Nil(t, ParseDatePtr(&invalid))
}
