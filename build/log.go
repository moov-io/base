package build

import (
	"runtime/debug"
	"strings"

	"github.com/moov-io/base/log"
)

func Log(logger log.Logger) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		logger.Error().Log("unable to read build info, pleasure ensure go module support")
	}

	logger = logger.With(log.Fields{
		"build_path":       log.String(info.Path),
		"build_go_version": log.String(info.GoVersion),
	})

	for _, mod := range info.Deps {
		mod = runningModule(mod)

		if strings.Contains(strings.ToLower(mod.Path), "/moov") {
			logger.With(log.Fields{
				"build_mod_path":    log.String(mod.Path),
				"build_mod_version": log.String(mod.Version),
			}).Log("")
		}
	}
}

// Recurse through all the replaces to find whats actually running
func runningModule(mod *debug.Module) *debug.Module {
	if mod.Replace != nil {
		return runningModule(mod.Replace)
	} else {
		return mod
	}
}
