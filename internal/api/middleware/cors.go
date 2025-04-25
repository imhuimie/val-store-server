package middleware

import (
	"net/http"
	"strings"

	"github.com/emper0r/val-store-server/internal/config"
	"github.com/gin-gonic/gin"
)

// CorsMiddleware 创建CORS中间件
func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取允许的域名列表
		allowedOrigins := config.GetEnv("ALLOWED_ORIGINS", "http://localhost:3000")
		origins := strings.Split(allowedOrigins, ",")

		// 获取请求的Origin
		origin := c.Request.Header.Get("Origin")

		// 检查请求的Origin是否在允许列表中
		allowOrigin := false
		for _, o := range origins {
			if origin == strings.TrimSpace(o) {
				allowOrigin = true
				break
			}
		}

		// 如果Origin在允许列表中，则设置CORS头
		if allowOrigin {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		}

		// 处理OPTIONS请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
