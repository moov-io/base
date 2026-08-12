package randx

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// Between will return a randomly generated integer within the lower (inclusive)
// and upper (exclusive) bounds provided.
func Between(lower, upper int) (int64, error) {
	if upper <= lower {
		return 0, fmt.Errorf("randx.Between: upper bound %d must be greater than lower bound %d", upper, lower)
	}

	n, err := rand.Int(rand.Reader, big.NewInt(int64(upper-lower)))
	if err != nil {
		return 0, err
	}
	return n.Int64() + int64(lower), nil
}

// Must is a helper that wraps a call to Between and panics if the error is non-nil.
func Must(n int64, err error) int64 {
	if err != nil {
		panic(err) //nolint:forbidigo
	}
	return n
}
