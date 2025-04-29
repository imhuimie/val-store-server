package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLoggerMiddleware 记录请求处理时间的中间件
func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 计算处理时间
		duration := time.Since(startTime)

		// 输出日志
		path := c.Request.URL.Path
		method := c.Request.Method
		status := c.Writer.Status()

		// 如果请求处理超过5秒，记录为警告
		if duration > 5*time.Second {
			log.Printf("[WARNING] 慢请求: %s %s | 状态: %d | 耗时: %v", method, path, status, duration)
		} else {
			log.Printf("[INFO] 请求: %s %s | 状态: %d | 耗时: %v", method, path, status, duration)
		}
	}
}
