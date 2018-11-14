// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/moov-io/base/idempotent/lru"
)

func TestResponse__Wrap(t *testing.T) {
	req := httptest.NewRequest("GET", "https://api.moov.io/v1/ach/ping", nil)
	req.Header.Set("Origin", "https://moov.io/demo")

	w := httptest.NewRecorder()
	ww := Wrap(nil, nil, w, req)
	ww.WriteHeader(http.StatusTeapot)
	w.Flush()

	if w.Code != http.StatusTeapot {
		t.Errorf("got HTTP code: %d", w.Code)
	}
	if v := w.Header().Get("Access-Control-Allow-Origin"); v == "" {
		t.Error("expected CORS heders")
	}
}

func TestResposne_EnsureHeaders(t *testing.T) {
	req := httptest.NewRequest("GET", "https://api.moov.io/v1/ach/ping", nil)
	req.Header.Set("x-user-id", "junk")
	req.Header.Set("Origin", "https://moov.io/demo")

	rec := lru.New()
	w := httptest.NewRecorder()

	ww, err := EnsureHeaders(nil, nil, rec, w, req)
	if err != nil {
		t.Error(err)
	}

	ww.WriteHeader(http.StatusTeapot)
	w.Flush()

	if w.Code != http.StatusTeapot {
		t.Errorf("got HTTP code: %d", w.Code)
	}
	if v := w.Header().Get("Access-Control-Allow-Origin"); v == "" {
		t.Error("expected CORS heders")
	}
}

func TestResponse__EnsureHeadersFail(t *testing.T) {
	req := httptest.NewRequest("GET", "https://api.moov.io/v1/ach/ping", nil)

	w := httptest.NewRecorder()
	ww, err := EnsureHeaders(nil, nil, nil, w, req)
	if err == nil {
		t.Errorf("expected error")
	}

	ww.WriteHeader(http.StatusTeapot)
	w.Flush()

	if w.Code != http.StatusForbidden {
		t.Errorf("got HTTP code: %d", w.Code)
	}
}
