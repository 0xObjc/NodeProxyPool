package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"proxyPool"
	"proxyPool/internal/api"
	"proxyPool/internal/config"
	"proxyPool/internal/database"
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

	// 初始化数据库并执行迁移
	db, err := database.New("./data/proxypool.db")
	if err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		fmt.Printf("Failed to migrate database: %v\n", err)
		os.Exit(1)
	}

	// 首次运行从 config.yaml 导入
	if db.IsFirstRun() {
		cfgFromFile, err := config.Load(*configFile)
		if err != nil {
			fmt.Printf("Failed to load config for seeding: %v\n", err)
			os.Exit(1)
		}
		if err := db.SeedFromConfig(config.BuildSeedData(cfgFromFile)); err != nil {
			fmt.Printf("Failed to seed database: %v\n", err)
			os.Exit(1)
		}
	}

	// 构建动态配置
	dynamicCfg, err := config.NewDynamicConfig(db)
	if err != nil {
		fmt.Printf("Failed to load dynamic config: %v\n", err)
		os.Exit(1)
	}
	cfg := dynamicCfg.Current()

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

	subRepo := database.NewSubscriptionRepository(db)
	nodeRepo := database.NewNodeRepository(db)

	// 创建节点池并尝试从数据库恢复
	nodePool := node.NewPool()
	if nodes, err := nodeRepo.GetAll(); err == nil && len(nodes) > 0 {
		nodePool.UpdateNodes(nodes)
		logger.Info("Loaded nodes from database", zap.Int("count", len(nodes)))
	}

	// 创建订阅管理器
	subManager := subscription.NewManager(dynamicCfg, db, nodePool, subRepo, nodeRepo)
	subManager.Start()
	defer subManager.Stop()

	// 创建并启动健康检查器
	var healthChecker *node.HealthChecker
	defer func() {
		if healthChecker != nil {
			healthChecker.Stop()
		}
	}()
	if cfg.HealthCheck.Enabled {
		healthChecker = node.NewHealthChecker(
			nodePool,
			cfg.HealthCheck.Interval,
			cfg.HealthCheck.Timeout,
			cfg.HealthCheck.URL,
			cfg.HealthCheck.MaxDelay,
			logger.GetLogger(),
		)
		healthChecker.Start()
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

	// 绑定配置变更回调
	dynamicCfg.SetOnProxyConfigChange(func(pc *config.ProxyConfig) {
		proxyManager.UpdateConfig(pc.MaxInstances, pc.DefaultTTL, pc.DefaultProtocol)
	})
	dynamicCfg.SetOnSubscriptionConfigChange(func(sc *config.SubscriptionConfig) {
		subManager.UpdateConfig(sc)
	})
	dynamicCfg.SetOnHealthCheckConfigChange(func(hc *config.HealthCheckConfig) {
		if hc.Enabled {
			if healthChecker == nil {
				healthChecker = node.NewHealthChecker(
					nodePool,
					hc.Interval,
					hc.Timeout,
					hc.URL,
					hc.MaxDelay,
					logger.GetLogger(),
				)
				healthChecker.Start()
				return
			}
			healthChecker.UpdateConfig(hc.Interval, hc.Timeout, hc.URL, hc.MaxDelay)
			return
		}

		if healthChecker != nil {
			healthChecker.Stop()
			healthChecker = nil
		}
	})

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

	// 开发环境下启用CORS
	if cfg.Server.Mode != "release" {
		r.Use(corsMiddleware())
	}

	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "ProxyPool is running",
		})
	})

	// 注册API路由
	apiHandler := api.NewHandler(subManager, proxyManager, dynamicCfg, subRepo)
	apiHandler.RegisterRoutes(r)

	// 配置静态文件服务
	setupStaticFiles(r)

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

// corsMiddleware 处理跨域请求(仅用于开发环境)
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// setupStaticFiles 配置静态文件服务
func setupStaticFiles(r *gin.Engine) {
	// 获取嵌入的web/dist目录
	distFS, err := fs.Sub(proxypool.WebDist, "web/dist")
	if err != nil {
		logger.Fatal("Failed to get dist sub filesystem", zap.Error(err))
	}

	// 1. 静态资源路由 (assets, favicon等)
	assetsFS, err := fs.Sub(distFS, "assets")
	if err != nil {
		logger.Fatal("Failed to get assets sub filesystem", zap.Error(err))
	}
	r.StaticFS("/assets", http.FS(assetsFS))

	// 预读 index.html，避免在根路径被 http.FileServer 301 重定向
	indexContent, err := fs.ReadFile(distFS, "index.html")
	if err != nil {
		logger.Fatal("Failed to read index.html from embedded dist", zap.Error(err))
	}

	// 2. SPA前端路由兜底
	r.NoRoute(func(c *gin.Context) {
		// 如果是API请求，返回404 JSON
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(404, gin.H{
				"code":    404,
				"message": "API endpoint not found",
			})
			return
		}

		// 其他请求返回index.html
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexContent)
	})
}
