// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package http

import (
	"net/http"
	"time"

	"github.com/moov-io/base/log"

	"github.com/go-kit/kit/metrics"
)

// ResponseWriter implements Go's standard library http.ResponseWriter to complete HTTP requests
type ResponseWriter struct {
	http.ResponseWriter

	start   time.Time
	request *http.Request
	metric  metrics.Histogram

	headersWritten bool // set on WriteHeader

	log log.Logger
}

// WriteHeader sends an HTTP response header with the provided status code, records response duration,
// and optionally records the HTTP metadata in a go-kit log.Logger
func (w *ResponseWriter) WriteHeader(code int) {
	if w == nil || w.headersWritten {
		return
	}
	w.headersWritten = true

	// Headers
	SetAccessControlAllowHeaders(w, w.request.Header.Get("Origin"))
	defer w.ResponseWriter.WriteHeader(code)

	// Record route timing
	diff := time.Since(w.start)
	if w.metric != nil {
		w.metric.Observe(diff.Seconds())
	}

	// Skip Go's content sniff here to speed up response timing for client
	if w.ResponseWriter.Header().Get("Content-Type") == "" {
		w.ResponseWriter.Header().Set("Content-Type", "text/plain")
		w.ResponseWriter.Header().Set("X-Content-Type-Options", "nosniff")
	}

	if requestID := GetRequestID(w.request); requestID != "" && w.log != nil {
		w.log.With(log.Fields{
			"method":    log.String(w.request.Method),
			"path":      log.String(w.request.URL.Path),
			"status":    log.Int(code),
			"duration":  log.TimeDuration(diff),
			"requestID": log.String(requestID),
		}).Send()
	}
}

// Wrap returns a ResponseWriter usable by applications. No parts of the Request are inspected or ResponseWriter modified.
func Wrap(logger log.Logger, m metrics.Histogram, w http.ResponseWriter, r *http.Request) *ResponseWriter {
	now := time.Now()
	return &ResponseWriter{
		ResponseWriter: w,
		start:          now,
		request:        r,
		metric:         m,
		log:            logger,
	}
}
