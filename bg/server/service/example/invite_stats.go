package example

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/example"
	exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
	"gorm.io/gorm"
)

type InviteStatsService struct{}

type inviterProfile struct {
	ID         uint64 `gorm:"column:id"`
	Username   string `gorm:"column:username"`
	InviteCode string `gorm:"column:invite_code"`
}

type inviteStatsDateTimeRange struct {
	Start time.Time
	End   time.Time
}

type inviteStatsRechargeCountRow struct {
	UserID uint64 `gorm:"column:user_id"`
	Count  int64  `gorm:"column:count"`
}

func (inviteStatsService *InviteStatsService) GetInviteStatsInfoList(ctx context.Context, info exampleReq.InviteStatsSearch) (list []example.InviteStats, total int64, summary example.InviteStatsSummary, err error) {
	inviter, found, err := inviteStatsService.resolveInviter(ctx, info)
	if err != nil || !found {
		return []example.InviteStats{}, 0, example.InviteStatsSummary{}, err
	}

	dateRange, err := normalizeInviteStatsDateRange(info.StatDateRange)
	if err != nil {
		return nil, 0, example.InviteStatsSummary{}, err
	}

	page, pageSize := normalizeInviteStatsPage(info.Page, info.PageSize)
	userDataService := UserDataService{}
	hasLastLoginAt, err := userDataService.columnExists(ctx, "users", "last_login_at")
	if err != nil {
		return nil, 0, example.InviteStatsSummary{}, err
	}
	hasLastLoginIP, err := userDataService.columnExists(ctx, "users", "last_login_ip")
	if err != nil {
		return nil, 0, example.InviteStatsSummary{}, err
	}
	hasRegisterIP, err := userDataService.columnExists(ctx, "users", "register_ip")
	if err != nil {
		return nil, 0, example.InviteStatsSummary{}, err
	}
	hasRegisterDomain, err := userDataService.columnExists(ctx, "users", "register_domain")
	if err != nil {
		return nil, 0, example.InviteStatsSummary{}, err
	}
	registerDomainSQL := userDataRegisterDomainSQL(hasRegisterDomain, "u")

	baseQuery := inviteStatsService.applyInviteeFilters(
		global.GVA_DB.WithContext(ctx).Table("users u"),
		inviter.ID,
		dateRange,
		info.RegisterDomain,
		registerDomainSQL,
	)
	if err = baseQuery.Count(&total).Error; err != nil {
		return nil, 0, example.InviteStatsSummary{}, err
	}

	summary, err = inviteStatsService.getInviteStatsSummary(ctx, inviter.ID, dateRange, info.RegisterDomain, registerDomainSQL, total)
	if err != nil {
		return nil, 0, example.InviteStatsSummary{}, err
	}

	if total == 0 {
		return []example.InviteStats{}, 0, summary, nil
	}

	selectParts := []string{
		"u.parent_id AS inviter_user_id",
		"u.id AS user_id",
		"u.username",
		"u.balance",
		"COALESCE(u.total_deposit, 0) AS recharge_amount",
		"u.created_at",
		inviteStatsRegisterIPSQL(hasRegisterIP, hasLastLoginIP, "u") + " AS register_ip",
		registerDomainSQL + " AS register_domain",
	}
	if hasLastLoginAt {
		selectParts = append(selectParts, "u.last_login_at")
	} else {
		selectParts = append(selectParts, "NULL AS last_login_at")
	}
	if hasLastLoginIP {
		selectParts = append(selectParts, "COALESCE(u.last_login_ip, '') AS last_login_ip")
	} else {
		selectParts = append(selectParts, "'' AS last_login_ip")
	}

	rows := make([]example.InviteStats, 0)
	err = global.GVA_DB.WithContext(ctx).
		Table("users u").
		Select(strings.Join(selectParts, ", ")).
		Scopes(func(db *gorm.DB) *gorm.DB {
			return inviteStatsService.applyInviteeFilters(db, inviter.ID, dateRange, info.RegisterDomain, registerDomainSQL)
		}).
		Order("u.created_at DESC, u.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, example.InviteStatsSummary{}, err
	}

	userIDs := make([]uint64, 0, len(rows))
	registerIPs := make([]string, 0, len(rows))
	for _, row := range rows {
		userIDs = append(userIDs, row.UserID)
		if registerIP := strings.TrimSpace(row.RegisterIP); registerIP != "" {
			registerIPs = append(registerIPs, registerIP)
		}
	}

	rechargeCountMap, err := inviteStatsService.getRechargeCountMap(ctx, userIDs)
	if err != nil {
		return nil, 0, example.InviteStatsSummary{}, err
	}
	sameRegisterIPCountMap, err := userDataService.getSameRegisterIPCountMap(ctx, registerIPs, inviteStatsRegisterIPSQL(hasRegisterIP, hasLastLoginIP, ""))
	if err != nil {
		return nil, 0, example.InviteStatsSummary{}, err
	}

	for i := range rows {
		rows[i].RegisterIP = strings.TrimSpace(rows[i].RegisterIP)
		rows[i].RegisterDomain = strings.TrimSpace(rows[i].RegisterDomain)
		rows[i].LastLoginIP = strings.TrimSpace(rows[i].LastLoginIP)
		rows[i].RechargeCount = rechargeCountMap[rows[i].UserID]
		rows[i].SameRegisterIPCount = sameRegisterIPCountMap[rows[i].RegisterIP]
		rows[i].IPCountryNote = userDataService.getIPCountryNote(ctx, rows[i].LastLoginIP)
	}

	return rows, total, summary, nil
}

func (inviteStatsService *InviteStatsService) applyInviteeFilters(db *gorm.DB, inviterID uint64, dateRange *inviteStatsDateTimeRange, registerDomain string, registerDomainSQL string) *gorm.DB {
	if inviterID > 0 {
		db = db.Where("u.parent_id = ?", inviterID)
	}
	if domain := normalizeUserDataDomain(registerDomain); domain != "" {
		db = db.Where("LOWER("+registerDomainSQL+") = ?", domain)
	}
	if dateRange != nil {
		db = db.Where("u.created_at >= ? AND u.created_at < ?", dateRange.Start, dateRange.End)
	}
	return db
}

func (inviteStatsService *InviteStatsService) getRechargeCountMap(ctx context.Context, userIDs []uint64) (map[uint64]int64, error) {
	result := make(map[uint64]int64)
	if len(userIDs) == 0 {
		return result, nil
	}

	rows := make([]inviteStatsRechargeCountRow, 0)
	err := global.GVA_DB.WithContext(ctx).
		Table("payment_orders").
		Select("user_id, COUNT(*) AS count").
		Where("user_id IN ? AND status = ?", userIDs, 1).
		Group("user_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		result[row.UserID] = row.Count
	}
	return result, nil
}

func (inviteStatsService *InviteStatsService) getInviteStatsSummary(ctx context.Context, inviterID uint64, dateRange *inviteStatsDateTimeRange, registerDomain string, registerDomainSQL string, registerCount int64) (example.InviteStatsSummary, error) {
	summary := example.InviteStatsSummary{
		RegisterCount: registerCount,
	}

	currencySQL := exampleCurrencyCaseSQL("po.dst_code")
	rateFallbackSQL := exampleRateFallbackSQL("po.dst_code")
	query := global.GVA_DB.WithContext(ctx).
		Table("payment_orders po").
		Joins("JOIN users u ON u.id = po.user_id").
		Joins(`
			LEFT JOIN (
				SELECT user_id, reference_id, SUM(amount) AS amount_u
				FROM transactions
				WHERE type = 1
				GROUP BY user_id, reference_id
			) tx ON tx.user_id = po.user_id AND tx.reference_id = po.order_id
		`).
		Joins("LEFT JOIN exchange_rates er ON er.code = "+currencySQL).
		Where("po.status = ? AND po.paid_at IS NOT NULL", 1)
	if inviterID > 0 {
		query = query.Where("u.parent_id = ?", inviterID)
	}
	if domain := normalizeUserDataDomain(registerDomain); domain != "" {
		query = query.Where("LOWER("+registerDomainSQL+") = ?", domain)
	}

	if dateRange != nil {
		query = query.Where("po.paid_at >= ? AND po.paid_at < ?", dateRange.Start, dateRange.End)
	}

	err := query.Select(`
			COALESCE(COUNT(DISTINCT po.user_id), 0) AS recharge_user_count,
			COALESCE(SUM(
				COALESCE(
					NULLIF(tx.amount_u, 0),
					po.amount / COALESCE(NULLIF(er.rate, 0), ` + rateFallbackSQL + `)
				)
			), 0) AS recharge_amount_u
		`).
		Scan(&summary).Error
	if err != nil {
		return example.InviteStatsSummary{}, err
	}

	summary.RegisterCount = registerCount
	return summary, nil
}

func (inviteStatsService *InviteStatsService) resolveInviter(ctx context.Context, info exampleReq.InviteStatsSearch) (inviter inviterProfile, found bool, err error) {
	inviteCode := strings.TrimSpace(info.InviteCode)
	hasInviterUserID := info.InviterUserId != nil && *info.InviterUserId > 0
	if !hasInviterUserID && inviteCode == "" {
		return inviterProfile{}, normalizeUserDataDomain(info.RegisterDomain) != "", nil
	}

	db := global.GVA_DB.WithContext(ctx).Table("users").Select("id, username, invite_code")
	if hasInviterUserID {
		err = db.Where("id = ?", *info.InviterUserId).Take(&inviter).Error
	} else {
		err = db.Where("invite_code = ?", inviteCode).Take(&inviter).Error
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return inviterProfile{}, false, fmt.Errorf("邀请人不存在")
	}
	if err != nil {
		return inviterProfile{}, false, err
	}
	return inviter, true, nil
}

func normalizeInviteStatsDateRange(dateRange []string) (*inviteStatsDateTimeRange, error) {
	if len(dateRange) == 0 {
		return nil, nil
	}
	if len(dateRange) != 2 {
		return nil, fmt.Errorf("日期范围格式不正确")
	}

	startAt, _, err := parseInviteStatsDateTime(dateRange[0])
	if err != nil {
		return nil, err
	}
	endAt, endIsDateOnly, err := parseInviteStatsDateTime(dateRange[1])
	if err != nil {
		return nil, err
	}
	if endIsDateOnly {
		endAt = endAt.Add(24 * time.Hour)
	}
	if startAt.After(endAt) {
		startAt, endAt = endAt, startAt
	}
	return &inviteStatsDateTimeRange{Start: startAt, End: endAt}, nil
}

func parseInviteStatsDateTime(raw string) (time.Time, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false, fmt.Errorf("日期不能为空")
	}

	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		location = time.FixedZone("CST", 8*3600)
	}

	layouts := []struct {
		layout   string
		dateOnly bool
	}{
		{time.RFC3339, false},
		{"2006-01-02 15:04:05", false},
		{"2006-01-02 15:04", false},
		{"2006-01-02T15:04:05", false},
		{"2006-01-02T15:04", false},
		{"2006-01-02", true},
	}

	for _, item := range layouts {
		var parsed time.Time
		var parseErr error
		if item.layout == time.RFC3339 {
			parsed, parseErr = time.Parse(item.layout, raw)
		} else {
			parsed, parseErr = time.ParseInLocation(item.layout, raw, location)
		}
		if parseErr == nil {
			return parsed, item.dateOnly, nil
		}
	}

	if len(raw) >= 10 {
		if parsed, err := time.ParseInLocation("2006-01-02", raw[:10], location); err == nil {
			return parsed, true, nil
		}
	}

	return time.Time{}, false, fmt.Errorf("无法解析日期: %s", raw)
}

func inviteStatsRegisterIPSQL(hasRegisterIP bool, hasLastLoginIP bool, qualifier string) string {
	column := func(name string) string {
		if strings.TrimSpace(qualifier) == "" {
			return name
		}
		return qualifier + "." + name
	}

	switch {
	case hasRegisterIP && hasLastLoginIP:
		return fmt.Sprintf("COALESCE(NULLIF(%s, ''), %s, '')", column("register_ip"), column("last_login_ip"))
	case hasRegisterIP:
		return fmt.Sprintf("COALESCE(%s, '')", column("register_ip"))
	case hasLastLoginIP:
		return fmt.Sprintf("COALESCE(%s, '')", column("last_login_ip"))
	default:
		return "''"
	}
}

func normalizeInviteStatsPage(page int, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	switch {
	case pageSize > 100:
		pageSize = 100
	case pageSize <= 0:
		pageSize = 10
	}
	return page, pageSize
}
