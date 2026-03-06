// bill-management/cmd/server/main.go
package main

import (
	"bill-management/internal/middleware"
	"bill-management/internal/model"
	"bill-management/internal/router"
	"bill-management/pkg/config"
	"bill-management/pkg/databaseutil"
	"bill-management/pkg/jwtutil"
	"bill-management/pkg/logger"
	"bill-management/pkg/redisutil"
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// ===================== 1. 初始化核心组件 =====================
	config.InitConfig("config") // 1.1 初始化配置（从文件/环境变量等加载）
	logger.InitLogger()         // 1.2 初始化日志组件（使用Zap）

	db := databaseutil.InitPostgreSQL() // 1.3 初始化数据库连接（使用GORM连接PostgreSQL）

	// 自动迁移表（创建/更新User表结构）
	err := db.AutoMigrate(&model.User{})
	if err != nil {
		logger.Fatal("表结构迁移失败", zap.Error(err))
	}
	logger.Info("PostgreSQL数据库初始化完成")

	// 1.4 初始化Redis客户端（用于JWT黑名单）
	redisClient := redisutil.NewRedisClient(&config.GetConfig().Database.Redis)
	// 测试Redis连接
	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		logger.Fatal("Redis连接失败", zap.Error(err))
	}
	logger.Info("Redis客户端初始化完成")

	// 1.5 初始化JWT组件（整合Redis黑名单）
	// 创建Redis黑名单存储
	redisStorage := jwtutil.NewRedisBlacklistStorage(redisClient)
	// 创建JWT服务
	jwtService := jwtutil.NewJWTService(config.GetConfig(), redisStorage)
	logger.Info("JWT组件初始化完成")
	logger.Info("业务层/控制器初始化完成")

	// ===================== 3. 初始化Gin引擎 & 注册路由 =====================
	// 3.1 设置Gin模式（生产环境设为ReleaseMode）
	gin.SetMode(gin.ReleaseMode)

	// 3.2 创建Gin引擎（替换默认日志/恢复中间件）
	r := gin.New()

	// 集成自定义日志中间件 + 异常恢复中间件
	r.Use(middleware.GinLogger(), middleware.GinRecovery())

	// 3.3 注册用户路由
	router.RegisterUserRouter(r, db, jwtService)
	logger.Info("路由注册完成")

	// ===================== 4. 启动HTTP服务 =====================
	// 4.1 创建HTTP服务器
	server := &http.Server{
		Addr:    ":" + strconv.Itoa(config.GetConfig().Server.Port), // 如 ":8080"
		Handler: r,
		// 可选：设置超时时间
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	// 4.2 异步启动服务（避免阻塞）
	go func() {
		logger.Info("HTTP服务启动成功", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP服务启动失败", zap.Error(err))
		}
	}()

	// ===================== 5. 优雅关闭服务 =====================
	// 监听系统信号（SIGINT/SIGTERM）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // 阻塞等待信号
	logger.Info("开始优雅关闭服务...")

	// 创建关闭上下文（5秒超时）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭HTTP服务
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("服务关闭失败", zap.Error(err))
	}
	logger.Info("服务已优雅关闭")
}
