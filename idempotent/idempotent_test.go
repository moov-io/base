// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package idempotent

import (
	"testing"
	"unicode/utf8"
)

func TestIdempotent__truncate(t *testing.T) {
	s1 := "1234567890123456789012345678901234567890" // 40 characters
	s2 := truncate(s1)
	if s1 == s2 {
		t.Errorf("strings shouldn't match")
	}
	if n := utf8.RuneCountInString(s2); n != maxIdempotencyKeyLength {
		t.Errorf("s2 length is %d", n)
	}
}
