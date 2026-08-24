
package example

import (
	"context"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/example"
    exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
)

type GameTypesService struct {}
// CreateGameTypes 创建gameTypes表记录
// Author [yourname](https://github.com/yourname)
func (gameTypesService *GameTypesService) CreateGameTypes(ctx context.Context, gameTypes *example.GameTypes) (err error) {
	err = global.GVA_DB.Create(gameTypes).Error
	return err
}

// DeleteGameTypes 删除gameTypes表记录
// Author [yourname](https://github.com/yourname)
func (gameTypesService *GameTypesService)DeleteGameTypes(ctx context.Context, id string) (err error) {
	err = global.GVA_DB.Delete(&example.GameTypes{},"id = ?",id).Error
	return err
}

// DeleteGameTypesByIds 批量删除gameTypes表记录
// Author [yourname](https://github.com/yourname)
func (gameTypesService *GameTypesService)DeleteGameTypesByIds(ctx context.Context, ids []string) (err error) {
	err = global.GVA_DB.Delete(&[]example.GameTypes{},"id in ?",ids).Error
	return err
}

// UpdateGameTypes 更新gameTypes表记录
// Author [yourname](https://github.com/yourname)
func (gameTypesService *GameTypesService)UpdateGameTypes(ctx context.Context, gameTypes example.GameTypes) (err error) {
	err = global.GVA_DB.Model(&example.GameTypes{}).Where("id = ?",gameTypes.Id).Updates(&gameTypes).Error
	return err
}

// GetGameTypes 根据id获取gameTypes表记录
// Author [yourname](https://github.com/yourname)
func (gameTypesService *GameTypesService)GetGameTypes(ctx context.Context, id string) (gameTypes example.GameTypes, err error) {
	err = global.GVA_DB.Where("id = ?", id).First(&gameTypes).Error
	return
}
// GetGameTypesInfoList 分页获取gameTypes表记录
// Author [yourname](https://github.com/yourname)
func (gameTypesService *GameTypesService)GetGameTypesInfoList(ctx context.Context, info exampleReq.GameTypesSearch) (list []example.GameTypes, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
    // 创建db
	db := global.GVA_DB.Model(&example.GameTypes{})
    var gameTypess []example.GameTypes
    // 如果有条件搜索 下方会自动创建搜索语句
    
	err = db.Count(&total).Error
	if err!=nil {
    	return
    }

	if limit != 0 {
       db = db.Limit(limit).Offset(offset)
    }

	err = db.Find(&gameTypess).Error
	return  gameTypess, total, err
}
func (gameTypesService *GameTypesService)GetGameTypesPublic(ctx context.Context) {
    // 此方法为获取数据源定义的数据
    // 请自行实现
}
