package example

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

func (api *DataStatisticsApi) GetDomainStats(c *gin.Context) {
	from, to := c.Query("from"), c.Query("to")
	if from == "" {
		from = to
	}
	if to == "" {
		to = from
	}
	if from == "" {
		from = time.Now().Format("2006-01-02")
	}
	if to == "" {
		to = from
	}
	domain := c.Query("domain")
	var rows []map[string]interface{}
	var domains []string
	global.GVA_DB.Raw("SELECT domain FROM domain_daily_stats WHERE domain <> '' UNION SELECT register_domain FROM users WHERE register_domain <> '' ORDER BY domain").Scan(&domains)
	query := global.GVA_DB.Table("(SELECT domain, stat_date FROM domain_daily_stats UNION SELECT register_domain AS domain, DATE(created_at) AS stat_date FROM users WHERE register_domain <> '') d").Select("d.domain, d.stat_date date, COALESCE(MAX(s.clicks), 0) clicks, COUNT(u.id) registrations").Joins("LEFT JOIN domain_daily_stats s ON s.domain = d.domain AND s.stat_date = d.stat_date").Joins("LEFT JOIN users u ON u.register_domain = d.domain AND DATE(u.created_at) = d.stat_date").Where("d.stat_date BETWEEN ? AND ?", from, to).Group("d.domain, d.stat_date").Order("d.stat_date DESC, d.domain")
	if domain != "" {
		query = query.Where("s.domain = ?", domain)
	}
	if err := query.Find(&rows).Error; err != nil {
		response.FailWithMessage("获取域名统计失败: "+err.Error(), c)
		return
	}
	for _, row := range rows {
		if v, ok := row["registrations"]; ok {
			row["registrations"] = v
		}
	}
	response.OkWithData(gin.H{"from": from, "to": to, "domain": domain, "domains": domains, "items": rows}, c)
}

type DataStatisticsApi struct{}

func (api *DataStatisticsApi) GetDataStatistics(c *gin.Context) {
	var info exampleReq.DataStatisticsSearch
	if err := c.ShouldBindQuery(&info); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	data, err := dataStatisticsService.GetDataStatistics(c.Request.Context(), info)
	if err != nil {
		global.GVA_LOG.Error("获取数据统计失败", zap.Error(err))
		response.FailWithMessage("获取数据统计失败: "+err.Error(), c)
		return
	}
	response.OkWithData(data, c)
}
