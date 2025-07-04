package initializer

import (
	"go.uber.org/zap"
	"nidavellir/internal/config"
	"nidavellir/internal/etcd"
)

type Global struct {
	Logger        *zap.Logger
	Client        *etcd.Client
	Cfg           *config.Config
	EnvCfg        *config.EnvConfig
	ConfigService *etcd.ConfigService
}
