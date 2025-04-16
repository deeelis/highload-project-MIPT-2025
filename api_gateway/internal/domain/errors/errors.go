package errors

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrContentNotFound    = errors.New("content not found")
	ErrInternalServer     = errors.New("internal server error")
	ErrInvalidContentType = errors.New("invalid content type")
)
