package example

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	model "github.com/flipped-aurora/gin-vue-admin/server/model/example"
	exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserDataService struct{}

var userDataGeoIPHTTPClient = &http.Client{
	Timeout: 1500 * time.Millisecond,
}

var userDataGeoIPCache sync.Map

type userDataBaseRow struct {
	ID             uint64     `gorm:"column:id"`
	Username       string     `gorm:"column:username"`
	Balance        float64    `gorm:"column:balance"`
	TotalDeposit   float64    `gorm:"column:total_deposit"`
	TotalWithdraw  float64    `gorm:"column:total_withdraw"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	RegisterIP     string     `gorm:"column:register_ip"`
	RegisterDomain string     `gorm:"column:register_domain"`
	LastLoginAt    *time.Time `gorm:"column:last_login_at"`
	LastLoginIP    string     `gorm:"column:last_login_ip"`
}

type userDataStatRow struct {
	UserID   uint64  `gorm:"column:user_id"`
	Turnover float64 `gorm:"column:turnover"`
	Winlose  float64 `gorm:"column:winlose"`
	RowCount int64   `gorm:"column:row_count"`
}

type userDataRechargeCountRow struct {
	UserID uint64 `gorm:"column:user_id"`
	Count  int64  `gorm:"column:count"`
}

type userDataGameCountRow struct {
	UserID         uint64 `gorm:"column:user_id"`
	TotalCount     int64  `gorm:"column:total_count"`
	SlotCount      int64  `gorm:"column:slot_count"`
	CasinoCount    int64  `gorm:"column:casino_count"`
	SportbookCount int64  `gorm:"column:sportbook_count"`
	OtherCount     int64  `gorm:"column:other_count"`
}

type userDataRegisterIPCountRow struct {
	RegisterIP string `gorm:"column:register_ip"`
	Count      int64  `gorm:"column:count"`
}

type userDataGeoIPResponse struct {
	Success       *bool  `json:"success"`
	Status        string `json:"status"`
	CountryCode   string `json:"country_code"`
	CountryCode2  string `json:"countryCode"`
	Country       string `json:"country"`
	Error         bool   `json:"error"`
	ErrorMessage  string `json:"error_message"`
	ErrorMessage2 string `json:"message"`
}

type userDataGeoIPLookupResult struct {
	CountryCode string
	CountryName string
}

type userDataDateTimeRange struct {
	Start        string
	End          string
	EndInclusive bool
}

func (userDataService *UserDataService) GetUserDataList(ctx context.Context, info exampleReq.UserDataSearch) (list []model.UserDataRecord, total int64, err error) {
	if info.Page <= 0 {
		info.Page = 1
	}
	if info.PageSize <= 0 {
		info.PageSize = 10
	}

	registerRange, err := parseRegisterRange(info)
	if err != nil {
		return nil, 0, err
	}
	rechargeRange, err := parseRechargeRange(info)
	if err != nil {
		return nil, 0, err
	}
	hasLastLoginAt, err := userDataService.columnExists(ctx, "users", "last_login_at")
	if err != nil {
		return nil, 0, err
	}
	hasLastLoginIP, err := userDataService.columnExists(ctx, "users", "last_login_ip")
	if err != nil {
		return nil, 0, err
	}
	hasRegisterIP, err := userDataService.columnExists(ctx, "users", "register_ip")
	if err != nil {
		return nil, 0, err
	}
	hasRegisterDomain, err := userDataService.columnExists(ctx, "users", "register_domain")
	if err != nil {
		return nil, 0, err
	}
	registerIPSQL := userDataService.registerIPSQL(hasRegisterIP, hasLastLoginIP)
	registerDomainSQL := userDataRegisterDomainSQL(hasRegisterDomain, "")

	baseQuery := userDataService.applyUserDataFilters(
		global.GVA_DB.WithContext(ctx).Table("users"),
		info,
		registerRange,
		rechargeRange,
		registerIPSQL,
		registerDomainSQL,
	)

	if err = baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []model.UserDataRecord{}, 0, nil
	}

	selectParts := []string{
		"id",
		"username",
		"balance",
		"COALESCE(total_deposit, 0) AS total_deposit",
		"COALESCE(total_withdraw, 0) AS total_withdraw",
		"created_at",
		registerIPSQL + " AS register_ip",
		registerDomainSQL + " AS register_domain",
	}
	if hasLastLoginAt {
		selectParts = append(selectParts, "last_login_at")
	} else {
		selectParts = append(selectParts, "NULL AS last_login_at")
	}
	if hasLastLoginIP {
		selectParts = append(selectParts, "COALESCE(last_login_ip, '') AS last_login_ip")
	} else {
		selectParts = append(selectParts, "'' AS last_login_ip")
	}

	var users []userDataBaseRow
	err = global.GVA_DB.WithContext(ctx).
		Table("users").
		Select(strings.Join(selectParts, ", ")).
		Scopes(func(db *gorm.DB) *gorm.DB {
			return userDataService.applyUserDataFilters(db, info, registerRange, rechargeRange, registerIPSQL, registerDomainSQL)
		}).
		Order("created_at DESC, id DESC").
		Offset((info.Page - 1) * info.PageSize).
		Limit(info.PageSize).
		Scan(&users).Error
	if err != nil {
		return nil, 0, err
	}

	userIDs := make([]uint64, 0, len(users))
	registerIPs := make([]string, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
		if registerIP := strings.TrimSpace(user.RegisterIP); registerIP != "" {
			registerIPs = append(registerIPs, registerIP)
		}
	}

	statMap, err := userDataService.getUserStatMap(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}
	rechargeMap, err := userDataService.getRechargeCountMap(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}
	gameCountMap, err := userDataService.getGameCountMap(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}
	sameRegisterIPCountMap, err := userDataService.getSameRegisterIPCountMap(ctx, registerIPs, registerIPSQL)
	if err != nil {
		return nil, 0, err
	}

	records := make([]model.UserDataRecord, 0, len(users))
	for _, userRow := range users {
		statRow := statMap[userRow.ID]
		gameCounts := gameCountMap[userRow.ID]
		registerIP := strings.TrimSpace(userRow.RegisterIP)
		if registerIP == "" {
			registerIP = strings.TrimSpace(userRow.LastLoginIP)
		}

		records = append(records, model.UserDataRecord{
			UserID:              userRow.ID,
			Username:            userRow.Username,
			Balance:             userRow.Balance,
			TotalTurnover:       statRow.Turnover,
			TotalWinlose:        statRow.Winlose,
			RechargeCount:       rechargeMap[userRow.ID],
			RechargeAmount:      userRow.TotalDeposit,
			WithdrawAmount:      userRow.TotalWithdraw,
			TotalGameCount:      gameCounts.TotalCount,
			SlotGameCount:       gameCounts.SlotCount,
			CasinoGameCount:     gameCounts.CasinoCount,
			SportbookGameCount:  gameCounts.SportbookCount,
			OtherGameCount:      gameCounts.OtherCount,
			CreatedAt:           userRow.CreatedAt,
			RegisterIP:          registerIP,
			RegisterDomain:      strings.TrimSpace(userRow.RegisterDomain),
			SameRegisterIPCount: sameRegisterIPCountMap[registerIP],
			LastLoginAt:         userRow.LastLoginAt,
			LastLoginIP:         strings.TrimSpace(userRow.LastLoginIP),
			IPCountryNote:       userDataService.getIPCountryNote(ctx, userRow.LastLoginIP),
		})
	}

	return records, total, nil
}

func (userDataService *UserDataService) applyUserDataFilters(db *gorm.DB, info exampleReq.UserDataSearch, registerRange *userDataDateTimeRange, rechargeRange *userDataDateTimeRange, registerIPSQL string, registerDomainSQL string) *gorm.DB {
	if info.UserID != nil && *info.UserID > 0 {
		db = db.Where("id = ?", *info.UserID)
	}
	if username := strings.TrimSpace(info.Username); username != "" {
		db = db.Where("username LIKE ?", "%"+username+"%")
	}
	if registerIP := strings.TrimSpace(info.RegisterIP); registerIP != "" {
		db = db.Where(registerIPSQL+" = ?", registerIP)
	}
	if registerDomain := normalizeUserDataDomain(info.RegisterDomain); registerDomain != "" {
		db = db.Where("LOWER("+registerDomainSQL+") = ?", registerDomain)
	}
	if registerRange != nil {
		if registerRange.EndInclusive {
			db = db.Where("created_at >= ? AND created_at <= ?", registerRange.Start, registerRange.End)
		} else {
			db = db.Where("created_at >= ? AND created_at < ?", registerRange.Start, registerRange.End)
		}
	}
	if rechargeRange != nil {
		db = db.Where(
			"EXISTS (SELECT 1 FROM payment_orders po WHERE po.user_id = users.id AND po.status = ? AND po.paid_at IS NOT NULL AND po.paid_at >= ? AND po.paid_at < ?)",
			1,
			rechargeRange.Start,
			rechargeRange.End,
		)
	}
	return db
}

func normalizeUserDataDomain(raw string) string {
	domain := strings.TrimSpace(strings.ToLower(raw))
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	if slash := strings.IndexByte(domain, '/'); slash >= 0 {
		domain = domain[:slash]
	}
	if colon := strings.IndexByte(domain, ':'); colon >= 0 {
		domain = domain[:colon]
	}
	domain = strings.TrimSuffix(domain, ".")
	return strings.TrimPrefix(domain, "www.")
}

func userDataRegisterDomainSQL(hasRegisterDomain bool, qualifier string) string {
	if !hasRegisterDomain {
		return "''"
	}
	if strings.TrimSpace(qualifier) == "" {
		return "COALESCE(register_domain, '')"
	}
	return "COALESCE(" + qualifier + ".register_domain, '')"
}

func newUserDataDateTimeRange(startAt time.Time, endAt time.Time) *userDataDateTimeRange {
	return &userDataDateTimeRange{
		Start: startAt.Format("2006-01-02 15:04:05"),
		End:   endAt.Format("2006-01-02 15:04:05"),
	}
}

func parseRegisterRange(info exampleReq.UserDataSearch) (*userDataDateTimeRange, error) {
	if len(info.RegisterDateRange) == 2 {
		startAt, err := parseDateTimeInLocation(info.RegisterDateRange[0], time.Local)
		if err != nil {
			return nil, fmt.Errorf("开始时间格式错误")
		}
		endAt, err := parseDateTimeInLocation(info.RegisterDateRange[1], time.Local)
		if err != nil {
			return nil, fmt.Errorf("结束时间格式错误")
		}
		if !endAt.After(startAt) {
			return nil, fmt.Errorf("结束时间必须大于开始时间")
		}
		rangeValue := newUserDataDateTimeRange(startAt, endAt)
		rangeValue.EndInclusive = true
		return rangeValue, nil
	}

	if strings.TrimSpace(info.RegisterDate) != "" {
		dayStart, err := time.Parse("2006-01-02", strings.TrimSpace(info.RegisterDate))
		if err != nil {
			return nil, fmt.Errorf("查询日期格式错误")
		}
		return newUserDataDateTimeRange(dayStart, dayStart.Add(24*time.Hour)), nil
	}

	if info.UserID != nil && *info.UserID > 0 {
		return nil, nil
	}

	if strings.TrimSpace(info.Username) != "" {
		return nil, nil
	}

	if strings.TrimSpace(info.RegisterIP) != "" {
		return nil, nil
	}

	if hasRechargeSearch(info) {
		return nil, nil
	}

	now := time.Now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	return newUserDataDateTimeRange(dayStart, dayStart.Add(24*time.Hour)), nil
}

func parseRechargeRange(info exampleReq.UserDataSearch) (*userDataDateTimeRange, error) {
	if len(info.RechargeDateRange) > 0 && len(info.RechargeDateRange) != 2 {
		return nil, fmt.Errorf("充值时间段格式错误")
	}

	if len(info.RechargeDateRange) == 2 {
		startAt, err := parseDateTimeInLocation(info.RechargeDateRange[0], time.Local)
		if err != nil {
			return nil, fmt.Errorf("充值开始时间格式错误")
		}
		endAt, err := parseDateTimeInLocation(info.RechargeDateRange[1], time.Local)
		if err != nil {
			return nil, fmt.Errorf("充值结束时间格式错误")
		}
		if !endAt.After(startAt) {
			return nil, fmt.Errorf("充值结束时间必须大于开始时间")
		}
		return newUserDataDateTimeRange(startAt, endAt), nil
	}

	if strings.TrimSpace(info.RechargeDate) != "" {
		dayStart, err := time.Parse("2006-01-02", strings.TrimSpace(info.RechargeDate))
		if err != nil {
			return nil, fmt.Errorf("充值日期格式错误")
		}
		return newUserDataDateTimeRange(dayStart, dayStart.Add(24*time.Hour)), nil
	}

	return nil, nil
}
func hasRechargeSearch(info exampleReq.UserDataSearch) bool {
	return strings.TrimSpace(info.RechargeDate) != "" || len(info.RechargeDateRange) > 0
}

func parseDateTimeInLocation(value string, location *time.Location) (time.Time, error) {
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

func (userDataService *UserDataService) getUserStatMap(ctx context.Context, userIDs []uint64) (map[uint64]userDataStatRow, error) {
	result := make(map[uint64]userDataStatRow)
	if len(userIDs) == 0 {
		return result, nil
	}

	detailRows := make([]userDataStatRow, 0)
	err := global.GVA_DB.WithContext(ctx).
		Table("user_game_transaction_details").
		Select(`
			user_id,
			COALESCE(SUM(bet), 0) AS turnover,
			COALESCE(SUM(winlose), 0) AS winlose,
			COALESCE(SUM(CASE WHEN turnover <> 0 OR bet <> 0 OR win <> 0 OR winlose <> 0 THEN 1 ELSE 0 END), 0) AS row_count
		`).
		Where("user_id IN ?", userIDs).
		Group("user_id").
		Scan(&detailRows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range detailRows {
		if row.RowCount > 0 {
			result[row.UserID] = row
		}
	}

	rows := make([]userDataStatRow, 0)
	err = global.GVA_DB.WithContext(ctx).
		Table("user_game_transaction_stats").
		Select(`
			user_id,
			COALESCE(SUM(bet), 0) AS turnover,
			COALESCE(SUM(winlose), 0) AS winlose,
			COALESCE(SUM(count), 0) AS row_count
		`).
		Where("user_id IN ? AND period_type = ?", userIDs, "daily").
		Group("user_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		if existing, ok := result[row.UserID]; !ok || existing.RowCount == 0 {
			result[row.UserID] = row
		}
	}

	totalRows := make([]userDataStatRow, 0)
	err = global.GVA_DB.WithContext(ctx).
		Table("user_game_transaction_stats").
		Select("user_id, COALESCE(bet, 0) AS turnover, COALESCE(winlose, 0) AS winlose, COALESCE(count, 0) AS row_count").
		Where("user_id IN ? AND period_type = ? AND period_key = ?", userIDs, "total", "all").
		Scan(&totalRows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range totalRows {
		if existing, ok := result[row.UserID]; !ok || existing.RowCount == 0 {
			result[row.UserID] = row
		}
	}
	return result, nil
}

func (userDataService *UserDataService) getRechargeCountMap(ctx context.Context, userIDs []uint64) (map[uint64]int64, error) {
	rows := make([]userDataRechargeCountRow, 0)
	err := global.GVA_DB.WithContext(ctx).
		Table("payment_orders").
		Select("user_id, COUNT(*) AS count").
		Where("user_id IN ? AND status = ?", userIDs, 1).
		Group("user_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uint64]int64, len(rows))
	for _, row := range rows {
		result[row.UserID] = row.Count
	}
	return result, nil
}

func (userDataService *UserDataService) getGameCountMap(ctx context.Context, userIDs []uint64) (map[uint64]userDataGameCountRow, error) {
	rows := make([]userDataGameCountRow, 0)
	sql := `
SELECT
	c.user_id,
	COALESCE(SUM(CASE WHEN c.category <> '' THEN 1 ELSE 0 END), 0) AS total_count,
	COALESCE(SUM(CASE WHEN c.category = 'SLOT' THEN 1 ELSE 0 END), 0) AS slot_count,
	COALESCE(SUM(CASE WHEN c.category = 'CASINO' THEN 1 ELSE 0 END), 0) AS casino_count,
	COALESCE(SUM(CASE WHEN c.category = 'SPORTBOOK' THEN 1 ELSE 0 END), 0) AS sportbook_count,
	COALESCE(SUM(CASE WHEN c.category = 'OTHER' THEN 1 ELSE 0 END), 0) AS other_count
FROM (
	SELECT
		d.user_id,
		CASE
			WHEN d.turnover = 0 AND d.bet = 0 AND d.win = 0 AND d.winlose = 0 THEN ''
			WHEN LOCATE('_SLOT', UPPER(CONCAT_WS('|',
				COALESCE(gp.code, ''),
				COALESCE(m7gp.code, ''),
				COALESCE(JSON_UNQUOTE(JSON_EXTRACT(d.remark, '$.thirdPartyType')), ''),
				COALESCE(JSON_UNQUOTE(JSON_EXTRACT(d.remark, '$.source')), '')
			))) > 0 THEN 'SLOT'
			WHEN LOCATE('_LIVE', UPPER(CONCAT_WS('|',
				COALESCE(gp.code, ''),
				COALESCE(m7gp.code, ''),
				COALESCE(JSON_UNQUOTE(JSON_EXTRACT(d.remark, '$.thirdPartyType')), ''),
				COALESCE(JSON_UNQUOTE(JSON_EXTRACT(d.remark, '$.source')), '')
			))) > 0 THEN 'CASINO'
			WHEN LOCATE('_SPORT', UPPER(CONCAT_WS('|',
				COALESCE(gp.code, ''),
				COALESCE(m7gp.code, ''),
				COALESCE(JSON_UNQUOTE(JSON_EXTRACT(d.remark, '$.thirdPartyType')), ''),
				COALESCE(JSON_UNQUOTE(JSON_EXTRACT(d.remark, '$.source')), '')
			))) > 0 THEN 'SPORTBOOK'
			WHEN gt.code IN ('SLOT', 'CASINO', 'SPORTBOOK') THEN gt.code
			ELSE 'OTHER'
		END AS category
	FROM user_game_transaction_details d
	LEFT JOIN game_providers gp
		ON gp.id = d.game_provider_code
	LEFT JOIN game_providers m7gp
		ON m7gp.platform_code = 'M7'
		AND m7gp.code = COALESCE(
			JSON_UNQUOTE(JSON_EXTRACT(d.remark, '$.thirdPartyType')),
			JSON_UNQUOTE(JSON_EXTRACT(d.remark, '$.source'))
		)
	LEFT JOIN games g
		ON g.game_code = d.game_code
		AND g.provider_id = COALESCE(NULLIF(d.game_provider_code, 0), m7gp.id)
		AND g.platform_code = COALESCE(gp.platform_code, m7gp.platform_code, 'HEDOC')
	LEFT JOIN game_types gt
		ON gt.id = COALESCE(g.game_type_id, gp.type_id, m7gp.type_id)
	WHERE d.user_id IN ?
) c
GROUP BY c.user_id
`
	err := global.GVA_DB.WithContext(ctx).Raw(sql, userIDs).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uint64]userDataGameCountRow, len(rows))
	for _, row := range rows {
		result[row.UserID] = row
	}
	return result, nil
}

func (userDataService *UserDataService) getSameRegisterIPCountMap(ctx context.Context, registerIPs []string, registerIPSQL string) (map[string]int64, error) {
	result := make(map[string]int64)
	if len(registerIPs) == 0 {
		return result, nil
	}

	seen := make(map[string]struct{}, len(registerIPs))
	uniqueIPs := make([]string, 0, len(registerIPs))
	for _, rawIP := range registerIPs {
		ip := strings.TrimSpace(rawIP)
		if ip == "" {
			continue
		}
		if _, exists := seen[ip]; exists {
			continue
		}
		seen[ip] = struct{}{}
		uniqueIPs = append(uniqueIPs, ip)
	}
	if len(uniqueIPs) == 0 {
		return result, nil
	}

	rows := make([]userDataRegisterIPCountRow, 0)
	err := global.GVA_DB.WithContext(ctx).
		Table("users").
		Select(registerIPSQL+" AS register_ip, COUNT(*) AS count").
		Where(registerIPSQL+" IN ?", uniqueIPs).
		Group("register_ip").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		if registerIP := strings.TrimSpace(row.RegisterIP); registerIP != "" {
			result[registerIP] = row.Count
		}
	}
	return result, nil
}

func (userDataService *UserDataService) registerIPSQL(hasRegisterIP bool, hasLastLoginIP bool) string {
	switch {
	case hasRegisterIP && hasLastLoginIP:
		return "COALESCE(NULLIF(register_ip, ''), last_login_ip, '')"
	case hasRegisterIP:
		return "COALESCE(register_ip, '')"
	case hasLastLoginIP:
		return "COALESCE(last_login_ip, '')"
	default:
		return "''"
	}
}

func (userDataService *UserDataService) columnExists(ctx context.Context, tableName string, columnName string) (bool, error) {
	var count int64
	err := global.GVA_DB.WithContext(ctx).
		Table("INFORMATION_SCHEMA.COLUMNS").
		Where("TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?", tableName, columnName).
		Count(&count).Error
	return count > 0, err
}

func (userDataService *UserDataService) getIPCountryNote(ctx context.Context, rawIP string) string {
	ipText := strings.TrimSpace(rawIP)
	if ipText == "" {
		return "未知"
	}

	ip := net.ParseIP(ipText)
	if ip == nil {
		return "未知"
	}

	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() {
		return "内网/其他"
	}

	if cached, ok := userDataGeoIPCache.Load(ipText); ok {
		if note, ok := cached.(string); ok && strings.TrimSpace(note) != "" && note != "未知" {
			return note
		}
		userDataGeoIPCache.Delete(ipText)
	}

	location, err := userDataService.lookupCountry(ctx, ipText)
	if err != nil {
		global.GVA_LOG.Warn("geoip lookup failed", zap.String("ip", ipText), zap.Error(err))
		return "未知"
	}

	note := "未知"
	countryCode := strings.ToUpper(strings.TrimSpace(location.CountryCode))
	switch countryCode {
	case "ID":
		note = "印尼"
	case "PH":
		note = "菲律宾"
	case "":
		note = "未知"
	default:
		countryName := strings.TrimSpace(location.CountryName)
		if countryName != "" {
			note = countryName + " (" + countryCode + ")"
		} else {
			note = countryCode
		}
	}
	if note != "未知" {
		userDataGeoIPCache.Store(ipText, note)
	}
	return note
}

func (userDataService *UserDataService) lookupCountry(ctx context.Context, ipText string) (userDataGeoIPLookupResult, error) {
	baseURL := strings.TrimSpace(global.GVA_CONFIG.CustomCfg.GeoIPURL)
	if baseURL == "" {
		baseURL = "https://ipwho.is/"
	}

	candidates := []string{baseURL}
	if !strings.Contains(baseURL, "ip-api.com") {
		candidates = append(candidates, "http://ip-api.com/json/")
	} else if strings.HasPrefix(strings.ToLower(baseURL), "https://ip-api.com/") {
		candidates = append(candidates, "http://"+strings.TrimPrefix(baseURL, "https://"))
	}

	timeout := global.GVA_CONFIG.CustomCfg.GeoIPTimeoutMs
	if timeout <= 0 {
		timeout = 3000
	}

	client := *userDataGeoIPHTTPClient
	client.Timeout = time.Duration(timeout) * time.Millisecond

	var firstErr error
	for _, currentBaseURL := range candidates {
		location, err := userDataService.lookupCountryFromEndpoint(ctx, &client, currentBaseURL, ipText)
		if err == nil {
			return location, nil
		}
		if firstErr == nil {
			firstErr = err
		}
		global.GVA_LOG.Warn("geoip endpoint failed", zap.String("endpoint", currentBaseURL), zap.String("ip", ipText), zap.Error(err))
	}

	if firstErr != nil {
		return userDataGeoIPLookupResult{}, firstErr
	}
	return userDataGeoIPLookupResult{}, fmt.Errorf("geoip lookup failed")
}

func (userDataService *UserDataService) lookupCountryFromEndpoint(ctx context.Context, client *http.Client, baseURL string, ipText string) (userDataGeoIPLookupResult, error) {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return userDataGeoIPLookupResult{}, fmt.Errorf("geoip endpoint is empty")
	}
	if !strings.HasSuffix(baseURL, "/") && !strings.Contains(baseURL, "?") {
		baseURL += "/"
	}

	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, client.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, buildGeoIPRequestURL(baseURL, ipText), nil)
	if err != nil {
		return userDataGeoIPLookupResult{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return userDataGeoIPLookupResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return userDataGeoIPLookupResult{}, fmt.Errorf("geoip status %d", resp.StatusCode)
	}

	var payload userDataGeoIPResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return userDataGeoIPLookupResult{}, err
	}

	if payload.Success != nil && !*payload.Success {
		return userDataGeoIPLookupResult{}, fmt.Errorf("geoip lookup failed: %s", firstNonEmptyString(payload.ErrorMessage, payload.ErrorMessage2))
	}
	if strings.EqualFold(payload.Status, "fail") || strings.EqualFold(payload.Status, "failed") {
		return userDataGeoIPLookupResult{}, fmt.Errorf("geoip lookup failed: %s", firstNonEmptyString(payload.ErrorMessage, payload.ErrorMessage2))
	}
	if payload.Error {
		return userDataGeoIPLookupResult{}, fmt.Errorf("geoip lookup failed: %s", firstNonEmptyString(payload.ErrorMessage, payload.ErrorMessage2))
	}

	countryCode := firstNonEmptyString(payload.CountryCode, payload.CountryCode2)
	if countryCode == "" && len(strings.TrimSpace(payload.Country)) == 2 {
		countryCode = strings.TrimSpace(payload.Country)
	}
	if countryCode == "" {
		return userDataGeoIPLookupResult{}, fmt.Errorf("geoip country code is empty")
	}

	return userDataGeoIPLookupResult{
		CountryCode: strings.TrimSpace(countryCode),
		CountryName: strings.TrimSpace(payload.Country),
	}, nil
}

func buildGeoIPRequestURL(baseURL string, ipText string) string {
	escapedIP := url.PathEscape(ipText)
	if strings.Contains(baseURL, "{ip}") {
		return strings.ReplaceAll(baseURL, "{ip}", escapedIP)
	}
	if strings.Contains(baseURL, "?") {
		separator := "&"
		if strings.HasSuffix(baseURL, "?") || strings.HasSuffix(baseURL, "&") {
			separator = ""
		}
		return baseURL + separator + "ip=" + url.QueryEscape(ipText)
	}
	return baseURL + escapedIP
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
