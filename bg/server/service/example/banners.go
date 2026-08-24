
package example

import (
	"context"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/example"
    exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
)

type BannersService struct {}
// CreateBanners 创建banners表记录
// Author [yourname](https://github.com/yourname)
func (bannersService *BannersService) CreateBanners(ctx context.Context, banners *example.Banners) (err error) {
	err = global.GVA_DB.Create(banners).Error
	return err
}

// DeleteBanners 删除banners表记录
// Author [yourname](https://github.com/yourname)
func (bannersService *BannersService)DeleteBanners(ctx context.Context, id string) (err error) {
	err = global.GVA_DB.Delete(&example.Banners{},"id = ?",id).Error
	return err
}

// DeleteBannersByIds 批量删除banners表记录
// Author [yourname](https://github.com/yourname)
func (bannersService *BannersService)DeleteBannersByIds(ctx context.Context, ids []string) (err error) {
	err = global.GVA_DB.Delete(&[]example.Banners{},"id in ?",ids).Error
	return err
}

// UpdateBanners 更新banners表记录
// Author [yourname](https://github.com/yourname)
func (bannersService *BannersService)UpdateBanners(ctx context.Context, banners example.Banners) (err error) {
	err = global.GVA_DB.Model(&example.Banners{}).Where("id = ?",banners.Id).Updates(&banners).Error
	return err
}

// GetBanners 根据id获取banners表记录
// Author [yourname](https://github.com/yourname)
func (bannersService *BannersService)GetBanners(ctx context.Context, id string) (banners example.Banners, err error) {
	err = global.GVA_DB.Where("id = ?", id).First(&banners).Error
	return
}
// GetBannersInfoList 分页获取banners表记录
// Author [yourname](https://github.com/yourname)
func (bannersService *BannersService)GetBannersInfoList(ctx context.Context, info exampleReq.BannersSearch) (list []example.Banners, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
    // 创建db
	db := global.GVA_DB.Model(&example.Banners{})
    var bannerss []example.Banners
    // 如果有条件搜索 下方会自动创建搜索语句
    
	err = db.Count(&total).Error
	if err!=nil {
    	return
    }

	if limit != 0 {
       db = db.Limit(limit).Offset(offset)
    }

	err = db.Find(&bannerss).Error
	return  bannerss, total, err
}
func (bannersService *BannersService)GetBannersPublic(ctx context.Context) {
    // 此方法为获取数据源定义的数据
    // 请自行实现
}
