package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/example"
)

func bizModel() error {
	db := global.GVA_DB
	err := db.AutoMigrate(example.Games{}, example.AgentRebateConfigs{}, example.Banners{}, example.GameTypes{}, example.GameProviders{}, example.DailyUserStats{}, example.DailyUserStats{}, example.StatsRetention{}, example.UserWagerLock{}, example.DesktopRewardClaim{}, example.ActivityWagerConfig{}, example.WithdrawFeeConfig{})
	if err != nil {
		return err
	}
	return nil
}
