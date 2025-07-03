package etcd

import (
	"context"
	"time"

	"nidavellir/internal/config"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// Client etcd客户端封装
type Client struct {
	client *clientv3.Client
}

// NewClient 创建新的etcd客户端
func NewClient(cfg config.EtcdConfig) (*Client, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: time.Duration(cfg.DialTimeout) * time.Second,
		Username:    cfg.Username,
		Password:    cfg.Password,
	})
	if err != nil {
		return nil, err
	}

	return &Client{client: client}, nil
}

// Close 关闭etcd客户端
func (c *Client) Close() error {
	return c.client.Close()
}

// Put 存储键值对
func (c *Client) Put(ctx context.Context, key, value string) error {
	_, err := c.client.Put(ctx, key, value)
	return err
}

// Get 获取键值
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	resp, err := c.client.Get(ctx, key)
	if err != nil {
		return "", err
	}

	if len(resp.Kvs) == 0 {
		return "", nil
	}

	return string(resp.Kvs[0].Value), nil
}

// GetWithPrefix 根据前缀获取所有键值对
func (c *Client) GetWithPrefix(ctx context.Context, prefix string) (map[string]string, error) {
	resp, err := c.client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, kv := range resp.Kvs {
		result[string(kv.Key)] = string(kv.Value)
	}

	return result, nil
}

// Delete 删除键
func (c *Client) Delete(ctx context.Context, key string) error {
	_, err := c.client.Delete(ctx, key)
	return err
}

// DeleteWithPrefix 根据前缀删除所有键
func (c *Client) DeleteWithPrefix(ctx context.Context, prefix string) error {
	_, err := c.client.Delete(ctx, prefix, clientv3.WithPrefix())
	return err
}

// Watch 监听键的变化
func (c *Client) Watch(ctx context.Context, key string) clientv3.WatchChan {
	return c.client.Watch(ctx, key)
}

// WatchWithPrefix 监听前缀的变化
func (c *Client) WatchWithPrefix(ctx context.Context, prefix string) clientv3.WatchChan {
	return c.client.Watch(ctx, prefix, clientv3.WithPrefix())
}