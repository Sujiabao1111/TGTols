package example

import (
	"context"
	"fmt"

	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	model "github.com/flipped-aurora/gin-vue-admin/server/model/example"
	exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
)

type DataStatisticsService struct{}

const platformWinLossStartDate = "2026-07-01"

// Recharge overview intentionally excludes legacy/test records before the
// current accounting period.
const rechargeStatsStartDate = "2026-09-01 00:00:00"

func (s *DataStatisticsService) GetDataStatistics(ctx context.Context, info exampleReq.DataStatisticsSearch) (model.DataStatisticsResult, error) {
	start, end, err := parseDataStatisticsRange(info)
	if err != nil {
		return model.DataStatisticsResult{}, err
	}
	result := model.DataStatisticsResult{StartDate: start.Format("2006-01-02"), EndDate: end.AddDate(0, 0, -1).Format("2006-01-02")}
	db := global.GVA_DB.WithContext(ctx)
	if err = db.Table("users").Count(&result.Overview.TotalRegisteredUsers).Error; err != nil {
		return result, err
	}
	// Keep the overview count consistent with recharge amount and daily details:
	// only transactions that were actually posted to the ledger count as a
	// successful recharge. Counting payment_orders also includes pending,
	// expired and duplicate/test orders.
	if err = db.Raw("SELECT COUNT(*) FROM (SELECT t.user_id FROM transactions t INNER JOIN users u ON u.id = t.user_id WHERE t.type = ? AND t.created_at >= ? GROUP BY t.user_id HAVING SUM(t.amount) > 0) AS recharge_users", 1, rechargeStatsStartDate).
		Scan(&result.Overview.TotalRechargeUsers).Error; err != nil {
		return result, err
	}
	if err = db.Table("transactions").Where("type = ? AND created_at >= ?", 1, rechargeStatsStartDate).Select("COALESCE(SUM(amount), 0)").Scan(&result.Overview.TotalRechargeAmount).Error; err != nil {
		return result, err
	}
	if err = db.Table("withdraw_orders wo").Joins("LEFT JOIN transactions tx ON tx.reference_id = wo.order_id AND tx.type = 2").Where("wo.status = ? AND wo.completed_at IS NOT NULL", 1).Select("COALESCE(SUM(CASE WHEN wo.cost > 0 THEN wo.cost ELSE ABS(COALESCE(tx.amount, 0)) END), 0)").Scan(&result.Overview.TotalWithdrawAmount).Error; err != nil {
		return result, err
	}
	result.Overview.TotalPaymentGap = result.Overview.TotalRechargeAmount - result.Overview.TotalWithdrawAmount
	platformWinLossStart, err := time.ParseInLocation("2006-01-02", platformWinLossStartDate, start.Location())
	if err != nil {
		return result, err
	}
	var playerWinLoss float64
	if err = db.Table("user_game_transaction_details").
		Where("stat_date >= ?", platformWinLossStart).
		Select("COALESCE(SUM(winlose), 0)").
		Scan(&playerWinLoss).Error; err != nil {
		return result, err
	}
	result.Overview.PlatformWinLoss = -playerWinLoss
	todayStart := time.Now().In(start.Location())
	todayStart = time.Date(todayStart.Year(), todayStart.Month(), todayStart.Day(), 0, 0, 0, 0, start.Location())
	var todayPlayerWinLoss float64
	if err = db.Table("user_game_transaction_details").Where("stat_date >= ? AND stat_date < ?", todayStart, todayStart.AddDate(0, 0, 1)).Select("COALESCE(SUM(winlose), 0)").Scan(&todayPlayerWinLoss).Error; err != nil {
		return result, err
	}
	result.Overview.TodayWinLoss = -todayPlayerWinLoss

	days := make(map[string]*model.DataStatisticsDaily)
	for day := start; day.Before(end); day = day.AddDate(0, 0, 1) {
		key := day.Format("2006-01-02")
		days[key] = &model.DataStatisticsDaily{Date: key}
	}
	type countRow struct {
		Date  string `gorm:"column:stat_date"`
		Count int64  `gorm:"column:item_count"`
	}
	var counts []countRow
	if err = db.Table("users").Select("DATE_FORMAT(created_at, '%Y-%m-%d') stat_date, COUNT(*) item_count").Where("created_at >= ? AND created_at < ?", start, end).Group("DATE(created_at)").Scan(&counts).Error; err != nil {
		return result, err
	}
	for _, row := range counts {
		if days[row.Date] != nil {
			days[row.Date].RegisterUsers = row.Count
		}
	}
	counts = nil
	if err = db.Table("users").Select("DATE_FORMAT(last_login_at, '%Y-%m-%d') stat_date, COUNT(*) item_count").Where("last_login_at >= ? AND last_login_at < ?", start, end).Group("DATE(last_login_at)").Scan(&counts).Error; err != nil {
		return result, err
	}
	for _, row := range counts {
		if days[row.Date] != nil {
			days[row.Date].ActiveUsers = row.Count
		}
	}

	type paymentRow struct {
		Date   string  `gorm:"column:stat_date"`
		Users  int64   `gorm:"column:user_count"`
		Amount float64 `gorm:"column:total_amount"`
	}
	var payments []paymentRow
	if err = db.Table("transactions").Select("DATE_FORMAT(created_at, '%Y-%m-%d') stat_date, COUNT(DISTINCT user_id) user_count, COALESCE(SUM(amount), 0) total_amount").Where("type = ? AND amount > 0 AND created_at >= ? AND created_at < ?", 1, start, end).Group("DATE(created_at)").Scan(&payments).Error; err != nil {
		return result, err
	}
	for _, row := range payments {
		if days[row.Date] != nil {
			days[row.Date].RechargeUsers = row.Users
			days[row.Date].RechargeAmount = row.Amount
		}
	}
	payments = nil
	if err = db.Table("withdraw_orders wo").Joins("LEFT JOIN transactions tx ON tx.reference_id = wo.order_id AND tx.type = 2").Select("DATE_FORMAT(wo.completed_at, '%Y-%m-%d') stat_date, COUNT(DISTINCT wo.user_id) user_count, COALESCE(SUM(CASE WHEN wo.cost > 0 THEN wo.cost ELSE ABS(COALESCE(tx.amount, 0)) END), 0) total_amount").Where("wo.status = ? AND wo.completed_at IS NOT NULL AND wo.completed_at >= ? AND wo.completed_at < ?", 1, start, end).Group("DATE(wo.completed_at)").Scan(&payments).Error; err != nil {
		return result, err
	}
	for _, row := range payments {
		if days[row.Date] != nil {
			days[row.Date].WithdrawUsers = row.Users
			days[row.Date].WithdrawAmount = row.Amount
		}
	}
	type winLossRow struct {
		Date          string  `gorm:"column:stat_date"`
		PlayerWinLoss float64 `gorm:"column:player_win_loss"`
	}
	var winLosses []winLossRow
	if err = db.Table("user_game_transaction_details").
		Select("DATE_FORMAT(stat_date, '%Y-%m-%d') stat_date, COALESCE(SUM(winlose), 0) player_win_loss").
		Where("stat_date >= ? AND stat_date < ?", start, end).
		Group("DATE(stat_date)").
		Scan(&winLosses).Error; err != nil {
		return result, err
	}
	for _, row := range winLosses {
		if days[row.Date] != nil {
			// 明细记录是玩家视角，经营报表统一转换为平台视角。
			days[row.Date].DailyWinLoss = -row.PlayerWinLoss
		}
	}
	for day := start; day.Before(end); day = day.AddDate(0, 0, 1) {
		row := days[day.Format("2006-01-02")]
		row.PaymentGap = row.RechargeAmount - row.WithdrawAmount
		result.Daily = append(result.Daily, *row)
	}
	return result, nil
}

func parseDataStatisticsRange(info exampleReq.DataStatisticsSearch) (time.Time, time.Time, error) {
	location, _ := time.LoadLocation("Asia/Shanghai")
	if location == nil {
		location = time.FixedZone("CST", 8*3600)
	}
	now := time.Now().In(location)
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	end := start.AddDate(0, 0, 1)
	var err error
	if info.StartDate != "" {
		start, err = time.ParseInLocation("2006-01-02", info.StartDate, location)
		if err != nil {
			return start, end, fmt.Errorf("开始日期格式错误")
		}
	}
	if info.EndDate != "" {
		var last time.Time
		last, err = time.ParseInLocation("2006-01-02", info.EndDate, location)
		if err != nil {
			return start, end, fmt.Errorf("结束日期格式错误")
		}
		end = last.AddDate(0, 0, 1)
	}
	if !end.After(start) {
		return start, end, fmt.Errorf("结束日期不能早于开始日期")
	}
	if end.Sub(start) > 366*24*time.Hour {
		return start, end, fmt.Errorf("单次最多查询366天")
	}
	return start, end, nil
}
