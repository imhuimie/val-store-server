package api

import (
	"github.com/emper0r/val-store-server/internal/api/handlers"
	"github.com/emper0r/val-store-server/internal/api/middleware"
	"github.com/emper0r/val-store-server/internal/repositories"
	"github.com/emper0r/val-store-server/internal/services"
	"github.com/gin-gonic/gin"
)

// SetupRouter 设置所有API路由
func SetupRouter(router *gin.Engine) *gin.Engine {
	// 配置CORS中间件
	router.Use(middleware.CorsMiddleware())

	// 初始化存储库
	valorantAPI, err := repositories.NewValorantAPI()
	if err != nil {
		panic(err)
	}

	// 初始化服务
	authService := services.NewAuthService(valorantAPI)

	// 初始化处理器
	authHandler := handlers.NewAuthHandler(authService)

	// API路由组
	api := router.Group("/api")
	{
		// 注册认证处理器的路由
		authHandler.RegisterRoutes(api)
	}

	return router
}
