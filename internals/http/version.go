package http

import (
	"bytes"
	"fmt"
)

type Version string

const (
	HTTP1_1 Version = "HTTP/1.1"
	HTTP1_0 Version = "HTTP/1.0"
	HTTP2   Version = "HTTP/2"
)

func (v *Version) String() string {
	switch *v {
	case HTTP1_0:
		return "HTTP/1.0"
	case HTTP1_1:
		return "HTTP/1.1"
	case HTTP2:
		return "HTTP/2.0"
	default:
		return "HTTP/1.1"
	}
}

func parseVersion(data []byte) (Version, error) {
	switch {
	case bytes.Equal(data, []byte(HTTP1_1)):
		return HTTP1_1, nil
	case bytes.Equal(data, []byte(HTTP2)):
		return HTTP2, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrUnsupportedVersion, data)
	}
}
