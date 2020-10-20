package database

import "fmt"

// ErrOpenConnections describes the number of open connections that should have been closed by a call to Close().
// All queries/transactions should call Close() to prevent unused, open connections.
type ErrOpenConnections struct {
	Database       string
	NumConnections int
}

func (e ErrOpenConnections) Error() string {
	return fmt.Sprintf("found %d open connection(s) in %s", e.NumConnections, e.Database)
}
