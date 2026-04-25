package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPool_InvalidURL(t *testing.T) {
	pool, err := NewPool("invalid_url")
	assert.Error(t, err)
	assert.Nil(t, pool)
}

func TestNewPool_ValidURL_ConnectionFailed(t *testing.T) {
	pool, err := NewPool("postgres://nonexistent:5432/db")
	assert.Error(t, err)
	assert.Nil(t, pool)
}
