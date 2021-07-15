// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.
package database

import (
	"fmt"
)

// ErrOpenConnections describes the number of open connections that should have been closed by a call to Close().
// All queries/transactions should call Close() to prevent unused, open connections.
type ErrOpenConnections struct {
	Database       string
	NumConnections int
}

func (e ErrOpenConnections) Error() string {
	return fmt.Sprintf("found %d open connection(s) in %s", e.NumConnections, e.Database)
}
