package domain

import "errors"

var (
	ErrEmptyURL         = errors.New("URL is required")
	ErrInvalidURL       = errors.New("invalid URL")
	ErrInvalidMethod    = errors.New("invalid HTTP method")
	ErrRequestTimeout   = errors.New("request timed out")
	ErrConnectionFailed = errors.New("connection failed")
	ErrVariableNotFound = errors.New("variable not found")
	ErrInvalidAuth      = errors.New("invalid auth configuration")
)
