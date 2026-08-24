package controllers

import "gogogo/models"

// RecordDailyStats 原子更新每日统计数据。
func RecordDailyStats(userId uint64, bet, win, deposit, withdraw float64, loginIncrement int) error {
	return models.RecordTodayDailyUserStats(userId, bet, win, deposit, withdraw, loginIncrement)
}
