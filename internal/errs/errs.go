package errs

import "errors"

var (
	ErrDuplicate               = errors.New("duplicate URL")
	ErrURLNotFound             = errors.New("url is not present")
	ErrInternalServerError     = errors.New("internal server error")
	ErrNotAuthorized           = errors.New("invalid token")
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
	ErrUserMismatch            = errors.New("user tried to delete not owned urls")
	ErrGone                    = errors.New("url was marked as deleted")
	ErrSigningToken            = errors.New("error signing token")
	ErrRefreshingToken         = errors.New("error refreshing token")
	ErrNoCert                  = errors.New("no certificate provided")
	ErrNoPK                    = errors.New("no private key provided")
)
