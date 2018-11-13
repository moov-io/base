// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package idempotent

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"unicode/utf8"
)

func TestIdempotent(t *testing.T) {
	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("X-Idempotency-Key", "test")

	key, seen := FromRequest(req, nil)
	if key != "test" {
		t.Errorf("got %q", key)
	}
	if seen {
		t.Errorf("shouldn't be marked as seen")
	}

	// Do it all again to make sure
	key, seen = FromRequest(req, nil)
	if key != "test" {
		t.Errorf("got %q", key)
	}
	if seen {
		t.Errorf("shouldn't be marked as seen")
	}
}

func TestIdempotent__Empty(t *testing.T) {
	req := httptest.NewRequest("GET", "/ping", nil)

	key, seen := FromRequest(req, nil)
	if key != "" {
		t.Errorf("got %q", key)
	}
	if seen {
		t.Errorf("shouldn't be marked as seen")
	}

	// Do it all again to make sure
	key, seen = FromRequest(req, nil)
	if key != "" {
		t.Errorf("got %q", key)
	}
	if seen {
		t.Errorf("shouldn't be marked as seen")
	}
}

func TestIdempotent__SeenBefore(t *testing.T) {
	w := httptest.NewRecorder()
	SeenBefore(w)
	w.Flush()

	if w.Code != http.StatusPreconditionFailed {
		t.Errorf("got %d", w.Code)
	}
}

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
