package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var gameLaunchErrorAuthorityIDs = []uint{888, 8881, 9528}

func EnsureGameLaunchErrorResources() {
	if global.GVA_DB == nil {
		return
	}

	if err := ensureGameLaunchErrorResources(global.GVA_DB); err != nil {
		global.GVA_LOG.Error("ensure game launch error resources failed", zap.Error(err))
	}
}

func ensureGameLaunchErrorResources(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		dataParentID, err := ensureRewardDataParentMenu(tx)
		if err != nil {
			return err
		}

		menuID, err := ensureRewardMenu(
			tx,
			"gameLaunchError",
			"gameLaunchError",
			dataParentID,
			"view/example/gameLaunchError/index.vue",
			8,
			"玩家报错信息",
			"warning",
		)
		if err != nil {
			return err
		}

		if err := ensureRewardAPI(tx, "/gameLaunchError/getGameLaunchErrorList", "GET", "获取玩家报错信息"); err != nil {
			return err
		}

		for _, authorityID := range gameLaunchErrorAuthorityIDs {
			if err := ensureRewardAuthorityMenu(tx, dataParentID, authorityID); err != nil {
				return err
			}
			if err := ensureRewardAuthorityMenu(tx, menuID, authorityID); err != nil {
				return err
			}
			if err := ensureRewardCasbin(tx, authorityID, "/gameLaunchError/getGameLaunchErrorList", "GET"); err != nil {
				return err
			}
		}

		return nil
	})
}
