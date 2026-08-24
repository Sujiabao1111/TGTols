package example

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	model "github.com/flipped-aurora/gin-vue-admin/server/model/example"
	exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
	"gorm.io/gorm"
)

type UserGameRecordService struct{}

func (userGameRecordService *UserGameRecordService) GetUserGameRecordList(ctx context.Context, info exampleReq.UserGameRecordSearch) (list []model.UserGameRecord, total int64, err error) {
	if info.UserID == nil || *info.UserID == 0 {
		return []model.UserGameRecord{}, 0, nil
	}
	if info.Page <= 0 {
		info.Page = 1
	}
	if info.PageSize <= 0 {
		info.PageSize = 10
	}

	db := userGameRecordService.baseQuery(ctx).
		Where("d.user_id = ?", *info.UserID)

	if len(info.DateRange) == 0 && strings.TrimSpace(info.Date) != "" {
		statDate, dateErr := time.ParseInLocation("2006-01-02", strings.TrimSpace(info.Date), userGameRecordLocation())
		if dateErr != nil {
			return nil, 0, fmt.Errorf("查询日期格式错误")
		}
		dateKey := statDate.Format("2006-01-02")
		db = db.Where(`(
			d.stat_date = ?
			OR (
				COALESCE(gp.platform_code, '') = 'M7PP'
				AND (
					DATE(DATE_ADD(d.start_date, INTERVAL 8 HOUR)) = ?
					OR DATE(DATE_ADD(d.created_at, INTERVAL 8 HOUR)) = ?
				)
			)
		)`, dateKey, dateKey, dateKey)
	} else if startAt, endAt, rangeErr := parseUserGameRecordRange(info); rangeErr != nil {
		return nil, 0, rangeErr
	} else if startAt != nil && endAt != nil {
		db = db.Where("COALESCE(d.start_date, TIMESTAMP(d.stat_date)) >= ? AND COALESCE(d.start_date, TIMESTAMP(d.stat_date)) < ?", *startAt, *endAt)
	}

	if platformCode := normalizeUserGameRecordPlatform(info.PlatformCode); platformCode != "" {
		db = db.Where("COALESCE(gp.platform_code, m7gp.platform_code, g.platform_code, '') = ?", platformCode)
	}

	if err = db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err = db.Scopes(info.Paginate()).
		Order("COALESCE(d.start_date, d.created_at) DESC, d.id DESC").
		Scan(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (userGameRecordService *UserGameRecordService) baseQuery(ctx context.Context) *gorm.DB {
	return global.GVA_DB.WithContext(ctx).Table("user_game_transaction_details AS d").
		Select(`
			d.id,
			d.user_id,
			COALESCE(u.username, '') AS username,
			COALESCE(NULLIF(gp.platform_code, ''), NULLIF(m7gp.platform_code, ''), NULLIF(g.platform_code, ''), CASE WHEN m7gp.id IS NOT NULL THEN 'M7' ELSE 'HEDOC' END) AS platform_code,
			COALESCE(gp.id, m7gp.id, d.game_provider_code, 0) AS provider_id,
			COALESCE(gp.code, m7gp.code, '') AS provider_code,
			COALESCE(gp.name, m7gp.name, '') AS provider_name,
			d.game_code,
			COALESCE(
				NULLIF(JSON_UNQUOTE(JSON_EXTRACT(g.name, '$.CN')), ''),
				NULLIF(JSON_UNQUOTE(JSON_EXTRACT(g.name, '$.ZH')), ''),
				NULLIF(JSON_UNQUOTE(JSON_EXTRACT(g.name, '$.cn')), ''),
				NULLIF(JSON_UNQUOTE(JSON_EXTRACT(g.name, '$.zh')), ''),
				d.game_code
			) AS game_name_cn,
			COALESCE(
				NULLIF(JSON_UNQUOTE(JSON_EXTRACT(g.name, '$.EN')), ''),
				NULLIF(JSON_UNQUOTE(JSON_EXTRACT(g.name, '$.en')), ''),
				d.game_code
			) AS game_name_en,
			d.external_id,
			COALESCE(d.turnover, 0) AS turnover,
			COALESCE(d.bet, 0) AS bet,
			COALESCE(d.win, 0) AS win,
			COALESCE(d.winlose, 0) AS win_lose,
			d.start_date,
			d.end_date,
			d.stat_date,
			d.created_at
		`).
		Joins("LEFT JOIN users u ON u.id = d.user_id").
		Joins("LEFT JOIN game_providers gp ON gp.id = d.game_provider_code").
		Joins(`
			LEFT JOIN game_providers m7gp
				ON m7gp.platform_code = 'M7'
				AND m7gp.code = COALESCE(
					CASE WHEN JSON_VALID(d.remark) THEN JSON_UNQUOTE(JSON_EXTRACT(d.remark, '$.thirdPartyType')) ELSE NULL END,
					CASE WHEN JSON_VALID(d.remark) THEN JSON_UNQUOTE(JSON_EXTRACT(d.remark, '$.source')) ELSE NULL END
				)
		`).
		Joins(`
			LEFT JOIN games g
				ON g.game_code = d.game_code
				AND g.provider_id = COALESCE(NULLIF(d.game_provider_code, 0), m7gp.id)
				AND g.platform_code = COALESCE(NULLIF(gp.platform_code, ''), NULLIF(m7gp.platform_code, ''), g.platform_code)
		`)
}

func parseUserGameRecordRange(info exampleReq.UserGameRecordSearch) (*time.Time, *time.Time, error) {
	location := userGameRecordLocation()

	if len(info.DateRange) > 0 && len(info.DateRange) != 2 {
		return nil, nil, fmt.Errorf("查询时间段格式错误")
	}
	if len(info.DateRange) == 2 {
		startAt, err := parseUserGameRecordDateTime(info.DateRange[0], location)
		if err != nil {
			return nil, nil, fmt.Errorf("开始时间格式错误")
		}
		endAt, err := parseUserGameRecordDateTime(info.DateRange[1], location)
		if err != nil {
			return nil, nil, fmt.Errorf("结束时间格式错误")
		}
		if !endAt.After(startAt) {
			return nil, nil, fmt.Errorf("结束时间必须大于开始时间")
		}
		return &startAt, &endAt, nil
	}

	if strings.TrimSpace(info.Date) != "" {
		dayStart, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(info.Date), location)
		if err != nil {
			return nil, nil, fmt.Errorf("查询日期格式错误")
		}
		dayEnd := dayStart.Add(24 * time.Hour)
		return &dayStart, &dayEnd, nil
	}

	return nil, nil, nil
}

func userGameRecordLocation() *time.Location {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("CST", 8*3600)
	}
	return location
}

func parseUserGameRecordDateTime(value string, location *time.Location) (time.Time, error) {
	text := strings.TrimSpace(value)
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, text, location); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid datetime")
}

func normalizeUserGameRecordPlatform(value string) string {
	platformCode := strings.ToUpper(strings.TrimSpace(value))
	switch platformCode {
	case "M7", "M7PP", "HEDOC":
		return platformCode
	default:
		return ""
	}
}
