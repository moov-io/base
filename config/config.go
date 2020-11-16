package config

import (
	"os"
	"strings"

	"github.com/moov-io/base/log"

	"github.com/markbates/pkger"
	"github.com/spf13/viper"
)

type Service struct {
	logger log.Logger
}

func NewService(logger log.Logger) Service {
	return Service{
		logger: logger.Set("component", log.String("Service")),
	}
}

func (s *Service) Load(config interface{}) error {
	err := s.LoadFile(pkger.Include("/configs/config.default.yml"), config)
	if err != nil {
		return err
	}

	if file, ok := os.LookupEnv("APP_CONFIG"); ok && strings.TrimSpace(file) != "" {
		logger := s.logger.Set("app_config", log.String(file))
		logger.Info().Logf("Loading APP_CONFIG config file")

		overrides := viper.New()
		overrides.SetConfigFile(file)

		if err := overrides.ReadInConfig(); err != nil {
			return logger.LogErrorf("Failed loading the specific app config: %w", err).Err()
		}

		if err := overrides.Unmarshal(config); err != nil {
			return logger.LogErrorf("Unable to unmarshal the specific app config: %w", err).Err()
		}
	}

	return nil
}

func (s *Service) LoadFile(file string, config interface{}) error {
	logger := s.logger.Set("file", log.String(file))
	logger.Info().Logf("loading config file")

	f, err := pkger.Open(file)
	if err != nil {
		return logger.LogErrorf("pkger unable to load: %w", err).Err()
	}

	deflt := viper.New()
	deflt.SetConfigType("yaml")
	if err := deflt.ReadConfig(f); err != nil {
		return logger.LogErrorf("unable to load the defaults: %w", err).Err()
	}

	if err := deflt.Unmarshal(config); err != nil {
		return logger.LogErrorf("unable to unmarshal the defaults: %w", err).Err()
	}

	return nil
}
