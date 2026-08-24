package example

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type BannersRouter struct {}

// InitBannersRouter 初始化 banners表 路由信息
func (s *BannersRouter) InitBannersRouter(Router *gin.RouterGroup,PublicRouter *gin.RouterGroup) {
	bannersRouter := Router.Group("banners").Use(middleware.OperationRecord())
	bannersRouterWithoutRecord := Router.Group("banners")
	bannersRouterWithoutAuth := PublicRouter.Group("banners")
	{
		bannersRouter.POST("createBanners", bannersApi.CreateBanners)   // 新建banners表
		bannersRouter.DELETE("deleteBanners", bannersApi.DeleteBanners) // 删除banners表
		bannersRouter.DELETE("deleteBannersByIds", bannersApi.DeleteBannersByIds) // 批量删除banners表
		bannersRouter.PUT("updateBanners", bannersApi.UpdateBanners)    // 更新banners表
	}
	{
		bannersRouterWithoutRecord.GET("findBanners", bannersApi.FindBanners)        // 根据ID获取banners表
		bannersRouterWithoutRecord.GET("getBannersList", bannersApi.GetBannersList)  // 获取banners表列表
	}
	{
	    bannersRouterWithoutAuth.GET("getBannersPublic", bannersApi.GetBannersPublic)  // banners表开放接口
	}
}
