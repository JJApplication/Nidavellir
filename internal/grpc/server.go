package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	grpcConfig "nidavellir/api/proto"
	"nidavellir/internal/config"
	"nidavellir/internal/etcd"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server gRPC服务器
type Server struct {
	grpcConfig.UnimplementedConfigServiceServer
	configService *etcd.ConfigService
	logger        *zap.Logger
	grpcServer    *grpc.Server
}

// NewServer 创建gRPC服务器
func NewServer(cfg config.GRPCConfig, configService *etcd.ConfigService, logger *zap.Logger) *Server {
	s := &Server{
		configService: configService,
		logger:        logger,
	}

	// 创建gRPC服务器
	s.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(s.unaryInterceptor),
		grpc.StreamInterceptor(s.streamInterceptor),
	)

	// 注册服务
	grpcConfig.RegisterConfigServiceServer(s.grpcServer, s)

	return s
}

// Serve 启动gRPC服务器
func (s *Server) Serve(lis net.Listener) error {
	s.logger.Info("Starting gRPC server", zap.String("addr", lis.Addr().String()))
	return s.grpcServer.Serve(lis)
}

// GracefulStop 优雅停止gRPC服务器
func (s *Server) GracefulStop() {
	s.grpcServer.GracefulStop()
}

// SetConfig 设置配置
func (s *Server) SetConfig(ctx context.Context, req *grpcConfig.SetConfigRequest) (*grpcConfig.SetConfigResponse, error) {
	if req.ServiceName == "" || req.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "service_name and key are required")
	}

	// 解析value为interface{}
	var value interface{}
	if err := json.Unmarshal([]byte(req.Value), &value); err != nil {
		// 如果解析失败，直接使用字符串值
		value = req.Value
	}

	if err := s.configService.SetConfig(ctx, req.ServiceName, req.Key, value, req.Description); err != nil {
		s.logger.Error("Failed to set config", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to set config")
	}

	return &grpcConfig.SetConfigResponse{
		Success: true,
		Message: "Config set successfully",
	}, nil
}

// GetConfig 获取配置
func (s *Server) GetConfig(ctx context.Context, req *grpcConfig.GetConfigRequest) (*grpcConfig.GetConfigResponse, error) {
	if req.ServiceName == "" || req.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "service_name and key are required")
	}

	configItem, err := s.configService.GetConfig(ctx, req.ServiceName, req.Key)
	if err != nil {
		s.logger.Error("Failed to get config", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to get config")
	}

	if configItem == nil {
		return &grpcConfig.GetConfigResponse{
			Found: false,
		}, nil
	}

	// 转换为protobuf格式
	valueBytes, _ := json.Marshal(configItem.Value)
	protoConfig := &grpcConfig.ConfigItem{
		Key:         configItem.Key,
		Value:       string(valueBytes),
		ServiceName: configItem.ServiceName,
		Description: configItem.Description,
		CreatedAt:   configItem.CreatedAt,
		UpdatedAt:   configItem.UpdatedAt,
	}

	return &grpcConfig.GetConfigResponse{
		Config: protoConfig,
		Found:  true,
	}, nil
}

// GetServiceConfigs 获取服务所有配置
func (s *Server) GetServiceConfigs(ctx context.Context, req *grpcConfig.GetServiceConfigsRequest) (*grpcConfig.GetServiceConfigsResponse, error) {
	if req.ServiceName == "" {
		return nil, status.Error(codes.InvalidArgument, "service_name is required")
	}

	configs, err := s.configService.GetServiceConfigs(ctx, req.ServiceName)
	if err != nil {
		s.logger.Error("Failed to get service configs", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to get service configs")
	}

	// 转换为protobuf格式
	protoConfigs := make(map[string]*grpcConfig.ConfigItem)
	for key, configItem := range configs {
		valueBytes, _ := json.Marshal(configItem.Value)
		protoConfigs[key] = &grpcConfig.ConfigItem{
			Key:         configItem.Key,
			Value:       string(valueBytes),
			ServiceName: configItem.ServiceName,
			Description: configItem.Description,
			CreatedAt:   configItem.CreatedAt,
			UpdatedAt:   configItem.UpdatedAt,
		}
	}

	return &grpcConfig.GetServiceConfigsResponse{
		Configs: protoConfigs,
	}, nil
}

// DeleteConfig 删除配置
func (s *Server) DeleteConfig(ctx context.Context, req *grpcConfig.DeleteConfigRequest) (*grpcConfig.DeleteConfigResponse, error) {
	if req.ServiceName == "" || req.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "service_name and key are required")
	}

	if err := s.configService.DeleteConfig(ctx, req.ServiceName, req.Key); err != nil {
		s.logger.Error("Failed to delete config", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to delete config")
	}

	return &grpcConfig.DeleteConfigResponse{
		Success: true,
		Message: "Config deleted successfully",
	}, nil
}

// DeleteServiceConfigs 删除服务所有配置
func (s *Server) DeleteServiceConfigs(ctx context.Context, req *grpcConfig.DeleteServiceConfigsRequest) (*grpcConfig.DeleteServiceConfigsResponse, error) {
	if req.ServiceName == "" {
		return nil, status.Error(codes.InvalidArgument, "service_name is required")
	}

	if err := s.configService.DeleteServiceConfigs(ctx, req.ServiceName); err != nil {
		s.logger.Error("Failed to delete service configs", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to delete service configs")
	}

	return &grpcConfig.DeleteServiceConfigsResponse{
		Success: true,
		Message: "Service configs deleted successfully",
	}, nil
}

// ListServices 列出所有服务
func (s *Server) ListServices(ctx context.Context, req *grpcConfig.ListServicesRequest) (*grpcConfig.ListServicesResponse, error) {
	services, err := s.configService.ListServices(ctx)
	if err != nil {
		s.logger.Error("Failed to list services", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to list services")
	}

	return &grpcConfig.ListServicesResponse{
		Services: services,
	}, nil
}

// WatchConfig 监听配置变化
func (s *Server) WatchConfig(req *grpcConfig.WatchConfigRequest, stream grpcConfig.ConfigService_WatchConfigServer) error {
	if req.ServiceName == "" {
		return status.Error(codes.InvalidArgument, "service_name is required")
	}

	// 构建监听键
	var watchKey string
	if req.Key != "" {
		watchKey = fmt.Sprintf("/nidavellir/config/%s/%s", req.ServiceName, req.Key)
	} else {
		watchKey = fmt.Sprintf("/nidavellir/config/%s/", req.ServiceName)
	}

	// 创建etcd客户端用于监听
	etcdClient := s.configService.GetEtcdClient() // 需要在ConfigService中添加此方法
	var watchChan clientv3.WatchChan
	if req.Key != "" {
		watchChan = etcdClient.Watch(stream.Context(), watchKey)
	} else {
		watchChan = etcdClient.WatchWithPrefix(stream.Context(), watchKey)
	}

	for watchResp := range watchChan {
		for _, event := range watchResp.Events {
			// 解析配置项
			var configItem etcd.ConfigItem
			if err := json.Unmarshal(event.Kv.Value, &configItem); err != nil {
				s.logger.Warn("Failed to unmarshal config item", zap.Error(err))
				continue
			}

			// 转换为protobuf格式
			valueBytes, _ := json.Marshal(configItem.Value)
			protoConfig := &grpcConfig.ConfigItem{
				Key:         configItem.Key,
				Value:       string(valueBytes),
				ServiceName: configItem.ServiceName,
				Description: configItem.Description,
				CreatedAt:   configItem.CreatedAt,
				UpdatedAt:   configItem.UpdatedAt,
			}

			// 确定事件类型
			eventType := "PUT"
			if event.Type == clientv3.EventTypeDelete {
				eventType = "DELETE"
			}

			// 发送响应
			response := &grpcConfig.WatchConfigResponse{
				EventType: eventType,
				Config:    protoConfig,
			}

			if err := stream.Send(response); err != nil {
				s.logger.Error("Failed to send watch response", zap.Error(err))
				return err
			}
		}
	}

	return nil
}

// unaryInterceptor 一元拦截器
func (s *Server) unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	s.logger.Info("gRPC unary call", zap.String("method", info.FullMethod))
	return handler(ctx, req)
}

// streamInterceptor 流拦截器
func (s *Server) streamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	s.logger.Info("gRPC stream call", zap.String("method", info.FullMethod))
	return handler(srv, ss)
}
