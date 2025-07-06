package session

// errWithCode wraps an error with an associated HTTP status code.
type errWithCode struct {
	err    error
	status int
}

func (e *errWithCode) Error() string { return e.err.Error() }

// StatusCode returns the associated HTTP status code.
func (e *errWithCode) StatusCode() int { return e.status }

// ErrWithCode creates a new error annotated with an HTTP status code.
func ErrWithCode(code int, err error) error {
	if err == nil {
		return nil
	}
	return &errWithCode{err: err, status: code}
}
