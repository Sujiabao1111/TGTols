package example

import (
	"context"
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/example"
	exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
)

type DailyUserStatsService struct{}

type dailyUserStatsGameAmountRow struct {
	UserID   int     `gorm:"column:user_id"`
	StatDate string  `gorm:"column:stat_date"`
	Bet      float64 `gorm:"column:bet_amount"`
	Win      float64 `gorm:"column:win_amount"`
	Count    int64   `gorm:"column:row_count"`
}

func (dailyUserStatsService *DailyUserStatsService) CreateDailyUserStats(ctx context.Context, dailyUserStats *example.DailyUserStats) (err error) {
	err = global.GVA_DB.Create(dailyUserStats).Error
	return err
}

func (dailyUserStatsService *DailyUserStatsService) DeleteDailyUserStats(ctx context.Context, id string) (err error) {
	err = global.GVA_DB.Delete(&example.DailyUserStats{}, "id = ?", id).Error
	return err
}

func (dailyUserStatsService *DailyUserStatsService) DeleteDailyUserStatsByIds(ctx context.Context, ids []string) (err error) {
	err = global.GVA_DB.Delete(&[]example.DailyUserStats{}, "id in ?", ids).Error
	return err
}

func (dailyUserStatsService *DailyUserStatsService) UpdateDailyUserStats(ctx context.Context, dailyUserStats example.DailyUserStats) (err error) {
	err = global.GVA_DB.Model(&example.DailyUserStats{}).Where("id = ?", dailyUserStats.Id).Updates(&dailyUserStats).Error
	return err
}

func (dailyUserStatsService *DailyUserStatsService) GetDailyUserStats(ctx context.Context, id string) (dailyUserStats example.DailyUserStats, err error) {
	err = global.GVA_DB.Where("id = ?", id).First(&dailyUserStats).Error
	if err != nil {
		return dailyUserStats, err
	}

	items := []example.DailyUserStats{dailyUserStats}
	if err = dailyUserStatsService.hydrateDailyUserStatsGameAmounts(ctx, items); err != nil {
		return dailyUserStats, err
	}
	return items[0], nil
}

func (dailyUserStatsService *DailyUserStatsService) GetDailyUserStatsInfoList(ctx context.Context, info exampleReq.DailyUserStatsSearch) (list []example.DailyUserStats, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	db := global.GVA_DB.Model(&example.DailyUserStats{})
	var dailyUserStatsList []example.DailyUserStats

	if info.UserId != nil {
		db = db.Where("user_id = ?", *info.UserId)
	}
	if info.StatDateRange != nil && len(info.StatDateRange) == 2 {
		db = db.Where("stat_date BETWEEN ? AND ? ", info.StatDateRange[0], info.StatDateRange[1])
	}
	err = db.Count(&total).Error
	if err != nil {
		return
	}

	if limit != 0 {
		db = db.Limit(limit).Offset(offset)
	}

	err = db.Find(&dailyUserStatsList).Error
	if err != nil {
		return dailyUserStatsList, total, err
	}
	err = dailyUserStatsService.hydrateDailyUserStatsGameAmounts(ctx, dailyUserStatsList)
	return dailyUserStatsList, total, err
}

func (dailyUserStatsService *DailyUserStatsService) hydrateDailyUserStatsGameAmounts(ctx context.Context, items []example.DailyUserStats) error {
	userIDs := make([]int, 0, len(items))
	statDates := make([]string, 0, len(items))
	seenUserIDs := make(map[int]struct{}, len(items))
	seenStatDates := make(map[string]struct{}, len(items))

	for _, item := range items {
		if item.UserId == nil || item.StatDate == nil {
			continue
		}
		userID := *item.UserId
		statDate := item.StatDate.Format("2006-01-02")
		if userID <= 0 {
			continue
		}
		if _, ok := seenUserIDs[userID]; !ok {
			seenUserIDs[userID] = struct{}{}
			userIDs = append(userIDs, userID)
		}
		if _, ok := seenStatDates[statDate]; !ok {
			seenStatDates[statDate] = struct{}{}
			statDates = append(statDates, statDate)
		}
	}
	if len(userIDs) == 0 || len(statDates) == 0 {
		return nil
	}

	amounts := make(map[string]dailyUserStatsGameAmountRow)
	detailRows := make([]dailyUserStatsGameAmountRow, 0)
	if err := global.GVA_DB.WithContext(ctx).
		Table("user_game_transaction_details").
		Select(`
			user_id,
			DATE_FORMAT(stat_date, '%Y-%m-%d') AS stat_date,
			COALESCE(SUM(bet), 0) AS bet_amount,
			COALESCE(SUM(winlose), 0) AS win_amount,
			COALESCE(SUM(CASE WHEN turnover <> 0 OR bet <> 0 OR win <> 0 OR winlose <> 0 THEN 1 ELSE 0 END), 0) AS row_count
		`).
		Where("user_id IN ? AND stat_date IN ?", userIDs, statDates).
		Group("user_id, stat_date").
		Scan(&detailRows).Error; err != nil {
		return err
	}
	for _, row := range detailRows {
		if row.Count > 0 {
			amounts[dailyUserStatsGameAmountKey(row.UserID, row.StatDate)] = row
		}
	}

	statRows := make([]dailyUserStatsGameAmountRow, 0)
	if err := global.GVA_DB.WithContext(ctx).
		Table("user_game_transaction_stats").
		Select(`
			user_id,
			period_key AS stat_date,
			COALESCE(SUM(bet), 0) AS bet_amount,
			COALESCE(SUM(winlose), 0) AS win_amount,
			COALESCE(SUM(count), 0) AS row_count
		`).
		Where("user_id IN ? AND period_type = ? AND period_key IN ?", userIDs, "daily", statDates).
		Group("user_id, period_key").
		Scan(&statRows).Error; err != nil {
		return err
	}
	for _, row := range statRows {
		key := dailyUserStatsGameAmountKey(row.UserID, row.StatDate)
		if existing, ok := amounts[key]; !ok || existing.Count == 0 {
			amounts[key] = row
		}
	}

	for index := range items {
		item := &items[index]
		if item.UserId == nil || item.StatDate == nil {
			continue
		}
		key := dailyUserStatsGameAmountKey(*item.UserId, item.StatDate.Format("2006-01-02"))
		row, ok := amounts[key]
		if !ok || row.Count == 0 {
			continue
		}
		item.BetAmount = dailyUserStatsFloat64Ptr(row.Bet)
		item.WinAmount = dailyUserStatsFloat64Ptr(row.Win)
	}
	return nil
}

func dailyUserStatsGameAmountKey(userID int, statDate string) string {
	return strconv.Itoa(userID) + "|" + statDate
}

func dailyUserStatsFloat64Ptr(value float64) *float64 {
	return &value
}

func (dailyUserStatsService *DailyUserStatsService) GetDailyUserStatsPublic(ctx context.Context) {
}
