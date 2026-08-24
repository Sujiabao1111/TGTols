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

var rewardManagementAuthorityIDs = []uint{888, 8881, 9528}

func EnsureRewardManagementResources() {
	if global.GVA_DB == nil {
		return
	}

	if err := ensureRewardManagementResources(global.GVA_DB); err != nil {
		global.GVA_LOG.Error("ensure reward management resources failed", zap.Error(err))
	}
}

func ensureRewardManagementResources(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		dataParentID, err := ensureRewardDataParentMenu(tx)
		if err != nil {
			return err
		}

		configParentID, err := ensureRewardConfigParentMenu(tx)
		if err != nil {
			return err
		}

		desktopMenuID, err := ensureRewardMenu(
			tx,
			"desktopReward",
			"desktopReward",
			dataParentID,
			"view/example/desktopReward/index.vue",
			4,
			"\u684c\u9762\u5956\u52b1",
			"present",
		)
		if err != nil {
			return err
		}

		manualMenuID, err := ensureRewardMenu(
			tx,
			"manualReward",
			"manualReward",
			dataParentID,
			"view/example/manualReward/index.vue",
			5,
			"\u53d1\u653e\u5956\u52b1",
			"gift",
		)
		if err != nil {
			return err
		}

		whatsAppMenuID, err := ensureRewardMenu(
			tx,
			"whatsappConfig",
			"whatsappConfig",
			configParentID,
			"view/example/whatsappConfig/index.vue",
			1,
			"\u5ba2\u670d\u94fe\u63a5",
			"service",
		)
		if err != nil {
			return err
		}

		activityWagerConfigMenuID, err := ensureRewardMenu(
			tx,
			"activityWagerConfig",
			"activityWagerConfig",
			configParentID,
			"view/example/activityWagerConfig/index.vue",
			2,
			"\u6253\u7801\u914d\u7f6e",
			"setting",
		)
		if err != nil {
			return err
		}

		withdrawFeeConfigMenuID, err := ensureRewardMenu(
			tx,
			"withdrawFeeConfig",
			"withdrawFeeConfig",
			configParentID,
			"view/example/withdrawFeeConfig/index.vue",
			3,
			"\u63d0\u73b0\u624b\u7eed\u8d39",
			"money",
		)
		if err != nil {
			return err
		}

		if err := ensureRewardAPI(tx, "/manualReward/grantDesktopReward", "POST", "\u684c\u9762\u5956\u52b1"); err != nil {
			return err
		}
		if err := ensureRewardAPI(tx, "/manualReward/grantReward", "POST", "\u53d1\u653e\u5956\u52b1"); err != nil {
			return err
		}
		if err := ensureRewardAPI(tx, "/activityWagerConfig/getActivityWagerConfig", "GET", "\u83b7\u53d6\u6253\u7801\u914d\u7f6e"); err != nil {
			return err
		}
		if err := ensureRewardAPI(tx, "/activityWagerConfig/updateActivityWagerConfig", "PUT", "\u66f4\u65b0\u6253\u7801\u914d\u7f6e"); err != nil {
			return err
		}
		if err := ensureRewardAPI(tx, "/withdrawFeeConfig/getWithdrawFeeConfig", "GET", "\u83b7\u53d6\u63d0\u73b0\u624b\u7eed\u8d39\u914d\u7f6e"); err != nil {
			return err
		}
		if err := ensureRewardAPI(tx, "/withdrawFeeConfig/updateWithdrawFeeConfig", "PUT", "\u66f4\u65b0\u63d0\u73b0\u624b\u7eed\u8d39\u914d\u7f6e"); err != nil {
			return err
		}

		for _, authorityID := range rewardManagementAuthorityIDs {
			for _, menuID := range []uint{dataParentID, desktopMenuID, manualMenuID, configParentID, whatsAppMenuID, activityWagerConfigMenuID, withdrawFeeConfigMenuID} {
				if err := ensureRewardAuthorityMenu(tx, menuID, authorityID); err != nil {
					return err
				}
			}
			if err := ensureRewardCasbin(tx, authorityID, "/manualReward/grantDesktopReward", "POST"); err != nil {
				return err
			}
			if err := ensureRewardCasbin(tx, authorityID, "/manualReward/grantReward", "POST"); err != nil {
				return err
			}
			if err := ensureRewardCasbin(tx, authorityID, "/activityWagerConfig/getActivityWagerConfig", "GET"); err != nil {
				return err
			}
			if err := ensureRewardCasbin(tx, authorityID, "/activityWagerConfig/updateActivityWagerConfig", "PUT"); err != nil {
				return err
			}
			if err := ensureRewardCasbin(tx, authorityID, "/withdrawFeeConfig/getWithdrawFeeConfig", "GET"); err != nil {
				return err
			}
			if err := ensureRewardCasbin(tx, authorityID, "/withdrawFeeConfig/updateWithdrawFeeConfig", "PUT"); err != nil {
				return err
			}
		}

		return nil
	})
}

func ensureRewardDataParentMenu(tx *gorm.DB) (uint, error) {
	return ensureRewardMenu(
		tx,
		"datas",
		"datas",
		0,
		"view/routerHolder.vue",
		3,
		"\u6570\u636e",
		"data-board",
	)
}

func ensureRewardConfigParentMenu(tx *gorm.DB) (uint, error) {
	return ensureRewardMenu(
		tx,
		"cfgs",
		"cfgs",
		0,
		"view/routerHolder.vue",
		2,
		"\u914d\u7f6e\u9879",
		"bowl",
	)
}

func ensureRewardMenu(tx *gorm.DB, path string, name string, parentID uint, component string, sort int, title string, icon string) (uint, error) {
	var menu system.SysBaseMenu
	err := tx.Where("path = ?", path).Take(&menu).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		menu = system.SysBaseMenu{
			MenuLevel: 0,
			ParentId:  parentID,
			Path:      path,
			Name:      name,
			Hidden:    false,
			Component: component,
			Sort:      sort,
			Meta: system.Meta{
				Title: title,
				Icon:  icon,
			},
		}
		if parentID != 0 {
			menu.MenuLevel = 1
		}
		if err := tx.Create(&menu).Error; err != nil {
			return 0, err
		}
		return menu.ID, nil
	}
	if err != nil {
		return 0, err
	}
	menuLevel := uint(0)
	if parentID != 0 {
		menuLevel = 1
	}
	updates := map[string]interface{}{
		"parent_id":  parentID,
		"name":       name,
		"component":  component,
		"sort":       sort,
		"title":      title,
		"icon":       icon,
		"menu_level": menuLevel,
	}
	if err := tx.Model(&menu).Updates(updates).Error; err != nil {
		return 0, err
	}
	menu.ParentId = parentID
	menu.Name = name
	menu.Component = component
	menu.Sort = sort
	menu.MenuLevel = menuLevel
	menu.Meta.Title = title
	menu.Meta.Icon = icon
	return menu.ID, nil
}

func ensureRewardAPI(tx *gorm.DB, path string, method string, description string) error {
	var api system.SysApi
	err := tx.Where("path = ? AND method = ?", path, method).Take(&api).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return tx.Create(&system.SysApi{
			Path:        path,
			Description: description,
			ApiGroup:    "\u5956\u52b1\u7ba1\u7406",
			Method:      method,
		}).Error
	}
	return err
}

func ensureRewardAuthorityMenu(tx *gorm.DB, menuID uint, authorityID uint) error {
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

func ensureRewardCasbin(tx *gorm.DB, authorityID uint, path string, method string) error {
	var rule adapter.CasbinRule
	authority := strconv.FormatUint(uint64(authorityID), 10)
	err := tx.Where(&adapter.CasbinRule{
		Ptype: "p",
		V0:    authority,
		V1:    path,
		V2:    method,
	}).Take(&rule).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return tx.Create(&adapter.CasbinRule{
			Ptype: "p",
			V0:    authority,
			V1:    path,
			V2:    method,
		}).Error
	}
	return err
}
