package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"nidavellir/internal/config"
	"nidavellir/internal/etcd"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Server HTTP服务器
type Server struct {
	server        *http.Server
	configService *etcd.ConfigService
	logger        *zap.Logger
}

// NewServer 创建HTTP服务器
func NewServer(cfg config.HTTPConfig, configService *etcd.ConfigService, logger *zap.Logger) *Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	router.Use(loggingMiddleware(logger))

	s := &Server{
		configService: configService,
		logger:        logger,
	}

	// 注册路由
	s.registerRoutes(router)

	s.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler: router,
	}

	return s
}

// Start 启动HTTP服务器
func (s *Server) Start() error {
	s.logger.Info("Starting HTTP server", zap.String("addr", s.server.Addr))
	return s.server.ListenAndServe()
}

// Shutdown 关闭HTTP服务器
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// registerRoutes 注册路由
func (s *Server) registerRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	{
		// 健康检查
		api.GET("/health", s.healthCheck)

		// 配置管理
		configs := api.Group("/configs")
		{
			// 设置配置
			configs.PUT("/:service/:key", s.setConfig)
			// 获取配置
			configs.GET("/:service/:key", s.getConfig)
			// 获取服务所有配置
			configs.GET("/:service", s.getServiceConfigs)
			// 删除配置
			configs.DELETE("/:service/:key", s.deleteConfig)
			// 删除服务所有配置
			configs.DELETE("/:service", s.deleteServiceConfigs)
		}

		// 服务管理
		api.GET("/services", s.listServices)
	}
}

// healthCheck 健康检查
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"service":   "nidavellir",
	})
}

// setConfig 设置配置
func (s *Server) setConfig(c *gin.Context) {
	service := c.Param("service")
	key := c.Param("key")

	var req struct {
		Value       interface{} `json:"value" binding:"required"`
		Description string      `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.configService.SetConfig(ctx, service, key, req.Value, req.Description); err != nil {
		s.logger.Error("Failed to set config", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Config set successfully"})
}

// getConfig 获取配置
func (s *Server) getConfig(c *gin.Context) {
	service := c.Param("service")
	key := c.Param("key")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	configItem, err := s.configService.GetConfig(ctx, service, key)
	if err != nil {
		s.logger.Error("Failed to get config", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get config"})
		return
	}

	if configItem == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		return
	}

	c.JSON(http.StatusOK, configItem)
}

// getServiceConfigs 获取服务所有配置
func (s *Server) getServiceConfigs(c *gin.Context) {
	service := c.Param("service")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	configs, err := s.configService.GetServiceConfigs(ctx, service)
	if err != nil {
		s.logger.Error("Failed to get service configs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get service configs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"configs": configs})
}

// deleteConfig 删除配置
func (s *Server) deleteConfig(c *gin.Context) {
	service := c.Param("service")
	key := c.Param("key")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.configService.DeleteConfig(ctx, service, key); err != nil {
		s.logger.Error("Failed to delete config", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Config deleted successfully"})
}

// deleteServiceConfigs 删除服务所有配置
func (s *Server) deleteServiceConfigs(c *gin.Context) {
	service := c.Param("service")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.configService.DeleteServiceConfigs(ctx, service); err != nil {
		s.logger.Error("Failed to delete service configs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete service configs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Service configs deleted successfully"})
}

// listServices 列出所有服务
func (s *Server) listServices(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	services, err := s.configService.ListServices(ctx)
	if err != nil {
		s.logger.Error("Failed to list services", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list services"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"services": services})
}

// corsMiddleware CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// loggingMiddleware 日志中间件
func loggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info("HTTP Request",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.String("ip", clientIP),
			zap.Duration("latency", latency),
		)
	}
}
