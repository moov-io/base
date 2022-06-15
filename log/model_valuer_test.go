package log_test

import (
	"testing"

	"github.com/moov-io/base/log"
)

type Item struct {
	Value string
}

type Foo struct {
	Name *Item
}

func (f Foo) Context() map[string]log.Valuer {
	return log.Fields{
		"name": log.String(
			f.Name.Value,
		),
	}
}

func TestValuer__String(t *testing.T) {
	logger := log.NewTestLogger()

	foo := Foo{}
	logger.With(foo).Log("shouldn't panic")
}
