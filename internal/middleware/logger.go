package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		start := time.Now()

		c.Next()
		// 记录请求结束时间
		latency := time.Since(start)
		// 获取请求的状态码和请求ID
		status := c.Writer.Status()
		requestID := c.GetString(requestIDKey)
		// 构建日志属性
		attrs := []any{
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", status,
			"latency", latency.String(),
			"client_ip", c.ClientIP(),
		}
		// 如果有错误信息，添加到日志属性中
		if len(c.Errors) > 0 {
			attrs = append(attrs, "errors", c.Errors.String())
		}
		// 根据状态码记录不同级别的日志,每个请求完成后都会记录一条日志，包含请求的基本信息和处理结果
		switch {
		case status >= http.StatusInternalServerError:
			logger.Error("request completed", attrs...)
		case status >= http.StatusBadRequest:
			logger.Warn("request completed", attrs...)
		default:
			logger.Info("request completed", attrs...)
		}
	}
}
