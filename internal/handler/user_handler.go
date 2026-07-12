package handler

import (
	"errors"

	"go-todo/internal/middleware"
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

// 刷新令牌请求
type RefreshRequest struct {
	SessionID    string `json:"session_id" binding:"required"`
	RefreshToken string `json:"refresh_token" binding:"required"`
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
		if errors.Is(err, service.ErrUsernameAlreadyExists) {
			response.FailConflict(c, "用户名已存在")
			return
		}

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
	result, err := h.service.Login(service.LoginInput{
		Username: req.Username,
		Password: req.Password,
	})

	// 3. 根据错误类型返回不同响应
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.FailInvalidArgument(c, "用户名或密码错误")
			return
		}

		response.FailInternalError(c, "登录失败")
		return
	}

	// 4. 返回用户信息和登录凭证
	response.Success(c, response.AuthResponse{
		User:         response.ToUserResponse(result.User),
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		SessionID:    result.SessionID,
	})
}

// 刷新登录令牌
func (h *UserHandler) Refresh(c *gin.Context) {
	var req RefreshRequest

	// 1. 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailInvalidArgument(c, "session_id 和 refresh_token 不能为空")
		return
	}

	// 2. 调用服务层刷新令牌
	result, err := h.service.Refresh(service.RefreshInput{
		SessionID:    req.SessionID,
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidSession) {
			response.FailUnauthorized(c, "登录状态已失效，请重新登录")
			return
		}

		response.FailInternalError(c, "刷新登录状态失败")
		return
	}

	// 3. 返回轮换后的新令牌
	response.Success(c, response.RefreshResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		SessionID:    result.SessionID,
	})
}

// 退出当前登录
func (h *UserHandler) Logout(c *gin.Context) {
	value, exists := c.Get(middleware.ContextSessionIDKey)
	if !exists {
		response.FailUnauthorized(c, "未登录或登录已过期")
		return
	}

	sessionID, ok := value.(string)
	if !ok || sessionID == "" {
		response.FailUnauthorized(c, "未登录或登录已过期")
		return
	}

	if err := h.service.Logout(sessionID); err != nil {
		if errors.Is(err, service.ErrInvalidSession) {
			response.FailUnauthorized(c, "登录状态已失效")
			return
		}

		response.FailInternalError(c, "退出登录失败")
		return
	}

	response.Success(c, nil)
}
