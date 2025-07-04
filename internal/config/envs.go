package config

import (
	"errors"
	"github.com/spf13/viper"
)

type EnvConfig struct {
	Service []struct {
		Name string `mapstructure:"name"`
		Envs []struct {
			Key         string `mapstructure:"key"`
			Val         string `mapstructure:"val"`
			Description string `mapstructure:"description"`
		} `mapstructure:"envs"`
	} `mapstructure:"service"`
}

func LoadEnvs() (*EnvConfig, error) {
	viper.SetConfigName("envs")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return &EnvConfig{}, err
		}
	}

	var cfg *EnvConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return &EnvConfig{}, err
	}

	return cfg, nil
}
