package controllers

import (
	"context"
	"errors"
	"fmt"
	"gogogo/models"
	"gogogo/models/dtos"
	"gogogo/provider"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const m7ppTransactionSyncPageSize = 2000
const m7ppTransactionMaxWindow = 6 * time.Hour

func SyncM7PPTransactions(ctx context.Context, from, to time.Time) (int, error) {
	if !to.After(from) {
		return 0, fmt.Errorf("invalid M7PP transaction range")
	}

	totalPersisted := 0
	for _, window := range splitM7PPTransactionWindows(from, to) {
		windowFrom, windowTo := window[0], window[1]
		persisted, err := syncM7PPTransactionWindow(ctx, windowFrom, windowTo)
		totalPersisted += persisted
		if err != nil {
			return totalPersisted, fmt.Errorf("M7PP transaction window %s to %s: %w", windowFrom.Format(time.RFC3339), windowTo.Format(time.RFC3339), err)
		}
	}
	return totalPersisted, nil
}

func splitM7PPTransactionWindows(from, to time.Time) [][2]time.Time {
	windows := make([][2]time.Time, 0)
	for windowFrom := from; windowFrom.Before(to); {
		windowTo := windowFrom.Add(m7ppTransactionMaxWindow)
		if windowTo.After(to) {
			windowTo = to
		}
		windows = append(windows, [2]time.Time{windowFrom, windowTo})
		windowFrom = windowTo
	}
	return windows
}

func syncM7PPTransactionWindow(ctx context.Context, from, to time.Time) (int, error) {
	client, err := provider.NewM7PPClient(m7ppProviderConfig())
	if err != nil {
		return 0, err
	}
	if err := ensurePlayerTransactionDetailTable(); err != nil {
		return 0, err
	}

	persistedCount := 0
	type affectedScope struct {
		userID   uint64
		loginID  string
		scope    playerTransactionScope
		previous *provider.PlayerSummaryResult
	}
	affected := make(map[string]*affectedScope)
	for page := 1; ; page++ {
		select {
		case <-ctx.Done():
			return persistedCount, ctx.Err()
		default:
		}
		transactions, totalPages, err := client.FetchTransactions(from.UnixMilli(), to.UnixMilli(), page, m7ppTransactionSyncPageSize)
		if err != nil {
			return persistedCount, err
		}
		for _, transaction := range transactions {
			if transaction.Status != 1 && transaction.Status != 3 {
				continue
			}
			userID, ok := parseLocalUserIDFromPlayerLoginID(transaction.Username)
			if !ok {
				continue
			}
			providerID, err := resolveM7PPTransactionProviderID(transaction.GameCode, transaction.GameCategoryCode)
			if err != nil {
				return persistedCount, fmt.Errorf("resolve M7PP provider game=%s category=%s: %w", transaction.GameCode, transaction.GameCategoryCode, err)
			}
			betTime := m7ppTransactionTime(transaction.VendorBetTime, transaction.CreateTime, transaction.UpdateTime)
			settleTime := m7ppTransactionTime(transaction.VendorSettleTime, transaction.UpdateTime, transaction.CreateTime)
			statDate := betTime.In(playerTransactionLocation()).Format("2006-01-02")
			scope := playerTransactionScope{PeriodType: playerTransactionPeriodDaily, PeriodKey: statDate,
				FromDate: startOfPlayerTransactionDay(betTime), ToDate: startOfPlayerTransactionDay(betTime).Add(24 * time.Hour)}
			affectedKey := fmt.Sprintf("%d|%s", userID, statDate)
			if _, exists := affected[affectedKey]; !exists {
				previous, err := aggregateM7PPDailySummary(userID, buildPlayerLoginID(userID), scope)
				if err != nil {
					return persistedCount, err
				}
				affected[affectedKey] = &affectedScope{userID: userID, loginID: buildPlayerLoginID(userID), scope: scope, previous: previous}
			}
			row := map[string]any{
				"source": "M7PP_TRANSACTION_LIST", "UserId": transaction.Username, "GameCode": transaction.GameCode,
				"ExternalId": transaction.BetID, "RoundId": transaction.RoundID, "Type": 2,
				"StartDate": betTime.Format(time.RFC3339Nano), "EndDate": settleTime.Format(time.RFC3339Nano),
				"Turnover": transaction.EffectiveTurnover, "Bet": transaction.BetAmount,
				"Win": transaction.WinAmount, "Winlose": transaction.WinLoss,
				"JPWin": transaction.JackpotAmount, "Currency": transaction.CurrencyCode,
				"Status": transaction.Status, "VendorCode": transaction.VendorCode,
			}
			recordKey := buildM7PPTransactionRecordKey(transaction)
			persisted, err := persistPlayerTransactionDetailRow(userID, buildPlayerLoginID(userID), scope, providerID, row, recordKey, "M7PP")
			if err != nil {
				log.Printf("[GameSummarySync][M7PP] user=%d bet=%s persist failed: %v", userID, transaction.BetID, err)
				continue
			}
			if persisted {
				persistedCount++
			}
		}
		if len(transactions) == 0 || (totalPages > 0 && page >= totalPages) || (totalPages == 0 && len(transactions) < m7ppTransactionSyncPageSize) {
			break
		}
	}
	for _, item := range affected {
		if err := mergeM7PPDailySummary(ctx, item.userID, item.loginID, item.scope, item.previous); err != nil {
			return persistedCount, fmt.Errorf("merge M7PP summary user=%d date=%s: %w", item.userID, item.scope.PeriodKey, err)
		}
	}
	log.Printf("[GameSummarySync][M7PP] from=%s to=%s persisted=%d", from.Format(time.RFC3339), to.Format(time.RFC3339), persistedCount)
	return persistedCount, nil
}

func SyncM7PPTransactionDate(ctx context.Context, statDate string) (int, error) {
	scope, err := buildDailyScope(statDate, time.Now().UTC())
	if err != nil {
		return 0, err
	}
	return SyncM7PPTransactions(ctx, scope.FromDate, scope.ToDate)
}

func SyncM7PPTransactionsNow(c *fiber.Ctx) error {
	if !QueryUserIsAdmin(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"code": fiber.StatusForbidden, "message": "administrator access required"})
	}
	var req struct {
		StatDate string `json:"stat_date"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": fiber.StatusBadRequest, "message": "invalid params"})
	}
	statDate, err := normalizeStatDate(req.StatDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": fiber.StatusBadRequest, "message": "invalid stat_date"})
	}
	count, err := SyncM7PPTransactionDate(c.Context(), statDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": fiber.StatusInternalServerError, "message": "M7PP transaction synchronization failed", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "M7PP transactions synchronized", "data": fiber.Map{"stat_date": statDate, "persisted": count}})
}

func aggregateM7PPDailySummary(userID uint64, loginID string, scope playerTransactionScope) (*provider.PlayerSummaryResult, error) {
	var providers []dtos.GameProvider
	if err := models.GetInstance().DbInstance.Select("id").
		Where("platform_code = ?", m7ppGamePlatformCode).Find(&providers).Error; err != nil {
		return nil, err
	}
	providerIDs := make(map[string]int, len(providers))
	for _, item := range providers {
		providerIDs[fmt.Sprint(item.ID)] = int(item.ID)
	}
	return aggregateM7PlayerDailySummaryFromStoredDetails(userID, loginID, scope, providerIDs)
}

func mergeM7PPDailySummary(ctx context.Context, userID uint64, loginID string, scope playerTransactionScope, previous *provider.PlayerSummaryResult) error {
	if _, err := SyncPlayerDailyGameSummaryWithCurrencyOptions(ctx, userID, scope.PeriodKey, "", true); err == nil {
		return nil
	} else {
		log.Printf("[GameSummarySync][M7PP] full daily rebuild failed; applying detail delta user=%d date=%s: %v", userID, scope.PeriodKey, err)
	}

	current, err := aggregateM7PPDailySummary(userID, loginID, scope)
	if err != nil {
		return err
	}
	stored, err := getStoredPlayerTransactionStat(userID, playerTransactionPeriodDaily, scope.PeriodKey)
	if err != nil {
		return err
	}
	combined := zeroPlayerSummary(loginID)
	if stored != nil {
		combined = &provider.PlayerSummaryResult{
			LoginId: loginID, Count: stored.Count, Turnover: stored.Turnover, Bet: stored.Bet,
			Win: stored.Win, Winlose: stored.Winlose, JPShare: stored.JPShare, JPWin: stored.JPWin,
		}
		combined = subtractPlayerSummary(combined, previous)
	}
	addPlayerSummary(combined, current)
	syncedAt := time.Now().UTC()
	if err := upsertPlayerTransactionStat(userID, loginID, scope, combined, syncedAt); err != nil {
		return err
	}
	if err := mirrorSummaryToDailyUserStats(userID, scope.PeriodKey, combined); err != nil {
		return err
	}
	return rebuildPlayerCurrentAggregateScopes(ctx, userID, loginID, time.Now().UTC(), syncedAt)
}

func buildM7PPTransactionRecordKey(transaction provider.M7PPTransaction) string {
	if externalID := strings.TrimSpace(transaction.ExternalTransactionID); externalID != "" {
		return "m7pp|tx|" + externalID
	}
	return fmt.Sprintf(
		"m7pp|fallback|%s|%s|%s|%d|%.8f|%.8f|%.8f|%.8f|%.8f",
		strings.TrimSpace(transaction.Username), strings.TrimSpace(transaction.BetID), strings.TrimSpace(transaction.RoundID),
		transaction.UpdateTime, transaction.BetAmount, transaction.WinAmount, transaction.WinLoss,
		transaction.EffectiveTurnover, transaction.JackpotAmount,
	)
}

func resolveM7PPTransactionProviderID(gameCode string, categoryCode string) (int, error) {
	var game dtos.Game
	err := models.GetInstance().DbInstance.Select("provider_id").
		Where("platform_code = ? AND game_code = ?", m7ppGamePlatformCode, gameCode).
		Order("status DESC, id DESC").First(&game).Error
	if err == nil {
		return int(game.ProviderID), nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}

	providerCode := strings.ToUpper(m7ppProviderConfig().VendorCode)
	categoryCode = strings.ToUpper(strings.TrimSpace(categoryCode))
	if strings.Contains(categoryCode, "CASINO") || strings.Contains(categoryCode, "LIVE") {
		providerCode += "_LIVE"
	}
	var gameProvider dtos.GameProvider
	if err := models.GetInstance().DbInstance.
		Where("platform_code = ? AND code = ?", m7ppGamePlatformCode, providerCode).
		Order("status DESC, id ASC").First(&gameProvider).Error; err == nil {
		return int(gameProvider.ID), nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}

	typeID := uint(2)
	name := "PP"
	if strings.HasSuffix(providerCode, "_LIVE") {
		typeID = gameTypeIDByProviderCode("CASINO", typeID)
		name = "PP真人"
	}
	apiConfig, _ := json.Marshal(map[string]string{
		"wallet_mode": "transfer",
		"api_url":     strings.TrimRight(m7ppProviderConfig().BaseURL, "/"),
		"vendor_code": strings.ToUpper(m7ppProviderConfig().VendorCode),
	})
	gameProvider = dtos.GameProvider{
		PlatformCode: m7ppGamePlatformCode, Code: providerCode, Name: name,
		Status: 1, TypeId: int(typeID), ApiConfig: datatypes.JSON(apiConfig),
	}
	if err := models.GetInstance().DbInstance.Create(&gameProvider).Error; err != nil {
		// Another synchronization goroutine may have created the provider first.
		if lookupErr := models.GetInstance().DbInstance.
			Where("platform_code = ? AND code = ?", m7ppGamePlatformCode, providerCode).
			Order("status DESC, id ASC").First(&gameProvider).Error; lookupErr != nil {
			return 0, err
		}
	}
	log.Printf("[GameSummarySync][M7PP] game=%s is not synced; preserving transaction under provider=%s", gameCode, providerCode)
	return int(gameProvider.ID), nil
}

func m7ppTransactionTime(values ...int64) time.Time {
	for _, value := range values {
		if value > 0 {
			return time.UnixMilli(value)
		}
	}
	return time.Now().UTC()
}

func startOfPlayerTransactionDay(value time.Time) time.Time {
	local := value.In(playerTransactionLocation())
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, playerTransactionLocation())
}
