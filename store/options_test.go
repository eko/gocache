package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOptionsCostValue(t *testing.T) {
	// Given
	options := &Options{
		cost: 7,
	}

	// When - Then
	assert.Equal(t, int64(7), options.cost)
}

func TestOptionsExpirationValue(t *testing.T) {
	// Given
	options := &Options{
		expiration: 25 * time.Second,
	}

	// When - Then
	assert.Equal(t, 25*time.Second, options.expiration)
}

func TestOptionsTagsValue(t *testing.T) {
	// Given
	options := &Options{
		tags: []string{"tag1", "tag2", "tag3"},
	}

	// When - Then
	assert.Equal(t, []string{"tag1", "tag2", "tag3"}, options.tags)
}

func Test_applyOptionsWithDefault(t *testing.T) {
	// Given
	defaultOptions := &Options{
		expiration: 25 * time.Second,
	}

	// When
	options := applyOptionsWithDefault(defaultOptions, WithCost(7))

	// Then
	assert.Equal(t, int64(7), options.cost)
	assert.Equal(t, 25*time.Second, options.expiration)
}
