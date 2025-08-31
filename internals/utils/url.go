package utils

import (
	"fmt"
)

var (
	ErrEmptyPath     = fmt.Errorf("path is empty")
	ErrParamNotFound = fmt.Errorf("parameter not found")
)

type Params map[string]string

func IsParam(segment string) bool {
	return len(segment) >= 2 && segment[0] == '{' && segment[len(segment)-1] == '}'
}

func (p *Params) Get(key string) (string, error) {
	value, ok := (*p)[key]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrParamNotFound, key)
	}
	return value, nil
}
