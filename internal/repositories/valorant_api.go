package repositories

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/emper0r/val-store-server/internal/models"
)

const (
	// API URLs
	loginURL        = "https://auth.riotgames.com/api/v1/authorization"
	entitlementsURL = "https://entitlements.auth.riotgames.com/api/token/v1"
	userInfoURL     = "https://auth.riotgames.com/userinfo"
	versionURL      = "https://valorant-api.com/v1/version"

	// 默认区域
	defaultRegion = "ap"
)

// 用于解析版本API响应的结构体
type versionResponse struct {
	Status int `json:"status"`
	Data   struct {
		RiotClientVersion string `json:"riotClientVersion"`
	} `json:"data"`
}

// ValorantAPI 处理与Valorant API的交互
type ValorantAPI struct {
	client        *http.Client
	region        string
	clientVersion string
}

// fetchLatestClientVersion 从valorant-api.com获取最新的客户端版本
func fetchLatestClientVersion(client *http.Client) (string, error) {
	req, err := http.NewRequest(http.MethodGet, versionURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建版本请求失败: %w", err)
	}
	req.Header.Set("User-Agent", "val-store-server")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求版本信息失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("获取版本信息失败，状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
	}

	var versionData versionResponse
	if err := json.NewDecoder(resp.Body).Decode(&versionData); err != nil {
		return "", fmt.Errorf("解析版本信息失败: %w", err)
	}

	if versionData.Status != 200 || versionData.Data.RiotClientVersion == "" {
		return "", errors.New("从版本信息响应中未找到有效的客户端版本")
	}

	return versionData.Data.RiotClientVersion, nil
}

// NewValorantAPI 创建一个新的ValorantAPI实例
func NewValorantAPI() (*ValorantAPI, error) {
	// 创建带cookie jar的HTTP客户端，以便能够维护会话
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	// TLS配置，提高安全性和兼容性
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// 创建具有更健壮配置的Transport
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true, // 启用IPv4/IPv6双栈
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		TLSClientConfig:       tlsConfig,
	}

	client := &http.Client{
		Jar:       jar,
		Timeout:   60 * time.Second, // 增加超时时间到60秒
		Transport: transport,
	}

	// 创建一个用于获取版本的临时客户端
	versionClient := &http.Client{
		Timeout:   15 * time.Second,
		Transport: transport,
	}

	// 备用的客户端版本，以防无法获取最新版本
	fallbackClientVersion := "release-10.07-shipping-6-3399868"
	fetchedClientVersion, err := fetchLatestClientVersion(versionClient)
	currentClientVersion := fallbackClientVersion

	if err != nil {
		log.Printf("警告: 无法获取最新的客户端版本: %v。将使用备用版本: %s", err, fallbackClientVersion)
	} else {
		currentClientVersion = fetchedClientVersion
		log.Printf("成功获取最新的客户端版本: %s", currentClientVersion)
	}

	api := &ValorantAPI{
		client:        client,
		region:        defaultRegion,
		clientVersion: currentClientVersion,
	}

	return api, nil
}

// SetRegion 设置用户的地区
func (v *ValorantAPI) SetRegion(region string) {
	if region == "" {
		v.region = defaultRegion
		return
	}

	// 转小写处理区域代码
	region = strings.ToLower(region)

	// 根据API文档更新区域代码映射
	switch region {
	case "na", "latam", "br":
		// latam和br使用na区域
		v.region = "na"
	case "eu":
		v.region = "eu"
	case "ap", "kr":
		v.region = region
	default:
		// 对于其他输入，使用默认区域
		v.region = defaultRegion
	}

	log.Printf("区域设置已更新: 输入=%s, 最终区域=%s\n", region, v.region)
}

// ParseCookieString 解析Cookie字符串为map
func ParseCookieString(cookieStr string) map[string]string {
	cookieMap := make(map[string]string)

	// 如果输入为空，返回空map
	if cookieStr == "" {
		return cookieMap
	}

	// 清理字符串 - 处理多行和特殊字符
	cleanCookieStr := strings.ReplaceAll(cookieStr, "\r", "")
	cleanCookieStr = strings.ReplaceAll(cleanCookieStr, "\n", "")

	// 尝试多种分隔符
	var parts []string
	if strings.Contains(cleanCookieStr, ";") {
		parts = strings.Split(cleanCookieStr, ";")
	} else if strings.Contains(cleanCookieStr, ",") {
		parts = strings.Split(cleanCookieStr, ",")
	} else {
		// 单个Cookie的情况
		parts = []string{cleanCookieStr}
	}

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// 尝试几种不同的分隔方式
		var name, value string

		// 标准的name=value格式
		equalsIndex := strings.Index(part, "=")
		if equalsIndex != -1 {
			name = strings.TrimSpace(part[:equalsIndex])
			value = strings.TrimSpace(part[equalsIndex+1:])

			// 移除值两端的引号（如果有）
			value = strings.Trim(value, `"'`)

			if name != "" {
				cookieMap[name] = value
			}
			continue
		}

		// 尝试其他可能的格式
		colonIndex := strings.Index(part, ":")
		if colonIndex != -1 {
			name = strings.TrimSpace(part[:colonIndex])
			value = strings.TrimSpace(part[colonIndex+1:])
			value = strings.Trim(value, `"'`)

			if name != "" {
				cookieMap[name] = value
			}
		}
	}

	return cookieMap
}

// FilterEssentialCookies 过滤保留必要的Cookie
func FilterEssentialCookies(cookies map[string]string) map[string]string {
	essentialCookies := make(map[string]string)

	// 优先检查的关键Cookie
	keyCookies := []string{"ssid", "csid", "clid", "sub", "tdid", "asid", "did"}

	// 检查并添加关键Cookie
	for _, key := range keyCookies {
		if value, exists := cookies[key]; exists && value != "" {
			essentialCookies[key] = value
		}
	}

	// 如果没有找到任何关键Cookie，返回所有Cookie
	if len(essentialCookies) == 0 {
		return cookies
	}

	return essentialCookies
}

// AuthenticateWithCookies 使用Cookie进行认证
func (v *ValorantAPI) AuthenticateWithCookies(cookies map[string]string) (*models.UserSession, error) {
	// 过滤保留有用的Cookie
	filteredCookies := FilterEssentialCookies(cookies)
	if len(filteredCookies) == 0 {
		return nil, errors.New("没有提供任何有效的Cookie")
	}

	// 创建带有cookie jar的新HTTP客户端
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	// 禁用重定向的客户端
	client := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// 禁止自动跟随重定向
			return http.ErrUseLastResponse
		},
	}

	// 使用这个带cookie的客户端替换当前的客户端
	v.client = client

	// 构建authorize URL
	authURL := "https://auth.riotgames.com/authorize?redirect_uri=https%3A%2F%2Fplayvalorant.com%2Fopt_in&client_id=play-valorant-web-prod&response_type=token%20id_token&scope=account%20openid&nonce=1"

	// 创建请求
	req, err := http.NewRequest(http.MethodGet, authURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 添加请求头
	v.setRiotRequestHeaders(req, filteredCookies)

	// 发送请求
	resp, err := v.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusSeeOther {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("认证请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
	}

	// 获取Location头部
	location := resp.Header.Get("Location")
	if location == "" {
		return nil, errors.New("响应中没有Location头")
	}

	if strings.Contains(location, "/login") {
		return nil, errors.New("Cookie无效或已过期")
	}

	if !strings.Contains(location, "access_token=") {
		return nil, errors.New("无法从响应中提取令牌")
	}

	// 从Location URL中提取访问令牌
	accessToken, err := parseAccessTokenFromURI(location)
	if err != nil {
		return nil, fmt.Errorf("提取访问令牌失败: %w", err)
	}

	// 获取授权令牌
	entitlementToken, err := v.getEntitlementToken(accessToken)
	if err != nil {
		return nil, fmt.Errorf("获取授权令牌失败: %w", err)
	}

	// 获取用户信息
	userInfo, err := v.getUserInfo(accessToken)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 创建用户会话
	session := &models.UserSession{
		UserID:       userInfo.Sub,
		Username:     userInfo.Email,
		AccessToken:  accessToken,
		Entitlement:  entitlementToken,
		RiotUsername: userInfo.Acct.GameName,
		RiotTagline:  userInfo.Acct.TagLine,
		Cookies:      filteredCookies,
	}

	return session, nil
}

// 设置Riot请求头
func (v *ValorantAPI) setRiotRequestHeaders(req *http.Request, cookies map[string]string) {
	// 设置常规请求头
	req.Header.Set("User-Agent", "RiotClient/62.0.1.4852791.4789131 rso-auth (Windows;10;;Professional, x64)")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Origin", "https://auth.riotgames.com")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("X-Riot-ClientVersion", v.clientVersion) // 添加客户端版本

	// 添加Cookie
	var cookieStrings []string
	for name, value := range cookies {
		cookieStrings = append(cookieStrings, fmt.Sprintf("%s=%s", name, value))
	}
	cookieHeader := strings.Join(cookieStrings, "; ")
	if cookieHeader != "" {
		req.Header.Set("Cookie", cookieHeader)
	}
}

// 从URI中提取访问令牌
func parseAccessTokenFromURI(uri string) (string, error) {
	startIndex := strings.Index(uri, "access_token=")
	if startIndex == -1 {
		return "", errors.New("URI中没有access_token参数")
	}
	startIndex += len("access_token=")

	endIndex := strings.Index(uri[startIndex:], "&")
	if endIndex == -1 {
		return uri[startIndex:], nil
	}

	return uri[startIndex : startIndex+endIndex], nil
}

// 获取授权令牌
func (v *ValorantAPI) getEntitlementToken(accessToken string) (string, error) {
	req, err := http.NewRequest(http.MethodPost, entitlementsURL, nil)
	if err != nil {
		return "", err
	}

	// 设置授权头
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Riot-ClientVersion", v.clientVersion) // 添加客户端版本

	// 发送请求
	resp, err := v.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("获取授权令牌失败，状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
	}

	// 解析响应
	var entitlementResp models.ValorantEntitlementResponse
	if err := json.NewDecoder(resp.Body).Decode(&entitlementResp); err != nil {
		return "", err
	}

	return entitlementResp.EntitlementToken, nil
}

// 获取用户信息
func (v *ValorantAPI) getUserInfo(accessToken string) (*models.ValorantUserInfoResponse, error) {
	req, err := http.NewRequest(http.MethodGet, userInfoURL, nil)
	if err != nil {
		return nil, err
	}

	// 设置授权头
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-Riot-ClientVersion", v.clientVersion) // 添加客户端版本

	// 发送请求
	resp, err := v.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取用户信息失败，状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
	}

	// 解析响应
	var userInfo models.ValorantUserInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}
