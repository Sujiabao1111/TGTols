package example

import (
	"context"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/example"
	exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
)

type GamesService struct{}

// CreateGames 创建games表记录
// Author [yourname](https://github.com/yourname)
func (gamesService *GamesService) CreateGames(ctx context.Context, games *example.Games) (err error) {
	err = global.GVA_DB.Create(games).Error
	return err
}

// DeleteGames 删除games表记录
// Author [yourname](https://github.com/yourname)
func (gamesService *GamesService) DeleteGames(ctx context.Context, id string) (err error) {
	err = global.GVA_DB.Delete(&example.Games{}, "id = ?", id).Error
	return err
}

// DeleteGamesByIds 批量删除games表记录
// Author [yourname](https://github.com/yourname)
func (gamesService *GamesService) DeleteGamesByIds(ctx context.Context, ids []string) (err error) {
	err = global.GVA_DB.Delete(&[]example.Games{}, "id in ?", ids).Error
	return err
}

// UpdateGames 更新games表记录
// Author [yourname](https://github.com/yourname)
func (gamesService *GamesService) UpdateGames(ctx context.Context, games example.Games) (err error) {
	err = global.GVA_DB.Model(&example.Games{}).Where("id = ?", games.Id).Updates(&games).Error
	return err
}

// GetGames 根据id获取games表记录
// Author [yourname](https://github.com/yourname)
func (gamesService *GamesService) GetGames(ctx context.Context, id string) (games example.Games, err error) {
	err = global.GVA_DB.Where("id = ?", id).First(&games).Error
	return
}

// GetGamesInfoList 分页获取games表记录
// Author [yourname](https://github.com/yourname)
func (gamesService *GamesService) GetGamesInfoList(ctx context.Context, info exampleReq.GamesSearch) (list []example.Games, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	// 创建db
	db := global.GVA_DB.Model(&example.Games{})
	var gamess []example.Games
	// 如果有条件搜索 下方会自动创建搜索语句
	if info.PlatformCode != nil && *info.PlatformCode != "" {
		db = db.Where("platform_code = ?", *info.PlatformCode)
	}
	if info.ProviderId != nil {
		db = db.Where("provider_id = ?", *info.ProviderId)
	}
	if info.GameTypeId != nil {
		db = db.Where("game_type_id = ?", *info.GameTypeId)
	}
	err = db.Count(&total).Error
	if err != nil {
		return
	}

	if limit != 0 {
		db = db.Limit(limit).Offset(offset)
	}

	err = db.Find(&gamess).Error
	return gamess, total, err
}
func (gamesService *GamesService) GetGamesPublic(ctx context.Context) {
	// 此方法为获取数据源定义的数据
	// 请自行实现
}
