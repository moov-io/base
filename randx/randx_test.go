package randx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBetween(t *testing.T) {
	lower, upper := 100, 250

	n, err := Between(lower, upper)
	require.NoError(t, err)

	if n < int64(lower) || n > int64(upper) {
		t.Fatalf("%d falls outside of %d and %d", n, lower, upper)
	}
}

func TestMust(t *testing.T) {
	lower, upper := 1000, 25000

	n := Must(Between(lower, upper))

	if n < int64(lower) || n > int64(upper) {
		t.Fatalf("%d falls outside of %d and %d", n, lower, upper)
	}
}
