// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package docker

import (
	"os"
	"runtime"
	"testing"
)

func TestDocker(t *testing.T) {
	osname := os.Getenv("TRAVIS_OS_NAME")
	if osname == "" {
		t.Skip("docker: only testing in CI")
	}

	if runtime.GOOS == "darwin" {
		if Enabled() {
			t.Error("docker on travis-ci osx/macOS available now?")
		}
	} else {
		if !Enabled() {
			t.Errorf("expected Docker to be enabled in %s CI", runtime.GOOS)
		}
	}
}
