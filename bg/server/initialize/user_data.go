package initialize

import (
	"errors"
	"strconv"

	adapter "github.com/casbin/gorm-adapter/v3"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var userDataAuthorityIDs = []uint{888, 8881, 9528}

func EnsureUserDataResources() {
	if global.GVA_DB == nil {
		return
	}

	if err := ensureUserDataResources(global.GVA_DB); err != nil {
		global.GVA_LOG.Error("ensure user data resources failed", zap.Error(err))
	}
}

func ensureUserDataResources(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		parentID, err := ensureRewardDataParentMenu(tx)
		if err != nil {
			return err
		}

		menuID, err := ensureRewardMenu(
			tx,
			"userData",
			"userData",
			parentID,
			"view/example/userData/index.vue",
			7,
			"用户列表",
			"user-filled",
		)
		if err != nil {
			return err
		}

		if err := ensureUserDataAPI(tx); err != nil {
			return err
		}

		for _, authorityID := range userDataAuthorityIDs {
			if err := ensureRewardAuthorityMenu(tx, parentID, authorityID); err != nil {
				return err
			}
			if err := ensureRewardAuthorityMenu(tx, menuID, authorityID); err != nil {
				return err
			}
			if err := ensureUserDataCasbin(tx, authorityID); err != nil {
				return err
			}
		}

		return nil
	})
}

func ensureUserDataAPI(tx *gorm.DB) error {
	var api system.SysApi
	err := tx.Where("path = ? AND method = ?", "/userData/getUserDataList", "GET").Take(&api).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return tx.Create(&system.SysApi{
			Path:        "/userData/getUserDataList",
			Description: "获取用户列表",
			ApiGroup:    "用户列表",
			Method:      "GET",
		}).Error
	}
	return err
}

func ensureUserDataCasbin(tx *gorm.DB, authorityID uint) error {
	var rule adapter.CasbinRule
	authority := strconv.FormatUint(uint64(authorityID), 10)
	err := tx.Where(&adapter.CasbinRule{
		Ptype: "p",
		V0:    authority,
		V1:    "/userData/getUserDataList",
		V2:    "GET",
	}).Take(&rule).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return tx.Create(&adapter.CasbinRule{
			Ptype: "p",
			V0:    authority,
			V1:    "/userData/getUserDataList",
			V2:    "GET",
		}).Error
	}
	return err
}
