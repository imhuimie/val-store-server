package middleware

import (
	"net/http"
	"strings"

	"github.com/emper0r/val-store-server/internal/models"
	"github.com/emper0r/val-store-server/internal/services"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware 创建认证中间件
func AuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.APIError{
				Status:  http.StatusUnauthorized,
				Message: "未授权",
				Error:   "缺少认证令牌",
			})
			c.Abort()
			return
		}

		// 提取令牌
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, models.APIError{
				Status:  http.StatusUnauthorized,
				Message: "未授权",
				Error:   "认证头格式无效",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 验证令牌
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.APIError{
				Status:  http.StatusUnauthorized,
				Message: "未授权",
				Error:   "无效的令牌: " + err.Error(),
			})
			c.Abort()
			return
		}

		// 将用户信息存储在上下文中，供后续处理使用
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}
