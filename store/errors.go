package store

const NOT_FOUND_ERR string = "Value not found in store"

type NotFound struct {
	cause error
}

func NotFoundWithCause(e error) error {
	return NotFound{
		cause: e,
	}
}

func (e NotFound) Cause() error {
	return e.cause
}

func (e NotFound) Is(err error) bool {
	return err.Error() == NOT_FOUND_ERR
}

func (e NotFound) Error() string {
	return NOT_FOUND_ERR
}
func (e NotFound) Unwrap() error { return e.cause }
