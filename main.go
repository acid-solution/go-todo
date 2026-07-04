package main

import (
	"log"
	"log/slog"
	"os"

	"go-todo/internal/config"
	"go-todo/internal/database"
	"go-todo/internal/handler"
	"go-todo/internal/middleware"
	"go-todo/internal/repository"
	"go-todo/internal/service"

	"github.com/gin-gonic/gin"
)

func initLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func main() {
	logger := initLogger()

	cfg := config.Load()

	db := database.InitMySQL(cfg.MySQLDSN)

	todoRepo := repository.NewTodoRepository(db)
	todoService := service.NewTodoService(todoRepo)
	todoHandler := handler.NewTodoHandler(todoService)

	r := gin.New()

	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger(logger))
	r.Use(gin.Recovery())

	handler.RegisterRoutes(r, todoHandler)

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("启动服务失败:", err)
	}
}
