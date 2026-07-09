package handler

import (
	"errors"

	"go-todo/internal/response"
	"go-todo/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

// 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// 注册用户
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	// 1. 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailInvalidArgument(c, "用户名和密码不能为空")
		return
	}
	// 2. 调用服务层注册用户
	user, err := h.service.Register(service.RegisterInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		response.FailInternalError(c, "注册失败")
		return
	}
	// 3. 返回注册成功的用户信息
	response.Success(c, response.ToUserResponse(user))
}

// 登录用户
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	// 1. 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailInvalidArgument(c, "用户名和密码不能为空")
		return
	}
	// 2. 调用服务层登录用户
	user, err := h.service.Login(service.LoginInput{
		Username: req.Username,
		Password: req.Password,
	})
	// 3. 根据错误类型返回不同的响应
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.FailInvalidArgument(c, "用户名或密码错误")
			return
		}

		response.FailInternalError(c, "登录失败")
		return
	}
	// 4. 返回登录成功的用户信息
	response.Success(c, response.ToUserResponse(user))
}
