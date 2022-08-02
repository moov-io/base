// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package base

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestID(t *testing.T) {
	for i := 0; i < 1000; i++ {
		id := ID()
		require.NotEmpty(t, id)
		require.Len(t, id, 40)
	}
}
