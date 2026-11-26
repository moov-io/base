# Config

The `config` package loads YAML into a struct.

## Load methods

- `Load(config)` reads `/configs/config.default.yml` through [pkger](https://github.com/markbates/pkger). Prefer this only when the service still embeds files with pkger.
- `LoadFromFS(config, fs)` reads `configs/config.default.yml` from an `fs.FS`. Use this with `go:embed`. `github.com/moov-io/base` exports `base.ConfigDefaults` as an example filesystem.

Both methods then call `MergeEnvironments`:

1. If `APP_CONFIG` is set to a file path, merge that YAML file.
2. If `APP_CONFIG_SECRETS` is set to a file path, merge that YAML file (this file wins when the same key is in both overlays).
3. Decode the result into `config` with `UnmarshalExact`.

`UnmarshalExact` fails when the YAML has a key that is not on the struct (`ErrorUnused = true`). A key in an overlay replaces the previous value (`ZeroFields = true`).

## Example

```go
package main

import (
	"fmt"

	"github.com/moov-io/base"
	"github.com/moov-io/base/config"
	"github.com/moov-io/base/log"
)

type GlobalConfig struct {
	Config struct {
		Default string
		Values  []string
		Zero    string
	}
}

func main() {
	var cfg GlobalConfig
	svc := config.NewService(log.NewDefaultLogger())
	if err := svc.LoadFromFS(&cfg, base.ConfigDefaults); err != nil {
		panic(err)
	}
	fmt.Println(cfg.Config.Default, cfg.Config.Zero)
}
```

Set `APP_CONFIG` and `APP_CONFIG_SECRETS` to file paths when you need overlays. A compile-checked walkthrough is in [example_test.go](example_test.go).
