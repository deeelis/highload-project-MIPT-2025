package errors

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidInput       = errors.New("invalid input")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrContentNotFound    = errors.New("content not found")
	ErrInternalServer     = errors.New("internal server error")
	ErrInvalidContentType = errors.New("invalid content type")
	ErrKafkaUnavailable   = errors.New("kafka unavailable")
)
