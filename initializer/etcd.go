package initializer

import (
	"go.uber.org/zap"
	"nidavellir/internal/config"
	"nidavellir/internal/etcd"
)

func InitializeEtcd(glb *Global) {
	// 初始化etcd客户端
	etcdClient, err := etcd.NewClient(glb.Cfg.Etcd)
	if err != nil {
		glb.Logger.Fatal("Failed to create etcd client", zap.Error(err))
	}
	defer etcdClient.Close()

	InitializeEnvs(glb.EnvCfg, etcdClient, glb.Logger)
	service := InitializeService(etcdClient, glb.Logger)

	glb.Client = etcdClient
	glb.ConfigService = service
}

func InitializeEnvs(envCfg *config.EnvConfig, client *etcd.Client, logger *zap.Logger) {
	etcd.InitServiceEnvs(envCfg, client, logger)
}

func InitializeService(client *etcd.Client, logger *zap.Logger) *etcd.ConfigService {
	return etcd.NewConfigService(client, logger)
}
