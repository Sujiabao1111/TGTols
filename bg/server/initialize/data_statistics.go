package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var dataStatisticsAuthorityIDs = []uint{888, 8881, 9528}
var domainStatsAuthorityIDs = []uint{888, 8881, 9528}

func EnsureDataStatisticsResources() {
	if global.GVA_DB == nil {
		return
	}
	if err := ensureDataStatisticsResources(global.GVA_DB); err != nil {
		global.GVA_LOG.Error("ensure data statistics resources failed", zap.Error(err))
	}
}

func EnsureDomainStatsResources() {
	if global.GVA_DB == nil { return }
	_ = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		parentID, err := ensureRewardDataParentMenu(tx); if err != nil { return err }
		menuID, err := ensureRewardMenu(tx, "domainStats", "domainStats", parentID, "view/example/domainStats/index.vue", 2, "域名统计", "data-analysis"); if err != nil { return err }
		if err = tx.Model(&system.SysBaseMenu{}).Where("id = ?", menuID).Updates(map[string]interface{}{"parent_id": parentID, "sort": 2, "component": "view/example/domainStats/index.vue", "title": "域名统计", "icon": "data-analysis", "hidden": false}).Error; err != nil { return err }
		if err = ensureRewardAPI(tx, "/admin/domain-stats", "GET", "获取域名统计"); err != nil { return err }
		for _, authorityID := range domainStatsAuthorityIDs {
			if err = ensureRewardAuthorityMenu(tx, parentID, authorityID); err != nil { return err }
			if err = ensureRewardAuthorityMenu(tx, menuID, authorityID); err != nil { return err }
			if err = ensureRewardCasbin(tx, authorityID, "/admin/domain-stats", "GET"); err != nil { return err }
		}
		return nil
	})
}

func ensureDataStatisticsResources(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		parentID, err := ensureRewardDataParentMenu(tx)
		if err != nil {
			return err
		}
		menuID, err := ensureRewardMenu(tx, "dataStatistics", "dataStatistics", parentID, "view/example/dataStatistics/index.vue", 1, "数据统计", "data-analysis")
		if err != nil {
			return err
		}
		if err = tx.Model(&system.SysBaseMenu{}).Where("id = ?", menuID).Updates(map[string]interface{}{"parent_id": parentID, "sort": 1, "component": "view/example/dataStatistics/index.vue", "title": "数据统计", "icon": "data-analysis", "hidden": false}).Error; err != nil {
			return err
		}
		if err = ensureRewardAPI(tx, "/dataStatistics/getDataStatistics", "GET", "获取数据统计"); err != nil {
			return err
		}
		for _, authorityID := range dataStatisticsAuthorityIDs {
			if err = ensureRewardAuthorityMenu(tx, parentID, authorityID); err != nil {
				return err
			}
			if err = ensureRewardAuthorityMenu(tx, menuID, authorityID); err != nil {
				return err
			}
			if err = ensureRewardCasbin(tx, authorityID, "/dataStatistics/getDataStatistics", "GET"); err != nil {
				return err
			}
		}
		return nil
	})
}
