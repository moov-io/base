// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package idempotent

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
