package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOptionsCostValue(t *testing.T) {
	// Given
	options := &options{
		cost: 7,
	}

	// When - Then
	assert.Equal(t, int64(7), options.cost)
}

func TestOptionsExpirationValue(t *testing.T) {
	// Given
	options := &options{
		expiration: 25 * time.Second,
	}

	// When - Then
	assert.Equal(t, 25*time.Second, options.expiration)
}

func TestOptionsTagsValue(t *testing.T) {
	// Given
	options := &options{
		tags: []string{"tag1", "tag2", "tag3"},
	}

	// When - Then
	assert.Equal(t, []string{"tag1", "tag2", "tag3"}, options.tags)
}
