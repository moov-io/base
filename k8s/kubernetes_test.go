// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package k8s

import (
	"os"
	"testing"
)

func TestK8SInside(t *testing.T) {
	if Inside() {
		t.Errorf("not inside k8s")
	}

	// Create a file and pretend it's the Kubernetes service account filepath
	fd, err := os.Create("k8s-service-account")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(fd.Name())
	if err := fd.Sync(); err != nil {
		t.Fatal(err)
	}

	// Pretend
	t.Setenv("KUBERNETES_SERVICE_ACCOUNT_FILEPATH", fd.Name())

	if !Inside() {
		t.Error("we should be pretending to be in a Kubernetes cluster")
	}
}
