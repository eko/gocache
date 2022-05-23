package store

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

func (e NotFound) Error() string {
	return "Value not found in store"
}
