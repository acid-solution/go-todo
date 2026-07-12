package middleware

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"go-todo/internal/repository"
	"go-todo/internal/response"
	"go-todo/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	ContextUserIDKey    = "auth_user_id"
	ContextSessionIDKey = "auth_session_id"
)

// Auth 校验 Access Token 和对应的登录 Session。
func Auth(
	sessionRepo *repository.SessionRepository,
	jwtSecret string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 提取 Authorization 请求头
		tokenString, ok := extractBearerToken(c)
		if !ok {
			abortUnauthorized(c)
			return
		}

		// 2. 解析并校验 JWT
		claims, err := service.ParseAccessToken(tokenString, jwtSecret)
		if err != nil {
			abortUnauthorized(c)
			return
		}

		// 3. 将 sub 转换成用户 ID
		userID, err := strconv.ParseUint(claims.Subject, 10, 64) //字符串，进制，位数
		if err != nil || userID == 0 {
			abortUnauthorized(c)
			return
		}

		// 4. 根据 sid 查询登录 Session
		session, err := sessionRepo.GetByID(claims.SessionID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				abortUnauthorized(c)
				return
			}

			response.FailInternalError(c, "认证服务异常")
			c.Abort()
			return
		}

		// 5. 校验 Session 状态
		now := time.Now()

		if session.UserID != userID ||
			session.RevokedAt != nil ||
			!now.Before(session.ExpiresAt) {
			abortUnauthorized(c)
			return
		}

		// 6. 把身份信息放入当前请求的 Context
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextSessionIDKey, session.ID)

		// 7. 继续执行后续中间件和 Handler
		c.Next()
	}
}

func extractBearerToken(c *gin.Context) (string, bool) {
	//得到请求头中的 Authorization 字段
	header := c.GetHeader("Authorization")
	// Authorization 字段的格式为 "Bearer <token>"
	parts := strings.Fields(header)
	//不是两部分，或者第一部分不是 Bearer，或者第二部分为空，则返回 false
	if len(parts) != 2 {
		return "", false
	}

	if !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}

	if parts[1] == "" {
		return "", false
	}

	return parts[1], true
}

func abortUnauthorized(c *gin.Context) {
	response.FailUnauthorized(c, "未登录或登录已过期")
	c.Abort()
}
