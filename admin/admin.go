// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

// Package admin implements an http.Server which can be used for operations
// and monitoring tools. It's designed to be shipped (and ran) inside
// an existing Go service.
package admin

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Opts struct {
	Addr    string
	Timeout time.Duration

	// Pprof configures /debug/pprof routes. A nil value keeps historical
	// behavior: all profiles are registered without authentication.
	Pprof *Pprof
}

// Pprof controls whether the admin server registers /debug/pprof routes.
type Pprof struct {
	// Enabled registers /debug/pprof handlers. The zero value omits them.
	Enabled bool

	// Secret, when non-empty, requires callers to present the value as
	// `Authorization: Bearer <secret>` or `X-Pprof-Token: <secret>`.
	// An empty Secret keeps unauthenticated access.
	Secret string

	// Block enables runtime block profiling and /debug/pprof/block.
	// Ignored when Opts.Pprof is nil, which keeps historical PPROF_BLOCK behavior.
	Block bool

	// Mutex enables runtime mutex profiling and /debug/pprof/mutex.
	// Ignored when Opts.Pprof is nil, which keeps historical PPROF_MUTEX behavior.
	Mutex bool
}

// New returns an admin.Server instance that handles Prometheus metrics and pprof requests.
// Callers can use ':0' to bind onto a random port and call BindAddr() for the address.
func New(opts Opts) (*Server, error) {
	timeout, _ := time.ParseDuration("45s")
	if opts.Timeout >= 0*time.Second {
		timeout = opts.Timeout
	}

	var listener net.Listener
	var err error
	if opts.Addr == "" || opts.Addr == ":0" {
		listener, err = net.Listen("tcp", "127.0.0.1:0")
	} else {
		listener, err = net.Listen("tcp", opts.Addr)
	}
	if err != nil {
		return nil, fmt.Errorf("listening on %s failed: %v", opts.Addr, err)
	}

	router := handler(opts)
	svc := &Server{
		router:   router,
		listener: listener,
		svc: &http.Server{
			Addr:         listener.Addr().String(),
			Handler:      router,
			ReadTimeout:  timeout,
			WriteTimeout: timeout,
			IdleTimeout:  timeout,
		},
	}

	svc.AddHandler("/live", svc.livenessHandler())
	svc.AddHandler("/ready", svc.readinessHandler())
	return svc, nil
}

// Server represents a holder around a net/http Server which
// is used for admin endpoints. (i.e. metrics, healthcheck)
type Server struct {
	router   *mux.Router
	svc      *http.Server
	listener net.Listener

	liveChecks  []*healthCheck
	readyChecks []*healthCheck
}

// BindAddr returns the server's bind address. This is in Go's format so :8080 is valid.
func (s *Server) BindAddr() string {
	if s == nil || s.svc == nil {
		return ""
	}
	return s.listener.Addr().String()
}

func (s *Server) SetReadTimeout(timeout time.Duration) {
	if s == nil || s.svc == nil {
		return
	}
	s.svc.ReadTimeout = timeout
}

func (s *Server) SetWriteTimeout(timeout time.Duration) {
	if s == nil || s.svc == nil {
		return
	}
	s.svc.WriteTimeout = timeout
}

func (s *Server) SetIdleTimeout(timeout time.Duration) {
	if s == nil || s.svc == nil {
		return
	}
	s.svc.IdleTimeout = timeout
}

// Listen brings up the admin HTTP server. This call blocks until the server is Shutdown or panics.
func (s *Server) Listen() error {
	if s == nil || s.svc == nil || s.listener == nil {
		return nil
	}
	return s.svc.Serve(s.listener)
}

// Shutdown unbinds the HTTP server.
func (s *Server) Shutdown() {
	if s == nil || s.svc == nil {
		return
	}
	s.svc.Shutdown(context.TODO())
}

// AddHandler will append an http.HandlerFunc to the admin Server
func (s *Server) AddHandler(path string, hf http.HandlerFunc) {
	s.router.HandleFunc(path, hf)
}

// AddVersionHandler will append 'GET /version' route returning the provided version
func (s *Server) AddVersionHandler(version string) {
	s.AddHandler("/version", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(version))
	})
}

// Subrouter creates and returns a subrouter with the specific prefix.
//
// The returned subrouter can use middleware without impacting
// the parent router. For example:
//
//	svr, err := New(Opts{
//		Addr: ":9090",
//	})
//
//	subRouter := svr.Subrouter("/prefix")
//	subRouter.Use(someMiddleware)
//	subRouter.HandleFunc("/resource", ResourceHandler)
//
// Here, requests for "/prefix/resource" would go through someMiddleware while
// the liveliness and readiness routes added to the parent router by New()
// would not.
func (s *Server) Subrouter(pathPrefix string) *mux.Router {
	return s.router.PathPrefix(pathPrefix).Subrouter()
}

// pprofEnabled reports whether /debug/pprof routes should be registered.
// A nil Opts.Pprof keeps historical behavior (enabled).
func pprofEnabled(opts Opts) bool {
	if opts.Pprof == nil {
		return true
	}
	return opts.Pprof.Enabled
}

func blockProfileEnabled(opts Opts) bool {
	if opts.Pprof != nil {
		return opts.Pprof.Enabled && opts.Pprof.Block
	}
	return profileEnabled("block")
}

func mutexProfileEnabled(opts Opts) bool {
	if opts.Pprof != nil {
		return opts.Pprof.Enabled && opts.Pprof.Mutex
	}
	return profileEnabled("mutex")
}

// profileEnabled returns if a given pprof handler should be
// enabled according to pprofHandlers and the PPROF_* environment
// variables.
//
// These profiles can be disabled by setting the appropriate PPROF_*
// environment variable. (i.e. PPROF_ALLOCS=no)
//
// An empty string, "yes", or "true" enables the profile. Any other
// value disables the profile.
func profileEnabled(name string) bool {
	k := fmt.Sprintf("PPROF_%s", strings.ToUpper(name))
	v := strings.ToLower(os.Getenv(k))
	return v == "" || v == "yes" || v == "true"
}

// Handler returns an http.Handler for the admin http service.
// This contains metrics and pprof handlers.
//
// No metrics specific to the handler are recorded.
//
// We only want to expose on the admin servlet because these
// profiles/dumps can contain sensitive info (raw memory).
// Handler uses default Opts, so pprof routes are registered.
func Handler() http.Handler {
	return handler(Opts{})
}

func handler(opts Opts) *mux.Router {
	r := mux.NewRouter()

	// prometheus metrics
	r.Path("/metrics").Handler(promhttp.Handler())

	if !pprofEnabled(opts) {
		return r
	}

	handle := func(path string, h http.Handler) {
		r.Handle(path, wrapPprof(opts, h))
	}

	// always register index and cmdline handlers
	handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))

	if profileEnabled("profile") {
		handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	}
	if profileEnabled("symbol") {
		handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	}
	if profileEnabled("trace") {
		handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	}

	// Register runtime/pprof handlers
	if profileEnabled("allocs") {
		handle("/debug/pprof/allocs", pprof.Handler("allocs"))
	}
	if blockProfileEnabled(opts) {
		runtime.SetBlockProfileRate(1)
		handle("/debug/pprof/block", pprof.Handler("block"))
	}
	if profileEnabled("goroutine") {
		handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	}
	if profileEnabled("heap") {
		handle("/debug/pprof/heap", pprof.Handler("heap"))
	}
	if mutexProfileEnabled(opts) {
		runtime.SetMutexProfileFraction(1)
		handle("/debug/pprof/mutex", pprof.Handler("mutex"))
	}
	if profileEnabled("threadcreate") {
		handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	}
	if profileEnabled("goroutineleak") {
		handle("/debug/pprof/goroutineleak", pprof.Handler("goroutineleak"))
	}

	return r
}

func wrapPprof(opts Opts, next http.Handler) http.Handler {
	if opts.Pprof == nil || opts.Pprof.Secret == "" {
		return next
	}
	want := sha256.Sum256([]byte(opts.Pprof.Secret))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := sha256.Sum256([]byte(pprofToken(r)))
		if subtle.ConstantTimeCompare(got[:], want[:]) != 1 {
			w.Header().Set("WWW-Authenticate", `Bearer realm="pprof"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func pprofToken(r *http.Request) string {
	if v := strings.TrimSpace(r.Header.Get("X-Pprof-Token")); v != "" {
		return v
	}
	const prefix = "Bearer "
	auth := r.Header.Get("Authorization")
	if len(auth) >= len(prefix) && strings.EqualFold(auth[:len(prefix)], prefix) {
		return strings.TrimSpace(auth[len(prefix):])
	}
	return ""
}
