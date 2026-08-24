package controllers

import (
	"context"
	stdjson "encoding/json"
	"gogogo/models"
	"gogogo/models/dtos"
	"log"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func retentionDateRange(statDate string) (string, string, error) {
	startAt, err := time.Parse("2006-01-02", statDate)
	if err != nil {
		return "", "", err
	}
	return startAt.Format("2006-01-02 15:04:05"), startAt.Add(24 * time.Hour).Format("2006-01-02 15:04:05"), nil
}

// CalculateRetention 每日执行一次 (计算昨日及历史的留存)
func CalculateRetention() {
	now := time.Now()
	today := now.Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")

	// 1. 统计昨日新增用户数 (基础数据)
	yesterdayStart, yesterdayEnd, err := retentionDateRange(yesterday)
	if err != nil {
		log.Println("[Stats] invalid retention date:", yesterday, err)
		return
	}
	var newUsersCount int64
	models.GetInstance().DbInstance.Model(&dtos.User{}).
		Where("created_at >= ? AND created_at < ?", yesterdayStart, yesterdayEnd).
		Count(&newUsersCount)

	// 2. 统计昨日充值人数、充值总金额和有输赢的游戏玩家数
	var dailySummary struct {
		DepositUsers     int64
		DepositAmount    float64
		GameWinloseUsers int64
	}
	models.GetInstance().DbInstance.Model(&dtos.DailyUserStats{}).
		Select(`
			COALESCE(SUM(CASE WHEN deposit_amount > 0 THEN 1 ELSE 0 END), 0) AS deposit_users,
			COALESCE(SUM(CASE WHEN deposit_amount > 0 THEN deposit_amount ELSE 0 END), 0) AS deposit_amount,
			COALESCE(SUM(CASE WHEN bet_amount > 0 AND win_amount <> 0 THEN 1 ELSE 0 END), 0) AS game_winlose_users
		`).
		Where("stat_date = ?", yesterday).
		Scan(&dailySummary)

	// 初始化/更新昨日的记录
	stat := dtos.StatsRetention{
		StatDate:         yesterday,
		NewUsers:         int(newUsersCount),
		DepositUsers:     int(dailySummary.DepositUsers),
		DepositAmount:    dailySummary.DepositAmount,
		GameWinloseUsers: int(dailySummary.GameWinloseUsers),
	}
	// Upsert
	models.GetInstance().DbInstance.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "stat_date"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"new_users",
			"deposit_users",
			"deposit_amount",
			"game_winlose_users",
		}),
	}).Create(&stat)

	// 3. 计算各个周期的留存
	// 逻辑：更新 StatDate = (今天 - N天 - 1) 的记录
	daysToCheck := []int{1, 3, 7, 30}

	for _, days := range daysToCheck {
		// 目标注册日期 = 今天 - (days + 1)
		// 例如计算次日留存(days=1): 注册日期就是昨天
		// 例如计算3日留存(days=3): 注册日期是前天
		targetRegDateObj := now.AddDate(0, 0, -(days + 1))
		targetRegDate := targetRegDateObj.Format("2006-01-02")
		targetStart, targetEnd, err := retentionDateRange(targetRegDate)
		if err != nil {
			log.Println("[Stats] invalid target retention date:", targetRegDate, err)
			continue
		}

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
		err = models.GetInstance().DbInstance.Model(&dtos.User{}).
			Joins("JOIN daily_user_stats d ON users.id = d.user_id").
			Where("users.created_at >= ? AND users.created_at < ?", targetStart, targetEnd).
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
	models.GetInstance().DbInstance.
		Select("DISTINCT user_id, stat_date").
		Where("stat_date = ?", targetDate).
		Find(&activeStats)

	log.Printf("[SyncStats] Start syncing for %s, total users: %d", targetDate, len(activeStats))

	// 2. 遍历用户调用统一同步入口，确保 summary 表和 daily_user_stats 一起刷新
	for _, stat := range activeStats {
		if _, err := SyncPlayerDailyGameSummaryWithCurrencyOptions(context.Background(), stat.UserID, targetDate, "", true); err != nil {
			log.Printf("[SyncStats] Failed for user %d: %v", stat.UserID, err)
		}
	}

	log.Printf("[SyncStats] Finished syncing for %s", targetDate)
}

// CleanupExpiredCoupons 清理过期优惠券
func CleanupExpiredCoupons() {
	now := time.Now()

	// 1. 将过期但未使用的优惠券标记为已过期
	result := models.GetInstance().DbInstance.Model(&dtos.UserCoupon{}).
		Where("status IN (?) AND valid_end < ?", []int{0, 1}, now).
		Update("status", 3)

	if result.Error != nil {
		log.Printf("[CleanupExpiredCoupons] Error: %v", result.Error)
		return
	}

	if result.RowsAffected > 0 {
		log.Printf("[CleanupExpiredCoupons] marked %d expired coupons", result.RowsAffected)
	}
}

// ResetDailyActivities 每日活动重置
// 1. 清理昨日过期的轮盘优惠券
// 2. 重置活动进度状态（如果需要）
func ResetDailyActivities() {
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// 1. 更新昨日轮盘进度中未激活的优惠券状态
	var wheelProgresses []dtos.UserActivityProgress
	models.GetInstance().DbInstance.Where("activity_type = ? AND DATE(progress_date) = ? AND status = ?",
		"coupon_wheel", yesterday, 1).Find(&wheelProgresses)

	expiredCount := 0
	for _, progress := range wheelProgresses {
		var progressData dtos.WheelProgressData
		if err := stdjson.Unmarshal(progress.ProgressData, &progressData); err == nil {
			if !progressData.Activated {
				// 标记为已过期
				progress.Status = 4 // 已过期
				models.GetInstance().DbInstance.Save(&progress)
				expiredCount++
			}
		}
	}

	if expiredCount > 0 {
		log.Printf("[ResetDailyActivities] expired %d inactive wheel progress records", expiredCount)
	}

	// 2. 检查并自动开始/结束活动（根据时间）
	AutoUpdateActivityStatus()
}

// AutoUpdateActivityStatus 根据时间自动更新活动状态
func AutoUpdateActivityStatus() {
	now := time.Now()

	// 1. 自动开始已到开始时间的活动
	result := models.GetInstance().DbInstance.Model(&dtos.Activity{}).
		Where("status = ? AND start_time <= ? AND end_time >= ?", 0, now, now).
		Update("status", 1)

	if result.RowsAffected > 0 {
		log.Printf("[AutoUpdateActivityStatus] auto-started %d activities", result.RowsAffected)
	}

	// 2. 自动结束已过结束时间的活动
	result = models.GetInstance().DbInstance.Model(&dtos.Activity{}).
		Where("status = ? AND end_time < ?", 1, now).
		Update("status", 0)

	if result.RowsAffected > 0 {
		log.Printf("[AutoUpdateActivityStatus] auto-ended %d activities", result.RowsAffected)
	}
}
