package handler

import "fmt"

func E(msg string, args ...interface{}) error {
	return fmt.Errorf(msg, args...)
}

func ErrWithCode(code int, e error) error {
	return &errWraper{
		err:    e,
		status: code,
	}
}

type errWraper struct {
	err    error
	status int
}

func (e *errWraper) Error() string {
	return e.err.Error()
}
