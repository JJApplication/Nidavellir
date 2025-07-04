package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"nidavellir/initializer"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"nidavellir/internal/grpc"
	httpSvr "nidavellir/internal/http"
)

func main() {
	// 初始化
	glb := initializer.InitialSequence()

	// 启动HTTP服务器
	httpServer := httpSvr.NewServer(glb.Cfg.HTTP, glb.ConfigService, glb.Logger)
	go func() {
		if glb.Cfg.HTTP.Enable {
			if err := httpServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				glb.Logger.Fatal("HTTP server failed", zap.Error(err))
			}
		}
	}()

	// 启动gRPC服务器
	grpcServer := grpc.NewServer(glb.Cfg.GRPC, glb.ConfigService, glb.Logger)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", glb.Cfg.GRPC.Port))
	if err != nil {
		glb.Logger.Fatal("Failed to listen gRPC port", zap.Error(err))
	}
	go func() {
		if glb.Cfg.GRPC.Enable {
			if err := grpcServer.Serve(lis); err != nil {
				glb.Logger.Fatal("gRPC server failed", zap.Error(err))
			}
		}
	}()

	go func() {
		if glb.Cfg.Twig.Enable {
			if err := grpcServer.ServeUDS(glb.Cfg.Twig.Address); err != nil {
				glb.Logger.Fatal("gRPC UDS server failed", zap.Error(err))
			}
		}
	}()

	glb.Logger.Info("Nidavellir config center started",
		zap.Int("http_port", glb.Cfg.HTTP.Port),
		zap.Int("grpc_port", glb.Cfg.GRPC.Port),
		zap.String("grpc_uds", glb.Cfg.Twig.Address))

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	glb.Logger.Info("Shutting down servers...")

	// 关闭HTTP服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		glb.Logger.Error("HTTP server shutdown error", zap.Error(err))
	}

	// 关闭gRPC服务器
	grpcServer.GracefulStop()

	glb.Logger.Info("Servers stopped")
}
