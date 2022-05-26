package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvalidateOptionsTagsValue(t *testing.T) {
	// Given
	options := invalidateOptions{
		tags: []string{"tag1", "tag2", "tag3"},
	}

	// When - Then
	assert.Equal(t, []string{"tag1", "tag2", "tag3"}, options.tags)
}
