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

type RechargeQueryService struct{}

func (rechargeQueryService *RechargeQueryService) GetRechargeQueryList(ctx context.Context, info exampleReq.RechargeQuerySearch) (list []model.RechargeQueryRecord, total int64, err error) {
	if info.Page <= 0 {
		info.Page = 1
	}
	if info.PageSize <= 0 {
		info.PageSize = 10
	}

	startAt, endAt, err := parseRechargeQueryRange(info)
	if err != nil {
		return nil, 0, err
	}

	db := rechargeQueryService.applyRechargeQueryFilters(
		global.GVA_DB.WithContext(ctx).Table("payment_orders po").
			Joins("LEFT JOIN users u ON u.id = po.user_id").
			Joins(`
				LEFT JOIN (
					SELECT user_id, reference_id, SUM(amount) AS amount_u
					FROM transactions
					WHERE type = 1
					GROUP BY user_id, reference_id
				) tx ON tx.user_id = po.user_id AND tx.reference_id = po.order_id
			`).
			Where("po.status = ? AND po.paid_at IS NOT NULL", 1),
		info,
		startAt,
		endAt,
	)

	if err = db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []model.RechargeQueryRecord{}, 0, nil
	}

	currencySQL := exampleCurrencyCaseSQL("po.dst_code")
	err = db.Select(`
			po.id,
			po.user_id,
			u.username,
			po.order_id,
			po.plat_order_id,
			po.amount AS local_amount,
			` + currencySQL + ` AS currency,
			COALESCE(tx.amount_u, 0) AS recharge_amount_u,
			COALESCE(u.total_deposit, 0) AS total_recharge_amount,
			po.type,
			po.dst_code,
			po.paid_at,
			po.created_at
		`).
		Order("po.paid_at DESC, po.id DESC").
		Scopes(info.Paginate()).
		Scan(&list).Error
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (rechargeQueryService *RechargeQueryService) applyRechargeQueryFilters(db *gorm.DB, info exampleReq.RechargeQuerySearch, startAt *time.Time, endAt *time.Time) *gorm.DB {
	if info.UserID != nil && *info.UserID > 0 {
		db = db.Where("po.user_id = ?", *info.UserID)
	}
	if startAt != nil && endAt != nil {
		db = db.Where("po.paid_at >= ? AND po.paid_at < ?", *startAt, *endAt)
	}
	return db
}

func parseRechargeQueryRange(info exampleReq.RechargeQuerySearch) (*time.Time, *time.Time, error) {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		location = time.FixedZone("CST", 8*3600)
	}

	if len(info.DateRange) > 0 && len(info.DateRange) != 2 {
		return nil, nil, fmt.Errorf("充值时间段格式错误")
	}

	if len(info.DateRange) == 2 {
		startAt, err := parseDateTimeInLocation(info.DateRange[0], location)
		if err != nil {
			return nil, nil, fmt.Errorf("充值开始时间格式错误")
		}
		endAt, err := parseDateTimeInLocation(info.DateRange[1], location)
		if err != nil {
			return nil, nil, fmt.Errorf("充值结束时间格式错误")
		}
		if !endAt.After(startAt) {
			return nil, nil, fmt.Errorf("充值结束时间必须大于开始时间")
		}
		return &startAt, &endAt, nil
	}

	if strings.TrimSpace(info.Date) != "" {
		dayStart, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(info.Date), location)
		if err != nil {
			return nil, nil, fmt.Errorf("充值日期格式错误")
		}
		dayEnd := dayStart.Add(24 * time.Hour)
		return &dayStart, &dayEnd, nil
	}

	if info.UserID != nil && *info.UserID > 0 {
		return nil, nil, nil
	}

	now := time.Now().In(location)
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	dayEnd := dayStart.Add(24 * time.Hour)
	return &dayStart, &dayEnd, nil
}
