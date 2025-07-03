package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	// ConfigPrefix 配置键前缀
	ConfigPrefix = "/config/"
)

// ConfigService 配置服务
type ConfigService struct {
	client *Client
	logger *zap.Logger
}

// ConfigItem 配置项
type ConfigItem struct {
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	ServiceName string      `json:"service_name"`
	Description string      `json:"description"`
	CreatedAt   int64       `json:"created_at"`
	UpdatedAt   int64       `json:"updated_at"`
}

// NewConfigService 创建配置服务
func NewConfigService(client *Client, logger *zap.Logger) *ConfigService {
	return &ConfigService{
		client: client,
		logger: logger,
	}
}

// SetConfig 设置服务配置
func (s *ConfigService) SetConfig(ctx context.Context, serviceName, key string, value interface{}, description string) error {
	configKey := s.buildConfigKey(serviceName, key)

	configItem := ConfigItem{
		Key:         key,
		Value:       value,
		ServiceName: serviceName,
		Description: description,
		CreatedAt:   getCurrentTimestamp(),
		UpdatedAt:   getCurrentTimestamp(),
	}

	// 检查是否已存在，如果存在则只更新时间戳
	existing, err := s.client.Get(ctx, configKey)
	if err != nil {
		return fmt.Errorf("failed to check existing config: %w", err)
	}

	if existing != "" {
		var existingItem ConfigItem
		if err := json.Unmarshal([]byte(existing), &existingItem); err == nil {
			configItem.CreatedAt = existingItem.CreatedAt
		}
	}

	data, err := json.Marshal(configItem)
	if err != nil {
		return fmt.Errorf("failed to marshal config item: %w", err)
	}

	if err := s.client.Put(ctx, configKey, string(data)); err != nil {
		return fmt.Errorf("failed to set config: %w", err)
	}

	s.logger.Info("Config set successfully",
		zap.String("service", serviceName),
		zap.String("key", key))

	return nil
}

// GetConfig 获取服务配置
func (s *ConfigService) GetConfig(ctx context.Context, serviceName, key string) (*ConfigItem, error) {
	configKey := s.buildConfigKey(serviceName, key)

	data, err := s.client.Get(ctx, configKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	if data == "" {
		return nil, nil
	}

	var configItem ConfigItem
	if err := json.Unmarshal([]byte(data), &configItem); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config item: %w", err)
	}

	return &configItem, nil
}

// GetServiceConfigs 获取服务的所有配置
func (s *ConfigService) GetServiceConfigs(ctx context.Context, serviceName string) (map[string]*ConfigItem, error) {
	prefix := s.buildServicePrefix(serviceName)

	data, err := s.client.GetWithPrefix(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to get service configs: %w", err)
	}

	result := make(map[string]*ConfigItem)
	for fullKey, value := range data {
		// 提取配置键名
		key := strings.TrimPrefix(fullKey, prefix)

		var configItem ConfigItem
		if err := json.Unmarshal([]byte(value), &configItem); err != nil {
			s.logger.Warn("Failed to unmarshal config item",
				zap.String("key", fullKey),
				zap.Error(err))
			continue
		}

		result[key] = &configItem
	}

	return result, nil
}

// DeleteConfig 删除服务配置
func (s *ConfigService) DeleteConfig(ctx context.Context, serviceName, key string) error {
	configKey := s.buildConfigKey(serviceName, key)

	if err := s.client.Delete(ctx, configKey); err != nil {
		return fmt.Errorf("failed to delete config: %w", err)
	}

	s.logger.Info("Config deleted successfully",
		zap.String("service", serviceName),
		zap.String("key", key))

	return nil
}

// DeleteServiceConfigs 删除服务的所有配置
func (s *ConfigService) DeleteServiceConfigs(ctx context.Context, serviceName string) error {
	prefix := s.buildServicePrefix(serviceName)

	if err := s.client.DeleteWithPrefix(ctx, prefix); err != nil {
		return fmt.Errorf("failed to delete service configs: %w", err)
	}

	s.logger.Info("Service configs deleted successfully",
		zap.String("service", serviceName))

	return nil
}

// ListServices 列出所有服务
func (s *ConfigService) ListServices(ctx context.Context) ([]string, error) {
	data, err := s.client.GetWithPrefix(ctx, ConfigPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	services := make(map[string]bool)
	for key := range data {
		// 提取服务名称
		relativeKey := strings.TrimPrefix(key, ConfigPrefix)
		parts := strings.Split(relativeKey, "/")
		if len(parts) > 0 {
			services[parts[0]] = true
		}
	}

	result := make([]string, 0, len(services))
	for service := range services {
		result = append(result, service)
	}

	return result, nil
}

// buildConfigKey 构建配置键
func (s *ConfigService) buildConfigKey(serviceName, key string) string {
	return fmt.Sprintf("%s%s/%s", ConfigPrefix, serviceName, key)
}

// buildServicePrefix 构建服务前缀
func (s *ConfigService) buildServicePrefix(serviceName string) string {
	return fmt.Sprintf("%s%s/", ConfigPrefix, serviceName)
}

// GetEtcdClient 获取etcd客户端（用于监听）
func (s *ConfigService) GetEtcdClient() *Client {
	return s.client
}

// getCurrentTimestamp 获取当前时间戳
func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}
