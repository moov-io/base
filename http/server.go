// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package http

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/moov-io/base/strx"
)

const (
	maxHeaderLength = 36
)

// Problem writes err to w while also setting the HTTP status code, content-type and marshaling
// err as the response body.
func Problem(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

// InternalError writes err to w while also setting the HTTP status code, content-type and marshaling
// err as the response body.
//
// Returned is the calling file and line number: server.go:33
func InternalError(w http.ResponseWriter, err error) string {
	w.WriteHeader(http.StatusInternalServerError)

	pcs := make([]uintptr, 5) // some limit
	_ = runtime.Callers(1, pcs)

	file, line := "", 0

	// Sometimes InternalError will be wrapped by helper methods inside an application.
	// We should linear search our callers until we find one outside github.com/moov-io
	// because that likely represents the stdlib.
	//
	// Note: This might not work for code already outside github.com/moov-io, please report
	// feedback if this works or not.
	i, frames := 0, runtime.CallersFrames(pcs)
	for {
		f, more := frames.Next()
		if !more {
			break
		}

		// f.Function can either be an absolute path (/Users/...) or a package
		// (i.e. github.com/moov-io/...) so check for either.
		if strings.Contains(f.Function, "github.com/moov-io") || strings.HasPrefix(f.Function, "main.") {
			_, file, line, _ = runtime.Caller(i) // next caller
		}
		i++
	}

	// Get the filename, file was a full path
	_, file = filepath.Split(file)
	return fmt.Sprintf("%s:%d", file, line)
}

// AddCORSHandler captures Corss Origin Resource Sharing (CORS) requests
// by looking at all OPTIONS requests for the Origin header, parsing that
// and responding back with the other Access-Control-Allow-* headers.
//
// Docs: https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS
func AddCORSHandler(r *mux.Router) {
	r.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		SetAccessControlAllowHeaders(w, r.Header.Get("Origin"))
		w.WriteHeader(http.StatusOK)
	})
}

// SetAccessControlAllowHeaders writes Access-Control-Allow-* headers to a response to allow
// for further CORS-allowed requests.
func SetAccessControlAllowHeaders(w http.ResponseWriter, origin string) {
	// Access-Control-Allow-Origin can't be '*' with requests that send credentials.
	// Instead, we need to explicitly set the domain (from request's Origin header)
	//
	// Allow requests from anyone's localhost and only from secure pages.
	if strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "https://") {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Cookie,X-User-Id,X-Request-Id,Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}
}

// GetRequestID returns the Moov header value for request IDs
func GetRequestID(r *http.Request) string {
	return r.Header.Get("X-Request-Id")
}

// GetUserID returns the Moov userId from HTTP headers
func GetUserID(r *http.Request) string {
	return strx.Or(r.Header.Get("X-User"), r.Header.Get("X-User-Id"))
}

// GetSkipAndCount returns the skip and count pagination values from the query parameters
// skip is the number of records to pass over before starting a search
// count is the number of records to retrieve in the search
// exists indicates if skip or count was passed into the request URL
func GetSkipAndCount(r *http.Request) (skip int, count int, exists bool, errors []error) {
	skipVal := r.URL.Query().Get("skip")
	var err error
	skip, err = strconv.Atoi(skipVal)
	if err != nil && len(skipVal) > 0 {
		errors = append(errors, err)
		skip = 0
	}
	skip = int(math.Min(float64(skip), 10000))
	skip = int(math.Max(0, float64(skip)))

	countVal := r.URL.Query().Get("count")
	count, err = strconv.Atoi(countVal)
	if err != nil && len(countVal) > 0 {
		errors = append(errors, err)
		count = 0
	}
	count = int(math.Min(float64(count), 200))
	count = int(math.Max(0, float64(count)))
	exists = skipVal != "" || countVal != ""

	// If no errors, set defaults
	if len(errors) == 0 {
		if count == 0 {
			// TODO - Add this to a config somewhere?
			count = 20
		}
		if skip > 1 {
			skip = int(int64(skip))
		}
	}
	return skip, count, exists, errors
}
