package types

type MultiError struct {
	errs   []error
	errCnt int
}

// Error .
func (c *MultiError) Error() string {
	if c.errCnt == 0 {
		return "empty chain error"
	}

	errmsg := ""
	for _, err := range c.errs {
		errmsg = errmsg + err.Error() + ";"
	}

	return errmsg
}

func (c *MultiError) Add(err error) {
	if len(c.errs) == 0 {
		c.errs = make([]error, 0, 5)
	}

	c.errs = append(c.errs, err)
	c.errCnt++
}

func (c *MultiError) Nil() bool {
	return c.errCnt == 0 && len(c.errs) == 0
}

func (c *MultiError) Reset() {
	c.errs = nil
	c.errCnt = 0
}
