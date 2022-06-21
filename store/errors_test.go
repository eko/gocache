package store

import (
	"errors"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func TestNotFoundIs(t *testing.T) {
	err := NotFoundWithCause(redis.Nil)
	assert.True(t, errors.Is(err, NotFound{}))
	assert.True(t, errors.Is(err, redis.Nil))

	err2 := &NotFound{}
	assert.True(t, errors.Is(err2, &NotFound{}))

	_, ok := err.(*NotFound)
	assert.True(t, ok)

	assert.True(t, err.Error() == NotFound{}.Error())
}
