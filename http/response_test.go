// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
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
