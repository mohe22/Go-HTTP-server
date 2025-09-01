package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	ErrEmptyPath     = fmt.Errorf("path is empty")
	ErrParamNotFound = fmt.Errorf("parameter not found")
)

type Params map[string]string

func IsParam(segment string) bool {
	return len(segment) >= 2 && segment[0] == '{' && segment[len(segment)-1] == '}'
}

// clean path from ? = &.
func CleanURL(url *string) string {
	if *url == "" {
		return ""
	}

	// Find the index of '?' which starts the query parameters
	if idx := strings.Index(*url, "?"); idx != -1 {
		*url = (*url)[:idx] // keep only the part before '?'
	}
	// Basic validation: must start with '/' and contain only allowed chars
	validPath := regexp.MustCompile(`^\/[a-zA-Z0-9\/\-\_\.]*$`)
	if !validPath.MatchString(*url) {
		return ""
	}

	return *url
}
func (p *Params) Get(key string) (string, error) {
	value, ok := (*p)[key]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrParamNotFound, key)
	}
	return value, nil
}
