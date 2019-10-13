package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOptionsExpirationValue(t *testing.T) {
	// Given
	options := Options{
		Expiration: 25 * time.Second,
	}

	// When - Then
	assert.Equal(t, 25*time.Second, options.ExpirationValue())
}
