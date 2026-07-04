package handler

import (
	"errors"
	"go-todo/internal/response"
	"go-todo/internal/service"

	"github.com/gin-gonic/gin"
)

type TodoHandler struct {
	service *service.TodoService
}

func NewTodoHandler(service *service.TodoService) *TodoHandler {
	return &TodoHandler{
		service: service,
	}
}

type CreateTodoRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

type UpdateTodoRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

type TodoIDRequest struct {
	ID int64 `uri:"id" binding:"required"`
}

type ListTodoRequest struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
	Completed string `form:"completed"`
}

// CreateTodo 创建一个新的 Todo 任务
func (h *TodoHandler) CreateTodo(c *gin.Context) {
	var req CreateTodoRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailInvalidArgument(c, "任务名称不能为空")
		return
	}

	todo, err := h.service.CreateTodo(service.CreateTodoInput{
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		response.FailInternalError(c, "创建任务失败")
		return
	}

	response.Success(c, response.ToTodoResponse(todo))
}

// ListTodos 查询 Todo 任务列表
func (h *TodoHandler) ListTodos(c *gin.Context) {
	var req ListTodoRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailInvalidArgument(c, "查询参数无效")
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}

	if req.PageSize == 0 {
		req.PageSize = 10
	}

	if req.Page < 1 {
		response.FailInvalidArgument(c, "page 必须大于等于 1")
		return
	}

	if req.PageSize < 1 || req.PageSize > 100 {
		response.FailInvalidArgument(c, "page_size 必须在 1 到 100 之间")
		return
	}

	if req.Completed != "" && req.Completed != "true" && req.Completed != "false" {
		response.FailInvalidArgument(c, "completed 只能是 true 或 false")
		return
	}

	result, err := h.service.ListTodos(service.ListTodosInput{
		Page:      req.Page,
		PageSize:  req.PageSize,
		Completed: req.Completed,
	})
	if err != nil {
		response.FailInternalError(c, "查询任务列表失败")
		return
	}

	items := make([]response.TodoResponse, 0, len(result.Items))
	for i := range result.Items {
		items = append(items, response.ToTodoResponse(&result.Items[i]))
	}

	response.Success(c, response.TodoListResponse{
		Items:      items,
		Page:       result.Page,
		PageSize:   result.PageSize,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	})
}

// UpdateTodo 更新任务
func (h *TodoHandler) UpdateTodo(c *gin.Context) {
	var idReq TodoIDRequest

	if err := c.ShouldBindUri(&idReq); err != nil {
		response.FailInvalidArgument(c, "无效的任务 ID")
		return
	}

	var req UpdateTodoRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailInvalidArgument(c, "任务名称不能为空")
		return
	}

	todo, err := h.service.UpdateTodo(idReq.ID, service.UpdateTodoInput{
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		if errors.Is(err, service.ErrTodoNotFound) {
			response.FailNotFound(c, "任务不存在")
			return
		}

		response.FailInternalError(c, "更新任务失败")
		return
	}

	response.Success(c, response.ToTodoResponse(todo))
}

// CompleteTodo 完成任务
func (h *TodoHandler) CompleteTodo(c *gin.Context) {
	var idReq TodoIDRequest

	if err := c.ShouldBindUri(&idReq); err != nil {
		response.FailInvalidArgument(c, "无效的任务 ID")
		return
	}

	todo, err := h.service.CompleteTodo(idReq.ID)
	if err != nil {
		if errors.Is(err, service.ErrTodoNotFound) {
			response.FailNotFound(c, "任务不存在")
			return
		}

		response.FailInternalError(c, "标记任务完成失败")
		return
	}

	response.Success(c, response.ToTodoResponse(todo))
}

// DeleteTodo 删除任务
func (h *TodoHandler) DeleteTodo(c *gin.Context) {
	var idReq TodoIDRequest

	if err := c.ShouldBindUri(&idReq); err != nil {
		response.FailInvalidArgument(c, "无效的任务 ID")
		return
	}

	if err := h.service.DeleteTodo(idReq.ID); err != nil {
		if errors.Is(err, service.ErrTodoNotFound) {
			response.FailNotFound(c, "任务不存在")
			return
		}

		response.FailInternalError(c, "删除任务失败")
		return
	}

	response.Success(c, nil)
}
