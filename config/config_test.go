package config_test

import (
	"embed"
	"os"
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
	Custom  string
}

func Test_Load(t *testing.T) {
	os.Setenv(config.APP_CONFIG, "../configs/config.app.yml")
	os.Setenv(config.APP_CONFIG_SECRETS, "../configs/config.secrets.yml")
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
}

//go:embed testdata/*.yml
var configs embed.FS

func Test_LoadFile(t *testing.T) {
	cfg := &GlobalConfigModel{}

	service := config.NewService(log.NewDefaultLogger())

	file, err := configs.Open("testdata/config.yml")
	require.NoError(t, err)

	err = service.LoadFromReader(file, cfg)
	require.NoError(t, err)

	require.Equal(t, "custom", cfg.Config.Custom)
}
