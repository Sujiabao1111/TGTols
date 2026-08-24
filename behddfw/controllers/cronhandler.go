package controllers

import (
	"fmt"
	"gogogo/helpers"
	"gogogo/models"
	"gogogo/models/dtos"
	"gogogo/provider"
	"log"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CalculateRetention 每日执行一次 (计算昨日及历史的留存)
func CalculateRetention() {
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// 1. 统计昨日新增用户数 (基础数据)
	var newUsersCount int64
	models.GetInstance().DbInstance.Model(&dtos.User{}).Where("DATE(created_at) = ?", yesterday).Count(&newUsersCount)

	// 初始化/更新昨日的记录
	stat := dtos.StatsRetention{
		StatDate: yesterday,
		NewUsers: int(newUsersCount),
	}
	// Upsert
	models.GetInstance().DbInstance.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "stat_date"}},
		DoUpdates: clause.AssignmentColumns([]string{"new_users"}),
	}).Create(&stat)

	// 2. 计算各个周期的留存
	// 逻辑：更新 StatDate = (今天 - N天) 的记录
	daysToCheck := []int{1, 3, 7, 30}

	for _, days := range daysToCheck {
		// 目标注册日期 = 今天 - (days)
		// 例如计算次日留存(days=1): 注册日期就是昨天
		// 例如计算3日留存(days=3): 注册日期是前天
		targetRegDateObj := time.Now().AddDate(0, 0, -days)
		targetRegDate := targetRegDateObj.Format("2006-01-02")

		// 活跃日期 = 昨天 (因为今天还没过完，我们统计的是截止到昨天的活跃)
		// 也就是：targetRegDate 注册的人，在 yesterday 是否活跃
		activeDate := yesterday

		var retentionCount int64
		// SQL:
		// SELECT COUNT(DISTINCT u.id)
		// FROM users u
		// JOIN daily_user_stats d ON u.id = d.user_id
		// WHERE DATE(u.created_at) = targetRegDate
		// AND d.stat_date = activeDate
		err := models.GetInstance().DbInstance.Model(&dtos.User{}).
			Joins("JOIN daily_user_stats d ON users.id = d.user_id").
			Where("DATE(users.created_at) = ?", targetRegDate).
			Where("d.stat_date = ?", activeDate).
			Count(&retentionCount).Error

		if err == nil {
			// 更新对应的 retention_N 字段
			column := ""
			switch days {
			case 1:
				column = "retention_1"
			case 3:
				column = "retention_3"
			case 7:
				column = "retention_7"
			case 30:
				column = "retention_30"
			}

			models.GetInstance().DbInstance.Model(&dtos.StatsRetention{}).
				Where("stat_date = ?", targetRegDate).
				Update(column, retentionCount)
		}
	}

	log.Println("[Stats] Retention calculation finished for", today)
}

// CalculateDailyRebate 结算昨日返佣 (定时任务调用)
func CalculateDailyRebate() {
	// 1. 获取返点配置 (Level 1-9)
	var configs []dtos.AgentRebateConfig
	models.GetInstance().DbInstance.Find(&configs)
	configMap := make(map[int]float64) // Level -> Rate
	for _, c := range configs {
		configMap[c.Level] = c.Rate
	}

	// 2. 获取昨日有有效投注的所有用户 (从 daily_user_stats 拿)
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	var stats []dtos.DailyUserStats

	// 预加载 User 以获取 Path
	models.GetInstance().DbInstance.Preload("User").Where("stat_date = ? AND bet_amount > 0", yesterday).Find(&stats)

	// 3. 遍历每个产生的业绩，向上级返点
	models.GetInstance().DbInstance.Transaction(func(tx *gorm.DB) error {
		for _, stat := range stats {
			// stat.User.Path 格式如 "1/5/12/" (假设 stat.User.ID 是 20)
			// 它的上级 ID 列表是 [1, 5, 12]
			pathStr := strings.Trim(stat.User.Path, "/")
			if pathStr == "" {
				continue // 顶层用户，无上级
			}

			ancestorIdsStr := strings.Split(pathStr, "/")

			// 倒序遍历 (最近的上级是 Level 1)
			// ancestorIdsStr: ["1", "5", "12"]
			// Level 1: ID 12
			// Level 2: ID 5
			// Level 3: ID 1
			currentLevel := 1

			for i := len(ancestorIdsStr) - 1; i >= 0; i-- {
				if currentLevel > 9 {
					break // 只返 9 级
				}

				ancestorId, _ := strconv.ParseUint(ancestorIdsStr[i], 10, 64)
				rate := configMap[currentLevel]

				if rate > 0 {
					rebateAmount := stat.BetAmount * rate // 返佣 = 下注额 * 比例

					if rebateAmount > 0.01 { // 忽略太小的金额
						// a. 增加余额
						tx.Model(&dtos.User{}).Where("id = ?", ancestorId).
							Update("balance", gorm.Expr("balance + ?", rebateAmount))

						// b. 记录流水
						tx.Create(&dtos.Transaction{
							UserID: uint64(ancestorId),
							Type:   5, // 5=分销返点
							Amount: rebateAmount,
							// 此处 Before/After Balance 估算略过，严谨需查库
							Remark: "下级返点: 来自用户 " + stat.User.Username + " (Lv." + strconv.Itoa(currentLevel) + ")",
						})
					}
				}

				currentLevel++
			}
		}
		return nil
	})
}

// SyncBetStatsFromProvider 从三方同步昨日打码量 (覆盖更新)
func SyncBetStatsFromProvider(targetDate string) {
	// targetDate 格式: "2025-01-01"

	// 1. 找出昨天可能活跃的用户 (优化性能，避免全表扫描)
	// 这里我们查询 daily_user_stats 表中昨天有记录的用户 (可能是登录了，或者本地有充提)
	// 如果你之前做的 Login 记录了 daily_user_stats，那么这里能查到所有登录过的人
	var activeStats []dtos.DailyUserStats
	// 预加载 User 获取 ID
	models.GetInstance().DbInstance.Preload("User").Where("stat_date = ?", targetDate).Find(&activeStats)

	client := provider.NewHedocClient()

	log.Printf("[SyncStats] Start syncing for %s, total users: %d", targetDate, len(activeStats))

	// 2. 遍历用户调用接口 (生产环境建议使用 worker pool 并发)
	for _, stat := range activeStats {
		thirdPartyLoginId := fmt.Sprintf("%su%d", helpers.GetCfgInstance().Conf.Prefix, stat.UserID)

		data, err := client.GetPlayerDailySummary(thirdPartyLoginId, targetDate, helpers.GetCfgInstance().Conf.Agentid, helpers.GetCfgInstance().Conf.Agentapi)
		if err != nil {
			log.Printf("[SyncStats] Failed for user %d: %v", stat.UserID, err)
			continue
		}

		// 3. 更新数据库
		// 注意：这里是覆盖更新 (Overwrite)，以三方数据为准
		// Turnover -> BetAmount (返佣依据)
		// Winlose -> WinAmount
		if data.Turnover > 0 || data.Winlose != 0 {
			models.GetInstance().DbInstance.Model(&dtos.DailyUserStats{}).
				Where("user_id = ? AND stat_date = ?", stat.UserID, targetDate).
				Updates(map[string]interface{}{
					"bet_amount": data.Turnover, // 核心：使用有效投注
					"win_amount": data.Winlose,  // 核心：使用实际输赢
					// 如果需要存 Count，可以在 Model 加字段
				})
		}
	}

	log.Printf("[SyncStats] Finished syncing for %s", targetDate)
}
