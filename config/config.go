package config

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/moov-io/base/log"

	"github.com/markbates/pkger"
	"github.com/spf13/viper"
)

const APP_CONFIG = "APP_CONFIG"
const APP_CONFIG_SECRETS = "APP_CONFIG_SECRETS"

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
		s.logger.LogErrorf("loading default config file using pkger: %w", err)
	}

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
		return fmt.Errorf("pkger unable to load %s: %w", file, err)
	}

	deflt := viper.New()
	deflt.SetConfigType("yaml")
	if err := deflt.ReadConfig(f); err != nil {
		return fmt.Errorf("unable to load the defaults: %w", err)
	}

	if err := deflt.Unmarshal(config); err != nil {
		return fmt.Errorf("unable to unmarshal the defaults: %w", err)
	}

	return nil
}

func (s *Service) LoadFromReader(in io.Reader, config interface{}) error {
	logger := s.logger
	logger.Info().Logf("loading config file from reader")

	deflt := viper.New()
	deflt.SetConfigType("yaml")
	if err := deflt.ReadConfig(in); err != nil {
		return logger.LogErrorf("unable to load the defaults: %w", err).Err()
	}

	if err := deflt.Unmarshal(config); err != nil {
		return logger.LogErrorf("unable to unmarshal the defaults: %w", err).Err()
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

		if err := overrides.Unmarshal(config); err != nil {
			return logger.LogErrorf("Unable to unmarshal the specific app config: %w", err).Err()
		}
	}

	return nil
}
