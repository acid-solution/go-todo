package handler

import "github.com/gin-gonic/gin"

func RegisterRoutes(
	r *gin.Engine,
	todoHandler *TodoHandler,
	userHandler *UserHandler,
	authMiddleware gin.HandlerFunc,
) {
	r.Static("/static", "./static")
	r.StaticFile("/", "./static/index.html")

	api := r.Group("/api")

	// 公共接口
	api.POST("/register", userHandler.Register)
	api.POST("/login", userHandler.Login)
	api.POST("/refresh", userHandler.Refresh)

	// 受保护接口
	protected := api.Group("")
	protected.Use(authMiddleware)

	protected.POST("/logout", userHandler.Logout)

	protected.POST("/todos", todoHandler.CreateTodo)
	protected.GET("/todos", todoHandler.ListTodos)
	protected.PUT("/todos/:id", todoHandler.UpdateTodo)
	protected.PATCH("/todos/:id/done", todoHandler.CompleteTodo)
	protected.DELETE("/todos/:id", todoHandler.DeleteTodo)
}
