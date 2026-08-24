package example

import (
	"context"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/example"
	exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
)

type GameProvidersService struct{}

// CreateGameProviders 创建gameProviders表记录
// Author [yourname](https://github.com/yourname)
func (gameProvidersService *GameProvidersService) CreateGameProviders(ctx context.Context, gameProviders *example.GameProviders) (err error) {
	err = global.GVA_DB.Create(gameProviders).Error
	return err
}

// DeleteGameProviders 删除gameProviders表记录
// Author [yourname](https://github.com/yourname)
func (gameProvidersService *GameProvidersService) DeleteGameProviders(ctx context.Context, id string) (err error) {
	err = global.GVA_DB.Delete(&example.GameProviders{}, "id = ?", id).Error
	return err
}

// DeleteGameProvidersByIds 批量删除gameProviders表记录
// Author [yourname](https://github.com/yourname)
func (gameProvidersService *GameProvidersService) DeleteGameProvidersByIds(ctx context.Context, ids []string) (err error) {
	err = global.GVA_DB.Delete(&[]example.GameProviders{}, "id in ?", ids).Error
	return err
}

// UpdateGameProviders 更新gameProviders表记录
// Author [yourname](https://github.com/yourname)
func (gameProvidersService *GameProvidersService) UpdateGameProviders(ctx context.Context, gameProviders example.GameProviders) (err error) {
	err = global.GVA_DB.Model(&example.GameProviders{}).Where("id = ?", gameProviders.Id).Updates(&gameProviders).Error
	return err
}

// GetGameProviders 根据id获取gameProviders表记录
// Author [yourname](https://github.com/yourname)
func (gameProvidersService *GameProvidersService) GetGameProviders(ctx context.Context, id string) (gameProviders example.GameProviders, err error) {
	err = global.GVA_DB.Where("id = ?", id).First(&gameProviders).Error
	return
}

// GetGameProvidersInfoList 分页获取gameProviders表记录
// Author [yourname](https://github.com/yourname)
func (gameProvidersService *GameProvidersService) GetGameProvidersInfoList(ctx context.Context, info exampleReq.GameProvidersSearch) (list []example.GameProviders, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	// 创建db
	db := global.GVA_DB.Model(&example.GameProviders{})
	var gameProviderss []example.GameProviders
	// 如果有条件搜索 下方会自动创建搜索语句
	if info.PlatformCode != nil && *info.PlatformCode != "" {
		db = db.Where("platform_code = ?", *info.PlatformCode)
	}
	err = db.Count(&total).Error
	if err != nil {
		return
	}

	if limit != 0 {
		db = db.Limit(limit).Offset(offset)
	}

	err = db.Find(&gameProviderss).Error
	return gameProviderss, total, err
}
func (gameProvidersService *GameProvidersService) GetGameProvidersPublic(ctx context.Context) {
	// 此方法为获取数据源定义的数据
	// 请自行实现
}
