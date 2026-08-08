// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

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

// CORSAllowedOriginsEnv is a comma-separated list of exact Origins permitted for
// credentialed CORS. Example: "https://moov.io,https://dashboard.moov.io".
//
// Services that use AddCORSHandler, SetAccessControlAllowHeaders, or Wrap honor this
// allowlist automatically after upgrading moov-io/base — set the env var in deploy.
// Local development origins (http://localhost[:port], http://127.0.0.1[:port]) remain allowed.
const CORSAllowedOriginsEnv = "MOOV_CORS_ALLOW_ORIGINS"

var (
	corsAllowlistMu   sync.RWMutex
	corsAllowlist     map[string]struct{}
	corsAllowlistInit sync.Once
)

func parseCORSAllowlist(value string) map[string]struct{} {
	out := make(map[string]struct{})
	for _, part := range strings.Split(value, ",") {
		origin := strings.TrimSpace(part)
		if origin != "" {
			out[origin] = struct{}{}
		}
	}
	return out
}

func loadCORSAllowlist() map[string]struct{} {
	corsAllowlistInit.Do(func() {
		corsAllowlistMu.Lock()
		defer corsAllowlistMu.Unlock()
		if corsAllowlist == nil {
			corsAllowlist = parseCORSAllowlist(os.Getenv(CORSAllowedOriginsEnv))
		}
	})
	corsAllowlistMu.RLock()
	defer corsAllowlistMu.RUnlock()
	return corsAllowlist
}

// SetCORSAllowedOrigins replaces the in-process CORS allowlist with the given exact
// Origins. Pass nil or empty to clear. Useful for tests and for services that load
// allowed origins from config rather than (or in addition to) the env var.
//
// Safe for concurrent use with SetAccessControlAllowHeaders / Wrap.
func SetCORSAllowedOrigins(origins []string) {
	next := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			next[origin] = struct{}{}
		}
	}
	corsAllowlistMu.Lock()
	defer corsAllowlistMu.Unlock()
	// Ensure loadCORSAllowlist will not overwrite a programmatic set from env later.
	corsAllowlistInit.Do(func() {})
	corsAllowlist = next
}

// ResetCORSAllowlistForTest clears the cached allowlist so the next use reloads from
// MOOV_CORS_ALLOW_ORIGINS. Intended for tests only.
func ResetCORSAllowlistForTest() {
	corsAllowlistMu.Lock()
	defer corsAllowlistMu.Unlock()
	corsAllowlistInit = sync.Once{}
	corsAllowlist = nil
}

// OriginAllowedForCORS reports whether origin may receive credentialed CORS headers.
// Empty origins are rejected. Localhost / 127.0.0.1 HTTP origins are always allowed
// for local development; all other origins must appear in the allowlist.
func OriginAllowedForCORS(origin string) bool {
	if origin == "" {
		return false
	}
	// Dev convenience: loopback HTTP origins (with or without explicit port).
	if isLocalDevOrigin(origin) {
		return true
	}
	// Allowlists are replaced wholesale (never mutated in place), so indexing the
	// snapshot returned by loadCORSAllowlist is safe without holding the lock.
	_, ok := loadCORSAllowlist()[origin]
	return ok
}

func isLocalDevOrigin(origin string) bool {
	switch {
	case origin == "http://localhost", origin == "http://127.0.0.1":
		return true
	case strings.HasPrefix(origin, "http://localhost:"),
		strings.HasPrefix(origin, "http://127.0.0.1:"):
		return true
	default:
		return false
	}
}

// SetAccessControlAllowHeaders writes Access-Control-Allow-* headers to a response to allow
// for further CORS-allowed requests. Used by AddCORSHandler and Wrap; services can call it
// directly (e.g. ACH go-kit ServerAfter) and share the same allowlist.
func SetAccessControlAllowHeaders(w http.ResponseWriter, origin string) {
	// Access-Control-Allow-Origin can't be '*' with requests that send credentials.
	// Reflect only explicitly allowlisted Origins (plus localhost for local dev).
	// Never reflect arbitrary https:// Origins with Allow-Credentials: true.
	if !OriginAllowedForCORS(origin) {
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Cookie,X-User-Id,X-Request-Id,Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	// Vary so caches don't serve one origin's ACAO to another.
	w.Header().Add("Vary", "Origin")
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
// - skip is the number of records to pass over before starting a search (max math.MaxInt32)
// - count is the number of records to retrieve in the search  (max 10,000)
// - exists indicates if skip or count was passed into the request URL
func GetSkipAndCount[T *http.Request | string](t T) (skip int, count int, exists bool, err error) {
	return readSkipCount(t, math.MaxInt32, 10000)
}

// LimitedSkipCount returns the skip and count pagination values from the request's query parameters
// See GetSkipAndCount for descriptions of each parameter
func LimitedSkipCount[T *http.Request | string](t T, skipLimit, countLimit int) (skip int, count int, exists bool, err error) {
	return readSkipCount(t, skipLimit, countLimit)
}

func readSkipCount[T *http.Request | string](t T, skipMax, countMax int) (skip int, count int, exists bool, err error) {
	var query url.Values

	switch tt := any(t).(type) {
	case *http.Request:
		query = tt.URL.Query()
	case string:
		query, err = url.ParseQuery(tt)
		if err != nil {
			return skip, count, exists, fmt.Errorf("parsing query string: %w", err)
		}
	default:
		return skip, count, exists, fmt.Errorf("unsupported type %T for reading skip and count", t)
	}

	skipVal := query.Get("skip")
	countVal := query.Get("count")
	exists = len(skipVal) > 0 || len(countVal) > 0

	// Parse skip
	skip, err = strconv.Atoi(skipVal)
	if err != nil && len(skipVal) > 0 {
		skip = 0
		return skip, count, exists, err
	}
	// Limit skip
	skip = int(math.Min(float64(skip), float64(skipMax)))
	skip = int(math.Max(0, float64(skip)))

	// Parse count
	count, err = strconv.Atoi(countVal)
	if err != nil && len(countVal) > 0 {
		count = 0
		return skip, count, exists, err
	}

	// Limit count
	count = int(math.Min(float64(count), float64(countMax)))
	count = int(math.Max(0, float64(count)))
	if count == 0 {
		count = 200
	}

	return skip, count, exists, nil
}

type Direction string

const (
	Ascending  Direction = "ASC"
	Descending Direction = "DESC"
)

type OrderBy struct {
	Name      string
	Direction Direction
}

// GetOrderBy returns the field names and direction to order the response by
func GetOrderBy(r *http.Request) ([]OrderBy, error) {
	orderByParam := r.URL.Query().Get("orderBy")
	if orderByParam == "" {
		return []OrderBy{}, nil
	}

	paramSplit := strings.Split(orderByParam, ",")
	var orderBys []OrderBy
	for _, split := range paramSplit {
		orderBy := strings.Split(split, ":")
		if len(orderBy) != 2 {
			return nil, fmt.Errorf("invalid orderBy: %s", orderBy)
		}

		name := strings.TrimSpace(orderBy[0])
		if name == "" {
			return nil, errors.New("missing orderBy name")
		}

		directionStr := strings.TrimSpace(orderBy[1])
		if directionStr == "" {
			return nil, errors.New("missing orderBy direction")
		}
		directionStr = strings.ToLower(directionStr)

		var direction Direction
		if strings.HasPrefix(directionStr, "asc") {
			direction = Ascending
		} else if strings.HasPrefix(directionStr, "desc") {
			direction = Descending
		} else {
			return nil, fmt.Errorf("invalid orderBy direction: %s", direction)
		}

		orderBys = append(orderBys, OrderBy{
			Name:      name,
			Direction: direction,
		})
	}
	return orderBys, nil
}
