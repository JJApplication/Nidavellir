package initializer

import (
	"go.uber.org/zap"
	"nidavellir/internal/config"
)

func InitializeConfig(glb *Global) {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		glb.Logger.Fatal("Failed to load config", zap.Error(err))
	}

	// 加载微服务默认的环境变量
	envCfg, err := config.LoadEnvs()
	if err != nil {
		glb.Logger.Fatal("Failed to load env config", zap.Error(err))
	}

	glb.Cfg = cfg
	glb.EnvCfg = envCfg
}
