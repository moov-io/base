// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package bind

import (
	"testing"
)

func TestBind(t *testing.T) {
	// valid
	http := HTTP("auth")
	if http != ":8081" {
		t.Errorf("got %s", http)
	}
	admin := Admin("auth")
	if admin != ":9091" {
		t.Errorf("got %s", admin)
	}

	// invalid
	if v := HTTP("other"); v != "" {
		t.Errorf("got %s", v)
	}
	if v := Admin("other"); v != "" {
		t.Errorf("got %s", v)
	}
}
