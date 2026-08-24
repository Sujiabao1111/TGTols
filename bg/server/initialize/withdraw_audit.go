package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var withdrawAuditAuthorityIDs = []uint{888, 8881, 9528}

func EnsureWithdrawAuditResources() {
	if global.GVA_DB == nil {
		return
	}

	if err := ensureWithdrawAuditResources(global.GVA_DB); err != nil {
		global.GVA_LOG.Error("ensure withdraw audit resources failed", zap.Error(err))
	}
}

func ensureWithdrawAuditResources(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		dataParentID, err := ensureRewardDataParentMenu(tx)
		if err != nil {
			return err
		}

		menuID, err := ensureRewardMenu(
			tx,
			"withdrawAudit",
			"withdrawAudit",
			dataParentID,
			"view/example/withdrawAudit/index.vue",
			6,
			"提现审核",
			"wallet-filled",
		)
		if err != nil {
			return err
		}

		if err := ensureRewardAPI(tx, "/withdrawAudit/getWithdrawAuditList", "GET", "获取提现审核列表"); err != nil {
			return err
		}
		if err := ensureRewardAPI(tx, "/withdrawAudit/approve", "POST", "提现审核通过"); err != nil {
			return err
		}
		if err := ensureRewardAPI(tx, "/withdrawAudit/reject", "POST", "提现审核驳回"); err != nil {
			return err
		}

		for _, authorityID := range withdrawAuditAuthorityIDs {
			if err := ensureRewardAuthorityMenu(tx, dataParentID, authorityID); err != nil {
				return err
			}
			if err := ensureRewardAuthorityMenu(tx, menuID, authorityID); err != nil {
				return err
			}
			if err := ensureRewardCasbin(tx, authorityID, "/withdrawAudit/getWithdrawAuditList", "GET"); err != nil {
				return err
			}
			if err := ensureRewardCasbin(tx, authorityID, "/withdrawAudit/approve", "POST"); err != nil {
				return err
			}
			if err := ensureRewardCasbin(tx, authorityID, "/withdrawAudit/reject", "POST"); err != nil {
				return err
			}
		}

		return nil
	})
}
