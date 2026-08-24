package example

import (
	"context"
	"log"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/example"
	exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StatsRetentionService struct{}

func getStatsRetentionQuery(ctx context.Context) *gorm.DB {
	return global.GVA_DB.WithContext(ctx).Model(&example.StatsRetention{})
}

func statsRetentionDateRange(statDate string) (string, string, error) {
	startAt, err := time.Parse("2006-01-02", statDate)
	if err != nil {
		return "", "", err
	}
	return startAt.Format("2006-01-02 15:04:05"), startAt.Add(24 * time.Hour).Format("2006-01-02 15:04:05"), nil
}

func buildStatsRetentionPayload(db *gorm.DB, statDate string) (map[string]interface{}, error) {
	startAt, endAt, err := statsRetentionDateRange(statDate)
	if err != nil {
		return nil, err
	}

	var newUsersCount int64
	if err = db.Table("users").
		Where("created_at >= ? AND created_at < ?", startAt, endAt).
		Count(&newUsersCount).Error; err != nil {
		return nil, err
	}

	var dailySummary struct {
		DepositUsers     int64   `gorm:"column:deposit_users"`
		DepositAmount    float64 `gorm:"column:deposit_amount"`
		GameWinloseUsers int64   `gorm:"column:game_winlose_users"`
	}
	if err := db.Table("daily_user_stats").
		Select(`
			COALESCE(SUM(CASE WHEN deposit_amount > 0 THEN 1 ELSE 0 END), 0) AS deposit_users,
			COALESCE(SUM(CASE WHEN deposit_amount > 0 THEN deposit_amount ELSE 0 END), 0) AS deposit_amount,
			COALESCE(SUM(CASE WHEN bet_amount > 0 AND win_amount <> 0 THEN 1 ELSE 0 END), 0) AS game_winlose_users
		`).
		Where("stat_date = ?", statDate).
		Scan(&dailySummary).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"new_users":          int(newUsersCount),
		"deposit_users":      int(dailySummary.DepositUsers),
		"deposit_amount":     dailySummary.DepositAmount,
		"game_winlose_users": int(dailySummary.GameWinloseUsers),
	}, nil
}

func deduplicateStatsRetentionByDate(ctx context.Context) error {
	db := global.GVA_DB.WithContext(ctx)
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`
			UPDATE stats_retention AS target
			INNER JOIN (
				SELECT
					DATE(stat_date) AS stat_day,
					MAX(id) AS keep_id,
					MAX(new_users) AS new_users,
					MAX(deposit_users) AS deposit_users,
					MAX(deposit_amount) AS deposit_amount,
					MAX(game_winlose_users) AS game_winlose_users,
					MAX(retention_1) AS retention_1,
					MAX(retention_3) AS retention_3,
					MAX(retention_7) AS retention_7,
					MAX(retention_30) AS retention_30
				FROM stats_retention
				GROUP BY DATE(stat_date)
			) AS merged
				ON target.id = merged.keep_id
			SET
				target.stat_date = merged.stat_day,
				target.new_users = merged.new_users,
				target.deposit_users = merged.deposit_users,
				target.deposit_amount = merged.deposit_amount,
				target.game_winlose_users = merged.game_winlose_users,
				target.retention_1 = merged.retention_1,
				target.retention_3 = merged.retention_3,
				target.retention_7 = merged.retention_7,
				target.retention_30 = merged.retention_30
		`).Error; err != nil {
			return err
		}

		return tx.Exec(`
			DELETE older
			FROM stats_retention AS older
			INNER JOIN stats_retention AS newer
				ON DATE(older.stat_date) = DATE(newer.stat_date)
				AND older.id < newer.id
		`).Error
	})
}

func deduplicateStatsRetentionByDateBestEffort(ctx context.Context) {
	if err := deduplicateStatsRetentionByDate(ctx); err != nil {
		log.Printf("[StatsRetention] deduplicate by date failed: %v", err)
	}
}

func refreshStatsRetentionByDate(ctx context.Context, statDate string) error {
	db := global.GVA_DB.WithContext(ctx)
	payload, err := buildStatsRetentionPayload(db, statDate)
	if err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		var existing []example.StatsRetention
		if err := tx.Model(&example.StatsRetention{}).
			Where("DATE(stat_date) = ?", statDate).
			Order("id DESC").
			Find(&existing).Error; err != nil {
			return err
		}

		if len(existing) > 0 {
			keepID := existing[0].Id
			if err := tx.Model(&example.StatsRetention{}).
				Where("id = ?", keepID).
				Updates(payload).Error; err != nil {
				return err
			}

			if len(existing) > 1 {
				duplicateIDs := make([]int, 0, len(existing)-1)
				for _, row := range existing[1:] {
					if row.Id != nil {
						duplicateIDs = append(duplicateIDs, *row.Id)
					}
				}
				if len(duplicateIDs) > 0 {
					if err := tx.Delete(&[]example.StatsRetention{}, "id IN ?", duplicateIDs).Error; err != nil {
						return err
					}
				}
			}
			return nil
		}

		statDateValue, err := time.Parse("2006-01-02", statDate)
		if err != nil {
			return err
		}

		return tx.Table("stats_retention").Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "stat_date"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"new_users",
				"deposit_users",
				"deposit_amount",
				"game_winlose_users",
			}),
		}).Create(map[string]interface{}{
			"stat_date":          statDateValue,
			"new_users":          payload["new_users"],
			"deposit_users":      payload["deposit_users"],
			"deposit_amount":     payload["deposit_amount"],
			"game_winlose_users": payload["game_winlose_users"],
		}).Error
	})
}

func refreshTodayStatsRetention(ctx context.Context) error {
	return refreshStatsRetentionByDate(ctx, time.Now().Format("2006-01-02"))
}

func refreshStatsRetentionDateBestEffort(ctx context.Context, statDate string) {
	if err := refreshStatsRetentionByDate(ctx, statDate); err != nil {
		log.Printf("[StatsRetention] refresh %s failed: %v", statDate, err)
	}
}

func refreshTodayStatsRetentionBestEffort(ctx context.Context) {
	if err := refreshTodayStatsRetention(ctx); err != nil {
		log.Printf("[StatsRetention] refresh today failed: %v", err)
	}
}

func refreshHistoricalStatsRetentionBestEffort(ctx context.Context, dateRange []string) {
	db := global.GVA_DB.WithContext(ctx)
	query := db.Table("stats_retention").Distinct("DATE(stat_date) AS stat_date")

	if len(dateRange) == 2 {
		query = query.Where("DATE(stat_date) BETWEEN ? AND ?", dateRange[0], dateRange[1])
	}

	var statDates []string
	if err := query.Order("DATE(stat_date) DESC").Scan(&statDates).Error; err != nil {
		log.Printf("[StatsRetention] load historical dates failed: %v", err)
		return
	}

	today := time.Now().Format("2006-01-02")
	for _, statDate := range statDates {
		if statDate == "" || statDate == today {
			continue
		}
		refreshStatsRetentionDateBestEffort(ctx, statDate)
	}
}

func (statsRetentionService *StatsRetentionService) CreateStatsRetention(ctx context.Context, statsRetention *example.StatsRetention) (err error) {
	err = getStatsRetentionQuery(ctx).Create(statsRetention).Error
	return err
}

func (statsRetentionService *StatsRetentionService) DeleteStatsRetention(ctx context.Context, id string) (err error) {
	err = getStatsRetentionQuery(ctx).Delete(&example.StatsRetention{}, "id = ?", id).Error
	return err
}

func (statsRetentionService *StatsRetentionService) DeleteStatsRetentionByIds(ctx context.Context, ids []string) (err error) {
	err = getStatsRetentionQuery(ctx).Delete(&[]example.StatsRetention{}, "id in ?", ids).Error
	return err
}

func (statsRetentionService *StatsRetentionService) UpdateStatsRetention(ctx context.Context, statsRetention example.StatsRetention) (err error) {
	err = getStatsRetentionQuery(ctx).Where("id = ?", statsRetention.Id).Updates(&statsRetention).Error
	return err
}

func (statsRetentionService *StatsRetentionService) GetStatsRetention(ctx context.Context, id string) (statsRetention example.StatsRetention, err error) {
	deduplicateStatsRetentionByDateBestEffort(ctx)
	refreshTodayStatsRetentionBestEffort(ctx)
	var statDateRow struct {
		StatDate time.Time `gorm:"column:stat_date"`
	}
	if lookupErr := global.GVA_DB.WithContext(ctx).
		Table("stats_retention").
		Select("DATE(stat_date) AS stat_date").
		Where("id = ?", id).
		Scan(&statDateRow).Error; lookupErr == nil && !statDateRow.StatDate.IsZero() {
		refreshStatsRetentionDateBestEffort(ctx, statDateRow.StatDate.Format("2006-01-02"))
	}
	err = getStatsRetentionQuery(ctx).Where("id = ?", id).First(&statsRetention).Error
	if err == nil && statsRetention.StatDate != nil {
		payload, payloadErr := buildStatsRetentionPayload(global.GVA_DB.WithContext(ctx), statsRetention.StatDate.Format("2006-01-02"))
		if payloadErr == nil {
			if newUsers, ok := payload["new_users"].(int); ok {
				statsRetention.NewUsers = &newUsers
			}
		}
	}
	return
}

func hydrateStatsRetentionNewUsers(ctx context.Context, items []example.StatsRetention) error {
	db := global.GVA_DB.WithContext(ctx)
	for index := range items {
		if items[index].StatDate == nil {
			continue
		}
		payload, err := buildStatsRetentionPayload(db, items[index].StatDate.Format("2006-01-02"))
		if err != nil {
			return err
		}
		if newUsers, ok := payload["new_users"].(int); ok {
			items[index].NewUsers = &newUsers
		}
	}
	return nil
}

func (statsRetentionService *StatsRetentionService) GetStatsRetentionInfoList(ctx context.Context, info exampleReq.StatsRetentionSearch) (list []example.StatsRetention, total int64, err error) {
	deduplicateStatsRetentionByDateBestEffort(ctx)
	refreshTodayStatsRetentionBestEffort(ctx)
	refreshHistoricalStatsRetentionBestEffort(ctx, info.StatDateRange)

	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	db := global.GVA_DB.WithContext(ctx).Table("stats_retention AS sr")
	var statsRetentions []example.StatsRetention

	if info.StatDateRange != nil && len(info.StatDateRange) == 2 {
		db = db.Where("DATE(sr.stat_date) BETWEEN ? AND ? ", info.StatDateRange[0], info.StatDateRange[1])
	}

	grouped := db.Select(`
		MAX(sr.id) AS id,
		DATE(sr.stat_date) AS stat_date,
		MAX(sr.new_users) AS new_users,
		MAX(sr.deposit_users) AS deposit_users,
		MAX(sr.deposit_amount) AS deposit_amount,
		MAX(sr.game_winlose_users) AS game_winlose_users,
		MAX(sr.retention_1) AS retention_1,
		MAX(sr.retention_3) AS retention_3,
		MAX(sr.retention_7) AS retention_7,
		MAX(sr.retention_30) AS retention_30,
		MAX(sr.created_at) AS created_at,
		MAX(sr.updated_at) AS updated_at
	`).Group("DATE(sr.stat_date)")

	err = grouped.Count(&total).Error
	if err != nil {
		return
	}

	if limit != 0 {
		grouped = grouped.Limit(limit).Offset(offset)
	}

	err = grouped.Order("DATE(sr.stat_date) DESC").Find(&statsRetentions).Error
	if err != nil {
		return nil, total, err
	}
	err = hydrateStatsRetentionNewUsers(ctx, statsRetentions)
	return statsRetentions, total, err
}

func (statsRetentionService *StatsRetentionService) GetStatsRetentionPublic(ctx context.Context) {
}
