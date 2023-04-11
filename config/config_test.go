package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/moov-io/base/config"
	"github.com/moov-io/base/log"
	"github.com/stretchr/testify/require"
)

type GlobalConfigModel struct {
	Config ConfigModel
}

type ConfigModel struct {
	Default string
	App     string
	Secret  string
	Values  []string
	Zero    string
}

func Test_Load(t *testing.T) {
	os.Setenv(config.APP_CONFIG, filepath.Join("..", "configs", "config.app.yml"))
	os.Setenv(config.APP_CONFIG_SECRETS, filepath.Join("..", "configs", "config.secrets.yml"))
	t.Cleanup(func() {
		os.Unsetenv(config.APP_CONFIG)
		os.Unsetenv(config.APP_CONFIG_SECRETS)
	})

	cfg := &GlobalConfigModel{}

	service := config.NewService(log.NewDefaultLogger())
	err := service.Load(cfg)
	require.Nil(t, err)

	require.Equal(t, "default", cfg.Config.Default)
	require.Equal(t, "app", cfg.Config.App)
	require.Equal(t, "keep secret!", cfg.Config.Secret)

	// This test documents some unexpected behavior where slices are merged and not overwritten.
	// Slices in secrets should overwrite the slice in app and default.
	require.Len(t, cfg.Config.Values, 1)
	require.Equal(t, "secret", cfg.Config.Values[0])

	require.Equal(t, "", cfg.Config.Zero)

	// Verify attempting to load from our default file errors on extra fields
	cfg = &GlobalConfigModel{}
	err = service.LoadFile("/configs/config.extra.yml", &cfg)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), `'Config' has invalid keys: extra`)

	// Verify attempting to load additional fields via env vars errors out
	os.Setenv(config.APP_CONFIG, filepath.Join("..", "configs", "config.extra.yml"))
	cfg = &GlobalConfigModel{}
	err = service.Load(cfg)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), `'Config' has invalid keys: extra`)
}
