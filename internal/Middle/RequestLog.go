package middle

import (
	"blog/internal/Log"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestLogMiddleware 记录用户操作、耗时和状态码
func RequestLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		userID, _ := c.Get("userID")
		Log.ZLog.Info("request_log",
			zap.String("method", c.Request.Method),
			zap.String("path", c.FullPath()),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", duration),
			zap.String("ip", c.ClientIP()),
			zap.Any("user_id", userID),
		)
	}
}
