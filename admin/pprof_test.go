// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.
package admin

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdmin__profileEnabled(t *testing.T) {
	cases := map[string]bool{
		// enable
		"yes":    true,
		" true ": true,
		"":       true,
		// disable
		"no":       false,
		"jsadlsaj": false,
	}
	for value, enabled := range cases {
		t.Setenv("PPROF_TESTING_VALUE", fmt.Sprintf("%v", enabled))

		if v := profileEnabled("TESTING_VALUE"); v != enabled {
			t.Errorf("value=%q, got=%v, expected=%v", value, v, enabled)
		}
	}
}

func TestPprofEnabled(t *testing.T) {
	require.True(t, pprofEnabled(Opts{}))
	require.True(t, pprofEnabled(Opts{Pprof: &Pprof{Enabled: true}}))
	require.False(t, pprofEnabled(Opts{Pprof: &Pprof{Enabled: false}}))
	require.False(t, pprofEnabled(Opts{Pprof: &Pprof{}}))
}

func TestHandler_PprofDefaultEnabled(t *testing.T) {
	rec := servePprof(t, handler(Opts{}), "/debug/pprof/cmdline")
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestHandler_PprofDisabled(t *testing.T) {
	h := handler(Opts{Pprof: &Pprof{Enabled: false}})

	rec := servePprof(t, h, "/debug/pprof/")
	require.Equal(t, http.StatusNotFound, rec.Code)

	rec = servePprof(t, h, "/debug/pprof/cmdline")
	require.Equal(t, http.StatusNotFound, rec.Code)

	rec = servePprof(t, h, "/debug/pprof/heap")
	require.Equal(t, http.StatusNotFound, rec.Code)

	rec = servePprof(t, h, "/metrics")
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestHandler_PprofExplicitlyEnabled(t *testing.T) {
	h := handler(Opts{Pprof: &Pprof{Enabled: true}})

	rec := servePprof(t, h, "/debug/pprof/")
	require.Equal(t, http.StatusOK, rec.Code)

	rec = servePprof(t, h, "/debug/pprof/cmdline")
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestHandler_PprofSecret(t *testing.T) {
	const secret = "test-pprof-secret"
	h := handler(Opts{Pprof: &Pprof{Enabled: true, Secret: secret}})

	rec := servePprof(t, h, "/debug/pprof/cmdline")
	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Equal(t, `Bearer realm="pprof"`, rec.Header().Get("WWW-Authenticate"))

	rec = servePprof(t, h, "/debug/pprof/cmdline", header{"Authorization", "Bearer " + secret})
	require.Equal(t, http.StatusOK, rec.Code)

	rec = servePprof(t, h, "/debug/pprof/cmdline", header{"X-Pprof-Token", secret})
	require.Equal(t, http.StatusOK, rec.Code)

	rec = servePprof(t, h, "/debug/pprof/cmdline", header{"Authorization", "Bearer wrong"})
	require.Equal(t, http.StatusUnauthorized, rec.Code)

	rec = servePprof(t, h, "/debug/pprof/cmdline", header{"X-Pprof-Token", "wrong"})
	require.Equal(t, http.StatusUnauthorized, rec.Code)

	rec = servePprof(t, h, "/metrics")
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestBlockProfileEnabled(t *testing.T) {
	require.True(t, blockProfileEnabled(Opts{}))
	require.False(t, blockProfileEnabled(Opts{Pprof: &Pprof{Enabled: true}}))
	require.False(t, blockProfileEnabled(Opts{Pprof: &Pprof{Block: true}}))
	require.True(t, blockProfileEnabled(Opts{Pprof: &Pprof{Enabled: true, Block: true}}))
}

func TestMutexProfileEnabled(t *testing.T) {
	require.True(t, mutexProfileEnabled(Opts{}))
	require.False(t, mutexProfileEnabled(Opts{Pprof: &Pprof{Enabled: true}}))
	require.False(t, mutexProfileEnabled(Opts{Pprof: &Pprof{Mutex: true}}))
	require.True(t, mutexProfileEnabled(Opts{Pprof: &Pprof{Enabled: true, Mutex: true}}))
}

func TestHandler_BlockMutexDefaultWhenPprofSet(t *testing.T) {
	h := handler(Opts{Pprof: &Pprof{Enabled: true}})

	rec := servePprof(t, h, "/debug/pprof/block")
	require.Equal(t, http.StatusNotFound, rec.Code)

	rec = servePprof(t, h, "/debug/pprof/mutex")
	require.Equal(t, http.StatusNotFound, rec.Code)

	rec = servePprof(t, h, "/debug/pprof/heap")
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestHandler_BlockMutexExplicitlyEnabled(t *testing.T) {
	h := handler(Opts{Pprof: &Pprof{Enabled: true, Block: true, Mutex: true}})

	rec := servePprof(t, h, "/debug/pprof/block")
	require.Equal(t, http.StatusOK, rec.Code)

	rec = servePprof(t, h, "/debug/pprof/mutex")
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestHandler_PprofEmptySecretUnauthenticated(t *testing.T) {
	h := handler(Opts{Pprof: &Pprof{Enabled: true, Secret: ""}})
	rec := servePprof(t, h, "/debug/pprof/cmdline")
	require.Equal(t, http.StatusOK, rec.Code)
}

type header struct {
	key, value string
}

func servePprof(t *testing.T, h http.Handler, path string, headers ...header) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	for _, hdr := range headers {
		req.Header.Set(hdr.key, hdr.value)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}
