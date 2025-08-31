package http

import "errors"

var (
	SEPARATOR         = "\r\n"
	DefaultBufferSize = 4096
	// Errors
	ErrUnsupportedVersion   = errors.New("unsupported HTTP version")
	ErrInvalidRequestLine   = errors.New("invalid request line")
	ErrInvalidRequestMethod = errors.New("invalid HTTP method")
	ErrRequestTooLarge      = errors.New("request too large")
	ErrInvalidHeader        = errors.New("invalid Headers")
	ErrInvalidHeaderName    = errors.New("invalid header name: contains whitespace or empty")
	ErrUnknownStatusCode    = errors.New("unknown status code")
	ErrMethodNotFound       = errors.New("method not found")
	ErrPathNotFound         = errors.New("path not found")

	// HEADER
	ErrKeyNotFound = errors.New("key not found in header")
	ErrEmptyKey    = errors.New("key cannot be empty")
)
