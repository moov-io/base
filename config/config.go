package config

import (
	"fmt"
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
		logger: logger.Set("component", "Service"),
	}
}

func (s *Service) Load(config interface{}) error {
	err := s.LoadFile(pkger.Include("/configs/config.default.yml"), config)
	if err != nil {
		return err
	}

	if file, ok := os.LookupEnv("APP_CONFIG"); ok && strings.TrimSpace(file) != "" {
		log := s.logger.Set("app_config", file)
		log.Info().Log("Loading APP_CONFIG config file")

		overrides := viper.New()
		overrides.SetConfigFile(file)

		if err := overrides.ReadInConfig(); err != nil {
			return log.LogError(fmt.Sprintf("Failed loading the specific app config - %s", err), err)
		}

		if err := overrides.Unmarshal(config); err != nil {
			return log.LogError(fmt.Sprintf("Unable to unmarshal the specific app config - %s", err), err)
		}
	}

	return nil
}

func (s *Service) LoadFile(file string, config interface{}) error {
	log := s.logger.Set("file", file)
	log.Info().Log("loading config file")

	f, err := pkger.Open(file)
	if err != nil {
		return log.LogError("pkger unable to load", err)
	}

	deflt := viper.New()
	deflt.SetConfigType("yaml")
	if err := deflt.ReadConfig(f); err != nil {
		return log.LogError("unable to load the defaults", err)
	}

	if err := deflt.Unmarshal(config); err != nil {
		return log.LogError("unable to unmarshal the defaults", err)
	}

	return nil
}
