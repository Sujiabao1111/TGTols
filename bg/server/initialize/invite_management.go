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

var inviteManagementAuthorityIDs = []uint{888, 8881, 9528}

func EnsureInviteManagementResources() {
	if global.GVA_DB == nil {
		return
	}

	if err := ensureInviteManagementResources(global.GVA_DB); err != nil {
		global.GVA_LOG.Error("ensure invite management resources failed", zap.Error(err))
	}
}

func ensureInviteManagementResources(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		parentID, err := ensureInviteManagementParentMenu(tx)
		if err != nil {
			return err
		}

		menuID, err := ensureInviteManagementMenu(tx, parentID)
		if err != nil {
			return err
		}

		if err := ensureInviteManagementAPI(tx); err != nil {
			return err
		}

		for _, authorityID := range inviteManagementAuthorityIDs {
			if err := ensureInviteManagementAuthorityMenu(tx, parentID, authorityID); err != nil {
				return err
			}
			if err := ensureInviteManagementAuthorityMenu(tx, menuID, authorityID); err != nil {
				return err
			}
			if err := ensureInviteManagementCasbin(tx, authorityID); err != nil {
				return err
			}
		}

		return nil
	})
}

func ensureInviteManagementParentMenu(tx *gorm.DB) (uint, error) {
	var menu system.SysBaseMenu
	err := tx.Where("path = ?", "datas").Take(&menu).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		menu = system.SysBaseMenu{
			MenuLevel: 0,
			ParentId:  0,
			Path:      "datas",
			Name:      "datas",
			Hidden:    false,
			Component: "view/routerHolder.vue",
			Sort:      2,
			Meta: system.Meta{
				Title: "\u6570\u636e",
				Icon:  "data-board",
			},
		}
		if err := tx.Create(&menu).Error; err != nil {
			return 0, err
		}
		return menu.ID, nil
	}
	if err != nil {
		return 0, err
	}
	return menu.ID, nil
}

func ensureInviteManagementMenu(tx *gorm.DB, parentID uint) (uint, error) {
	var menu system.SysBaseMenu
	err := tx.Where("path = ?", "inviteManagement").Take(&menu).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		menu = system.SysBaseMenu{
			MenuLevel: 1,
			ParentId:  parentID,
			Path:      "inviteManagement",
			Name:      "inviteManagement",
			Hidden:    false,
			Component: "view/example/inviteManagement/inviteManagement.vue",
			Sort:      3,
			Meta: system.Meta{
				Title: "\u9080\u8bf7\u7ba1\u7406",
				Icon:  "promotion",
			},
		}
		if err := tx.Create(&menu).Error; err != nil {
			return 0, err
		}
		return menu.ID, nil
	}
	if err != nil {
		return 0, err
	}
	return menu.ID, nil
}

func ensureInviteManagementAPI(tx *gorm.DB) error {
	var api system.SysApi
	err := tx.Where("path = ? AND method = ?", "/inviteStats/getInviteStatsList", "GET").Take(&api).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return tx.Create(&system.SysApi{
			Path:        "/inviteStats/getInviteStatsList",
			Description: "\u83b7\u53d6\u9080\u8bf7\u7edf\u8ba1\u5217\u8868",
			ApiGroup:    "\u9080\u8bf7\u7ba1\u7406",
			Method:      "GET",
		}).Error
	}
	return err
}

func ensureInviteManagementAuthorityMenu(tx *gorm.DB, menuID uint, authorityID uint) error {
	var record system.SysAuthorityMenu
	err := tx.Where("sys_base_menu_id = ? AND sys_authority_authority_id = ?", menuID, authorityID).Take(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return tx.Table("sys_authority_menus").Create(map[string]interface{}{
			"sys_base_menu_id":           menuID,
			"sys_authority_authority_id": authorityID,
		}).Error
	}
	return err
}

func ensureInviteManagementCasbin(tx *gorm.DB, authorityID uint) error {
	var rule adapter.CasbinRule
	authority := strconv.FormatUint(uint64(authorityID), 10)
	err := tx.Where(&adapter.CasbinRule{
		Ptype: "p",
		V0:    authority,
		V1:    "/inviteStats/getInviteStatsList",
		V2:    "GET",
	}).Take(&rule).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return tx.Create(&adapter.CasbinRule{
			Ptype: "p",
			V0:    authority,
			V1:    "/inviteStats/getInviteStatsList",
			V2:    "GET",
		}).Error
	}
	return err
}
