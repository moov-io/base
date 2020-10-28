// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/moov-io/base"
	"github.com/moov-io/base/idempotent/lru"

	"github.com/go-kit/kit/metrics/prometheus"
	"github.com/gorilla/mux"
	stdprometheus "github.com/prometheus/client_golang/prometheus"

	"github.com/moov-io/base/log"
)

var (
	routeHistogram = prometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
		Name: "http_response_duration_seconds",
		Help: "Histogram representing the http response durations",
	}, nil)
)

func TestResponse__Wrap(t *testing.T) {
	req := httptest.NewRequest("GET", "https://api.moov.io/v1/ach/ping", nil)
	req.Header.Set("Origin", "https://moov.io/demo")

	w := httptest.NewRecorder()

	ww := Wrap(nil, routeHistogram, w, req)
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
	req.Header.Set("x-request-id", base.ID())
	req.Header.Set("Origin", "https://moov.io/demo")

	rec := lru.New()
	w := httptest.NewRecorder()

	logger := log.NewDefaultLogger()
	ww, err := EnsureHeaders(logger, nil, rec, w, req)
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

func TestResposne__Idempotency(t *testing.T) {
	logger := log.NewNopLogger()
	idempot := lru.New()

	router := mux.NewRouter()
	router.Methods("GET").Path("/test").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w, err := EnsureHeaders(logger, nil, idempot, w, r)
		if err != nil {
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("PONG"))
	})

	key := base.ID()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("x-idempotency-key", key)
	req.Header.Set("x-user-id", base.ID())

	// mark the key as seen
	if seen := idempot.SeenBefore(key); seen {
		t.Errorf("shouldn't have been seen before")
	}

	// make our request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusPreconditionFailed {
		t.Errorf("got %d", w.Code)
	}

	// Key should be seen now
	if seen := idempot.SeenBefore(key); !seen {
		t.Errorf("should have seen %q", key)
	}
}
