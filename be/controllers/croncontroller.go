package controllers

import (
	"context"
	"log"
	"time"

	"gogogo/services"

	"github.com/robfig/cron/v3"
)

// 全局汇率服务实例
var exchangeRateService *services.ExchangeRateService

func mustAddCron(c *cron.Cron, spec string, cmd func()) {
	if _, err := c.AddFunc(spec, cmd); err != nil {
		log.Fatalf("failed to register cron %q: %v", spec, err)
	}
}

// 日常定时任务
func DailyGrabDatas() {

}

func SetupAndGo() *cron.Cron {
	c := cron.New(cron.WithSeconds())
	// c := cron.New()
	// no seconds
	// c.AddFunc("* * * * *", ClearSignQuest)
	// Funcs are invoked in their own goroutine, asynchronously.
	mustAddCron(c, "@daily", DailyGrabDatas)
	mustAddCron(c, "@every 30s", func() {
		_ = services.GetPaymentService().ExpireTonOrders(context.Background())
		if err := services.GetPaymentService().ScanTonPendingOrders(context.Background()); err != nil {
			log.Printf("[TON] pending order scan failed: %v", err)
		}
	})
	mustAddCron(c, "@every 15m", func() {
		if err := SyncAllGamesIfStale(gameListSyncInterval); err != nil {
			log.Printf("[Sync] Scheduled game sync failed: %v", err)
		}
	})
	mustAddCron(c, "0 */5 * * * *", func() {
		if err := SyncM7CompletedWindowForActiveUsers(context.Background()); err != nil {
			log.Printf("[GameSummarySync][M7] scheduled completed window sync failed: %v", err)
		}
		now := time.Now().UTC()
		if _, err := SyncM7PPTransactions(context.Background(), now.Add(-20*time.Minute), now); err != nil {
			log.Printf("[GameSummarySync][M7PP] scheduled sync failed: %v", err)
		}
	})
	mustAddCron(c, "0 30 0 * * *", func() {
		log.Println("=== 开始执行昨日结算 ===")

		// 0. 数据校准 (The Source of Truth)
		// 从三方拉取准确的 Turnover 和 Winlose，覆盖本地估算值
		yesterday := time.Now().In(playerTransactionLocation()).AddDate(0, 0, -1).Format("2006-01-02")
		if _, err := SyncM7PPTransactionDate(context.Background(), yesterday); err != nil {
			log.Printf("[GameSummarySync][M7PP] yesterday backfill failed date=%s: %v", yesterday, err)
		}
		SyncBetStatsFromProvider(yesterday)

		// 1. 计算返佣 (Distribution Rebate)
		// 读取昨天的 daily_user_stats -> 算出 bet_amount -> 给上级发钱
		CalculateDailyRebate()

		// 2. 计算留存 (Retention)
		// 读取昨天的 daily_user_stats -> 看谁活跃了 -> 对比注册时间 -> 更新 stats_retention 表
		CalculateRetention()

		log.Println("=== 昨日结算完成 ===")
	})

	// 活动相关定时任务
	// 每小时清理过期优惠券
	mustAddCron(c, "0 0 * * * *", func() {
		log.Println("=== 开始清理过期优惠券 ===")
		CleanupExpiredCoupons()
		log.Println("=== 过期优惠券清理完成 ===")
	})

	// 每天凌晨1点重置轮盘活动（标记过期未使用的优惠券）
	mustAddCron(c, "0 0 1 * * *", func() {
		log.Println("=== 开始处理每日活动重置 ===")
		ResetDailyActivities()
		log.Println("=== 每日活动重置完成 ===")
	})
	// 初始化汇率服务并执行首次更新
	exchangeRateService = services.NewExchangeRateService()
	if err := exchangeRateService.FetchAndCache(); err != nil {
		log.Printf("[ExchangeRate] Initial fetch failed: %v", err)
	}

	// 每小时更新汇率
	mustAddCron(c, "0 0 * * * *", func() {
		log.Println("=== 开始更新汇率 ===")
		if err := exchangeRateService.FetchAndCache(); err != nil {
			log.Printf("[ExchangeRate] Update failed: %v", err)
		} else {
			log.Println("[ExchangeRate] Update success")
		}
		log.Println("=== 汇率更新完成 ===")
	})

	c.Start()
	go func() {
		if err := SyncM7PPGames(); err != nil {
			log.Printf("[Sync][M7PP] initial synchronization failed: %v", err)
		}
		now := time.Now().UTC()
		if _, err := SyncM7PPTransactions(context.Background(), now.AddDate(0, 0, -7), now); err != nil {
			log.Printf("[GameSummarySync][M7PP] initial transaction sync failed: %v", err)
		}
	}()
	// Funcs may also be added to a running Cron
	return c
}

// GetExchangeRateService 获取汇率服务实例
func GetExchangeRateService() *services.ExchangeRateService {
	return exchangeRateService
}

func StopAndClean(c *cron.Cron) {
	es := c.Entries()
	for i := 0; i < len(es); i++ {
		c.Remove(es[i].ID)
	}
	c.Stop()
}
