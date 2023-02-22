package log_test

import (
	"testing"

	"github.com/moov-io/base/log"
	"github.com/stretchr/testify/require"
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

type Mode int

func (m Mode) String() string {
	switch m {
	case 1:
		return "SANDBOX"
	case 2:
		return "PRODUCTION"
	}
	return "UNSPECIFIED"
}

func TestValuer_Stringer(t *testing.T) {
	out, logger := log.NewBufferLogger()

	m := Mode(2)

	logger.With(log.Fields{
		"mode": log.Stringer(m),
	}).Log("log with .String() key/value pair")

	require.Contains(t, out.String(), `mode=PRODUCTION`)
}
