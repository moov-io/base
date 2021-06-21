// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package mask

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMaskPassword(t *testing.T) {
	cases := []struct {
		input, expected string
	}{
		{"", "*****"},
		{"ab", "*****"},
		{"abcde", "a*****e"},
		{"123456", "1*****6"},
		{"password", "p*****d"},
	}
	for i := range cases {
		output := Password(cases[i].input)
		require.Equal(t, cases[i].expected, output)
	}
}
