package main

import (
	"log"
	"log/slog"
	"os"

	"go-todo/internal/cache"
	"go-todo/internal/config"
	"go-todo/internal/database"
	"go-todo/internal/handler"
	"go-todo/internal/middleware"
	"go-todo/internal/repository"
	"go-todo/internal/service"

	"github.com/gin-gonic/gin"
)

func initLogger() *slog.Logger {
	// 初始化日志记录器，使用 JSON 格式输出到标准输出，只记录级别为Info以上的日志
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func main() {
	// 初始化日志记录器
	logger := initLogger()

	// 加载配置
	cfg := config.Load()
	// 初始化数据库连接
	db := database.InitMySQL(cfg.MySQLDSN)
	// 初始化 Redis 客户端
	redisClient := database.InitRedis(
		cfg.RedisAddr,
		cfg.RedisPassword,
		cfg.RedisDB,
	)
	defer redisClient.Close()

	//手动封装依赖（就是要用到依赖结构的一些函数或变量）
	todoRepo := repository.NewTodoRepository(db)
	todoCache := cache.NewJSONCache(redisClient)
	todoService := service.NewTodoService(
		todoRepo,
		todoCache,
		cfg.TodoListCacheTTL,
		logger,
	)
	todoHandler := handler.NewTodoHandler(todoService)

	userRepo := repository.NewUserRepository(db)       //repo依赖数据库实例
	sessionRepo := repository.NewSessionRepository(db) //repo依赖数据库实例
	userService := service.NewUserService(             //service依赖两个repo和cfg配置
		userRepo,
		sessionRepo,
		cfg.JWTSecret,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)
	userHandler := handler.NewUserHandler(userService) //handler依赖service
	authMiddleware := middleware.Auth(
		sessionRepo,
		cfg.JWTSecret,
	)

	// 创建 Gin 引擎
	r := gin.New()

	// 注册中间件，给所有请求添加统一处理逻辑
	// 1. RequestID 中间件：为每个请求生成唯一的请求 ID，并将其添加到请求上下文中，方便后续日志记录和追踪。
	r.Use(middleware.RequestID())
	// 2. RequestLogger 中间件：记录每个请求的详细信息，包括请求方法、路径、状态码、处理时间等，方便调试和监控。
	r.Use(middleware.RequestLogger(logger))
	// 3. Recovery 中间件：在请求处理过程中，如果发生 panic，能够捕获并记录错误信息，防止服务崩溃。
	r.Use(gin.Recovery())

	// 注册路由，将请求路径与处理函数进行映射
	handler.RegisterRoutes(
		r,
		todoHandler,
		userHandler,
		authMiddleware, //因为是部分用到，所有不能像其他中间件一样放在外面
	)

	// 启动 HTTP 服务，监听指定端口
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("启动服务失败:", err)
	}
}
