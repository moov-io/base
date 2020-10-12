package http2

import (
	"errors"
	"fmt"
)

var (
	errInvalidJSON = errors.New("invalid JSON in request body")
)

type errPathVarNotFound struct {
	key string
}

func (e errPathVarNotFound) Error() string {
	return fmt.Sprintf("path variable '%s' not found in request URL", e.key)
}

type errHeaderNotFound struct {
	key string
}

func (e errHeaderNotFound) Error() string {
	return fmt.Sprintf("header '%s' not found in request", e.key)
}
