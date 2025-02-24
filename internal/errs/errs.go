package errs

import "errors"

var (
	ErrDuplicate = errors.New("duplicate URL")
	ErrNotFound  = errors.New("URL is not present")
)
