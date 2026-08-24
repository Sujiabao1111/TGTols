package controllers

import (
	"gogogo/models"
	"gogogo/models/dtos"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RecordDailyStats 原子更新每日统计数据
// 利用 MySQL 的 ON DUPLICATE KEY UPDATE 特性
func RecordDailyStats(userId uint64, bet, win, deposit, withdraw float64, loginIncrement int) error {
	today := time.Now().Format("2006-01-02")

	// SQL 逻辑:
	// 如果记录不存在 -> INSERT ...
	// 如果记录已存在 -> UPDATE set bet_amount = bet_amount + new_bet ...
	err := models.GetInstance().DbInstance.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "stat_date"}}, // 冲突检测列
		DoUpdates: clause.Assignments(map[string]interface{}{
			"bet_amount":      gorm.Expr("bet_amount + ?", bet),
			"win_amount":      gorm.Expr("win_amount + ?", win),
			"deposit_amount":  gorm.Expr("deposit_amount + ?", deposit),
			"withdraw_amount": gorm.Expr("withdraw_amount + ?", withdraw),
			"login_count":     gorm.Expr("login_count + ?", loginIncrement),
			// updated_at 会由 GORM 自动处理
		}),
	}).Create(&dtos.DailyUserStats{
		UserID:         userId,
		StatDate:       today,
		BetAmount:      bet,
		WinAmount:      win,
		DepositAmount:  deposit,
		WithdrawAmount: withdraw,
		LoginCount:     loginIncrement,
	}).Error

	return err
}
