package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger 日志中间件
func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		logger.Info("API request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}

// Recovery 错误恢复中间件
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
				)

				c.JSON(500, gin.H{
					"code":    500,
					"message": "Internal server error",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
