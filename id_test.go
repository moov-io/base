// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package base

import (
	"testing"
)

func TestID(t *testing.T) {
	for i := 0; i < 1000; i++ {
		if v := ID(); v == "" {
			t.Error("got empty ID")
		}
	}
}
