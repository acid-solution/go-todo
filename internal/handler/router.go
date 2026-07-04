package handler

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, todoHandler *TodoHandler) {
	r.Static("/static", "./static")
	r.StaticFile("/", "./static/index.html")

	r.POST("/api/todos", todoHandler.CreateTodo)
	r.GET("/api/todos", todoHandler.ListTodos)
	r.PUT("/api/todos/:id", todoHandler.UpdateTodo)
	r.PATCH("/api/todos/:id/done", todoHandler.CompleteTodo)
	r.DELETE("/api/todos/:id", todoHandler.DeleteTodo)
}
