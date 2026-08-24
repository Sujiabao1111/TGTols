package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var rechargeQueryAuthorityIDs = []uint{888, 8881, 9528}

func EnsureRechargeQueryResources() {
	if global.GVA_DB == nil {
		return
	}

	if err := ensureRechargeQueryResources(global.GVA_DB); err != nil {
		global.GVA_LOG.Error("ensure recharge query resources failed", zap.Error(err))
	}
}

func ensureRechargeQueryResources(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		dataParentID, err := ensureRewardDataParentMenu(tx)
		if err != nil {
			return err
		}

		menuID, err := ensureRewardMenu(
			tx,
			"rechargeQuery",
			"rechargeQuery",
			dataParentID,
			"view/example/rechargeQuery/index.vue",
			8,
			"充值查询",
			"money",
		)
		if err != nil {
			return err
		}

		if err := ensureRewardAPI(tx, "/rechargeQuery/getRechargeQueryList", "GET", "获取充值查询列表"); err != nil {
			return err
		}

		for _, authorityID := range rechargeQueryAuthorityIDs {
			if err := ensureRewardAuthorityMenu(tx, dataParentID, authorityID); err != nil {
				return err
			}
			if err := ensureRewardAuthorityMenu(tx, menuID, authorityID); err != nil {
				return err
			}
			if err := ensureRewardCasbin(tx, authorityID, "/rechargeQuery/getRechargeQueryList", "GET"); err != nil {
				return err
			}
		}

		return nil
	})
}
