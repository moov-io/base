package config

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/moov-io/base/log"

	"github.com/markbates/pkger"
	"github.com/spf13/viper"
)

const APP_CONFIG = "APP_CONFIG"
const APP_CONFIG_SECRETS = "APP_CONFIG_SECRETS" //nolint:gosec

type Service struct {
	logger log.Logger
}

func NewService(logger log.Logger) Service {
	return Service{
		logger: logger.Set("component", log.String("Service")),
	}
}

func (s *Service) Load(config interface{}) error {
	if err := s.LoadFile(pkger.Include("/configs/config.default.yml"), config); err != nil {
		return err
	}

	return s.MergeEnvironments(config)
}

func (s *Service) LoadFromFS(config interface{}, fs fs.FS) error {
	if err := s.LoadEmbeddedFile("configs/config.default.yml", config, fs); err != nil {
		return err
	}

	return s.MergeEnvironments(config)
}

func (s *Service) MergeEnvironments(config interface{}) error {

	if err := LoadEnvironmentFile(s.logger, APP_CONFIG, config); err != nil {
		return err
	}

	if err := LoadEnvironmentFile(s.logger, APP_CONFIG_SECRETS, config); err != nil {
		return err
	}

	return nil
}

func (s *Service) LoadFile(file string, config interface{}) error {
	logger := s.logger.Set("file", log.String(file))
	logger.Info().Logf("loading config file")

	f, err := pkger.Open(file)
	if err != nil {
		return logger.LogErrorf("pkger unable to load %s: %w", file, err).Err()
	}

	if err := configFromReader(config, f); err != nil {
		return logger.LogError(err).Err()
	}

	return nil
}

func (s *Service) LoadEmbeddedFile(file string, config interface{}, fs fs.FS) error {
	logger := s.logger.Set("file", log.String(file))
	logger.Info().Logf("loading config file")

	f, err := fs.Open(file)
	if err != nil {
		return logger.LogErrorf("go:embed FS unable to load %s: %w", file, err).Err()
	}

	if err := configFromReader(config, f); err != nil {
		return logger.LogError(err).Err()
	}

	return nil
}

func configFromReader(config interface{}, f io.Reader) error {
	deflt := viper.New()
	deflt.SetConfigType("yaml")
	if err := deflt.ReadConfig(f); err != nil {
		return fmt.Errorf("unable to load the defaults: %w", err)
	}

	if err := deflt.UnmarshalExact(config, overwriteConfig); err != nil {
		return fmt.Errorf("unable to unmarshal the defaults: %w", err)
	}

	return nil
}

func LoadEnvironmentFile(logger log.Logger, envVar string, config interface{}) error {
	if file, ok := os.LookupEnv(envVar); ok && strings.TrimSpace(file) != "" {

		logger := logger.Set(envVar, log.String(file))
		logger.Info().Logf("Loading %s config file", envVar)

		logger = logger.Set("file", log.String(file))
		logger.Info().Logf("loading config file")

		overrides := viper.New()

		overrides.SetConfigFile(file)

		if err := overrides.ReadInConfig(); err != nil {
			return logger.LogErrorf("Failed loading the specific app config: %w", err).Err()
		}

		if err := overrides.UnmarshalExact(config, overwriteConfig); err != nil {
			return logger.LogErrorf("Unable to unmarshal the specific app config: %w", err).Err()
		}
	}

	return nil
}

func overwriteConfig(cfg *mapstructure.DecoderConfig) {
	cfg.ErrorUnused = true
	cfg.ZeroFields = true
}
