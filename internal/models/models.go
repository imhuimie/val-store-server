package models

import (
	"github.com/golang-jwt/jwt/v5"
)

// 区域常量
const (
	RegionAP    = "ap"    // 亚太地区
	RegionNA    = "na"    // 北美
	RegionEU    = "eu"    // 欧洲
	RegionKR    = "kr"    // 韩国
	RegionLATAM = "latam" // 拉丁美洲
	RegionBR    = "br"    // 巴西
)

// CookieLoginRequest Cookie登录请求
type CookieLoginRequest struct {
	Cookies string `json:"cookies" binding:"required"`
	Region  string `json:"region"` // 可选的区域设置参数
}

// UserSession 用户会话信息
type UserSession struct {
	UserID       string            `json:"user_id"`
	Username     string            `json:"username"`
	AccessToken  string            `json:"access_token"`
	Entitlement  string            `json:"entitlement_token"`
	RiotUsername string            `json:"riot_username"`
	RiotTagline  string            `json:"riot_tagline"`
	Region       string            `json:"region"` // 用户区域
	Cookies      map[string]string `json:"-"`      // Cookie不会返回给客户端
}

// JWTClaims 定义JWT令牌的声明
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// UserTokensResponse 登录成功后的响应
type UserTokensResponse struct {
	Token string `json:"token"` // JWT令牌
	User  struct {
		Username string `json:"username"`
		UserID   string `json:"user_id"`
	} `json:"user"`
}

// APIError 统一API错误响应格式
type APIError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// APISuccess 统一API成功响应格式
type APISuccess struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ValorantUserInfoResponse 包含用户ID和其他信息
type ValorantUserInfoResponse struct {
	Sub      string `json:"sub"` // 用户ID
	Email    string `json:"email"`
	Name     string `json:"name"`
	Tag      string `json:"tag"`
	Picture  string `json:"picture,omitempty"`
	Country  string `json:"country,omitempty"`
	Locale   string `json:"locale,omitempty"`
	PhoneID  string `json:"phone_id,omitempty"`
	Verified bool   `json:"email_verified"`
	// 游戏名称和标签
	Acct struct {
		GameName string `json:"game_name"`
		TagLine  string `json:"tag_line"`
	} `json:"acct"`
}

// ValorantEntitlementResponse 包含Valorant的授权令牌
type ValorantEntitlementResponse struct {
	EntitlementToken string `json:"entitlements_token"`
}
