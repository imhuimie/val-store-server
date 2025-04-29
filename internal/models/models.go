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
	UserID           string `json:"user_id"`
	Username         string `json:"username"`
	AccessToken      string `json:"access_token"`
	EntitlementToken string `json:"entitlement_token"`
	Region           string `json:"region"`
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

// ValorantStoreResponse 包含商店中的皮肤信息
type ValorantStoreResponse struct {
	FeaturedBundle       FeaturedBundle       `json:"FeaturedBundle"`
	SkinsPanelLayout     SkinsPanelLayout     `json:"SkinsPanelLayout"`
	BonusStore           BonusStore           `json:"BonusStore,omitempty"`
	AccessoryStore       AccessoryStore       `json:"AccessoryStore"`
	UpgradeCurrencyStore UpgradeCurrencyStore `json:"UpgradeCurrencyStore"`
}

// FeaturedBundle 精选套装信息
type FeaturedBundle struct {
	Bundle                           Bundle   `json:"Bundle"`
	Bundles                          []Bundle `json:"Bundles"`
	BundleRemainingDurationInSeconds int64    `json:"BundleRemainingDurationInSeconds"`
}

// Bundle 套装信息
type Bundle struct {
	ID                         string            `json:"ID"`
	DataAssetID                string            `json:"DataAssetID"`
	CurrencyID                 string            `json:"CurrencyID"`
	Items                      []BundleItem      `json:"Items"`
	ItemOffers                 []BundleItemOffer `json:"ItemOffers,omitempty"`
	TotalBaseCost              map[string]int    `json:"TotalBaseCost"`
	TotalDiscountedCost        map[string]int    `json:"TotalDiscountedCost"`
	TotalDiscountPercent       float64           `json:"TotalDiscountPercent"`
	DurationRemainingInSeconds int64             `json:"DurationRemainingInSeconds"`
	WholesaleOnly              bool              `json:"WholesaleOnly"`
	IsGiftable                 int               `json:"IsGiftable"`
}

// BundleItem 套装中的单个物品信息
type BundleItem struct {
	Item            ItemInfo `json:"Item"`
	BasePrice       int      `json:"BasePrice"`
	DiscountPercent int      `json:"DiscountPercent"`
	DiscountedPrice int      `json:"DiscountedPrice"`
	Quantity        int      `json:"Quantity"`
}

// BundleItemOffer 套装中的单个物品报价
type BundleItemOffer struct {
	BundleItemOfferID string         `json:"BundleItemOfferID"`
	Offer             Offer          `json:"Offer"`
	DiscountPercent   int            `json:"DiscountPercent"`
	DiscountedCost    map[string]int `json:"DiscountedCost"`
}

// ItemInfo 物品的详细信息
type ItemInfo struct {
	ItemTypeID string `json:"ItemTypeID"`
	ItemID     string `json:"ItemID"`
	Amount     int    `json:"Amount"`
}

// Offer 商品报价信息
type Offer struct {
	OfferID          string         `json:"OfferID"`
	IsDirectPurchase bool           `json:"IsDirectPurchase"`
	StartDate        string         `json:"StartDate"`
	Cost             map[string]int `json:"Cost"`
	Rewards          []ItemReward   `json:"Rewards"`
}

// ItemReward 物品奖励信息
type ItemReward struct {
	ItemTypeID string `json:"ItemTypeID"`
	ItemID     string `json:"ItemID"`
	Quantity   int    `json:"Quantity"`
}

// SkinsPanelLayout 每日商店皮肤展示区布局
type SkinsPanelLayout struct {
	SingleItemOffers                           []string `json:"SingleItemOffers"`      // 皮肤ID列表
	SingleItemStoreOffers                      []Offer  `json:"SingleItemStoreOffers"` // 详细的皮肤报价
	SingleItemOffersRemainingDurationInSeconds int64    `json:"SingleItemOffersRemainingDurationInSeconds"`
}

// BonusStore 额外商店/特惠商店（如有夜市）
type BonusStore struct {
	BonusStoreOffers                     []BonusStoreOffer `json:"BonusStoreOffers"`
	BonusStoreRemainingDurationInSeconds int64             `json:"BonusStoreRemainingDurationInSeconds"`
}

// BonusStoreOffer 特惠商店中的物品
type BonusStoreOffer struct {
	BonusOfferID    string         `json:"BonusOfferID"`
	Offer           Offer          `json:"Offer"`
	DiscountPercent int            `json:"DiscountPercent"`
	DiscountCosts   map[string]int `json:"DiscountCosts"`
	IsSeen          bool           `json:"IsSeen"`
}

// AccessoryStore 配件商店
type AccessoryStore struct {
	AccessoryStoreOffers                     []AccessoryStoreOffer `json:"AccessoryStoreOffers"`
	AccessoryStoreRemainingDurationInSeconds int64                 `json:"AccessoryStoreRemainingDurationInSeconds"`
}

// AccessoryStoreOffer 配件商店中的物品
type AccessoryStoreOffer struct {
	Offer      ItemInfo `json:"Offer"`
	ContractID string   `json:"ContractID"`
}

// UpgradeCurrencyStore 升级币商店
type UpgradeCurrencyStore struct {
	UpgradeCurrencyOffers []UpgradeCurrencyOffer `json:"UpgradeCurrencyOffers"`
}

// UpgradeCurrencyOffer 升级币商店中的物品
type UpgradeCurrencyOffer struct {
	OfferID          string   `json:"OfferID"`
	StorefrontItemID string   `json:"StorefrontItemID"`
	Offer            ItemInfo `json:"Offer"`
	Cost             int      `json:"Cost"`
}

// ValorantWalletResponse 用户钱包/余额信息
type ValorantWalletResponse struct {
	Balances map[string]int `json:"Balances"` // 键是货币ID，值是数量
}
