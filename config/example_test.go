package config_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/moov-io/base"
	"github.com/moov-io/base/config"
	"github.com/moov-io/base/log"
)

func ExampleService_LoadFromFS() {
	type GlobalConfig struct {
		Config struct {
			Default string
			Values  []string
			Zero    string
		}
	}

	os.Unsetenv(config.APP_CONFIG)
	os.Unsetenv(config.APP_CONFIG_SECRETS)

	var cfg GlobalConfig
	svc := config.NewService(log.NewNopLogger())
	if err := svc.LoadFromFS(&cfg, base.ConfigDefaults); err != nil {
		panic(err)
	}
	fmt.Println(cfg.Config.Default, cfg.Config.Zero, cfg.Config.Values)
	// Output:
	// default hero [default]
}

func ExampleService_MergeEnvironments() {
	type GlobalConfig struct {
		Config struct {
			Default string
			App     string
			Secret  string
			Values  []string
			Zero    string
		}
	}

	dir, err := os.MkdirTemp("", "base-config-example")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	appPath := filepath.Join(dir, "app.yml")
	secretPath := filepath.Join(dir, "secrets.yml")
	if err := os.WriteFile(appPath, []byte("Config:\n  App: app\n"), 0o600); err != nil {
		panic(err)
	}
	if err := os.WriteFile(secretPath, []byte("Config:\n  Secret: keep secret!\n"), 0o600); err != nil {
		panic(err)
	}

	// Overlay files replace matching keys after the embedded defaults.
	os.Setenv(config.APP_CONFIG, appPath)
	os.Setenv(config.APP_CONFIG_SECRETS, secretPath)
	defer os.Unsetenv(config.APP_CONFIG)
	defer os.Unsetenv(config.APP_CONFIG_SECRETS)

	var cfg GlobalConfig
	svc := config.NewService(log.NewNopLogger())
	if err := svc.LoadFromFS(&cfg, base.ConfigDefaults); err != nil {
		panic(err)
	}
	fmt.Println(cfg.Config.Default, cfg.Config.App, cfg.Config.Secret)
	// Output:
	// default app keep secret!
}
