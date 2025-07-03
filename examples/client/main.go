package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"nidavellir/api/proto/config"
)

// HTTPClient HTTP客户端示例
type HTTPClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPClient 创建HTTP客户端
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// SetConfig 设置配置
func (c *HTTPClient) SetConfig(service, key string, value interface{}, description string) error {
	url := fmt.Sprintf("%s/api/v1/configs/%s/%s", c.baseURL, service, key)
	
	reqBody := map[string]interface{}{
		"value":       value,
		"description": description,
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}
	
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// GetConfig 获取配置
func (c *HTTPClient) GetConfig(service, key string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/v1/configs/%s/%s", c.baseURL, service, key)
	
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("config not found")
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	
	return result, nil
}

// GRPCClient gRPC客户端示例
type GRPCClient struct {
	client config.ConfigServiceClient
	conn   *grpc.ClientConn
}

// NewGRPCClient 创建gRPC客户端
func NewGRPCClient(address string) (*GRPCClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	
	client := config.NewConfigServiceClient(conn)
	
	return &GRPCClient{
		client: client,
		conn:   conn,
	}, nil
}

// Close 关闭连接
func (c *GRPCClient) Close() error {
	return c.conn.Close()
}

// SetConfig 设置配置
func (c *GRPCClient) SetConfig(ctx context.Context, service, key, value, description string) error {
	req := &config.SetConfigRequest{
		ServiceName: service,
		Key:         key,
		Value:       value,
		Description: description,
	}
	
	resp, err := c.client.SetConfig(ctx, req)
	if err != nil {
		return err
	}
	
	if !resp.Success {
		return fmt.Errorf("failed to set config: %s", resp.Message)
	}
	
	return nil
}

// GetConfig 获取配置
func (c *GRPCClient) GetConfig(ctx context.Context, service, key string) (*config.ConfigItem, error) {
	req := &config.GetConfigRequest{
		ServiceName: service,
		Key:         key,
	}
	
	resp, err := c.client.GetConfig(ctx, req)
	if err != nil {
		return nil, err
	}
	
	if !resp.Found {
		return nil, fmt.Errorf("config not found")
	}
	
	return resp.Config, nil
}

// WatchConfig 监听配置变化
func (c *GRPCClient) WatchConfig(ctx context.Context, service, key string) error {
	req := &config.WatchConfigRequest{
		ServiceName: service,
		Key:         key,
	}
	
	stream, err := c.client.WatchConfig(ctx, req)
	if err != nil {
		return err
	}
	
	for {
		resp, err := stream.Recv()
		if err != nil {
			return err
		}
		
		fmt.Printf("配置变化: %s - %s/%s = %s\n", 
			resp.EventType, 
			resp.Config.ServiceName, 
			resp.Config.Key, 
			resp.Config.Value)
	}
}

func main() {
	fmt.Println("Nidavellir 配置中心客户端示例")
	fmt.Println("================================")
	
	// HTTP客户端示例
	fmt.Println("\n1. HTTP客户端示例")
	httpClient := NewHTTPClient("http://localhost:8080")
	
	// 设置配置
	if err := httpClient.SetConfig("user-service", "database_url", "mysql://user:pass@localhost:3306/userdb", "用户服务数据库连接"); err != nil {
		log.Printf("HTTP设置配置失败: %v", err)
	} else {
		fmt.Println("✓ HTTP设置配置成功")
	}
	
	// 获取配置
	if configData, err := httpClient.GetConfig("user-service", "database_url"); err != nil {
		log.Printf("HTTP获取配置失败: %v", err)
	} else {
		fmt.Printf("✓ HTTP获取配置成功: %+v\n", configData)
	}
	
	// gRPC客户端示例
	fmt.Println("\n2. gRPC客户端示例")
	grpcClient, err := NewGRPCClient("localhost:9090")
	if err != nil {
		log.Printf("创建gRPC客户端失败: %v", err)
		return
	}
	defer grpcClient.Close()
	
	ctx := context.Background()
	
	// 设置配置
	if err := grpcClient.SetConfig(ctx, "order-service", "redis_url", `"redis://localhost:6379"`, "订单服务Redis连接"); err != nil {
		log.Printf("gRPC设置配置失败: %v", err)
	} else {
		fmt.Println("✓ gRPC设置配置成功")
	}
	
	// 获取配置
	if configItem, err := grpcClient.GetConfig(ctx, "order-service", "redis_url"); err != nil {
		log.Printf("gRPC获取配置失败: %v", err)
	} else {
		fmt.Printf("✓ gRPC获取配置成功: %+v\n", configItem)
	}
	
	// 监听配置变化示例（注释掉，因为会阻塞）
	/*
	fmt.Println("\n3. 监听配置变化示例")
	go func() {
		if err := grpcClient.WatchConfig(ctx, "order-service", "redis_url"); err != nil {
			log.Printf("监听配置变化失败: %v", err)
		}
	}()
	
	// 等待一段时间以观察变化
	time.Sleep(30 * time.Second)
	*/
	
	fmt.Println("\n示例完成！")
}