package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var userGameRecordAuthorityIDs = []uint{888, 8881, 9528}

func EnsureUserGameRecordResources() {
	if global.GVA_DB == nil {
		return
	}

	if err := ensureUserGameRecordResources(global.GVA_DB); err != nil {
		global.GVA_LOG.Error("ensure user game record resources failed", zap.Error(err))
	}
}

func ensureUserGameRecordResources(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		dataParentID, err := ensureRewardDataParentMenu(tx)
		if err != nil {
			return err
		}

		menuID, err := ensureRewardMenu(
			tx,
			"userGameRecord",
			"userGameRecord",
			dataParentID,
			"view/example/userGameRecord/index.vue",
			9,
			"\u7528\u6237\u6e38\u620f\u8bb0\u5f55",
			"tickets",
		)
		if err != nil {
			return err
		}

		if err := ensureRewardAPI(tx, "/userGameRecord/getUserGameRecordList", "GET", "\u83b7\u53d6\u7528\u6237\u6e38\u620f\u8bb0\u5f55"); err != nil {
			return err
		}

		for _, authorityID := range userGameRecordAuthorityIDs {
			if err := ensureRewardAuthorityMenu(tx, dataParentID, authorityID); err != nil {
				return err
			}
			if err := ensureRewardAuthorityMenu(tx, menuID, authorityID); err != nil {
				return err
			}
			if err := ensureRewardCasbin(tx, authorityID, "/userGameRecord/getUserGameRecordList", "GET"); err != nil {
				return err
			}
		}

		return nil
	})
}
