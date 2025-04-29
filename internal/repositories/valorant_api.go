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
	storeURL        = "https://pd.%s.a.pvp.net/store/v3/storefront/%s"

	// HTTP Headers
	clientPlatform = "ew0KCSJwbGF0Zm9ybVR5cGUiOiAiUEMiLA0KCSJwbGF0Zm9ybU9TIjogIldpbmRvd3MiLA0KCSJwbGF0Zm9ybU9TVmVyc2lvbiI6ICIxMC4wLjE5MDQyLjEuMjU2LjY0Yml0IiwNCgkicGxhdGZvcm1DaGlwc2V0IjogIlVua25vd24iDQp9"

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

	// 备用的客户端版本，以防无法获取最新版本
	currentClientVersion := "release-10.07-shipping-6-3399868"

	// 创建API实例
	api := &ValorantAPI{
		client:        client,
		region:        defaultRegion,
		clientVersion: currentClientVersion,
	}

	// 异步获取最新版本，避免阻塞API初始化
	go func() {
		// 创建一个用于获取版本的临时客户端
		versionClient := &http.Client{
			Timeout:   15 * time.Second,
			Transport: transport,
		}

		fetchedClientVersion, err := fetchLatestClientVersion(versionClient)
		if err != nil {
			log.Printf("警告: 无法获取最新的客户端版本: %v。将使用备用版本: %s", err, currentClientVersion)
		} else {
			// 更新客户端版本
			api.clientVersion = fetchedClientVersion
			log.Printf("成功获取最新的客户端版本: %s", fetchedClientVersion)
		}
	}()

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

// EnhancedParseCookieString 增强版Cookie解析函数，处理更多格式
func EnhancedParseCookieString(cookieStr string) map[string]string {
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

// ParseCookieString 解析Cookie字符串为map
func ParseCookieString(cookieStr string) map[string]string {
	return EnhancedParseCookieString(cookieStr)
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
	// 直接使用authorize端点进行认证
	log.Println("正在进行Cookie认证...")
	session, err := v.authenticateWithCookiesViaAuthorizeEndpoint(cookies)
	if err != nil {
		return nil, fmt.Errorf("认证失败: %w", err)
	}

	// 获取用户信息并设置Riot用户名
	userInfo, err := v.getUserInfo(session.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	session.RiotUsername = userInfo.Name
	session.RiotTagline = userInfo.Tag
	session.Region = v.region

	return session, nil
}

// authenticateWithCookiesViaAuthorizeEndpoint 通过authorize端点进行认证
func (v *ValorantAPI) authenticateWithCookiesViaAuthorizeEndpoint(cookies map[string]string) (*models.UserSession, error) {
	// 创建带有cookie jar的新HTTP客户端
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	// 使用禁用重定向的客户端 - 这一点很重要
	client := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// 禁止自动跟随重定向
			return http.ErrUseLastResponse
		},
	}

	// 保存原始客户端
	originalClient := v.client
	// 使用新客户端替换当前的客户端
	v.client = client
	// 在函数结束时恢复原始客户端
	defer func() {
		v.client = originalClient
	}()

	// 过滤保留有用的Cookie
	essentialCookies := FilterEssentialCookies(cookies)
	if len(essentialCookies) == 0 {
		return nil, errors.New("没有提供任何有效的Cookie")
	}

	// 通过authorize端点获取token
	// 构建请求URL
	authURL := "https://auth.riotgames.com/authorize?redirect_uri=https%3A%2F%2Fplayvalorant.com%2Fopt_in&client_id=play-valorant-web-prod&response_type=token%20id_token&scope=account%20openid&nonce=1"

	// 创建请求
	req, err := http.NewRequest(http.MethodGet, authURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建authorize请求失败: %w", err)
	}

	// 添加所有必要的头部和Cookie
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "RiotClient/"+v.clientVersion)
	for name, value := range essentialCookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}

	// 发送请求
	resp, err := v.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送authorize请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码，应该是302或303（重定向）
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
		RiotUsername: userInfo.Name,
		RiotTagline:  userInfo.Tag,
		Cookies:      essentialCookies,
	}

	return session, nil
}

// setRiotRequestHeaders 设置Riot API请求头
func (v *ValorantAPI) setRiotRequestHeaders(req *http.Request, cookies map[string]string) {
	// 添加通用头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "RiotClient/"+v.clientVersion)

	// 添加Cookie
	for name, value := range cookies {
		req.Header.Add("Cookie", name+"="+value)
	}
}

// parseAccessTokenFromURI 从URI解析访问令牌
func parseAccessTokenFromURI(uri string) (string, error) {
	if uri == "" {
		return "", errors.New("空URI")
	}

	// 例如: https://playvalorant.com/opt_in#access_token=eyJh...&scope=...
	hashParts := strings.Split(uri, "#")
	if len(hashParts) < 2 {
		return "", errors.New("URI中没有找到#部分")
	}

	params := strings.Split(hashParts[1], "&")
	for _, param := range params {
		if strings.HasPrefix(param, "access_token=") {
			token := strings.TrimPrefix(param, "access_token=")
			return token, nil
		}
	}

	return "", errors.New("未在URI中找到访问令牌")
}

// getEntitlementToken 获取授权令牌
func (v *ValorantAPI) getEntitlementToken(accessToken string) (string, error) {
	req, err := http.NewRequest(http.MethodPost, entitlementsURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+accessToken)

	resp, err := v.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("获取授权令牌失败，状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
	}

	var entitlementResp models.ValorantEntitlementResponse
	if err := json.NewDecoder(resp.Body).Decode(&entitlementResp); err != nil {
		return "", err
	}

	return entitlementResp.EntitlementToken, nil
}

// getUserInfo 获取用户信息
func (v *ValorantAPI) getUserInfo(accessToken string) (*models.ValorantUserInfoResponse, error) {
	req, err := http.NewRequest(http.MethodGet, userInfoURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+accessToken)

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取用户信息失败，状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
	}

	var userInfo models.ValorantUserInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	// 如果Acct字段中有数据，使用它作为最优先的游戏名称和标签
	if userInfo.Acct.GameName != "" {
		userInfo.Name = userInfo.Acct.GameName
		userInfo.Tag = userInfo.Acct.TagLine
	}

	return &userInfo, nil
}

// GetStoreOffers 获取商店物品
func (v *ValorantAPI) GetStoreOffers(userID, accessToken, entitlementToken string) (*models.ValorantStoreResponse, error) {
	// 确保使用有效的区域设置
	if v.region == "" {
		v.region = defaultRegion
		log.Printf("警告：未设置区域，将使用默认区域: %s\n", defaultRegion)
	}

	log.Printf("区域检查: 当前使用的区域是 %s\n", v.region)

	// 构建URL
	url := fmt.Sprintf(storeURL, v.region, userID)

	// 打印基本日志（不包含可能的敏感信息）
	log.Printf("正在请求商店数据，URL: %s\n", url)

	// 创建请求 - 使用POST方法
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader("{}"))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置完整的请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Riot-ClientPlatform", clientPlatform)
	req.Header.Set("X-Riot-ClientVersion", v.clientVersion)
	req.Header.Set("X-Riot-Entitlements-JWT", entitlementToken)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// 不输出请求头详细信息以避免泄露敏感数据
	log.Println("请求头已设置（敏感信息已隐藏）")

	resp, err := v.client.Do(req)
	if err != nil {
		log.Printf("获取商店数据失败: %v\n", err)
		return nil, fmt.Errorf("获取商店物品失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyStr := string(bodyBytes)

		log.Printf("请求失败: 状态码: %d\n", resp.StatusCode)
		log.Printf("响应体: %s\n", bodyStr)

		// 针对特定错误提供更具体的错误信息
		if resp.StatusCode == 404 {
			log.Printf("警告: 获取到404错误，这可能意味着API接口路径已更改或用户区域不正确\n")
		} else if resp.StatusCode == 405 {
			log.Printf("警告: 获取到405错误，这表明请求方法不被允许\n")
		}

		return nil, fmt.Errorf("获取商店数据失败，状态码: %d, 响应: %s", resp.StatusCode, bodyStr)
	}

	// 解析响应
	var storeResp models.ValorantStoreResponse
	if err := json.NewDecoder(resp.Body).Decode(&storeResp); err != nil {
		return nil, fmt.Errorf("解析商店数据失败: %w", err)
	}

	log.Printf("成功获取商店数据\n")
	return &storeResp, nil
}

// GetStoreOffersRaw 获取商店物品的原始JSON响应
func (v *ValorantAPI) GetStoreOffersRaw(userID, accessToken, entitlementToken string) ([]byte, error) {
	// 确保使用有效的区域设置
	if v.region == "" {
		v.region = defaultRegion
		log.Printf("警告：未设置区域，将使用默认区域: %s\n", defaultRegion)
	}

	log.Printf("区域检查: 当前使用的区域是 %s\n", v.region)

	// 构建URL
	url := fmt.Sprintf(storeURL, v.region, userID)

	// 打印基本日志（不包含可能的敏感信息）
	log.Printf("正在请求商店数据，URL: %s\n", url)

	// 创建请求 - 使用POST方法
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader("{}"))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置完整的请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Riot-ClientPlatform", clientPlatform)
	req.Header.Set("X-Riot-ClientVersion", v.clientVersion)
	req.Header.Set("X-Riot-Entitlements-JWT", entitlementToken)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// 不输出请求头详细信息以避免泄露敏感数据
	log.Println("请求头已设置（敏感信息已隐藏）")

	resp, err := v.client.Do(req)
	if err != nil {
		log.Printf("获取商店数据失败: %v\n", err)
		return nil, fmt.Errorf("获取商店物品失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取完整响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	// 检查响应状态
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("请求失败: 状态码: %d\n", resp.StatusCode)
		log.Printf("响应体: %s\n", string(bodyBytes))

		// 针对特定错误提供更具体的错误信息
		if resp.StatusCode == 404 {
			log.Printf("警告: 获取到404错误，这可能意味着API接口路径已更改或用户区域不正确\n")
		} else if resp.StatusCode == 405 {
			log.Printf("警告: 获取到405错误，这表明请求方法不被允许\n")
		}

		return nil, fmt.Errorf("获取商店数据失败，状态码: %d", resp.StatusCode)
	}

	log.Printf("成功获取商店数据\n")
	return bodyBytes, nil
}
