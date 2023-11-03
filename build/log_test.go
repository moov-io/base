package build_test

import (
	"testing"

	"github.com/moov-io/base/build"
	"github.com/moov-io/base/log"
)

func Test_LogDeps(t *testing.T) {
	_, logger := log.NewBufferLogger()

	// Running it purely to make sure it doesn't panic as it requires a compiled binary to work.
	build.Log(logger)
}
