package etcd

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
	"nidavellir/internal/config"
)

// InitServiceEnvs 初始化服务的环境变量, 如果已经存在了任何配置则不执行
func InitServiceEnvs(envs *config.EnvConfig, client *Client, logger *zap.Logger) {
	ctx := context.Background()
	if len(envs.Service) <= 0 {
		return
	}

	data, err := client.GetWithPrefix(ctx, ConfigPrefix)
	if err != nil {
		logger.Error("fail to get config", zap.Error(err))
		return
	}

	if len(data) > 0 {
		logger.Info("Service config already init")
		return
	}
	for _, env := range envs.Service {
		service := env.Name
		if env.Envs == nil || len(env.Envs) <= 0 {
			continue
		}
		for _, envCfg := range env.Envs {
			if envCfg.Key == "" {
				continue
			}
			configKey := fmt.Sprintf("%s%s/%s", ConfigPrefix, service, envCfg.Key)
			configItem := ConfigItem{
				Key:         envCfg.Key,
				Value:       envCfg.Val,
				ServiceName: service,
				Description: envCfg.Description,
				CreatedAt:   getCurrentTimestamp(),
				UpdatedAt:   getCurrentTimestamp(),
			}

			// 检查是否已存在，如果存在则只更新时间戳
			existing, err := client.Get(ctx, configKey)
			if err != nil {
				logger.Error("failed to get config", zap.String("key", configKey), zap.Error(err))
				return
			}

			if existing != "" {
				var existingItem ConfigItem
				if err := json.Unmarshal([]byte(existing), &existingItem); err == nil {
					configItem.CreatedAt = existingItem.CreatedAt
				}
			}

			data, err := json.Marshal(configItem)
			if err != nil {
				logger.Error("failed to marshal config item", zap.Error(err))
				return
			}

			if err := client.Put(ctx, configKey, string(data)); err != nil {
				logger.Error("failed to set config", zap.Error(err))
				return
			}
		}
	}

	logger.Info("InitConfig set successfully")
}
