package config

import (
	"errors"
	"github.com/spf13/viper"
)

type EnvConfig struct {
	Service []struct {
		Name string `mapstructure:"name"`
		Envs []struct {
			Key string `mapstructure:"key"`
			Val string `mapstructure:"val"`
		} `mapstructure:"envs"`
	} `mapstructure:"service"`
}

func LoadEnvs() (*[]EnvConfig, error) {
	viper.SetConfigName("envs")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, err
		}
	}

	var cfg []EnvConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
