package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"nidavellir/internal/config"
	"nidavellir/internal/etcd"
	"nidavellir/internal/grpc"
	httpSvr "nidavellir/internal/http"
	"nidavellir/pkg/logger"

	"go.uber.org/zap"
)

func main() {
	// 初始化日志
	logger := logger.NewLogger()
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			fmt.Printf("sync logger error: %v\n", err)
		}
	}(logger)

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// 加载微服务默认的环境变量
	envCfg, err := config.LoadEnvs()
	if err != nil {
		logger.Fatal("Failed to load env config", zap.Error(err))
	}
	fmt.Printf("%+v\n", envCfg)

	// 初始化etcd客户端
	etcdClient, err := etcd.NewClient(cfg.Etcd)
	if err != nil {
		logger.Fatal("Failed to create etcd client", zap.Error(err))
	}
	defer etcdClient.Close()

	// 创建配置服务
	configService := etcd.NewConfigService(etcdClient, logger)

	// 启动HTTP服务器
	httpServer := httpSvr.NewServer(cfg.HTTP, configService, logger)
	go func() {
		if err := httpServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("HTTP server failed", zap.Error(err))
		}
	}()

	// 启动gRPC服务器
	grpcServer := grpc.NewServer(cfg.GRPC, configService, logger)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		logger.Fatal("Failed to listen gRPC port", zap.Error(err))
	}
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("gRPC server failed", zap.Error(err))
		}
	}()

	logger.Info("Nidavellir config center started",
		zap.Int("http_port", cfg.HTTP.Port),
		zap.Int("grpc_port", cfg.GRPC.Port))

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down servers...")

	// 关闭HTTP服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("HTTP server shutdown error", zap.Error(err))
	}

	// 关闭gRPC服务器
	grpcServer.GracefulStop()

	logger.Info("Servers stopped")
}
