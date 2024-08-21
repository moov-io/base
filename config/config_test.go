package config_test

import (
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/moov-io/base"
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

	Widgets map[string]Widget

	Search   SearchConfig
	Security SecurityConfig
}

type Widget struct {
	Name        string
	Credentials Credentials
	Nested      Nested
}

type Credentials struct {
	Username string
	Password string
}

type Nested struct {
	Nested2 Nested2
}

type Nested2 struct {
	Nested3 Nested3
}

type Nested3 struct {
	Value string
}

type SearchConfig struct {
	Patterns []*regexp.Regexp

	MaxResults int
	Timeout    time.Duration
}

type SecurityConfig struct {
	Audience []string `yaml:"x-audience"`
	Cluster  string   `yaml:"x-cluster"`
	Service  string   `yaml:"x-service"`
}

func Test_Load(t *testing.T) {
	t.Setenv(config.APP_CONFIG, filepath.Join("..", "configs", "config.app.yml"))
	t.Setenv(config.APP_CONFIG_SECRETS, filepath.Join("..", "configs", "config.secrets.yml"))

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
	t.Setenv(config.APP_CONFIG, filepath.Join("..", "configs", "config.extra.yml"))
	cfg = &GlobalConfigModel{}
	err = service.Load(cfg)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), `'Config' has invalid keys: extra`)
}

func Test_Embedded_Load(t *testing.T) {
	t.Setenv(config.APP_CONFIG, filepath.Join("..", "configs", "config.app.yml"))
	t.Setenv(config.APP_CONFIG_SECRETS, filepath.Join("..", "configs", "config.secrets.yml"))

	cfg := &GlobalConfigModel{}

	service := config.NewService(log.NewDefaultLogger())
	err := service.LoadFromFS(cfg, base.ConfigDefaults)
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
	t.Setenv(config.APP_CONFIG, filepath.Join("..", "configs", "config.extra.yml"))
	cfg = &GlobalConfigModel{}
	err = service.Load(cfg)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), `'Config' has invalid keys: extra`)
}

func Test_WidgetsConfig(t *testing.T) {
	t.Setenv(config.APP_CONFIG, filepath.Join("testdata", "with-widgets.yml"))
	t.Setenv(config.APP_CONFIG_SECRETS, filepath.Join("testdata", "with-widget-secrets.yml"))

	cfg := &GlobalConfigModel{}

	service := config.NewService(log.NewDefaultLogger())
	err := service.LoadFromFS(cfg, base.ConfigDefaults)
	require.Nil(t, err)

	w, ok := cfg.Config.Widgets["aaa"]
	require.True(t, ok)

	require.Equal(t, "aaa", w.Name)
	require.Equal(t, "u1", w.Credentials.Username)
	require.Equal(t, "p2", w.Credentials.Password)
	require.Equal(t, "v1", w.Nested.Nested2.Nested3.Value)
}

func Test_SearchAndSecurityConfig(t *testing.T) {
	t.Setenv(config.APP_CONFIG, filepath.Join("testdata", "with-search-and-security.yml"))
	t.Setenv(config.APP_CONFIG_SECRETS, "")

	cfg := &GlobalConfigModel{}

	service := config.NewService(log.NewDefaultLogger())
	err := service.LoadFromFS(cfg, base.ConfigDefaults)
	require.Nil(t, err)

	// Search
	patterns := cfg.Config.Search.Patterns
	require.Len(t, patterns, 1)
	require.Equal(t, "a(b+)c", patterns[0].String())

	require.Equal(t, 100, cfg.Config.Search.MaxResults)
	require.Equal(t, 30*time.Second, cfg.Config.Search.Timeout)

	// Security
	require.Len(t, cfg.Config.Security.Audience, 1)
	require.Equal(t, "platform", cfg.Config.Security.Cluster)
	require.Equal(t, "roles", cfg.Config.Security.Service)
}
