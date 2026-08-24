package models

import (
	"gogogo/models/dtos"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UpsertDailyUserStatsWithDB increments a user's daily aggregated stats inside the provided DB context.
func UpsertDailyUserStatsWithDB(db *gorm.DB, userID uint64, statDate string, bet, win, deposit, withdraw float64, loginIncrement int) error {
	if statDate == "" {
		statDate = time.Now().Format("2006-01-02")
	}

	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "stat_date"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"bet_amount":      gorm.Expr("bet_amount + ?", bet),
			"win_amount":      gorm.Expr("win_amount + ?", win),
			"deposit_amount":  gorm.Expr("deposit_amount + ?", deposit),
			"withdraw_amount": gorm.Expr("withdraw_amount + ?", withdraw),
			"login_count":     gorm.Expr("login_count + ?", loginIncrement),
			"updated_at":      gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(&dtos.DailyUserStats{
		UserID:         userID,
		StatDate:       statDate,
		BetAmount:      bet,
		WinAmount:      win,
		DepositAmount:  deposit,
		WithdrawAmount: withdraw,
		LoginCount:     loginIncrement,
	}).Error
}

func UpsertDailyUserStats(userID uint64, statDate string, bet, win, deposit, withdraw float64, loginIncrement int) error {
	return UpsertDailyUserStatsWithDB(GetInstance().DbInstance, userID, statDate, bet, win, deposit, withdraw, loginIncrement)
}

func RecordTodayDailyUserStats(userID uint64, bet, win, deposit, withdraw float64, loginIncrement int) error {
	return UpsertDailyUserStats(userID, time.Now().Format("2006-01-02"), bet, win, deposit, withdraw, loginIncrement)
}
