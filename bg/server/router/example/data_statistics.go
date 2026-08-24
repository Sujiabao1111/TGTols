package example

import "github.com/gin-gonic/gin"

type DataStatisticsRouter struct{}

func (s *DataStatisticsRouter) InitDataStatisticsRouter(router *gin.RouterGroup, publicRouter *gin.RouterGroup) {
	router.Group("dataStatistics").GET("getDataStatistics", dataStatisticsApi.GetDataStatistics)
	router.GET("/admin/domain-stats", dataStatisticsApi.GetDomainStats)
}
