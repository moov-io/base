// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package k8s

import (
	"os"
)

var serviceAccountFilepaths = []string{
	// https://stackoverflow.com/a/49045575
	"/var/run/secrets/kubernetes.io",

	// https://github.com/hashicorp/vault/blob/master/command/agent/auth/kubernetes/kubernetes.go#L20
	"/var/run/secrets/kubernetes.io/serviceaccount/token",
}

// Inside returns true if ran from inside a Kubernetes cluster.
func Inside() bool {
	// Allow a user override path
	paths := append(serviceAccountFilepaths, os.Getenv("KUBERNETES_SERVICE_ACCOUNT_FILEPATH"))

	for i := range paths {
		if _, err := os.Stat(paths[i]); err == nil {
			return true
		}
	}

	return false
}
