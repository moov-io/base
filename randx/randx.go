package randx

import (
	"crypto/rand"
	"math/big"
)

// Between will return a randomly generated integer within the lower and upper bounds provided.
func Between(lower, upper int) (int64, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(lower)))
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
