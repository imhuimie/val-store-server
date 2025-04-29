package handlers

import (
	"net/http"

	"github.com/emper0r/val-store-server/internal/models"
	"github.com/emper0r/val-store-server/internal/services"
	"github.com/gin-gonic/gin"
)

// StoreHandler 处理商店相关的请求
type StoreHandler struct {
	storeService *services.StoreService
	authService  *services.AuthService
}

// NewStoreHandler 创建一个新的商店处理器
func NewStoreHandler(storeService *services.StoreService, authService *services.AuthService) *StoreHandler {
	return &StoreHandler{
		storeService: storeService,
		authService:  authService,
	}
}

// GetStoreData 获取用户的商店数据
func (h *StoreHandler) GetStoreData(c *gin.Context) {
	// 从请求中提取令牌
	tokenString := extractTokenFromRequest(c)
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, models.APIError{
			Status:  http.StatusUnauthorized,
			Message: "未提供认证令牌",
		})
		return
	}

	// 验证令牌并获取用户信息
	claims, err := h.authService.ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIError{
			Status:  http.StatusUnauthorized,
			Message: "无效的认证令牌",
			Error:   err.Error(),
		})
		return
	}

	// 从声明中获取必要的信息
	userID := claims.UserID
	accessToken := claims.AccessToken
	entitlementToken := claims.EntitlementToken

	// 从请求获取区域，默认使用JWT中的区域或AP
	region := c.DefaultQuery("region", claims.Region)
	if region == "" {
		region = "ap"
	}

	// 获取商店数据
	storeData, err := h.storeService.GetStoreData(userID, accessToken, entitlementToken, region)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIError{
			Status:  http.StatusInternalServerError,
			Message: "获取商店数据失败",
			Error:   err.Error(),
		})
		return
	}

	// 返回商店数据
	c.JSON(http.StatusOK, models.APISuccess{
		Status:  http.StatusOK,
		Message: "成功获取商店数据",
		Data:    storeData,
	})
}

// GetRawStoreData 获取原始JSON格式的商店数据
func (h *StoreHandler) GetRawStoreData(c *gin.Context) {
	// 从请求中提取令牌
	tokenString := extractTokenFromRequest(c)
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, models.APIError{
			Status:  http.StatusUnauthorized,
			Message: "未提供认证令牌",
		})
		return
	}

	// 验证令牌并获取用户信息
	claims, err := h.authService.ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIError{
			Status:  http.StatusUnauthorized,
			Message: "无效的认证令牌",
			Error:   err.Error(),
		})
		return
	}

	// 从声明中获取必要的信息
	userID := claims.UserID
	accessToken := claims.AccessToken
	entitlementToken := claims.EntitlementToken

	// 从请求获取区域，默认使用JWT中的区域或AP
	region := c.DefaultQuery("region", claims.Region)
	if region == "" {
		region = "ap"
	}

	// 获取原始商店数据
	rawData, err := h.storeService.GetStoreDataRaw(userID, accessToken, entitlementToken, region)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIError{
			Status:  http.StatusInternalServerError,
			Message: "获取商店数据失败",
			Error:   err.Error(),
		})
		return
	}

	// 设置Content-Type并返回原始JSON数据
	c.Data(http.StatusOK, "application/json", rawData)
}

// RegisterRoutes 注册商店相关的路由
func (h *StoreHandler) RegisterRoutes(router *gin.RouterGroup) {
	store := router.Group("/store")
	{
		store.GET("/data", h.GetStoreData)
		store.GET("/raw", h.GetRawStoreData)
	}
}

// extractTokenFromRequest 从请求中提取JWT令牌
func extractTokenFromRequest(c *gin.Context) string {
	// 首先尝试从Authorization头获取
	tokenString := c.GetHeader("Authorization")
	if tokenString != "" {
		// 移除"Bearer "前缀（如果存在）
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			return tokenString[7:]
		}
		return tokenString
	}

	// 然后尝试从cookie中获取
	tokenCookie, err := c.Cookie("token")
	if err == nil {
		return tokenCookie
	}

	// 最后尝试从查询参数获取
	return c.Query("token")
}
