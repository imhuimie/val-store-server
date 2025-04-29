package services

import (
	"github.com/emper0r/val-store-server/internal/models"
	"github.com/emper0r/val-store-server/internal/repositories"
)

// StoreService 处理商店数据相关的业务逻辑
type StoreService struct {
	valorantAPI *repositories.ValorantAPI
}

// NewStoreService 创建一个新的商店服务
func NewStoreService(valorantAPI *repositories.ValorantAPI) *StoreService {
	return &StoreService{
		valorantAPI: valorantAPI,
	}
}

// GetStoreData 获取完整的商店数据
func (s *StoreService) GetStoreData(userID, accessToken, entitlementToken, region string) (*models.ValorantStoreResponse, error) {
	// 设置区域
	if region != "" {
		s.valorantAPI.SetRegion(region)
	}

	// 获取商店数据
	return s.valorantAPI.GetStoreOffers(userID, accessToken, entitlementToken)
}

// GetStoreDataRaw 获取原始JSON格式的商店数据
func (s *StoreService) GetStoreDataRaw(userID, accessToken, entitlementToken, region string) ([]byte, error) {
	// 设置区域
	if region != "" {
		s.valorantAPI.SetRegion(region)
	}

	// 获取原始商店数据
	return s.valorantAPI.GetStoreOffersRaw(userID, accessToken, entitlementToken)
}
