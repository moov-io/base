// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package strx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOr(t *testing.T) {
	if v := Or(); v != "" {
		t.Errorf("got %q", v)
	}
	if v := Or("", "backup"); v != "backup" {
		t.Errorf("got %q", v)
	}
	if v := Or("primary", ""); v != "primary" {
		t.Errorf("got %q", v)
	}
	if v := Or("primary", "backup"); v != "primary" {
		t.Errorf("got %q", v)
	}
}

func TestYes(t *testing.T) {
	// accepted values
	require.True(t, Yes("yes"))
	require.True(t, Yes(" true "))

	// common, but unsupported
	require.False(t, Yes("on"))
	require.False(t, Yes("y"))
	require.False(t, Yes("no"))

	// explicit no values
	require.False(t, Yes("no"))
	require.False(t, Yes("false"))
	require.False(t, Yes(""))
}
