package services

import (
	"fmt"
	"time"

	"github.com/emper0r/val-store-server/internal/config"
	"github.com/emper0r/val-store-server/internal/models"
	"github.com/emper0r/val-store-server/internal/repositories"
	"github.com/golang-jwt/jwt/v5"
)

// AuthService 处理认证相关的业务逻辑
type AuthService struct {
	valorantAPI *repositories.ValorantAPI
	jwtSecret   string
	tokenExpiry time.Duration
}

// NewAuthService 创建新的认证服务
func NewAuthService(valorantAPI *repositories.ValorantAPI) *AuthService {
	// 从环境变量获取JWT密钥，默认为一个随机字符串（仅用于开发环境）
	jwtSecret := config.GetEnv("JWT_SECRET", "val-store-server-secret-key-development-only")
	// JWT令牌有效期，默认24小时
	tokenExpiry := 24 * time.Hour

	return &AuthService{
		valorantAPI: valorantAPI,
		jwtSecret:   jwtSecret,
		tokenExpiry: tokenExpiry,
	}
}

// LoginWithCookies 使用Cookie进行登录，返回JWT令牌
func (s *AuthService) LoginWithCookies(cookieStr string, region string) (*models.UserTokensResponse, error) {
	// 解析Cookie字符串
	cookies := repositories.ParseCookieString(cookieStr)

	// 如果没有解析出任何Cookie，返回错误
	if len(cookies) == 0 {
		return nil, fmt.Errorf("无法解析Cookie字符串，请确保格式正确")
	}

	// 如果提供了区域，则设置
	if region != "" {
		s.valorantAPI.SetRegion(region)
	}

	// 调用认证方法
	session, err := s.valorantAPI.AuthenticateWithCookies(cookies)
	if err != nil {
		return nil, fmt.Errorf("Cookie认证失败: %w", err)
	}

	// 设置区域
	if region != "" {
		session.Region = region
	} else {
		// 如果未提供区域，使用默认值AP
		session.Region = models.RegionAP
	}

	// 生成JWT令牌
	token, err := s.generateJWT(session)
	if err != nil {
		return nil, fmt.Errorf("生成JWT失败: %w", err)
	}

	// 构建格式化的用户名
	formattedUsername := session.RiotUsername
	if session.RiotTagline != "" {
		formattedUsername = fmt.Sprintf("%s#%s", session.RiotUsername, session.RiotTagline)
	}

	// 构建响应
	response := &models.UserTokensResponse{
		Token: token,
		User: struct {
			Username string `json:"username"`
			UserID   string `json:"user_id"`
		}{
			Username: formattedUsername,
			UserID:   session.UserID,
		},
	}

	return response, nil
}

// generateJWT 生成JWT令牌
func (s *AuthService) generateJWT(session *models.UserSession) (string, error) {
	// 构建格式化的用户名
	formattedUsername := session.RiotUsername
	if session.RiotTagline != "" {
		formattedUsername = fmt.Sprintf("%s#%s", session.RiotUsername, session.RiotTagline)
	}

	// 设置JWT声明
	claims := models.JWTClaims{
		UserID:   session.UserID,
		Username: formattedUsername,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// 创建JWT令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用密钥签名令牌
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken 验证JWT令牌的有效性并返回声明
func (s *AuthService) ValidateToken(tokenString string) (*models.JWTClaims, error) {
	// 解析JWT令牌
	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("非预期的签名方法: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("解析令牌失败: %w", err)
	}

	// 验证令牌有效性并提取声明
	if claims, ok := token.Claims.(*models.JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("无效的令牌")
}
