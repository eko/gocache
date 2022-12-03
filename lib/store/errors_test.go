package store

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotFoundIs(t *testing.T) {
	expectedErr := errors.New(("this is an expected error cause"))
	err := NotFoundWithCause(nil)
	assert.True(t, errors.Is(err, NotFound{cause: expectedErr}))

	err2 := &NotFound{}
	assert.True(t, errors.Is(err2, &NotFound{}))

	_, ok := err.(*NotFound)
	assert.True(t, ok)

	assert.True(t, err.Error() == NotFound{}.Error())
}
