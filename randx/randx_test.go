package randx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBetween(t *testing.T) {
	lower, upper := 100, 250

	for i := 0; i < 1000; i++ {
		n, err := Between(lower, upper)
		require.NoError(t, err)

		require.GreaterOrEqual(t, n, int64(lower))
		require.Less(t, n, int64(upper))
	}
}

func TestBetween_ZeroLower(t *testing.T) {
	for i := 0; i < 1000; i++ {
		n, err := Between(0, 10)
		require.NoError(t, err)

		require.GreaterOrEqual(t, n, int64(0))
		require.Less(t, n, int64(10))
	}
}

func TestBetween_InvalidBounds(t *testing.T) {
	n, err := Between(250, 100)
	require.Error(t, err)
	require.Equal(t, int64(0), n)

	n, err = Between(100, 100)
	require.Error(t, err)
	require.Equal(t, int64(0), n)
}

func TestMust(t *testing.T) {
	lower, upper := 1000, 25000

	for i := 0; i < 1000; i++ {
		n := Must(Between(lower, upper))

		require.GreaterOrEqual(t, n, int64(lower))
		require.Less(t, n, int64(upper))
	}
}

func TestMust_Panics(t *testing.T) {
	require.Panics(t, func() {
		Must(Between(100, 100))
	})
}
