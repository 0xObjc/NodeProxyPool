package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"proxyPool/internal/api"
	"proxyPool/internal/config"
	"proxyPool/internal/mihomo"
	"proxyPool/internal/node"
	"proxyPool/internal/proxy"
	"proxyPool/internal/subscription"
	"proxyPool/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	configFile = flag.String("c", "configs/config.yaml", "配置文件路径")
)

func main() {
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configFile)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	if err := logger.Init(cfg.Log.Level, cfg.Log.File); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("ProxyPool starting...",
		zap.String("version", "1.0.0"),
		zap.String("config", *configFile),
	)

	// 创建节点池
	nodePool := node.NewPool()

	// 创建订阅管理器
	subManager := subscription.NewManager(cfg, nodePool)
	subManager.Start()
	defer subManager.Stop()

	// 创建并启动健康检查器
	if cfg.HealthCheck.Enabled {
		healthChecker := node.NewHealthChecker(
			nodePool,
			cfg.HealthCheck.Interval,
			cfg.HealthCheck.Timeout,
			cfg.HealthCheck.URL,
			cfg.HealthCheck.MaxDelay,
			logger.GetLogger(),
		)
		healthChecker.Start()
		defer healthChecker.Stop()
	}

	// 创建Mihomo适配器
	mihomoAdapter, err := mihomo.NewAdapter(logger.GetLogger())
	if err != nil {
		logger.Fatal("Failed to create mihomo adapter", zap.Error(err))
	}

	// 创建端口分配器
	portAllocator, err := proxy.NewPortAllocator(
		cfg.Proxy.PortRange.Min,
		cfg.Proxy.PortRange.Max,
	)
	if err != nil {
		logger.Fatal("Failed to create port allocator", zap.Error(err))
	}

	// 创建代理管理器
	proxyManager := proxy.NewManager(
		portAllocator,
		nodePool,
		mihomoAdapter,
		cfg.Proxy.MaxInstances,
		cfg.Proxy.DefaultTTL,
		cfg.Proxy.DefaultProtocol,
		logger.GetLogger(),
	)

	// 创建并启动TTL清理器
	ttlCleaner := proxy.NewTTLCleaner(
		proxyManager,
		30*time.Second,
		logger.GetLogger(),
	)
	ttlCleaner.Start()
	defer ttlCleaner.Stop()

	// 设置Gin模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin引擎
	r := gin.New()
	r.Use(api.Recovery(logger.GetLogger()))
	r.Use(api.Logger(logger.GetLogger()))

	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"message": "ProxyPool is running",
		})
	})

	// 注册API路由
	apiHandler := api.NewHandler(subManager, proxyManager)
	apiHandler.RegisterRoutes(r)

	// 启动HTTP服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Info("HTTP server listening",
		zap.String("addr", addr),
	)

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// 在goroutine中启动服务器
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down ProxyPool...")

	// 优雅关闭HTTP服务器
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("ProxyPool stopped")
}
