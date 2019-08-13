// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

// Package lru is a simple inmemory Recorder implementation. This implementation
// is intended for simple usecases (local dev) and not production workloads.
package lru

import (
	hashlru "github.com/hashicorp/golang-lru"
)

var (
	// This value is intended for local dev / inmem caching.
	defaultLRUSize = 1024

	// Used for LRU value
	defaultValue = struct{}{}
)

// New returns an in-memory LRU instance
func New() *Mem {
	cache, _ := hashlru.New(defaultLRUSize)
	return &Mem{
		cache: cache,
	}
}

// Mem represents an in-memory LRU
type Mem struct {
	cache *hashlru.Cache
}

// SeenBefore sets a HTTP response code as an error for previously seen idempotency keys.
func (m *Mem) SeenBefore(key string) bool {
	if m == nil {
		return false
	}

	seen := m.cache.Contains(key)
	if !seen {
		m.cache.Add(key, defaultValue)
	}
	return seen
}
