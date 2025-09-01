package types

import (
	"bytes"
	"errors"
	"fmt"
)

type Method string

const (
	GET     Method = "GET"
	POST    Method = "POST"
	PUT     Method = "PUT"
	DELETE  Method = "DELETE"
	HEAD    Method = "HEAD"
	OPTIONS Method = "OPTIONS"
	PATCH   Method = "PATCH"
	CONNECT Method = "CONNECT"
	TRACE   Method = "TRACE"
)

var ErrInvalidRequestMethod = errors.New("invalid HTTP method")

func ParseMethod(data []byte) (Method, error) {
	switch {
	case bytes.Equal(data, []byte(GET)):
		return GET, nil
	case bytes.Equal(data, []byte(POST)):
		return POST, nil
	case bytes.Equal(data, []byte(DELETE)):
		return DELETE, nil
	case bytes.Equal(data, []byte(PUT)):
		return PUT, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrInvalidRequestMethod, data)
	}
}
