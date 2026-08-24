package controllers

import (
	"log"

	"github.com/robfig/cron/v3"
)

// 日常定时任务
func DailyGrabDatas() {

}

func SetupAndGo() *cron.Cron {
	c := cron.New(cron.WithSeconds())
	// c := cron.New()
	// no seconds
	// c.AddFunc("* * * * *", ClearSignQuest)
	// Funcs are invoked in their own goroutine, asynchronously.
	c.AddFunc("@daily", DailyGrabDatas)
	c.AddFunc("@every 15m", func() {
		if err := SyncAllGamesIfStale(gameListSyncInterval); err != nil {
			log.Printf("[Sync] Scheduled game sync failed: %v", err)
		}
	})
	c.AddFunc("30 0 * * *", func() {
		log.Println("=== 开始执行昨日结算 ===")

		// 0. 数据校准 (The Source of Truth)
		// 从三方拉取准确的 Turnover 和 Winlose，覆盖本地估算值
		// yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		// SyncBetStatsFromProvider(yesterday)

		// 1. 计算返佣 (Distribution Rebate)
		// 读取昨天的 daily_user_stats -> 算出 bet_amount -> 给上级发钱
		CalculateDailyRebate()

		// 2. 计算留存 (Retention)
		// 读取昨天的 daily_user_stats -> 看谁活跃了 -> 对比注册时间 -> 更新 stats_retention 表
		CalculateRetention()

		log.Println("=== 昨日结算完成 ===")
	})
	c.Start()
	// Funcs may also be added to a running Cron
	return c
}

func StopAndClean(c *cron.Cron) {
	es := c.Entries()
	for i := 0; i < len(es); i++ {
		c.Remove(es[i].ID)
	}
	c.Stop()
}
