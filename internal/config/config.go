package config

import (
	"errors"
	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	HTTP HTTPConfig `mapstructure:"http"`
	GRPC GRPCConfig `mapstructure:"grpc"`
	Twig TwigConfig `mapstructure:"twig"`
	Etcd EtcdConfig `mapstructure:"etcd"`
	Log  LogConfig  `mapstructure:"log"`
}

// HTTPConfig HTTP服务器配置
type HTTPConfig struct {
	Port   int    `mapstructure:"port"`
	Host   string `mapstructure:"host"`
	Enable bool   `mapstructure:"enable"`
}

// GRPCConfig gRPC服务器配置
type GRPCConfig struct {
	Port   int    `mapstructure:"port"`
	Host   string `mapstructure:"host"`
	Enable bool   `mapstructure:"enable"`
}

type TwigConfig struct {
	Address string `mapstructure:"address"`
	Enable  bool   `mapstructure:"enable"`
}

// EtcdConfig etcd配置
type EtcdConfig struct {
	Endpoints   []string `mapstructure:"endpoints"`
	DialTimeout int      `mapstructure:"dial_timeout"`
	Username    string   `mapstructure:"username"`
	Password    string   `mapstructure:"password"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Load 加载配置
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")

	// 设置默认值
	setDefaults()

	// 读取环境变量
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, err
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// setDefaults 设置默认配置值
func setDefaults() {
	viper.SetDefault("http.port", 8080)
	viper.SetDefault("http.host", "0.0.0.0")
	viper.SetDefault("grpc.port", 9090)
	viper.SetDefault("grpc.host", "0.0.0.0")
	viper.SetDefault("etcd.endpoints", []string{"localhost:2379"})
	viper.SetDefault("etcd.dial_timeout", 5)
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
}
