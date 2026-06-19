package middleware

import (
	"net/http"
	"time"

	"iot-platform/pkg/errors"
	"iot-platform/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RequestID 请求 ID 中间件
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// Logger 请求日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		fields := []zap.Field{
			zap.Int("status", statusCode),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.Duration("latency", latency),
		}

		requestID, _ := c.Get("request_id")
		if rid, ok := requestID.(string); ok {
			fields = append(fields, zap.String("request_id", rid))
		}

		if statusCode >= 500 {
			logger.Error("Request error", fields...)
		} else if statusCode >= 400 {
			logger.Warn("Request warning", fields...)
		} else {
			logger.Info("Request", fields...)
		}
	}
}

// ErrorHandler 全局错误处理
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			requestID, _ := c.Get("request_id")

			if appErr, ok := err.(*errors.AppError); ok {
				logger.Warn("App error",
					zap.String("code", appErr.Code),
					zap.String("message", appErr.Message),
					zap.Any("request_id", requestID),
				)
				c.JSON(appErr.StatusCode, gin.H{
					"code":       appErr.Code,
					"message":    appErr.Message,
					"request_id": requestID,
				})
				return
			}

			// 未知错误
			logger.Error("Unexpected error",
				zap.Error(err),
				zap.Any("request_id", requestID),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":       "INTERNAL_ERROR",
				"message":    "服务器内部错误",
				"request_id": requestID,
			})
		}
	}
}

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
