package controllers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	stdjson "encoding/json"
	"fmt"
	"gogogo/common"
	"gogogo/helpers"
	"gogogo/models"
	"gogogo/models/dtos"
	"gogogo/provider"
	"gogogo/services"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	playerTransactionPeriodDaily    = "daily"
	playerTransactionPeriodWeekly   = "weekly"
	playerTransactionPeriodTotal    = "total"
	playerTransactionPeriodTotalKey = "all"
)

type syncGameTransactionsRequest struct {
	Date string `json:"date"`
}

type syncAllGameTransactionsRequest struct {
	Date        string   `json:"date"`
	Limit       int      `json:"limit"`
	StartUserID uint64   `json:"start_user_id"`
	UserIDs     []uint64 `json:"user_ids"`
	Force       bool     `json:"force"`
}

type backfillGameTransactionsRequest struct {
	UserID    uint64 `json:"user_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Currency  string `json:"currency"`
}

type backfillLegacyGameTransactionsRequest struct {
	UserID    uint64 `json:"user_id"`
	Limit     int    `json:"limit"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type debugM7GameHistoryRequest struct {
	UserID         uint64 `json:"user_id"`
	LoginID        string `json:"login_id"`
	Date           string `json:"date"`
	StartTime      string `json:"start_time"`
	EndTime        string `json:"end_time"`
	ThirdPartyType string `json:"third_party_type"`
	PageIndex      int    `json:"page_index"`
	PageSize       int    `json:"page_size"`
	Sync           bool   `json:"sync"`
}

type m7CompletedWindowSyncResult struct {
	affectedUserIDs    []uint64
	previousM7ByUserID map[uint64]*provider.PlayerSummaryResult
}

type m7PendingGameHistoryDetail struct {
	userID           uint64
	loginID          string
	gameProviderCode int
	row              map[string]any
	pageIndex        int
	rowIndex         int
}

type playerTransactionSummaryView struct {
	Currency  string  `json:"currency"`
	Count     int     `json:"count"`
	Turnover  float64 `json:"turnover"`
	Bet       float64 `json:"bet"`
	Win       float64 `json:"win"`
	Winlose   float64 `json:"winlose"`
	JPShare   float64 `json:"jp_share"`
	JPWin     float64 `json:"jp_win"`
	TurnoverU float64 `json:"turnover_u"`
	BetU      float64 `json:"bet_u"`
	WinU      float64 `json:"win_u"`
	WinloseU  float64 `json:"winlose_u"`
	JPShareU  float64 `json:"jp_share_u"`
	JPWinU    float64 `json:"jp_win_u"`
}

type playerTransactionSyncResult struct {
	StatDate string                       `json:"stat_date"`
	LoginID  string                       `json:"login_id"`
	Summary  playerTransactionSummaryView `json:"summary"`
	SyncedAt time.Time                    `json:"synced_at"`
}

type playerTransactionBatchSyncItem struct {
	UserID    uint64    `json:"user_id"`
	LoginID   string    `json:"login_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message,omitempty"`
	SyncedAt  time.Time `json:"synced_at,omitempty"`
	StatDate  string    `json:"stat_date"`
	PeriodKey string    `json:"period_key,omitempty"`
}

type playerTransactionBatchSyncResult struct {
	StatDate        string                           `json:"stat_date"`
	StartUserID     uint64                           `json:"start_user_id"`
	NextStartUserID uint64                           `json:"next_start_user_id"`
	UsersScanned    int                              `json:"users_scanned"`
	UsersSynced     int                              `json:"users_synced"`
	UsersSkipped    int                              `json:"users_skipped"`
	UsersFailed     int                              `json:"users_failed"`
	Items           []playerTransactionBatchSyncItem `json:"items"`
	SyncedAt        time.Time                        `json:"synced_at"`
}

type playerTransactionBackfillResult struct {
	LoginID       string    `json:"login_id"`
	StartDate     string    `json:"start_date"`
	EndDate       string    `json:"end_date"`
	DaysRequested int       `json:"days_requested"`
	DaysSynced    int       `json:"days_synced"`
	FailedDates   []string  `json:"failed_dates"`
	SyncedAt      time.Time `json:"synced_at"`
}

type playerTransactionLegacyBackfillItem struct {
	UserID        uint64   `json:"user_id"`
	LoginID       string   `json:"login_id"`
	StartDate     string   `json:"start_date"`
	EndDate       string   `json:"end_date"`
	DaysRequested int      `json:"days_requested"`
	DaysSynced    int      `json:"days_synced"`
	FailedDates   []string `json:"failed_dates"`
}

type playerTransactionLegacyBackfillResult struct {
	UsersScanned int                                   `json:"users_scanned"`
	UsersFixed   int                                   `json:"users_fixed"`
	UsersFailed  int                                   `json:"users_failed"`
	Items        []playerTransactionLegacyBackfillItem `json:"items"`
	SyncedAt     time.Time                             `json:"synced_at"`
}

type playerTransactionOverviewBlock struct {
	PeriodType string                       `json:"period_type"`
	PeriodKey  string                       `json:"period_key"`
	FromDate   string                       `json:"from_date"`
	ToDate     string                       `json:"to_date"`
	Summary    playerTransactionSummaryView `json:"summary"`
}

type playerTransactionOverviewResult struct {
	Currency    string                         `json:"currency"`
	LoginID     string                         `json:"login_id"`
	Daily       playerTransactionOverviewBlock `json:"daily"`
	Weekly      playerTransactionOverviewBlock `json:"weekly"`
	Total       playerTransactionOverviewBlock `json:"total"`
	RefreshedAt time.Time                      `json:"refreshed_at"`
}

type playerTransactionScope struct {
	PeriodType string
	PeriodKey  string
	FromDate   time.Time
	ToDate     time.Time
}

var (
	playerTransactionStatsTableOnce    sync.Once
	playerTransactionStatsTableErr     error
	playerTransactionDetailTableOnce   sync.Once
	playerTransactionDetailTableErr    error
	playerTransactionCurrencyTableOnce sync.Once
	playerTransactionCurrencyTableErr  error
	playerTransactionSyncStateMu       sync.Mutex
	playerTransactionSyncLastRun       = map[uint64]time.Time{}
	m7GameHistoryRateLimitMu           sync.Mutex
	m7GameHistoryLastRequestAt         time.Time
)

const (
	m7GameHistoryMaxWindow  = 20 * time.Minute
	m7GameHistoryRequestGap = 12 * time.Second
)

var m7GameHistoryProviderWhitelist = map[string]struct{}{
	"JDB_SLOT":     {},
	"PP_SLOT":      {},
	"PG_SLOT":      {},
	"ONE_PT_SLOT":  {},
	"ONE_MG_SLOT":  {},
	"DNG_HS_SLOT":  {},
	"DNG_STM_SLOT": {},
	"PP_LIVE":      {},
	"ONE_PT_LIVE":  {},
	"EVO_LIVE":     {},
}

func SyncMyGameTransactions(c *fiber.Ctx) error {
	if err := ensurePlayerTransactionStatsTable(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": err.Error(),
		})
	}

	var req syncGameTransactionsRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    400,
				"message": "Invalid JSON",
			})
		}
	}

	statDate, err := normalizeStatDate(req.Date)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    400,
			"message": "Invalid date, expected YYYY-MM-DD",
		})
	}

	result, err := SyncPlayerDailyGameSummary(c.Context(), uint64(QueryUserIdFromJwt(c)), statDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data":    result,
	})
}

func SyncAllGameTransactions(c *fiber.Ctx) error {
	if err := ensurePlayerTransactionStatsTable(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": err.Error(),
		})
	}

	if !QueryUserIsAdmin(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"code":    403,
			"message": "forbidden",
		})
	}

	var req syncAllGameTransactionsRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    400,
				"message": "Invalid JSON",
			})
		}
	}

	result, err := SyncLatestGameTransactionsForUsers(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data":    result,
	})
}

func BackfillMyGameTransactions(c *fiber.Ctx) error {
	if err := ensurePlayerTransactionStatsTable(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": err.Error(),
		})
	}

	var req backfillGameTransactionsRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    400,
				"message": "Invalid JSON",
			})
		}
	}

	currentUserID := uint64(QueryUserIdFromJwt(c))
	targetUserID := currentUserID
	if req.UserID > 0 {
		if !QueryUserIsAdmin(c) && req.UserID != currentUserID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"code":    403,
				"message": "forbidden",
			})
		}
		targetUserID = req.UserID
	}

	result, err := BackfillPlayerGameTransactionsWithCurrency(
		c.Context(),
		targetUserID,
		req.StartDate,
		req.EndDate,
		req.Currency,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data":    result,
	})
}

func TempBackfillGameTransactions(c *fiber.Ctx) error {
	if err := ensurePlayerTransactionStatsTable(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": err.Error(),
		})
	}

	var req backfillGameTransactionsRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    400,
				"message": "Invalid JSON",
			})
		}
	}

	if req.UserID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    400,
			"message": "user_id is required",
		})
	}

	result, err := BackfillPlayerGameTransactionsWithCurrency(
		c.Context(),
		req.UserID,
		req.StartDate,
		req.EndDate,
		req.Currency,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data":    result,
	})
}

func BackfillLegacyGameTransactions(c *fiber.Ctx) error {
	if err := ensurePlayerTransactionStatsTable(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": err.Error(),
		})
	}

	var req backfillLegacyGameTransactionsRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    400,
				"message": "Invalid JSON",
			})
		}
	}

	currentUserID := uint64(QueryUserIdFromJwt(c))
	if !QueryUserIsAdmin(c) {
		if req.UserID != 0 && req.UserID != currentUserID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"code":    403,
				"message": "forbidden",
			})
		}
		req.UserID = currentUserID
		req.Limit = 1
	}

	result, err := BackfillLegacyPlayerGameTransactions(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data":    result,
	})
}

func TempBackfillLegacyGameTransactions(c *fiber.Ctx) error {
	if err := ensurePlayerTransactionStatsTable(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": err.Error(),
		})
	}

	var req backfillLegacyGameTransactionsRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    400,
				"message": "Invalid JSON",
			})
		}
	}

	if req.UserID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    400,
			"message": "user_id is required",
		})
	}

	result, err := BackfillLegacyPlayerGameTransactions(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data":    result,
	})
}

func TempDebugM7GameHistory(c *fiber.Ctx) error {
	var req debugM7GameHistoryRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    400,
				"message": "Invalid JSON",
			})
		}
	}

	loginID := strings.TrimSpace(req.LoginID)
	if loginID == "" {
		if req.UserID == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    400,
				"message": "user_id or login_id is required",
			})
		}
		loginID = buildPlayerLoginID(req.UserID)
	}

	loc := playerTransactionLocation()
	var start time.Time
	var end time.Time
	var err error
	if strings.TrimSpace(req.StartTime) != "" || strings.TrimSpace(req.EndTime) != "" {
		start, err = parseM7DebugTime(req.StartTime, loc)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": 400, "message": "invalid start_time"})
		}
		end, err = parseM7DebugTime(req.EndTime, loc)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": 400, "message": "invalid end_time"})
		}
	} else {
		statDate, err := normalizeStatDate(req.Date)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": 400, "message": "invalid date, expected YYYY-MM-DD"})
		}
		start, _ = time.ParseInLocation("2006-01-02", statDate, loc)
		if statDate == time.Now().In(loc).Format("2006-01-02") {
			end = time.Now().In(loc)
			start = end.Add(-m7GameHistoryMaxWindow).Add(time.Second)
			dayStart, _ := time.ParseInLocation("2006-01-02", statDate, loc)
			if start.Before(dayStart) {
				start = dayStart
			}
		} else {
			end = start.Add(m7GameHistoryMaxWindow).Add(-time.Second)
		}
	}

	client, err := provider.NewM7Client(m7ProviderConfig())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	pageIndex := req.PageIndex
	if pageIndex <= 0 {
		pageIndex = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	resp, err := client.GetGameHistory(loginID, req.ThirdPartyType, start, end, pageIndex, pageSize)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":       500,
			"message":    err.Error(),
			"login_id":   loginID,
			"start_time": start.In(loc).Format("2006-01-02 15:04:05"),
			"end_time":   end.In(loc).Format("2006-01-02 15:04:05"),
		})
	}

	rows := resp.Rows()
	syncResult := any(nil)
	if req.Sync && req.UserID > 0 {
		statDate := start.In(loc).Format("2006-01-02")
		if result, syncErr := SyncPlayerDailyGameSummaryWithCurrencyOptions(c.Context(), req.UserID, statDate, "", true); syncErr != nil {
			syncResult = fiber.Map{"success": false, "message": syncErr.Error()}
		} else {
			syncResult = fiber.Map{"success": true, "data": result}
		}
	}

	return c.JSON(fiber.Map{
		"code":          0,
		"message":       "success",
		"login_id":      loginID,
		"start_time":    start.In(loc).Format("2006-01-02 15:04:05"),
		"end_time":      end.In(loc).Format("2006-01-02 15:04:05"),
		"m7_code":       resp.Code,
		"m7_message":    m7ResponseMessage(resp),
		"row_count":     len(rows),
		"total_records": resp.TotalRecords(),
		"total_pages":   resp.TotalPages(),
		"rows":          rows,
		"sync_result":   syncResult,
	})
}

func parseM7DebugTime(value string, loc *time.Location) (time.Time, error) {
	text := strings.TrimSpace(value)
	for _, layout := range []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		time.RFC3339,
	} {
		if parsed, err := time.ParseInLocation(layout, text, loc); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time")
}

func m7ResponseMessage(resp *provider.M7GameHistoryResponse) string {
	if resp == nil {
		return ""
	}
	if strings.TrimSpace(resp.Message) != "" {
		return resp.Message
	}
	return resp.Msg
}

func GetMyGameTransactions(c *fiber.Ctx) error {
	if err := ensurePlayerTransactionStatsTable(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": err.Error(),
		})
	}

	statDate, err := normalizeStatDate(c.Query("date"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    400,
			"message": "Invalid date, expected YYYY-MM-DD",
		})
	}

	userID := uint64(QueryUserIdFromJwt(c))
	displayCurrency := resolvePlayerTransactionCurrency(userID, c.Query("currency"))
	summary, err := getStoredPlayerDailyGameSummary(userID, statDate, displayCurrency)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": "Failed to load transaction summary",
		})
	}

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data":    summary,
	})
}

func GetMyGameTransactionOverview(c *fiber.Ctx) error {
	if err := ensurePlayerTransactionStatsTable(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": err.Error(),
		})
	}

	userID := uint64(QueryUserIdFromJwt(c))
	user, err := getUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": "Failed to load user",
		})
	}

	shouldSync := isTruthyFlag(c.Query("sync"))
	displayCurrency := resolvePlayerTransactionCurrency(userID, c.Query("currency"))
	result, err := loadPlayerTransactionOverview(c.Context(), user, shouldSync, displayCurrency)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data":    result,
	})
}

func TriggerTodayPlayerDailyGameSummarySync(userID uint64) {
	TriggerTodayPlayerDailyGameSummarySyncIfStale(userID, 0, true)
}

func TriggerTodayPlayerDailyGameSummarySyncAfter(userID uint64, delay time.Duration) {
	if delay <= 0 {
		TriggerTodayPlayerDailyGameSummarySync(userID)
		return
	}

	go func() {
		time.Sleep(delay)
		TriggerTodayPlayerDailyGameSummarySyncIfStale(userID, 0, true)
	}()
}

func BackfillPlayerGameTransactions(ctx context.Context, userID uint64, startDate string, endDate string) (*playerTransactionBackfillResult, error) {
	return BackfillPlayerGameTransactionsWithCurrency(ctx, userID, startDate, endDate, "")
}

func BackfillPlayerGameTransactionsWithCurrency(
	ctx context.Context,
	userID uint64,
	startDate string,
	endDate string,
	sourceCurrency string,
) (*playerTransactionBackfillResult, error) {
	user, err := getUserByID(userID)
	if err != nil {
		return nil, err
	}
	if err := ensurePlayerTransactionDetailTable(); err != nil {
		return nil, err
	}
	if normalizedCurrency := normalizePlayerTransactionCurrency(sourceCurrency); normalizedCurrency != "" {
		if err := savePlayerTransactionCurrencyOverride(ctx, userID, normalizedCurrency); err != nil {
			return nil, err
		}
	}

	start, end, err := normalizePlayerTransactionBackfillRange(user, startDate, endDate)
	if err != nil {
		return nil, err
	}

	loginID := buildPlayerLoginID(userID)
	daysRequested := 0
	daysSynced := 0
	failedDates := make([]string, 0)

	for current := start; !current.After(end); current = current.AddDate(0, 0, 1) {
		statDate := current.Format("2006-01-02")
		daysRequested++
		if err := resetPlayerTransactionDetailRows(ctx, userID, statDate); err != nil {
			log.Printf("[GameSummaryBackfill] user_id=%d login_id=%s stat_date=%s clear details failed: %v\n", userID, loginID, statDate, err)
			failedDates = append(failedDates, statDate)
			continue
		}
		if _, err := SyncPlayerDailyGameSummaryWithCurrencyOptions(ctx, userID, statDate, sourceCurrency, true); err != nil {
			log.Printf("[GameSummaryBackfill] user_id=%d login_id=%s stat_date=%s failed: %v\n", userID, loginID, statDate, err)
			failedDates = append(failedDates, statDate)
			continue
		}
		daysSynced++
	}

	return &playerTransactionBackfillResult{
		LoginID:       loginID,
		StartDate:     start.Format("2006-01-02"),
		EndDate:       end.Format("2006-01-02"),
		DaysRequested: daysRequested,
		DaysSynced:    daysSynced,
		FailedDates:   failedDates,
		SyncedAt:      time.Now().UTC(),
	}, nil
}

func SyncLatestGameTransactionsForUsers(
	ctx context.Context,
	req syncAllGameTransactionsRequest,
) (*playerTransactionBatchSyncResult, error) {
	statDate, err := normalizeStatDate(req.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date, expected YYYY-MM-DD")
	}

	users, nextStartUserID, err := listUsersForLatestGameTransactionSync(req)
	if err != nil {
		return nil, err
	}

	result := &playerTransactionBatchSyncResult{
		StatDate:        statDate,
		StartUserID:     req.StartUserID,
		NextStartUserID: nextStartUserID,
		UsersScanned:    len(users),
		Items:           make([]playerTransactionBatchSyncItem, 0, len(users)),
		SyncedAt:        time.Now().UTC(),
	}

	for _, user := range users {
		item := playerTransactionBatchSyncItem{
			UserID:   user.ID,
			LoginID:  buildPlayerLoginID(user.ID),
			StatDate: statDate,
		}

		if !req.Force && isCurrentPlayerTransactionStatFresh(statDate, &user, 2*time.Minute) {
			item.Status = "skipped"
			item.Message = "already synced recently"
			result.UsersSkipped++
			result.Items = append(result.Items, item)
			continue
		}

		syncResult, syncErr := SyncPlayerDailyGameSummary(ctx, user.ID, statDate)
		if syncErr != nil {
			item.Status = "failed"
			item.Message = syncErr.Error()
			result.UsersFailed++
			result.Items = append(result.Items, item)
			log.Printf("[GameSummaryBatchSync] user_id=%d login_id=%s stat_date=%s failed: %v\n", user.ID, item.LoginID, statDate, syncErr)
			continue
		}

		item.Status = "synced"
		item.SyncedAt = syncResult.SyncedAt
		item.PeriodKey = syncResult.StatDate
		result.UsersSynced++
		result.Items = append(result.Items, item)
	}

	result.SyncedAt = time.Now().UTC()
	return result, nil
}

func BackfillLegacyPlayerGameTransactions(
	ctx context.Context,
	req backfillLegacyGameTransactionsRequest,
) (*playerTransactionLegacyBackfillResult, error) {
	type legacyUserRange struct {
		UserID    uint64 `gorm:"column:user_id"`
		StartDate string `gorm:"column:start_date"`
		EndDate   string `gorm:"column:end_date"`
	}

	query := models.GetInstance().DbInstance.
		Model(&dtos.UserGameTransactionStat{}).
		Select("user_id, MIN(period_key) AS start_date, MAX(period_key) AS end_date").
		Where("period_type = ? AND amount_unit = ''", playerTransactionPeriodDaily).
		Group("user_id").
		Order("user_id ASC")

	if req.UserID > 0 {
		query = query.Where("user_id = ?", req.UserID)
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}
	if req.UserID == 0 {
		query = query.Limit(limit)
	}

	var ranges []legacyUserRange
	if err := query.Scan(&ranges).Error; err != nil {
		return nil, err
	}

	result := &playerTransactionLegacyBackfillResult{
		UsersScanned: len(ranges),
		Items:        make([]playerTransactionLegacyBackfillItem, 0, len(ranges)),
		SyncedAt:     time.Now().UTC(),
	}

	for _, item := range ranges {
		startDate := item.StartDate
		endDate := item.EndDate
		if strings.TrimSpace(req.StartDate) != "" {
			startDate = req.StartDate
		}
		if strings.TrimSpace(req.EndDate) != "" {
			endDate = req.EndDate
		}

		backfillResult, err := BackfillPlayerGameTransactions(ctx, item.UserID, startDate, endDate)
		if err != nil {
			result.UsersFailed++
			result.Items = append(result.Items, playerTransactionLegacyBackfillItem{
				UserID:      item.UserID,
				StartDate:   startDate,
				EndDate:     endDate,
				FailedDates: []string{err.Error()},
			})
			log.Printf("[GameSummaryLegacyBackfill] user_id=%d failed: %v\n", item.UserID, err)
			continue
		}

		if err := markPlayerTransactionAmountUnit(ctx, item.UserID, backfillResult.StartDate, backfillResult.EndDate, "U"); err != nil {
			return nil, err
		}

		result.UsersFixed++
		result.Items = append(result.Items, playerTransactionLegacyBackfillItem{
			UserID:        item.UserID,
			LoginID:       backfillResult.LoginID,
			StartDate:     backfillResult.StartDate,
			EndDate:       backfillResult.EndDate,
			DaysRequested: backfillResult.DaysRequested,
			DaysSynced:    backfillResult.DaysSynced,
			FailedDates:   backfillResult.FailedDates,
		})
	}

	result.SyncedAt = time.Now().UTC()
	return result, nil
}

func SyncPlayerDailyGameSummary(ctx context.Context, userID uint64, statDate string) (*playerTransactionSyncResult, error) {
	return SyncPlayerDailyGameSummaryWithCurrency(ctx, userID, statDate, "")
}

func SyncPlayerDailyGameSummaryWithCurrency(
	ctx context.Context,
	userID uint64,
	statDate string,
	sourceCurrency string,
) (*playerTransactionSyncResult, error) {
	return SyncPlayerDailyGameSummaryWithCurrencyOptions(ctx, userID, statDate, sourceCurrency, false)
}

func SyncPlayerDailyGameSummaryWithCurrencyOptions(
	ctx context.Context,
	userID uint64,
	statDate string,
	sourceCurrency string,
	allowHistoricalDetailFallback bool,
) (*playerTransactionSyncResult, error) {
	user, err := getUserByID(userID)
	if err != nil {
		return nil, err
	}

	loginID := buildPlayerLoginID(userID)
	client := provider.NewHedocClient()
	registerProviderAccountIfNeeded(client, userID, loginID)

	nowUTC := time.Now().UTC()
	dailyScope, err := buildDailyScope(statDate, nowUTC)
	if err != nil {
		return nil, err
	}
	displayCurrency := resolvePlayerTransactionCurrency(userID, sourceCurrency)

	syncedAt := time.Now().UTC()
	data, err := syncPlayerTransactionScope(ctx, user, loginID, client, dailyScope, syncedAt, allowHistoricalDetailFallback)
	if err != nil {
		return nil, err
	}

	if err := rebuildPlayerCurrentAggregateScopes(ctx, user.ID, loginID, nowUTC, syncedAt); err != nil {
		return nil, err
	}

	if isCurrentUTCDate(statDate) {
		if err := services.GetActivityService().EnsureWeeklySpinWheelTickets(ctx, userID); err != nil {
			return nil, fmt.Errorf("failed to auto issue weekly spin tickets: %w", err)
		}
	}

	return &playerTransactionSyncResult{
		StatDate: statDate,
		LoginID:  loginID,
		Summary:  toPlayerTransactionSummaryView(data, displayCurrency),
		SyncedAt: syncedAt,
	}, nil
}

func loadPlayerTransactionOverview(ctx context.Context, user *dtos.User, shouldSync bool, displayCurrency string) (*playerTransactionOverviewResult, error) {
	nowUTC := time.Now().UTC()
	scopes := buildCurrentPlayerTransactionScopes(user, nowUTC)
	loginID := buildPlayerLoginID(user.ID)

	if shouldSync {
		if err := syncCurrentPlayerTransactionOverview(ctx, user, loginID, nowUTC); err != nil {
			return nil, err
		}
	} else {
		missing, err := hasMissingPlayerTransactionScope(user.ID, scopes)
		if err != nil {
			return nil, err
		}
		needsRepair, err := needsPlayerTransactionScopeRepair(user.ID, scopes)
		if err != nil {
			return nil, err
		}
		if missing || needsRepair {
			if err := syncCurrentPlayerTransactionOverview(ctx, user, loginID, nowUTC); err != nil {
				return nil, err
			}
		}
	}

	return buildPlayerTransactionOverviewResult(user.ID, loginID, scopes, displayCurrency)
}

func syncCurrentPlayerTransactionOverview(ctx context.Context, user *dtos.User, loginID string, nowUTC time.Time) error {
	client := provider.NewHedocClient()
	registerProviderAccountIfNeeded(client, user.ID, loginID)

	syncedAt := time.Now().UTC()
	statDate := nowUTC.In(playerTransactionLocation()).Format("2006-01-02")
	dailyScope, err := buildDailyScope(statDate, nowUTC)
	if err != nil {
		return err
	}

	if _, err := syncPlayerTransactionScope(ctx, user, loginID, client, dailyScope, syncedAt, false); err != nil {
		return err
	}

	if err := rebuildPlayerCurrentAggregateScopes(ctx, user.ID, loginID, nowUTC, syncedAt); err != nil {
		return err
	}

	if err := services.GetActivityService().EnsureWeeklySpinWheelTickets(ctx, user.ID); err != nil {
		return fmt.Errorf("failed to auto issue weekly spin tickets: %w", err)
	}

	return nil
}

func rebuildPlayerCurrentAggregateScopes(ctx context.Context, userID uint64, loginID string, nowUTC time.Time, syncedAt time.Time) error {
	user, err := getUserByID(userID)
	if err != nil {
		return err
	}

	scopes := buildCurrentPlayerTransactionScopes(user, nowUTC)[1:]
	for _, scope := range scopes {
		data, aggregateSyncedAt, err := aggregatePlayerTransactionScopeFromDailyStats(userID, loginID, scope)
		if err != nil {
			return err
		}

		if aggregateSyncedAt.IsZero() {
			aggregateSyncedAt = syncedAt
		}

		if err := upsertPlayerTransactionStat(userID, loginID, scope, data, aggregateSyncedAt); err != nil {
			return err
		}
	}

	if _, err := services.RecalculateAndPersistUserVIPLevel(ctx, userID); err != nil {
		return err
	}

	return nil
}

func aggregatePlayerTransactionScopeFromDailyStats(
	userID uint64,
	loginID string,
	scope playerTransactionScope,
) (*provider.PlayerSummaryResult, time.Time, error) {
	type aggregateRow struct {
		Count          int64      `gorm:"column:count"`
		Turnover       float64    `gorm:"column:turnover"`
		Bet            float64    `gorm:"column:bet"`
		Win            float64    `gorm:"column:win"`
		Winlose        float64    `gorm:"column:winlose"`
		JPShare        float64    `gorm:"column:jp_share"`
		JPWin          float64    `gorm:"column:jp_win"`
		LatestSyncedAt *time.Time `gorm:"column:latest_synced_at"`
	}

	loc := playerTransactionLocation()
	fromKey := scope.FromDate.In(loc).Format("2006-01-02")
	toKey := scope.ToDate.In(loc).Format("2006-01-02")
	var row aggregateRow
	if err := models.GetInstance().DbInstance.
		Model(&dtos.UserGameTransactionStat{}).
		Where("user_id = ? AND period_type = ? AND period_key >= ? AND period_key <= ?",
			userID,
			playerTransactionPeriodDaily,
			fromKey,
			toKey,
		).
		Select(`
			COALESCE(SUM(count), 0) AS count,
			COALESCE(SUM(turnover), 0) AS turnover,
			COALESCE(SUM(bet), 0) AS bet,
			COALESCE(SUM(win), 0) AS win,
			COALESCE(SUM(winlose), 0) AS winlose,
			COALESCE(SUM(jp_share), 0) AS jp_share,
			COALESCE(SUM(jp_win), 0) AS jp_win,
			MAX(synced_at) AS latest_synced_at
		`).
		Scan(&row).Error; err != nil {
		return nil, time.Time{}, err
	}

	latestSyncedAt := time.Time{}
	if row.LatestSyncedAt != nil {
		latestSyncedAt = row.LatestSyncedAt.UTC()
	}

	return &provider.PlayerSummaryResult{
		LoginId:  loginID,
		Count:    int(row.Count),
		Turnover: row.Turnover,
		Bet:      row.Bet,
		Win:      row.Win,
		Winlose:  row.Winlose,
		JPShare:  row.JPShare,
		JPWin:    row.JPWin,
	}, latestSyncedAt, nil
}

func syncPlayerTransactionScope(
	ctx context.Context,
	user *dtos.User,
	loginID string,
	client *provider.HedoClient,
	scope playerTransactionScope,
	syncedAt time.Time,
	allowHistoricalDetailFallback bool,
) (*provider.PlayerSummaryResult, error) {
	var summaryErr error
	data, err := fetchPlayerSummaryByRangeOrZero(client, loginID, scope.FromDate, scope.ToDate)
	if err != nil {
		if shouldTryDetailFallbackOnSummaryError(scope, allowHistoricalDetailFallback) {
			fallbackData, fallbackErr := aggregatePlayerDailySummaryFromDetails(client, user.ID, loginID, scope)
			if fallbackErr == nil {
				data = fallbackData
				log.Printf(
					"[GameSummarySync] user_id=%d login_id=%s stat_date=%s used detail fallback after summary error count=%d turnover=%.2f bet=%.2f summary_error=%v\n",
					user.ID,
					loginID,
					scope.PeriodKey,
					data.Count,
					data.Turnover,
					data.Bet,
					err,
				)
			} else {
				summaryErr = fmt.Errorf("failed to fetch %s player summary: %w", scope.PeriodType, err)
				if scope.PeriodType != playerTransactionPeriodDaily {
					return nil, summaryErr
				}
				data = zeroPlayerSummary(loginID)
				log.Printf(
					"[GameSummarySync] user_id=%d login_id=%s stat_date=%s continue daily sync with zero HEDOC summary after error: %v\n",
					user.ID,
					loginID,
					scope.PeriodKey,
					summaryErr,
				)
			}
		} else {
			summaryErr = fmt.Errorf("failed to fetch %s player summary: %w", scope.PeriodType, err)
			if scope.PeriodType != playerTransactionPeriodDaily {
				return nil, summaryErr
			}
			data = zeroPlayerSummary(loginID)
			log.Printf(
				"[GameSummarySync] user_id=%d login_id=%s stat_date=%s continue daily sync with zero HEDOC summary after error: %v\n",
				user.ID,
				loginID,
				scope.PeriodKey,
				summaryErr,
			)
		}
	}
	if shouldFallbackToDetailSummary(scope, data, allowHistoricalDetailFallback) {
		fallbackData, fallbackErr := aggregatePlayerDailySummaryFromDetails(client, user.ID, loginID, scope)
		if fallbackErr != nil {
			log.Printf(
				"[GameSummarySync] user_id=%d login_id=%s stat_date=%s detail fallback failed: %v\n",
				user.ID,
				loginID,
				scope.PeriodKey,
				fallbackErr,
			)
		} else {
			data = fallbackData
			log.Printf(
				"[GameSummarySync] user_id=%d login_id=%s stat_date=%s used detail fallback count=%d turnover=%.2f bet=%.2f\n",
				user.ID,
				loginID,
				scope.PeriodKey,
				data.Count,
				data.Turnover,
				data.Bet,
			)
		}
	}

	if summaryErr != nil && isZeroPlayerSummary(data) {
		return nil, summaryErr
	}

	if scope.PeriodType == playerTransactionPeriodDaily {
		combinedData, err := combineDailySummaryWithStoredTransferWalletDetails(user.ID, loginID, scope, data)
		if err != nil {
			return nil, err
		}
		data = combinedData
	}

	if err := upsertPlayerTransactionStat(user.ID, loginID, scope, data, syncedAt); err != nil {
		return nil, err
	}

	if scope.PeriodType == playerTransactionPeriodDaily {
		if err := mirrorSummaryToDailyUserStats(user.ID, scope.PeriodKey, data); err != nil {
			return nil, err
		}
	}

	return data, nil
}

func SyncM7CompletedWindowForActiveUsers(ctx context.Context) error {
	loc := playerTransactionLocation()
	now := time.Now().In(loc)
	window := previousCompletedM7GameHistoryWindow(now)
	statDate := window.from.In(loc).Format("2006-01-02")

	log.Printf(
		"[GameSummarySync][M7] completed window sync stat_date=%s from=%s to=%s\n",
		statDate,
		window.from.Format("2006-01-02 15:04:05"),
		window.to.Format("2006-01-02 15:04:05"),
	)

	dailyScope, err := buildDailyScope(statDate, time.Now().UTC())
	if err != nil {
		return err
	}

	syncResult, err := syncM7CompletedWindowByProviders(ctx, dailyScope, window)
	if err != nil {
		return err
	}
	if syncResult == nil || len(syncResult.affectedUserIDs) == 0 {
		return nil
	}
	log.Printf("[GameSummarySync][M7] completed window affected users=%d\n", len(syncResult.affectedUserIDs))

	for _, userID := range syncResult.affectedUserIDs {
		loginID := buildPlayerLoginID(userID)
		hedocData, err := getStoredHedocDailySummaryForM7Merge(
			userID,
			loginID,
			dailyScope,
			syncResult.previousM7ByUserID[userID],
		)
		if err != nil {
			log.Printf("[GameSummarySync][M7] user_id=%d load HEDOC summary failed: %v\n", userID, err)
			continue
		}
		if err := mergeStoredM7DailySummary(ctx, userID, loginID, dailyScope, hedocData); err != nil {
			log.Printf("[GameSummarySync][M7] user_id=%d completed window merge failed: %v\n", userID, err)
		}
	}

	return nil
}

func previousCompletedM7GameHistoryWindow(now time.Time) m7GameHistoryWindow {
	loc := playerTransactionLocation()
	current := now.In(loc)
	currentSlotStart := floorToM7GameHistoryWindow(current)
	previousSlotStart := currentSlotStart.Add(-m7GameHistoryMaxWindow)

	return m7GameHistoryWindow{
		from: previousSlotStart,
		to:   currentSlotStart.Add(-time.Second),
	}
}

func floorToM7GameHistoryWindow(value time.Time) time.Time {
	loc := playerTransactionLocation()
	current := value.In(loc)
	dayStart := startOfDayInLocation(current, loc)
	elapsed := current.Sub(dayStart)
	slot := elapsed / m7GameHistoryMaxWindow
	return dayStart.Add(slot * m7GameHistoryMaxWindow)
}

func getStoredHedocDailySummaryForM7Merge(
	userID uint64,
	loginID string,
	scope playerTransactionScope,
	previousM7Data *provider.PlayerSummaryResult,
) (*provider.PlayerSummaryResult, error) {
	record, err := getStoredPlayerTransactionStat(userID, playerTransactionPeriodDaily, scope.PeriodKey)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return zeroPlayerSummary(loginID), nil
	}

	if previousM7Data == nil {
		previousM7Data = zeroPlayerSummary(loginID)
	}

	base := &provider.PlayerSummaryResult{
		LoginId:  loginID,
		Count:    record.Count,
		Turnover: record.Turnover,
		Bet:      record.Bet,
		Win:      record.Win,
		Winlose:  record.Winlose,
		JPShare:  record.JPShare,
		JPWin:    record.JPWin,
	}

	if isZeroPlayerSummary(previousM7Data) || !playerSummaryLooksLikeItIncludesM7(base, previousM7Data) {
		return base, nil
	}

	return subtractPlayerSummary(base, previousM7Data), nil
}

func getStoredM7DailySummaryForMerge(userID uint64, loginID string, scope playerTransactionScope) (*provider.PlayerSummaryResult, error) {
	providerIDs, err := loadActiveM7GameProviderIDByCode()
	if err != nil {
		return nil, err
	}
	return aggregateM7PlayerDailySummaryFromStoredDetails(userID, loginID, scope, providerIDs)
}

func combineDailySummaryWithStoredM7Details(
	userID uint64,
	loginID string,
	scope playerTransactionScope,
	data *provider.PlayerSummaryResult,
) (*provider.PlayerSummaryResult, error) {
	if data == nil || scope.PeriodType != playerTransactionPeriodDaily {
		return data, nil
	}

	m7Data, err := getStoredM7DailySummaryForMerge(userID, loginID, scope)
	if err != nil {
		return nil, err
	}
	if isZeroPlayerSummary(m7Data) {
		return data, nil
	}

	combined := *data
	addPlayerSummary(&combined, m7Data)
	return &combined, nil
}

// combineDailySummaryWithStoredTransferWalletDetails rebuilds the combined
// daily value from the HEDOC summary and locally persisted transfer-wallet
// details. Keeping this in the common sync path prevents login, recycle and
// settlement refreshes from overwriting M7PP totals.
func combineDailySummaryWithStoredTransferWalletDetails(
	userID uint64,
	loginID string,
	scope playerTransactionScope,
	data *provider.PlayerSummaryResult,
) (*provider.PlayerSummaryResult, error) {
	combined, err := combineDailySummaryWithStoredM7Details(userID, loginID, scope, data)
	if err != nil || combined == nil || scope.PeriodType != playerTransactionPeriodDaily {
		return combined, err
	}

	m7ppData, err := aggregateM7PPDailySummary(userID, loginID, scope)
	if err != nil {
		return nil, err
	}
	addPlayerSummary(combined, m7ppData)
	return combined, nil
}

func playerSummaryLooksLikeItIncludesM7(base *provider.PlayerSummaryResult, m7Data *provider.PlayerSummaryResult) bool {
	if base == nil || m7Data == nil || isZeroPlayerSummary(m7Data) {
		return false
	}

	return base.Count >= m7Data.Count &&
		base.Turnover+0.0001 >= m7Data.Turnover &&
		base.Bet+0.0001 >= m7Data.Bet
}

func subtractPlayerSummary(base *provider.PlayerSummaryResult, minus *provider.PlayerSummaryResult) *provider.PlayerSummaryResult {
	if base == nil || minus == nil {
		return base
	}

	base.Count -= minus.Count
	if base.Count < 0 {
		base.Count = 0
	}
	base.Turnover -= minus.Turnover
	base.Bet -= minus.Bet
	base.Win -= minus.Win
	base.Winlose -= minus.Winlose
	base.JPShare -= minus.JPShare
	base.JPWin -= minus.JPWin
	if base.Turnover < 0 {
		base.Turnover = 0
	}
	if base.Bet < 0 {
		base.Bet = 0
	}
	if base.Win < 0 {
		base.Win = 0
	}
	if base.JPShare < 0 {
		base.JPShare = 0
	}
	if base.JPWin < 0 {
		base.JPWin = 0
	}
	return base
}

func syncM7CompletedWindowByProviders(
	ctx context.Context,
	scope playerTransactionScope,
	window m7GameHistoryWindow,
) (*m7CompletedWindowSyncResult, error) {
	const pageSize = 5000

	client, err := provider.NewM7Client(m7ProviderConfig())
	if err != nil {
		return nil, nil
	}
	if err := ensurePlayerTransactionDetailTable(); err != nil {
		return nil, err
	}

	providerIDs, err := loadActiveM7GameProviderIDByCode()
	if err != nil {
		return nil, err
	}
	if len(providerIDs) == 0 {
		return nil, nil
	}

	affected := make(map[uint64]struct{})
	previousM7ByUserID := make(map[uint64]*provider.PlayerSummaryResult)
	m7AgentID := strings.TrimSpace(m7ProviderConfig().AgentID)
	providerCodes := sortedM7ProviderCodes(providerIDs)
	for _, providerCode := range providerCodes {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		fetchedRows := 0
		acceptedRows := 0
		persistedRows := 0
		pendingDetails := make(map[string]*m7PendingGameHistoryDetail)
		for pageIndex := 1; ; pageIndex++ {
			resp, err := getM7GameHistoryWithRateLimit(client, "", providerCode, window.from, window.to, pageIndex, pageSize)
			if err != nil {
				log.Printf(
					"[GameSummarySync][M7] provider=%s stat_date=%s window=%s-%s failed: %v\n",
					providerCode,
					scope.PeriodKey,
					window.from.In(playerTransactionLocation()).Format("15:04:05"),
					window.to.In(playerTransactionLocation()).Format("15:04:05"),
					err,
				)
				break
			}
			rows := resp.Rows()
			fetchedRows += len(rows)
			if len(rows) == 0 {
				break
			}

			for index, row := range rows {
				if !isSettledM7GameHistoryRow(row) {
					continue
				}
				userID, loginID := resolveM7HistoryLocalUser(row)
				if userID == 0 || strings.TrimSpace(loginID) == "" {
					log.Printf(
						"[GameSummarySync][M7] provider=%s stat_date=%s skip row without local user: keys=%s\n",
						providerCode,
						scope.PeriodKey,
						strings.Join(transactionRowKeys(row), ","),
					)
					continue
				}

				gameProviderCode := providerIDs[providerCode]
				normalizeM7GameHistoryRow(row)
				if _, exists := previousM7ByUserID[userID]; !exists {
					previousM7Data, err := getStoredM7DailySummaryForMerge(userID, loginID, scope)
					if err != nil {
						return nil, err
					}
					previousM7ByUserID[userID] = previousM7Data
				}
				addM7PendingGameHistoryDetail(pendingDetails, userID, loginID, gameProviderCode, row, pageIndex, index)
				acceptedRows++
			}

			if totalPages := resp.TotalPages(); totalPages > 0 && pageIndex >= totalPages {
				break
			}
			if len(rows) < pageSize && resp.TotalPages() <= 0 {
				break
			}
			if pageIndex > 10000 {
				return nil, fmt.Errorf("m7 game history page overflow for provider %s on %s", providerCode, scope.PeriodKey)
			}
		}

		pendingKeys := make([]string, 0, len(pendingDetails))
		for key := range pendingDetails {
			pendingKeys = append(pendingKeys, key)
		}
		sort.Strings(pendingKeys)
		for _, key := range pendingKeys {
			detail := pendingDetails[key]
			recordKey := buildM7GameHistoryTransactionDetailRecordKey(detail.row, detail.gameProviderCode, detail.pageIndex, detail.rowIndex)
			persisted, err := persistPlayerTransactionDetailRow(detail.userID, detail.loginID, scope, detail.gameProviderCode, detail.row, recordKey, m7AgentID)
			if err != nil {
				log.Printf(
					"[GameSummarySync][M7] provider=%s user_id=%d login_id=%s stat_date=%s persist detail row failed: %v\n",
					providerCode,
					detail.userID,
					detail.loginID,
					scope.PeriodKey,
					err,
				)
				continue
			}
			if !persisted {
				continue
			}
			affected[detail.userID] = struct{}{}
			persistedRows++
		}
		log.Printf(
			"[GameSummarySync][M7] provider=%s stat_date=%s from=%s to=%s fetched=%d accepted=%d persisted=%d\n",
			providerCode,
			scope.PeriodKey,
			window.from.In(playerTransactionLocation()).Format("2006-01-02 15:04:05"),
			window.to.In(playerTransactionLocation()).Format("2006-01-02 15:04:05"),
			fetchedRows,
			acceptedRows,
			persistedRows,
		)
	}

	result := make([]uint64, 0, len(affected))
	for userID := range affected {
		result = append(result, userID)
	}
	return &m7CompletedWindowSyncResult{
		affectedUserIDs:    result,
		previousM7ByUserID: previousM7ByUserID,
	}, nil
}

func mergeStoredM7DailySummary(
	ctx context.Context,
	userID uint64,
	loginID string,
	scope playerTransactionScope,
	hedocData *provider.PlayerSummaryResult,
) error {
	m7Data, err := getStoredM7DailySummaryForMerge(userID, loginID, scope)
	if err != nil {
		return err
	}
	if isZeroPlayerSummary(m7Data) {
		return nil
	}
	combined := *hedocData
	addPlayerSummary(&combined, m7Data)
	syncedAt := time.Now().UTC()
	if err := upsertPlayerTransactionStat(userID, loginID, scope, &combined, syncedAt); err != nil {
		return err
	}
	if err := mirrorSummaryToDailyUserStats(userID, scope.PeriodKey, &combined); err != nil {
		return err
	}
	return rebuildPlayerCurrentAggregateScopes(ctx, userID, loginID, time.Now().UTC(), syncedAt)
}

func sortedM7ProviderCodes(providerIDs map[string]int) []string {
	codes := make([]string, 0, len(providerIDs))
	for code := range providerIDs {
		if strings.TrimSpace(code) != "" {
			codes = append(codes, code)
		}
	}
	sort.Strings(codes)
	return codes
}

func resolveM7HistoryLocalUser(row map[string]any) (uint64, string) {
	candidateKeys := []string{
		"userViewId",
		"UserViewId",
		"userId",
		"UserId",
		"userID",
		"UserID",
		"memberId",
		"MemberId",
		"playerId",
		"PlayerId",
	}

	firstCandidate := ""
	for _, key := range candidateKeys {
		loginID := strings.TrimSpace(readAnyString(row, key))
		if loginID == "" {
			continue
		}
		if firstCandidate == "" {
			firstCandidate = loginID
		}
		userID, ok := parseLocalUserIDFromPlayerLoginID(loginID)
		if ok {
			return userID, buildPlayerLoginID(userID)
		}
	}

	return 0, firstCandidate
}

func parseLocalUserIDFromPlayerLoginID(loginID string) (uint64, bool) {
	trimmed := strings.TrimSpace(loginID)
	prefix := strings.TrimSpace(helpers.GetCfgInstance().Conf.Prefix)
	if prefix != "" && strings.HasPrefix(trimmed, prefix) {
		trimmed = strings.TrimPrefix(trimmed, prefix)
	}
	if !strings.HasPrefix(trimmed, "u") {
		return 0, false
	}
	parsed, err := strconv.ParseUint(strings.TrimPrefix(trimmed, "u"), 10, 64)
	if err != nil || parsed == 0 {
		return 0, false
	}
	return parsed, true
}

func transactionRowKeys(row map[string]any) []string {
	keys := make([]string, 0, len(row))
	for key := range row {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func normalizeM7GameHistoryRow(row map[string]any) {
	if row == nil {
		return
	}

	row["source"] = "M7_GAME_HISTORY"
	row["Type"] = 2
	row["GameCode"] = firstNonEmptyTransactionString(
		readAnyString(row, "gamesId", "GamesId"),
		readAnyString(row, "gameId", "GameId"),
	)
	row["ExternalId"] = readAnyString(row, "txId", "TxId", "id", "ID")
	row["RoundId"] = readAnyString(row, "gameRound", "GameRound")
	row["StartDate"] = readAnyString(row, "betTime", "BetTime")
	row["EndDate"] = readAnyString(row, "settleTime", "SettleTime")
	row["Turnover"] = readAnyFloat(row, "validBetAmount", "ValidBetAmount")
	row["Bet"] = readAnyFloat(row, "betAmount", "BetAmount")
	row["Win"] = readAnyFloat(row, "win", "Win")
	row["Winlose"] = readAnyFloat(row, "winLose", "WinLose", "winLoss", "WinLoss")
	row["StartBalance"] = readAnyFloat(row, "StartBalance", "startBalance", "start_balance")
	row["EndBalance"] = readAnyFloat(row, "EndBalance", "endBalance", "end_balance", "balance", "Balance")
}

func addM7PendingGameHistoryDetail(
	pending map[string]*m7PendingGameHistoryDetail,
	userID uint64,
	loginID string,
	gameProviderCode int,
	row map[string]any,
	pageIndex int,
	rowIndex int,
) {
	if pending == nil || row == nil {
		return
	}

	key := buildM7GameHistoryAggregateKey(userID, loginID, gameProviderCode, row, pageIndex, rowIndex)
	if existing, ok := pending[key]; ok {
		mergeM7GameHistoryRows(existing.row, row)
		return
	}

	pending[key] = &m7PendingGameHistoryDetail{
		userID:           userID,
		loginID:          loginID,
		gameProviderCode: gameProviderCode,
		row:              cloneTransactionRow(row),
		pageIndex:        pageIndex,
		rowIndex:         rowIndex,
	}
}

func buildM7GameHistoryAggregateKey(userID uint64, loginID string, gameProviderCode int, row map[string]any, pageIndex int, rowIndex int) string {
	roundID := strings.TrimSpace(readAnyString(row, "RoundId", "RoundID", "gameRound", "GameRound"))
	if roundID != "" {
		return fmt.Sprintf("m7-round|%d|%s|%d|%s", userID, loginID, gameProviderCode, roundID)
	}
	return fmt.Sprintf("m7-row|%d|%s|%d|%d|%d", userID, loginID, gameProviderCode, pageIndex, rowIndex)
}

func cloneTransactionRow(row map[string]any) map[string]any {
	cloned := make(map[string]any, len(row))
	for key, value := range row {
		cloned[key] = value
	}
	return cloned
}

func mergeM7GameHistoryRows(base map[string]any, extra map[string]any) {
	if base == nil || extra == nil {
		return
	}

	base["Turnover"] = roundPlayerTransactionAmount(readTransactionDetailTurnover(base) + readTransactionDetailTurnover(extra))
	base["Bet"] = roundPlayerTransactionAmount(readTransactionDetailBet(base) + readTransactionDetailBet(extra))
	base["Win"] = roundPlayerTransactionAmount(readTransactionDetailWin(base) + readTransactionDetailWin(extra))
	base["Winlose"] = roundPlayerTransactionAmount(
		readAnyFloat(base, "Winlose", "WinLose", "winLose", "WinLoss", "winLoss") +
			readAnyFloat(extra, "Winlose", "WinLose", "winLose", "WinLoss", "winLoss"),
	)
	base["JPShare"] = roundPlayerTransactionAmount(readAnyFloat(base, "JPShare", "JackpotShare") + readAnyFloat(extra, "JPShare", "JackpotShare"))
	base["JPWin"] = roundPlayerTransactionAmount(readAnyFloat(base, "JPWin", "JackpotWin") + readAnyFloat(extra, "JPWin", "JackpotWin"))

	if start := earliestTransactionDetailTime(readTransactionDetailTime(base, "StartDate", "betTime", "BetTime"), readTransactionDetailTime(extra, "StartDate", "betTime", "BetTime")); start != "" {
		base["StartDate"] = start
	}
	if end := latestTransactionDetailTime(readTransactionDetailTime(base, "EndDate", "settleTime", "SettleTime"), readTransactionDetailTime(extra, "EndDate", "settleTime", "SettleTime")); end != "" {
		base["EndDate"] = end
	}

	for _, key := range []string{"ExternalId", "externalId", "txId", "TxId"} {
		if strings.TrimSpace(readAnyString(extra, key)) != "" {
			base["ExternalId"] = readAnyString(extra, key)
			break
		}
	}
	base["source"] = "M7_GAME_HISTORY_AGGREGATED"
}

func earliestTransactionDetailTime(left *time.Time, right *time.Time) string {
	if left == nil && right == nil {
		return ""
	}
	if left == nil {
		return formatTransactionDetailLocalTime(*right)
	}
	if right == nil || left.Before(*right) {
		return formatTransactionDetailLocalTime(*left)
	}
	return formatTransactionDetailLocalTime(*right)
}

func latestTransactionDetailTime(left *time.Time, right *time.Time) string {
	if left == nil && right == nil {
		return ""
	}
	if left == nil {
		return formatTransactionDetailLocalTime(*right)
	}
	if right == nil || left.After(*right) {
		return formatTransactionDetailLocalTime(*left)
	}
	return formatTransactionDetailLocalTime(*right)
}

func formatTransactionDetailLocalTime(value time.Time) string {
	return value.In(playerTransactionLocation()).Format("2006-01-02 15:04:05")
}

func firstNonEmptyTransactionString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func buildPlayerTransactionOverviewResult(
	userID uint64,
	loginID string,
	scopes []playerTransactionScope,
	displayCurrency string,
) (*playerTransactionOverviewResult, error) {
	blocks := make(map[string]playerTransactionOverviewBlock, len(scopes))
	refreshedAt := time.Time{}

	for _, scope := range scopes {
		record, err := getStoredPlayerTransactionStat(userID, scope.PeriodType, scope.PeriodKey)
		if err != nil {
			return nil, err
		}

		block := playerTransactionOverviewBlock{
			PeriodType: scope.PeriodType,
			PeriodKey:  scope.PeriodKey,
			FromDate:   scope.FromDate.Format(time.RFC3339),
			ToDate:     scope.ToDate.Format(time.RFC3339),
			Summary:    buildPlayerTransactionSummaryViewFromStoredValues(0, 0, 0, 0, 0, 0, 0, displayCurrency),
		}
		if record != nil {
			block.FromDate = record.PeriodStart.UTC().Format(time.RFC3339)
			block.ToDate = record.PeriodEnd.UTC().Format(time.RFC3339)
			block.Summary = buildPlayerTransactionSummaryViewFromStoredValues(
				record.Count,
				record.Turnover,
				record.Bet,
				record.Win,
				record.Winlose,
				record.JPShare,
				record.JPWin,
				displayCurrency,
			)
			if record.SyncedAt.After(refreshedAt) {
				refreshedAt = record.SyncedAt
			}
		}

		blocks[scope.PeriodType] = block
	}

	if refreshedAt.IsZero() {
		refreshedAt = time.Now().UTC()
	}

	return &playerTransactionOverviewResult{
		Currency:    displayCurrency,
		LoginID:     loginID,
		Daily:       blocks[playerTransactionPeriodDaily],
		Weekly:      blocks[playerTransactionPeriodWeekly],
		Total:       blocks[playerTransactionPeriodTotal],
		RefreshedAt: refreshedAt,
	}, nil
}

func getStoredPlayerDailyGameSummary(userID uint64, statDate string, displayCurrency string) (*playerTransactionSyncResult, error) {
	record, err := getStoredPlayerTransactionStat(userID, playerTransactionPeriodDaily, statDate)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return &playerTransactionSyncResult{
			StatDate: statDate,
			LoginID:  buildPlayerLoginID(userID),
			Summary:  buildPlayerTransactionSummaryViewFromStoredValues(0, 0, 0, 0, 0, 0, 0, displayCurrency),
		}, nil
	}

	return &playerTransactionSyncResult{
		StatDate: record.PeriodKey,
		LoginID:  record.LoginID,
		Summary: buildPlayerTransactionSummaryViewFromStoredValues(
			record.Count,
			record.Turnover,
			record.Bet,
			record.Win,
			record.Winlose,
			record.JPShare,
			record.JPWin,
			displayCurrency,
		),
		SyncedAt: record.SyncedAt,
	}, nil
}

func getStoredPlayerTransactionStat(userID uint64, periodType string, periodKey string) (*dtos.UserGameTransactionStat, error) {
	var record dtos.UserGameTransactionStat
	err := models.GetInstance().DbInstance.
		Where("user_id = ? AND period_type = ? AND period_key = ?", userID, periodType, periodKey).
		First(&record).Error
	if err != nil {
		if errorsIsRecordNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &record, nil
}

func hasMissingPlayerTransactionScope(userID uint64, scopes []playerTransactionScope) (bool, error) {
	for _, scope := range scopes {
		record, err := getStoredPlayerTransactionStat(userID, scope.PeriodType, scope.PeriodKey)
		if err != nil {
			return false, err
		}
		if record == nil {
			return true, nil
		}
	}

	return false, nil
}

func needsPlayerTransactionScopeRepair(userID uint64, scopes []playerTransactionScope) (bool, error) {
	for _, scope := range scopes {
		record, err := getStoredPlayerTransactionStat(userID, scope.PeriodType, scope.PeriodKey)
		if err != nil {
			return false, err
		}
		if record == nil {
			continue
		}
		if playerTransactionScopeStartMismatch(record, scope) {
			return true, nil
		}
	}

	return false, nil
}

func playerTransactionScopeStartMismatch(record *dtos.UserGameTransactionStat, scope playerTransactionScope) bool {
	if record == nil {
		return false
	}

	loc := playerTransactionLocation()
	recordStart := record.PeriodStart.In(loc).Format("2006-01-02 15:04:05")
	expectedStart := scope.FromDate.In(loc).Format("2006-01-02 15:04:05")
	return recordStart != expectedStart
}

func upsertPlayerTransactionStat(
	userID uint64,
	loginID string,
	scope playerTransactionScope,
	data *provider.PlayerSummaryResult,
	syncedAt time.Time,
) error {
	return models.GetInstance().DbInstance.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "period_type"}, {Name: "period_key"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"period_start": scope.FromDate.UTC(),
			"period_end":   scope.ToDate.UTC(),
			"login_id":     loginID,
			"amount_unit":  "U",
			"count":        data.Count,
			"turnover":     data.Turnover,
			"bet":          data.Bet,
			"win":          data.Win,
			"winlose":      data.Winlose,
			"jp_share":     data.JPShare,
			"jp_win":       data.JPWin,
			"synced_at":    syncedAt,
			"updated_at":   gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(&dtos.UserGameTransactionStat{
		UserID:      userID,
		PeriodType:  scope.PeriodType,
		PeriodKey:   scope.PeriodKey,
		PeriodStart: scope.FromDate.UTC(),
		PeriodEnd:   scope.ToDate.UTC(),
		LoginID:     loginID,
		AmountUnit:  "U",
		Count:       data.Count,
		Turnover:    data.Turnover,
		Bet:         data.Bet,
		Win:         data.Win,
		Winlose:     data.Winlose,
		JPShare:     data.JPShare,
		JPWin:       data.JPWin,
		SyncedAt:    syncedAt,
	}).Error
}

func mirrorSummaryToDailyUserStats(userID uint64, statDate string, data *provider.PlayerSummaryResult) error {
	if !isCurrentUTCDate(statDate) && isZeroPlayerSummary(data) {
		var existing dtos.DailyUserStats
		err := models.GetInstance().DbInstance.
			Where("user_id = ? AND stat_date = ?", userID, statDate).
			First(&existing).Error
		if err == nil && (existing.BetAmount > 0 || existing.WinAmount != 0) {
			log.Printf(
				"[GameSummarySync] skip overwriting historical daily_user_stats with zero summary user_id=%d stat_date=%s existing_bet=%.2f existing_win=%.2f\n",
				userID,
				statDate,
				existing.BetAmount,
				existing.WinAmount,
			)
			return nil
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
	}

	return models.GetInstance().DbInstance.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "stat_date"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"bet_amount": data.Turnover,
			"win_amount": data.Winlose,
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(&dtos.DailyUserStats{
		UserID:    userID,
		StatDate:  statDate,
		BetAmount: data.Turnover,
		WinAmount: data.Winlose,
	}).Error
}

func toPlayerTransactionSummaryView(data *provider.PlayerSummaryResult, displayCurrency string) playerTransactionSummaryView {
	if data == nil {
		return buildPlayerTransactionSummaryViewFromStoredValues(0, 0, 0, 0, 0, 0, 0, displayCurrency)
	}

	return buildPlayerTransactionSummaryViewFromStoredValues(
		data.Count,
		data.Turnover,
		data.Bet,
		data.Win,
		data.Winlose,
		data.JPShare,
		data.JPWin,
		displayCurrency,
	)
}

func normalizeStatDate(value string) (string, error) {
	loc := playerTransactionLocation()
	if strings.TrimSpace(value) == "" {
		return time.Now().In(loc).Format("2006-01-02"), nil
	}

	t, err := time.ParseInLocation("2006-01-02", value, loc)
	if err != nil {
		return "", err
	}

	return t.Format("2006-01-02"), nil
}

func normalizePlayerTransactionBackfillRange(user *dtos.User, startDate string, endDate string) (time.Time, time.Time, error) {
	loc := playerTransactionLocation()

	var start time.Time
	var end time.Time
	var err error

	if strings.TrimSpace(startDate) == "" {
		start = startOfDayInLocation(user.CreatedAt.In(loc), loc)
	} else {
		start, err = time.ParseInLocation("2006-01-02", startDate, loc)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid start_date, expected YYYY-MM-DD")
		}
	}

	if strings.TrimSpace(endDate) == "" {
		end = startOfDayInLocation(time.Now().In(loc), loc)
	} else {
		end, err = time.ParseInLocation("2006-01-02", endDate, loc)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid end_date, expected YYYY-MM-DD")
		}
	}

	if end.Before(start) {
		return time.Time{}, time.Time{}, fmt.Errorf("end_date must be on or after start_date")
	}

	return start, end, nil
}

func isCurrentUTCDate(statDate string) bool {
	return statDate == time.Now().In(playerTransactionLocation()).Format("2006-01-02")
}

func isIgnorablePlayerSummaryError(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(strings.TrimSpace(err.Error()))
	if message == "" {
		return false
	}

	return strings.HasSuffix(message, "provider error:") ||
		strings.HasSuffix(message, "message=") ||
		(strings.Contains(message, "not exist") && strings.Contains(message, "login")) ||
		(strings.Contains(message, "not found") && strings.Contains(message, "login")) ||
		strings.Contains(message, "member not exist") ||
		strings.Contains(message, "user not exist")
}

func fetchPlayerSummaryByRangeOrZero(client *provider.HedoClient, loginID string, fromDate time.Time, toDate time.Time) (*provider.PlayerSummaryResult, error) {
	data, err := client.GetPlayerSummaryByRange(
		loginID,
		fromDate,
		toDate,
		helpers.GetCfgInstance().Conf.Agentid,
		helpers.GetCfgInstance().Conf.Agentapi,
	)
	if err != nil {
		if isIgnorablePlayerSummaryError(err) {
			return zeroPlayerSummary(loginID), nil
		}
		return nil, err
	}

	return data, nil
}

func shouldFallbackToDetailSummary(scope playerTransactionScope, data *provider.PlayerSummaryResult, allowHistoricalDetailFallback bool) bool {
	if scope.PeriodType != playerTransactionPeriodDaily {
		return false
	}
	if !allowHistoricalDetailFallback && !isCurrentUTCDate(scope.PeriodKey) {
		return false
	}
	return isZeroPlayerSummary(data)
}

func shouldTryDetailFallbackOnSummaryError(scope playerTransactionScope, allowHistoricalDetailFallback bool) bool {
	return scope.PeriodType == playerTransactionPeriodDaily &&
		(allowHistoricalDetailFallback || isCurrentUTCDate(scope.PeriodKey))
}

func isZeroPlayerSummary(data *provider.PlayerSummaryResult) bool {
	if data == nil {
		return true
	}

	return data.Count == 0 &&
		data.Turnover == 0 &&
		data.Bet == 0 &&
		data.Win == 0 &&
		data.Winlose == 0 &&
		data.JPShare == 0 &&
		data.JPWin == 0
}

func aggregatePlayerDailySummaryFromDetails(
	client *provider.HedoClient,
	userID uint64,
	loginID string,
	scope playerTransactionScope,
) (*provider.PlayerSummaryResult, error) {
	const rowPerPage = 200

	summary := zeroPlayerSummary(loginID)
	seen := make(map[string]struct{})
	providerCodes, err := loadActiveGameProviderCodes()
	if err != nil {
		return nil, err
	}

	var firstNonIgnorableErr error
	successCount := 0
	for _, providerCode := range providerCodes {
		providerSuccess, providerErr := aggregatePlayerDailySummaryFromSingleProviderDetails(
			client,
			userID,
			loginID,
			scope,
			providerCode,
			rowPerPage,
			summary,
			seen,
		)
		if providerErr != nil {
			if isInvalidGameProviderError(providerErr) || isIgnorablePlayerSummaryError(providerErr) {
				continue
			}
			if firstNonIgnorableErr == nil {
				firstNonIgnorableErr = providerErr
			}
			continue
		}
		if providerSuccess {
			successCount++
		}
	}

	if successCount == 0 && firstNonIgnorableErr != nil {
		return nil, firstNonIgnorableErr
	}

	if summary.Count > 0 && summary.Turnover == 0 && summary.Bet == 0 && summary.Win == 0 && summary.Winlose == 0 {
		log.Printf(
			"[GameSummarySync] user_id=%d login_id=%s stat_date=%s detail fallback rows found but mapped amounts are zero\n",
			userID,
			loginID,
			scope.PeriodKey,
		)
	}

	return summary, nil
}

type m7GameHistoryWindow struct {
	from time.Time
	to   time.Time
}

func loadActiveM7GameProviderIDByCode() (map[string]int, error) {
	var providers []dtos.GameProvider
	if err := models.GetInstance().DbInstance.
		Select("id, code").
		Where("platform_code = ? AND status = ?", m7GamePlatformCode, 1).
		Find(&providers).Error; err != nil {
		return nil, err
	}

	result := make(map[string]int, len(providers))
	for _, item := range providers {
		code := strings.ToUpper(strings.TrimSpace(item.Code))
		if code == "" {
			continue
		}
		if !isAllowedM7GameHistoryProvider(code) {
			continue
		}
		result[code] = int(item.ID)
	}
	return result, nil
}

func isAllowedM7GameHistoryProvider(code string) bool {
	_, ok := m7GameHistoryProviderWhitelist[strings.ToUpper(strings.TrimSpace(code))]
	return ok
}

func getM7GameHistoryWithRateLimit(
	client *provider.M7Client,
	loginID string,
	thirdPartyType string,
	from time.Time,
	to time.Time,
	pageIndex int,
	pageSize int,
) (*provider.M7GameHistoryResponse, error) {
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		resp, err := getM7GameHistoryWithRateLimitOnce(client, loginID, thirdPartyType, from, to, pageIndex, pageSize)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if !isRetryableM7GameHistoryError(err) {
			break
		}
		log.Printf(
			"[GameSummarySync][M7] thirdPartyType=%s page=%d retry=%d after error: %v\n",
			thirdPartyType,
			pageIndex,
			attempt,
			err,
		)
	}
	return nil, lastErr
}

func getM7GameHistoryWithRateLimitOnce(
	client *provider.M7Client,
	loginID string,
	thirdPartyType string,
	from time.Time,
	to time.Time,
	pageIndex int,
	pageSize int,
) (*provider.M7GameHistoryResponse, error) {
	m7GameHistoryRateLimitMu.Lock()
	waitFor := m7GameHistoryRequestGap - time.Since(m7GameHistoryLastRequestAt)
	if waitFor > 0 {
		time.Sleep(waitFor)
	}
	m7GameHistoryLastRequestAt = time.Now()
	m7GameHistoryRateLimitMu.Unlock()

	return client.GetGameHistory(loginID, thirdPartyType, from, to, pageIndex, pageSize)
}

func isRetryableM7GameHistoryError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "context deadline exceeded") ||
		strings.Contains(message, "client.timeout") ||
		strings.Contains(message, "timeout") ||
		strings.Contains(message, "code=1049") ||
		strings.Contains(message, "frequency")
}

func aggregateM7PlayerDailySummaryFromStoredDetails(
	userID uint64,
	loginID string,
	scope playerTransactionScope,
	providerIDs map[string]int,
) (*provider.PlayerSummaryResult, error) {
	providerIDList := m7ProviderIDList(providerIDs)
	if len(providerIDList) == 0 {
		return zeroPlayerSummary(loginID), nil
	}

	type aggregateRow struct {
		Count    int64   `gorm:"column:count"`
		Turnover float64 `gorm:"column:turnover"`
		Bet      float64 `gorm:"column:bet"`
		Win      float64 `gorm:"column:win"`
		Winlose  float64 `gorm:"column:winlose"`
		JPShare  float64 `gorm:"column:jp_share"`
		JPWin    float64 `gorm:"column:jp_win"`
	}

	var row aggregateRow
	if err := models.GetInstance().DbInstance.
		Model(&dtos.UserGameTransactionDetail{}).
		Where("user_id = ? AND stat_date = ? AND game_provider_code IN ?", userID, scope.PeriodKey, providerIDList).
		Select(`
			COALESCE(COUNT(DISTINCT CASE
				WHEN turnover <> 0 OR bet <> 0 OR win <> 0 OR winlose <> 0
				THEN COALESCE(NULLIF(round_id, ''), NULLIF(external_id, ''), record_hash)
			END), 0) AS count,
			COALESCE(SUM(turnover), 0) AS turnover,
			COALESCE(SUM(bet), 0) AS bet,
			COALESCE(SUM(win), 0) AS win,
			COALESCE(SUM(winlose), 0) AS winlose,
			COALESCE(SUM(jp_share), 0) AS jp_share,
			COALESCE(SUM(jp_win), 0) AS jp_win
		`).
		Scan(&row).Error; err != nil {
		return nil, err
	}

	return &provider.PlayerSummaryResult{
		LoginId:  loginID,
		Count:    int(row.Count),
		Turnover: row.Turnover,
		Bet:      row.Bet,
		Win:      row.Win,
		Winlose:  row.Winlose,
		JPShare:  row.JPShare,
		JPWin:    row.JPWin,
	}, nil
}

func m7ProviderIDList(providerIDs map[string]int) []int {
	ids := make([]int, 0, len(providerIDs))
	seen := make(map[int]struct{}, len(providerIDs))
	for _, id := range providerIDs {
		if id <= 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids
}

func m7GameProviderIDForHistoryRow(row map[string]any, providerIDs map[string]int) int {
	code := strings.ToUpper(strings.TrimSpace(readAnyString(row, "thirdPartyType", "ThirdPartyType")))
	if code == "" {
		code = strings.ToUpper(strings.TrimSpace(readAnyString(row, "source", "Source")))
	}
	if code == "" {
		code = inferM7ProviderCodeFromThirdParty(row, providerIDs)
	}
	return providerIDs[code]
}

func inferM7ProviderCodeFromThirdParty(row map[string]any, providerIDs map[string]int) string {
	thirdParty := strings.ToUpper(strings.TrimSpace(readAnyString(row, "thirdParty", "ThirdParty")))
	if thirdParty == "" {
		return ""
	}

	for _, suffix := range []string{"_SLOT", "_LIVE", "_SPORT", "_FISH"} {
		candidate := thirdParty + suffix
		if _, exists := providerIDs[candidate]; exists {
			return candidate
		}
	}
	return thirdParty
}

func resolveTransactionDetailGameCode(row map[string]any, gameProviderCode int) string {
	candidates := []string{
		readAnyString(row, "GamesId", "gamesId"),
		readAnyString(row, "GameId", "gameId"),
		readAnyString(row, "GameCode", "ProviderGameCode", "gameNo", "GameNo"),
	}
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		if gameProviderCode <= 0 {
			return candidate
		}
		var count int64
		if err := models.GetInstance().DbInstance.
			Model(&dtos.Game{}).
			Where("provider_id = ? AND game_code = ?", gameProviderCode, candidate).
			Count(&count).Error; err == nil && count > 0 {
			return candidate
		}
	}
	for _, candidate := range candidates {
		if strings.TrimSpace(candidate) != "" {
			return strings.TrimSpace(candidate)
		}
	}
	return ""
}

func isSettledM7GameHistoryRow(row map[string]any) bool {
	value, exists := lookupRowValue(row, "status")
	if !exists || value == nil {
		return true
	}

	switch v := value.(type) {
	case float64:
		return int(v) == 1
	case float32:
		return int(v) == 1
	case int:
		return v == 1
	case int8:
		return v == 1
	case int16:
		return v == 1
	case int32:
		return v == 1
	case int64:
		return v == 1
	case uint:
		return v == 1
	case uint8:
		return v == 1
	case uint16:
		return v == 1
	case uint32:
		return v == 1
	case uint64:
		return v == 1
	case string:
		return strings.TrimSpace(v) == "1"
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v)) == "1"
	}
}

func addPlayerSummary(base *provider.PlayerSummaryResult, extra *provider.PlayerSummaryResult) {
	if base == nil || extra == nil {
		return
	}

	base.Count += extra.Count
	base.Turnover += extra.Turnover
	base.Bet += extra.Bet
	base.Win += extra.Win
	base.Winlose += extra.Winlose
	base.JPShare += extra.JPShare
	base.JPWin += extra.JPWin
}

func aggregatePlayerDailySummaryFromSingleProviderDetails(
	client *provider.HedoClient,
	userID uint64,
	loginID string,
	scope playerTransactionScope,
	gameProviderCode int,
	rowPerPage int,
	summary *provider.PlayerSummaryResult,
	seen map[string]struct{},
) (bool, error) {
	pageIndex := 1
	hadSuccess := false

	for {
		resp, err := client.GetPlayerTransactionDetails(
			loginID,
			scope.FromDate,
			scope.ToDate,
			gameProviderCode,
			pageIndex,
			rowPerPage,
			helpers.GetCfgInstance().Conf.Agentid,
			helpers.GetCfgInstance().Conf.Agentapi,
		)
		if err != nil {
			return hadSuccess, err
		}
		hadSuccess = true
		if resp == nil || len(resp.Data) == 0 {
			break
		}

		for index, row := range resp.Data {
			key := buildTransactionDetailRecordKey(row, gameProviderCode, pageIndex, index)
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			if _, err := persistPlayerTransactionDetailRow(userID, loginID, scope, gameProviderCode, row, key, helpers.GetCfgInstance().Conf.Agentid); err != nil {
				log.Printf(
					"[GameSummarySync] user_id=%d login_id=%s stat_date=%s persist detail row failed: %v\n",
					userID,
					loginID,
					scope.PeriodKey,
					err,
				)
			}
			mergeTransactionDetailIntoSummary(summary, row)
		}

		if resp.TotalPage > 0 && pageIndex >= resp.TotalPage {
			break
		}
		if len(resp.Data) < rowPerPage && resp.TotalPage <= 0 {
			break
		}
		pageIndex++
		if pageIndex > 10000 {
			return hadSuccess, fmt.Errorf("detail page overflow for provider %d on %s", gameProviderCode, scope.PeriodKey)
		}
	}

	return hadSuccess, nil
}

func persistPlayerTransactionDetailRow(
	userID uint64,
	loginID string,
	scope playerTransactionScope,
	gameProviderCode int,
	row map[string]any,
	recordKey string,
	agentID string,
) (bool, error) {
	bet := readTransactionDetailBet(row)
	turnover := readTransactionDetailTurnover(row)
	rawWin := readTransactionDetailWin(row)
	rawWinlose, hasRawWinlose := readTransactionDetailRawWinlose(row)
	winlose := normalizeTransactionDetailPlayerWinlose(rawWinlose, hasRawWinlose, bet, rawWin)
	win := normalizeTransactionDetailWin(rawWin, bet, winlose)
	if !isCountableGameTransaction(turnover, bet, win, winlose) {
		return false, nil
	}

	rawJSON, err := stdjson.Marshal(row)
	if err != nil {
		return false, err
	}

	recordHash := sha256.Sum256([]byte(fmt.Sprintf("%d|%s|%s", userID, scope.PeriodKey, recordKey)))
	if strings.TrimSpace(agentID) == "" {
		agentID = helpers.GetCfgInstance().Conf.Agentid
	}

	detail := dtos.UserGameTransactionDetail{
		UserID:           userID,
		StatDate:         scope.PeriodKey,
		AgentID:          agentID,
		ProviderUserID:   readAnyString(row, "UserId", "ProviderUserId", "ProviderUserID", "userViewId", "userId"),
		LoginID:          loginID,
		GameProviderCode: gameProviderCode,
		GameCode:         resolveTransactionDetailGameCode(row, gameProviderCode),
		ExternalID:       readAnyString(row, "ExternalId", "ExternalID", "TransactionId", "TxnId", "TransId", "TxId", "txId", "BetId", "Id", "ID"),
		RoundID:          readAnyString(row, "RoundId", "RoundID", "GameRoundId", "GameRound", "gameRound"),
		Type:             resolveTransactionDetailType(row),
		StartDate:        readTransactionDetailTime(row, "StartDate", "DateTime", "CreatedAt", "createdAt", "CreateTime", "createTime", "TransactionTime", "BetTime", "betTime"),
		EndDate:          readTransactionDetailTime(row, "EndDate", "UpdatedAt", "updatedAt", "SettleTime", "settleTime", "CreateTime", "createTime"),
		StartBalance:     readAnyFloat(row, "StartBalance", "startBalance", "start_balance", "BeforeBalance", "beforeBalance", "before_balance", "BalanceBefore", "balanceBefore", "balance_before"),
		EndBalance:       readAnyFloat(row, "EndBalance", "endBalance", "end_balance", "AfterBalance", "afterBalance", "after_balance", "BalanceAfter", "balanceAfter", "balance_after"),
		Deposit:          readAnyFloat(row, "Deposit", "deposit", "Amount", "amount"),
		Turnover:         turnover,
		Bet:              bet,
		Win:              win,
		Winlose:          winlose,
		JPShare:          readAnyFloat(row, "JPShare", "JackpotShare"),
		JPWin:            readAnyFloat(row, "JPWin", "JackpotWin"),
		Remark:           string(rawJSON),
		RecordHash:       hex.EncodeToString(recordHash[:]),
	}

	err = models.GetInstance().DbInstance.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "record_hash"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"agent_id":           detail.AgentID,
			"provider_user_id":   detail.ProviderUserID,
			"login_id":           detail.LoginID,
			"game_provider_code": detail.GameProviderCode,
			"game_code":          detail.GameCode,
			"external_id":        detail.ExternalID,
			"round_id":           detail.RoundID,
			"type":               detail.Type,
			"start_date":         detail.StartDate,
			"end_date":           detail.EndDate,
			"start_balance":      detail.StartBalance,
			"end_balance":        detail.EndBalance,
			"deposit":            detail.Deposit,
			"turnover":           detail.Turnover,
			"bet":                detail.Bet,
			"win":                detail.Win,
			"winlose":            detail.Winlose,
			"jp_share":           detail.JPShare,
			"jp_win":             detail.JPWin,
			"remark":             detail.Remark,
			"updated_at":         gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(&detail).Error
	if err != nil {
		return false, err
	}
	return true, nil
}

func resolveTransactionDetailType(row map[string]any) int {
	return int(readAnyFloat(row, "Type", "type", "TransactionType", "transactionType"))
}

func loadActiveGameProviderCodes() ([]int, error) {
	var providers []dtos.GameProvider
	if err := models.GetInstance().DbInstance.
		Select("id").
		Where("platform_code = ? AND status = ?", defaultGamePlatformCode, 1).
		Order("id ASC").
		Find(&providers).Error; err != nil {
		return nil, err
	}

	codes := make([]int, 0, len(providers))
	for _, item := range providers {
		codes = append(codes, int(item.ID))
	}
	if len(codes) == 0 {
		return nil, fmt.Errorf("no active game providers found")
	}

	return codes, nil
}

func isInvalidGameProviderError(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(message, "invalid_game_provider")
}

func buildTransactionDetailRecordKey(row map[string]any, gameProviderCode int, pageIndex int, rowIndex int) string {
	parts := []string{
		strconv.Itoa(gameProviderCode),
		readAnyString(row, "ExternalId", "ExternalID", "TransactionId", "TxnId", "TransId", "TxId", "txId", "BetId", "Id", "ID"),
		readAnyString(row, "RoundId", "RoundID", "GameRoundId", "GameRound", "gameRound"),
		readAnyString(row, "GameCode", "ProviderGameCode", "GamesId", "gamesId", "GameId", "gameId"),
		readAnyString(row, "StartDate", "DateTime", "CreatedAt", "TransactionTime", "BetTime", "betTime"),
		readAnyString(row, "EndDate", "UpdatedAt", "SettleTime", "settleTime"),
		fmt.Sprintf("%.4f", readTransactionDetailTurnover(row)),
		fmt.Sprintf("%.4f", readTransactionDetailBet(row)),
		fmt.Sprintf("%.4f", readTransactionDetailWinloseForRecordKey(row)),
	}

	key := strings.Join(parts, "|")
	if strings.Trim(key, "|") != "" {
		return key
	}

	raw, err := stdjson.Marshal(row)
	if err == nil && len(raw) > 0 {
		return fmt.Sprintf("%d|%s", gameProviderCode, string(raw))
	}

	return fmt.Sprintf("page:%d:row:%d", pageIndex, rowIndex)
}

func buildM7GameHistoryTransactionDetailRecordKey(row map[string]any, gameProviderCode int, pageIndex int, rowIndex int) string {
	roundID := strings.TrimSpace(readAnyString(row, "RoundId", "RoundID", "GameRoundId", "GameRound", "gameRound"))
	if roundID != "" {
		return strings.Join([]string{
			"m7-round",
			strconv.Itoa(gameProviderCode),
			roundID,
		}, "|")
	}
	return buildTransactionDetailRecordKey(row, gameProviderCode, pageIndex, rowIndex)
}

func mergeTransactionDetailIntoSummary(summary *provider.PlayerSummaryResult, row map[string]any) {
	if summary == nil {
		return
	}

	bet := readTransactionDetailBet(row)
	turnover := readTransactionDetailTurnover(row)
	rawWin := readTransactionDetailWin(row)
	rawWinlose, hasRawWinlose := readTransactionDetailRawWinlose(row)
	winlose := normalizeTransactionDetailPlayerWinlose(rawWinlose, hasRawWinlose, bet, rawWin)
	win := normalizeTransactionDetailWin(rawWin, bet, winlose)

	if isCountableGameTransaction(turnover, bet, win, winlose) {
		summary.Count++
	}
	summary.Turnover += turnover
	summary.Bet += bet
	summary.Win += win
	summary.Winlose += winlose
	summary.JPShare += readAnyFloat(row, "JPShare", "JackpotShare")
	summary.JPWin += readAnyFloat(row, "JPWin", "JackpotWin")
}

func isCountableGameTransaction(turnover float64, bet float64, win float64, winlose float64) bool {
	return turnover != 0 || bet != 0 || win != 0 || winlose != 0
}

func readTransactionDetailTurnover(row map[string]any) float64 {
	return readAnyFloat(row,
		"Turnover",
		"TurnOver",
		"TurnoverAmount",
		"ValidBet",
		"ValidBetAmount",
		"ValidBetMoney",
		"ValidAmount",
		"ValidStakeAmount",
		"ValidStake",
		"EffectiveBet",
		"EffectiveBetAmount",
		"EffectiveStake",
		"RealBetAmount",
		"Rolling",
		"RollingAmount",
		"Water",
		"StakeValid",
	)
}

func readTransactionDetailBet(row map[string]any) float64 {
	return readAnyFloat(row,
		"Bet",
		"Stake",
		"StakeAmount",
		"BetAmount",
		"BetMoney",
		"BetValue",
		"BetAmt",
		"BettingAmount",
		"Wager",
		"WagerAmount",
		"Amount",
		"TurnoverAmount",
	)
}

func readTransactionDetailWin(row map[string]any) float64 {
	win := readAnyFloat(row,
		"Win",
		"win",
		"Payout",
		"PayOut",
		"Payoff",
		"PayOff",
		"Prize",
		"PrizeAmount",
		"PayoutAmount",
		"PayOutAmount",
		"PayAmount",
		"SettlementAmount",
		"SettleAmount",
		"WinAmount",
		"WinMoney",
		"ReturnAmount",
	)
	if win < 0 {
		return 0
	}
	return win
}

func readTransactionDetailRawWinlose(row map[string]any) (float64, bool) {
	for _, key := range []string{
		"Winlose",
		"WinLoss",
		"WinLose",
		"WinLossAmount",
		"WinLoseAmount",
		"WinloseAmount",
		"WinLossMoney",
		"ProfitLoss",
		"ProfitAndLoss",
		"ProfitLossAmount",
		"NetAmount",
		"NetWin",
		"NetWinAmount",
		"NetProfit",
		"Profit",
	} {
		value, exists := lookupRowValue(row, key)
		if !exists || value == nil {
			continue
		}
		return readAnyFloat(row, key), true
	}

	return 0, false
}

func readTransactionDetailWinloseForRecordKey(row map[string]any) float64 {
	value, exists := readTransactionDetailRawWinlose(row)
	if exists {
		return value
	}
	return readAnyFloat(row, "Winlose", "WinLoss", "WinLose", "ProfitLoss", "NetAmount", "NetWin")
}

func normalizeTransactionDetailPlayerWinlose(rawWinlose float64, hasRawWinlose bool, bet float64, win float64) float64 {
	if hasRawWinlose {
		expectedFromPayout := roundPlayerTransactionAmount(win - bet)
		if win > 0 || bet > 0 {
			if almostEqualTransactionAmount(rawWinlose, expectedFromPayout) {
				return rawWinlose
			}
			if almostEqualTransactionAmount(-rawWinlose, expectedFromPayout) {
				return -rawWinlose
			}
		}

		// Third-party detail rows usually expose platform profit, so flip to player perspective by default.
		return -rawWinlose
	}

	if bet == 0 && win == 0 {
		return 0
	}

	return roundPlayerTransactionAmount(win - bet)
}

func normalizeTransactionDetailWin(rawWin float64, bet float64, playerWinlose float64) float64 {
	if rawWin > 0 {
		return rawWin
	}

	if bet == 0 && playerWinlose == 0 {
		return 0
	}

	derivedWin := roundPlayerTransactionAmount(bet + playerWinlose)
	if derivedWin < 0 {
		return 0
	}

	return derivedWin
}

func almostEqualTransactionAmount(left float64, right float64) bool {
	return math.Abs(left-right) < 0.01
}

func readTransactionDetailTime(row map[string]any, keys ...string) *time.Time {
	text := readAnyString(row, keys...)
	if strings.TrimSpace(text) == "" {
		return nil
	}

	loc := playerTransactionLocation()
	localLayouts := []string{
		"2006-01-02 15:04:05.000",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05.000",
		"2006-01-02T15:04:05",
	}
	for _, layout := range localLayouts {
		if parsed, err := time.ParseInLocation(layout, text, loc); err == nil {
			value := parsed.UTC()
			return &value
		}
	}

	zoneLayouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.000Z07:00",
	}
	for _, layout := range zoneLayouts {
		if parsed, err := time.Parse(layout, text); err == nil {
			value := parsed.UTC()
			return &value
		}
	}

	return nil
}

func readTransactionDetailWinlose(row map[string]any, bet float64, win float64) float64 {
	winlose := readAnyFloat(row,
		"Winlose",
		"WinLoss",
		"WinLose",
		"WinLossAmount",
		"WinLoseAmount",
		"WinloseAmount",
		"WinLossMoney",
		"ProfitLoss",
		"ProfitAndLoss",
		"ProfitLossAmount",
		"NetAmount",
		"NetWin",
		"NetWinAmount",
		"NetProfit",
		"Profit",
	)
	if winlose != 0 {
		return winlose
	}
	if bet == 0 && win == 0 {
		return 0
	}
	return roundPlayerTransactionAmount(win - bet)
}

func readAnyString(row map[string]any, keys ...string) string {
	for _, key := range keys {
		value, exists := lookupRowValue(row, key)
		if !exists || value == nil {
			continue
		}
		switch v := value.(type) {
		case string:
			if trimmed := strings.TrimSpace(v); trimmed != "" {
				return trimmed
			}
		case float64:
			return formatNumericTransactionID(v)
		case float32:
			return formatNumericTransactionID(float64(v))
		case int:
			return strconv.FormatInt(int64(v), 10)
		case int8:
			return strconv.FormatInt(int64(v), 10)
		case int16:
			return strconv.FormatInt(int64(v), 10)
		case int32:
			return strconv.FormatInt(int64(v), 10)
		case int64:
			return strconv.FormatInt(v, 10)
		case uint:
			return strconv.FormatUint(uint64(v), 10)
		case uint8:
			return strconv.FormatUint(uint64(v), 10)
		case uint16:
			return strconv.FormatUint(uint64(v), 10)
		case uint32:
			return strconv.FormatUint(uint64(v), 10)
		case uint64:
			return strconv.FormatUint(v, 10)
		default:
			text := strings.TrimSpace(fmt.Sprintf("%v", v))
			if text != "" && text != "<nil>" {
				return text
			}
		}
	}

	return ""
}

func formatNumericTransactionID(value float64) string {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return ""
	}
	if value == math.Trunc(value) {
		return strconv.FormatInt(int64(value), 10)
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func readAnyFloat(row map[string]any, keys ...string) float64 {
	for _, key := range keys {
		value, exists := lookupRowValue(row, key)
		if !exists || value == nil {
			continue
		}

		switch v := value.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int8:
			return float64(v)
		case int16:
			return float64(v)
		case int32:
			return float64(v)
		case int64:
			return float64(v)
		case uint:
			return float64(v)
		case uint8:
			return float64(v)
		case uint16:
			return float64(v)
		case uint32:
			return float64(v)
		case uint64:
			return float64(v)
		case string:
			text := strings.TrimSpace(v)
			if text == "" {
				continue
			}
			text = strings.ReplaceAll(text, ",", "")
			parsed, err := strconv.ParseFloat(text, 64)
			if err == nil {
				return parsed
			}
		case stdjson.Number:
			parsed, err := v.Float64()
			if err == nil {
				return parsed
			}
		default:
			text := strings.TrimSpace(fmt.Sprintf("%v", v))
			if text == "" || text == "<nil>" {
				continue
			}
			text = strings.ReplaceAll(text, ",", "")
			parsed, err := strconv.ParseFloat(text, 64)
			if err == nil {
				return parsed
			}
		}
	}

	return 0
}

func buildPlayerTransactionSummaryViewFromStoredValues(
	count int,
	turnoverU float64,
	betU float64,
	winU float64,
	winloseU float64,
	jpShareU float64,
	jpWinU float64,
	displayCurrency string,
) playerTransactionSummaryView {
	currency := normalizePlayerTransactionCurrency(displayCurrency)

	return playerTransactionSummaryView{
		Currency:  currency,
		Count:     count,
		Turnover:  convertUToPlayerCurrency(turnoverU, currency),
		Bet:       convertUToPlayerCurrency(betU, currency),
		Win:       convertUToPlayerCurrency(winU, currency),
		Winlose:   convertUToPlayerCurrency(winloseU, currency),
		JPShare:   convertUToPlayerCurrency(jpShareU, currency),
		JPWin:     convertUToPlayerCurrency(jpWinU, currency),
		TurnoverU: turnoverU,
		BetU:      betU,
		WinU:      winU,
		WinloseU:  winloseU,
		JPShareU:  jpShareU,
		JPWinU:    jpWinU,
	}
}

func convertUToPlayerCurrency(amount float64, currency string) float64 {
	currency = normalizePlayerTransactionCurrency(currency)
	if amount == 0 || currency == "USD" {
		return amount
	}

	rate, err := models.GetInstance().GetExchangeRate(currency)
	if err != nil || rate <= 0 {
		return amount
	}

	return roundPlayerTransactionAmount(amount * rate)
}

func roundPlayerTransactionAmount(amount float64) float64 {
	return math.Round(amount*100) / 100
}

func resolvePlayerTransactionCurrency(userID uint64, requested string) string {
	if normalized := normalizePlayerTransactionCurrency(requested); normalized != "" {
		return normalized
	}
	if overridden := getPlayerTransactionCurrencyOverride(userID); overridden != "" {
		return overridden
	}
	if inferred := inferPlayerTransactionCurrency(userID); inferred != "" {
		return inferred
	}
	return "IDR"
}

func normalizePlayerTransactionCurrency(currency string) string {
	normalized := strings.ToUpper(strings.TrimSpace(currency))
	switch normalized {
	case "PHP", "IDR", "USD":
		return normalized
	default:
		return ""
	}
}

func getPlayerTransactionCurrencyOverride(userID uint64) string {
	if userID == 0 {
		return ""
	}
	if err := ensurePlayerTransactionCurrencyOverrideTable(); err != nil {
		return ""
	}

	type currencyRow struct {
		Currency string `gorm:"column:currency"`
	}
	var record currencyRow
	if err := models.GetInstance().DbInstance.
		Model(&dtos.UserGameTransactionCurrencyOverride{}).
		Select("currency").
		Where("user_id = ?", userID).
		Limit(1).
		Scan(&record).Error; err != nil {
		return ""
	}

	return normalizePlayerTransactionCurrency(record.Currency)
}

func savePlayerTransactionCurrencyOverride(ctx context.Context, userID uint64, currency string) error {
	currency = normalizePlayerTransactionCurrency(currency)
	if userID == 0 || currency == "" {
		return nil
	}
	if err := ensurePlayerTransactionCurrencyOverrideTable(); err != nil {
		return err
	}

	return models.GetInstance().DbInstance.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"currency": currency, "updated_at": gorm.Expr("CURRENT_TIMESTAMP")}),
	}).Create(&dtos.UserGameTransactionCurrencyOverride{
		UserID:   userID,
		Currency: currency,
	}).Error
}

func markPlayerTransactionAmountUnit(
	ctx context.Context,
	userID uint64,
	startDate string,
	endDate string,
	amountUnit string,
) error {
	updateQuery := models.GetInstance().DbInstance.WithContext(ctx).
		Model(&dtos.UserGameTransactionStat{}).
		Where("user_id = ? AND amount_unit = ''", userID)

	if strings.TrimSpace(startDate) != "" && strings.TrimSpace(endDate) != "" {
		updateQuery = updateQuery.Where(
			"(period_type = ? AND period_key >= ? AND period_key <= ?) OR period_type IN (?, ?)",
			playerTransactionPeriodDaily,
			startDate,
			endDate,
			playerTransactionPeriodWeekly,
			playerTransactionPeriodTotal,
		)
	}

	return updateQuery.Update("amount_unit", amountUnit).Error
}

func inferPlayerTransactionCurrency(userID uint64) string {
	type orderCurrencyRow struct {
		DstCode   string    `gorm:"column:dst_code"`
		CreatedAt time.Time `gorm:"column:created_at"`
	}

	var latestPayment orderCurrencyRow
	if err := models.GetInstance().DbInstance.
		Model(&dtos.PaymentOrder{}).
		Select("dst_code, created_at").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(1).
		Scan(&latestPayment).Error; err == nil {
		if currency := playerTransactionCurrencyFromDstCode(latestPayment.DstCode); currency != "" {
			return currency
		}
	}

	var latestWithdraw orderCurrencyRow
	if err := models.GetInstance().DbInstance.
		Model(&dtos.WithdrawOrder{}).
		Select("dst_code, created_at").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(1).
		Scan(&latestWithdraw).Error; err == nil {
		if currency := playerTransactionCurrencyFromDstCode(latestWithdraw.DstCode); currency != "" {
			return currency
		}
	}

	return ""
}

func playerTransactionCurrencyFromDstCode(dstCode string) string {
	normalized := strings.ToUpper(strings.TrimSpace(dstCode))
	switch {
	case normalized == "":
		return ""
	case strings.HasPrefix(normalized, "GCASH"),
		strings.HasPrefix(normalized, "BDO"),
		strings.HasPrefix(normalized, "BPI"),
		strings.HasPrefix(normalized, "METROBANK"),
		strings.HasPrefix(normalized, "LANDBANK"),
		strings.HasPrefix(normalized, "PNB"),
		strings.HasPrefix(normalized, "SECURITY"),
		strings.HasPrefix(normalized, "UNIONBANK"),
		strings.HasPrefix(normalized, "CHINABANK"),
		strings.HasPrefix(normalized, "RCBC"),
		strings.HasPrefix(normalized, "EASTWEST"):
		return "PHP"
	case strings.HasPrefix(normalized, "USDT"), normalized == "PAYPAL":
		return "USD"
	default:
		return "IDR"
	}
}

func lookupRowValue(row map[string]any, key string) (any, bool) {
	if row == nil {
		return nil, false
	}
	if value, exists := row[key]; exists {
		return value, true
	}

	target := normalizeTransactionDetailKey(key)
	for currentKey, value := range row {
		if normalizeTransactionDetailKey(currentKey) == target {
			return value, true
		}
	}

	return nil, false
}

func normalizeTransactionDetailKey(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, "_", "")
	value = strings.ReplaceAll(value, "-", "")
	return value
}

func zeroPlayerSummary(loginID string) *provider.PlayerSummaryResult {
	return &provider.PlayerSummaryResult{
		LoginId: loginID,
	}
}

func buildPlayerLoginID(userID uint64) string {
	return fmt.Sprintf("%su%d", helpers.GetCfgInstance().Conf.Prefix, userID)
}

func registerProviderAccountIfNeeded(client *provider.HedoClient, userID uint64, loginID string) {
	thirdPartyPassword := common.GenerateGamePassword(userID)
	if err := client.RegisterUser(
		helpers.GetCfgInstance().Conf.Agentid,
		loginID,
		thirdPartyPassword,
		helpers.GetCfgInstance().Conf.Agentapi,
	); err != nil {
		log.Printf("[GameSummarySync] user_id=%d login_id=%s register skipped: %v\n", userID, loginID, err)
	}
}

func getUserByID(userID uint64) (*dtos.User, error) {
	var user dtos.User
	if err := models.GetInstance().DbInstance.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func buildDailyScope(statDate string, nowUTC time.Time) (playerTransactionScope, error) {
	loc := playerTransactionLocation()
	start, err := time.ParseInLocation("2006-01-02", statDate, loc)
	if err != nil {
		return playerTransactionScope{}, err
	}

	end := start.Add(24 * time.Hour).Add(-time.Millisecond)

	return playerTransactionScope{
		PeriodType: playerTransactionPeriodDaily,
		PeriodKey:  statDate,
		FromDate:   start,
		ToDate:     end,
	}, nil
}

func buildCurrentPlayerTransactionScopes(user *dtos.User, nowUTC time.Time) []playerTransactionScope {
	loc := playerTransactionLocation()
	nowLocal := nowUTC.In(loc)
	dailyStart := startOfDayInLocation(nowLocal, loc)
	weeklyStart := weekStartInLocation(nowLocal, loc)
	totalStart := startOfDayInLocation(user.CreatedAt.In(loc), loc)

	return []playerTransactionScope{
		{
			PeriodType: playerTransactionPeriodDaily,
			PeriodKey:  dailyStart.Format("2006-01-02"),
			FromDate:   dailyStart,
			ToDate:     nowLocal,
		},
		{
			PeriodType: playerTransactionPeriodWeekly,
			PeriodKey:  weeklyStart.Format("2006-01-02"),
			FromDate:   weeklyStart,
			ToDate:     nowLocal,
		},
		{
			PeriodType: playerTransactionPeriodTotal,
			PeriodKey:  playerTransactionPeriodTotalKey,
			FromDate:   totalStart,
			ToDate:     nowLocal,
		},
	}
}

func startOfDayInLocation(value time.Time, loc *time.Location) time.Time {
	current := value.In(loc)
	return time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, loc)
}

func weekStartInLocation(value time.Time, loc *time.Location) time.Time {
	current := startOfDayInLocation(value, loc)
	weekday := int(current.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	return current.AddDate(0, 0, -(weekday - 1))
}

func isTruthyFlag(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func errorsIsRecordNotFound(err error) bool {
	return err == gorm.ErrRecordNotFound
}

func ensurePlayerTransactionStatsTable() error {
	playerTransactionStatsTableOnce.Do(func() {
		db := models.GetInstance().DbInstance
		if db == nil {
			playerTransactionStatsTableErr = fmt.Errorf("db is nil")
			return
		}

		playerTransactionStatsTableErr = db.AutoMigrate(
			&dtos.UserGameTransactionStat{},
			&dtos.UserGameTransactionCurrencyOverride{},
		)
	})

	return playerTransactionStatsTableErr
}

func ensurePlayerTransactionDetailTable() error {
	playerTransactionDetailTableOnce.Do(func() {
		db := models.GetInstance().DbInstance
		if db == nil {
			playerTransactionDetailTableErr = fmt.Errorf("db is nil")
			return
		}

		playerTransactionDetailTableErr = db.AutoMigrate(&dtos.UserGameTransactionDetail{})
	})

	return playerTransactionDetailTableErr
}

func ensurePlayerTransactionCurrencyOverrideTable() error {
	playerTransactionCurrencyTableOnce.Do(func() {
		db := models.GetInstance().DbInstance
		if db == nil {
			playerTransactionCurrencyTableErr = fmt.Errorf("db is nil")
			return
		}

		playerTransactionCurrencyTableErr = db.AutoMigrate(&dtos.UserGameTransactionCurrencyOverride{})
	})

	return playerTransactionCurrencyTableErr
}

func resetPlayerTransactionDetailRows(ctx context.Context, userID uint64, statDate string) error {
	return models.GetInstance().DbInstance.WithContext(ctx).
		Where("user_id = ? AND stat_date = ?", userID, statDate).
		Delete(&dtos.UserGameTransactionDetail{}).Error
}

func TriggerTodayPlayerDailyGameSummarySyncIfStale(userID uint64, minInterval time.Duration, force bool) {
	if !force {
		shouldRun, err := shouldSyncTodayPlayerDailyGameSummary(userID, minInterval)
		if err != nil {
			log.Printf("[GameSummarySync] user_id=%d freshness check failed: %v\n", userID, err)
		} else if !shouldRun {
			return
		}
	}

	nowUTC := time.Now().UTC()

	playerTransactionSyncStateMu.Lock()
	lastRun := playerTransactionSyncLastRun[userID]
	if nowUTC.Sub(lastRun) < 5*time.Second {
		playerTransactionSyncStateMu.Unlock()
		return
	}
	playerTransactionSyncLastRun[userID] = nowUTC
	playerTransactionSyncStateMu.Unlock()

	statDate := nowUTC.In(playerTransactionLocation()).Format("2006-01-02")
	go func() {
		if _, err := SyncPlayerDailyGameSummary(context.Background(), userID, statDate); err != nil {
			log.Printf("[GameSummarySync] user_id=%d stat_date=%s failed: %v\n", userID, statDate, err)
		}
	}()
}

func shouldSyncTodayPlayerDailyGameSummary(userID uint64, minInterval time.Duration) (bool, error) {
	if minInterval <= 0 {
		return true, nil
	}

	if err := ensurePlayerTransactionStatsTable(); err != nil {
		return false, err
	}

	todayKey := time.Now().In(playerTransactionLocation()).Format("2006-01-02")
	record, err := getStoredPlayerTransactionStat(userID, playerTransactionPeriodDaily, todayKey)
	if err != nil {
		return false, err
	}
	if record == nil {
		return true, nil
	}

	return time.Since(record.SyncedAt.UTC()) >= minInterval, nil
}

func listUsersForLatestGameTransactionSync(req syncAllGameTransactionsRequest) ([]dtos.User, uint64, error) {
	db := models.GetInstance().DbInstance

	if len(req.UserIDs) > 0 {
		userIDs := make([]uint64, 0, len(req.UserIDs))
		seen := make(map[uint64]struct{}, len(req.UserIDs))
		for _, userID := range req.UserIDs {
			if userID == 0 {
				continue
			}
			if _, ok := seen[userID]; ok {
				continue
			}
			seen[userID] = struct{}{}
			userIDs = append(userIDs, userID)
		}

		if len(userIDs) == 0 {
			return []dtos.User{}, 0, nil
		}

		var users []dtos.User
		if err := db.Where("id IN ? AND status = 1", userIDs).Order("id ASC").Find(&users).Error; err != nil {
			return nil, 0, err
		}
		return users, 0, nil
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	query := db.Model(&dtos.User{}).
		Where("status = 1")
	if req.StartUserID > 0 {
		query = query.Where("id > ?", req.StartUserID)
	}

	var users []dtos.User
	if err := query.Order("id ASC").Limit(limit + 1).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	nextStartUserID := uint64(0)
	if len(users) > limit {
		nextStartUserID = users[limit-1].ID
		users = users[:limit]
	}

	return users, nextStartUserID, nil
}

func isCurrentPlayerTransactionStatFresh(statDate string, user *dtos.User, minInterval time.Duration) bool {
	if user == nil || minInterval <= 0 {
		return false
	}

	record, err := getStoredPlayerTransactionStat(user.ID, playerTransactionPeriodDaily, statDate)
	if err != nil || record == nil {
		return false
	}

	return time.Since(record.SyncedAt.UTC()) < minInterval
}

func playerTransactionLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("GMT+8", 8*60*60)
	}

	return loc
}
