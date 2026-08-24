package controllers

import (
	"errors"
	"fmt"
	"gogogo/common"
	"gogogo/helpers"
	"gogogo/models"
	"gogogo/models/dtos"
	"gogogo/models/requests"
	"gogogo/models/responses"
	"gogogo/provider"
	"gogogo/services"
	"log"
	"math"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const gameListSyncInterval = 15 * time.Minute
const debugGameProviderID uint = 23
const defaultGamePlatformCode = "HEDOC"
const m7GamePlatformCode = "M7"
const m7ppGamePlatformCode = "M7PP"
const m7ProviderUnavailableMessage = "M7 provider is unavailable from the current server region"
const pgProviderOrderSQL = "CASE WHEN UPPER(game_providers.code) IN ('PG', 'PG_SLOT', 'PGSOFT') OR UPPER(REPLACE(game_providers.name, ' ', '')) IN ('PG', 'PGSOFT') THEN 0 ELSE 1 END ASC"
const pgProviderOptionsOrderSQL = "CASE WHEN UPPER(code) IN ('PG', 'PG_SLOT', 'PGSOFT') OR UPPER(REPLACE(name, ' ', '')) IN ('PG', 'PGSOFT') THEN 0 ELSE 1 END ASC"

func normalizeGamePlatformCode(raw string) string {
	code := strings.ToUpper(strings.TrimSpace(raw))
	if code == "" {
		return defaultGamePlatformCode
	}
	return code
}

func normalizeHedocGameImageURL(rawURL, apiBaseURL string) string {
	rawURL = strings.TrimSpace(strings.ReplaceAll(rawURL, `\`, "/"))
	if rawURL == "" {
		return ""
	}
	if strings.HasPrefix(rawURL, "//") {
		rawURL = "https:" + rawURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	if !parsed.IsAbs() {
		base, baseErr := url.Parse(strings.TrimSpace(apiBaseURL))
		if baseErr == nil && base.IsAbs() {
			parsed = base.ResolveReference(parsed)
		}
	}

	// Some HEDOC responses prefix an already-hosted CDN path with the API host,
	// for example https://api.example.com/cdn.example.com/games/a.png.
	pathParts := strings.Split(strings.TrimPrefix(parsed.Path, "/"), "/")
	if len(pathParts) > 1 && strings.Contains(pathParts[0], ".") {
		parsed.Host = pathParts[0]
		parsed.Path = "/" + strings.Join(pathParts[1:], "/")
	}
	if parsed.Scheme == "http" {
		parsed.Scheme = "https"
	}
	return parsed.String()
}

func isPGProvider(providerInfo *dtos.GameProvider) bool {
	if providerInfo == nil {
		return false
	}
	code := strings.ToUpper(strings.TrimSpace(providerInfo.Code))
	name := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(providerInfo.Name), " ", ""))
	return code == "PG" ||
		code == "PG_SLOT" ||
		code == "PGSOFT" ||
		name == "PG" ||
		name == "PGSOFT"
}

func buildLaunchResponseData(url string, platformCode string, game dtos.Game, extra fiber.Map) fiber.Map {
	data := fiber.Map{
		"url":                 url,
		"platform_code":       platformCode,
		"launch_content_type": "url",
	}
	for key, value := range extra {
		data[key] = value
	}

	if normalizeGamePlatformCode(game.PlatformCode) == m7GamePlatformCode && isPGProvider(game.Provider) {
		if htmlSource, ok := provider.NormalizeLaunchHTML(url); ok {
			data["url"] = ""
			data["html"] = htmlSource
			data["launch_content_type"] = "html"
		}
	}

	return data
}

var gameListSyncState = struct {
	mu          sync.Mutex
	running     bool
	done        chan struct{}
	lastSuccess time.Time
}{}

// SyncAllGames 执行同步任务
func SyncHedocGames() error {
	log.Println("[Sync] Starting game synchronization...")

	// 1. 获取所有状态开启的厂商
	var providers []dtos.GameProvider
	if err := models.GetInstance().DbInstance.
		Where("platform_code = ? AND status = ?", defaultGamePlatformCode, 1).
		Find(&providers).Error; err != nil {
		log.Printf("[Sync] Failed to fetch providers: %v\n", err)
		return err
	}

	var debugProvider dtos.GameProvider
	debugProviderFound := models.GetInstance().DbInstance.Select("id, platform_code, code, name, status").First(&debugProvider, debugGameProviderID).Error == nil
	if debugProviderFound {
		log.Printf("[Sync][Provider %d] DB record found: platform=%s code=%s name=%s status=%d\n", debugProvider.ID, debugProvider.PlatformCode, debugProvider.Code, debugProvider.Name, debugProvider.Status)
	} else {
		log.Printf("[Sync][Provider %d] DB record not found\n", debugGameProviderID)
	}

	debugProviderIncluded := false
	for _, p := range providers {
		if p.ID == debugGameProviderID {
			debugProviderIncluded = true
			break
		}
	}
	if debugProviderIncluded {
		log.Printf("[Sync][Provider %d] Included in enabled providers list, will fetch games this round\n", debugGameProviderID)
	} else {
		log.Printf("[Sync][Provider %d] Not included in enabled providers list, current sync only fetches providers with status=1\n", debugGameProviderID)
	}

	// 2. 预加载所有游戏分类
	var gameTypes []dtos.GameType
	// 查询所有分类 (不管状态如何先查出来，后面在内存里判断)
	models.GetInstance().DbInstance.Find(&gameTypes)

	// Set: ID(uint) -> bool (是否存在)
	validTypeIds := make(map[uint]bool)
	// 兜底 ID (如果 API 返回的 Type 在本地找不到，归类到 Other)
	defaultTypeId := uint(0)

	for _, t := range gameTypes {
		// 只有 Status = 1 (启用) 的分类才被视为有效目标
		if t.Status == 1 {
			validTypeIds[t.ID] = true
			// 寻找默认分类 (例如 Code 为 OTHER 或 ID 为 1)
			if t.Code == "OTHER" || defaultTypeId == 0 {
				defaultTypeId = t.ID
			}
		}
	}

	client := provider.NewHedocClient()

	// 3. 遍历厂商进行同步
	for _, p := range providers {
		log.Printf("[Sync] Fetching games for provider: %s (ID: %d)\n", p.Name, p.ID)

		// 调用 API (传入 ID)
		externalGames, err := client.FetchGameList(int(p.ID), helpers.GetCfgInstance().Conf.Agentid, helpers.GetCfgInstance().Conf.Agentapi)
		if err != nil {
			if p.ID == debugGameProviderID {
				log.Printf("[Sync][Provider %d] FetchGameList failed: %v\n", p.ID, err)
			}
			log.Printf("[Sync] Error fetching provider %d: %v\n", p.ID, err)
			continue
		}
		if p.ID == debugGameProviderID {
			log.Printf("[Sync][Provider %d] FetchGameList succeeded, received %d games\n", p.ID, len(externalGames))
		}

		if len(externalGames) == 0 {
			if p.ID == debugGameProviderID {
				log.Printf("[Sync][Provider %d] Third-party returned 0 games\n", p.ID)
			}
			continue
		}

		var gamesToUpsert []dtos.Game

		// 4. 数据转换与整理
		for _, item := range externalGames {
			// 4.1 映射分类
			targetTypeId := uint(item.Type)
			// 校验 ID 是否有效，无效则使用兜底 ID
			if !validTypeIds[targetTypeId] {
				// 如果不想用兜底，而是直接跳过该游戏，可以使用 continue
				targetTypeId = defaultTypeId
			}

			// 4.2 处理多语言 Name (Map -> JSON)
			nameJsonBytes, _ := json.Marshal(item.Name)

			// 4.3 构造本地模型
			game := dtos.Game{
				PlatformCode: defaultGamePlatformCode,
				ProviderID:   p.ID,         // 使用本地循环的 ID
				GameTypeID:   targetTypeId, // 映射后的分类 ID
				GameCode:     item.Code,    // 唯一标识
				Name:         datatypes.JSON(nameJsonBytes),
				ImgUrl:       normalizeHedocGameImageURL(item.GameIcon, helpers.GetCfgInstance().Conf.Agentapi),
				Sort:         item.Order, // 使用 API 返回的排序
				Status:       1,          // 默认上架
			}
			gamesToUpsert = append(gamesToUpsert, game)
		}

		// 5. 批量 Upsert (存在则更新，不存在则插入)
		// 需要 game_code 上有唯一索引
		if len(gamesToUpsert) > 0 {
			err := models.GetInstance().DbInstance.Clauses(clause.OnConflict{
				// 修改关键点: 冲突检测基于 [provider_id, game_code] 组合
				Columns: []clause.Column{
					{Name: "platform_code"},
					{Name: "provider_id"},
					{Name: "game_code"},
				},
				// 如果冲突（即该厂商下已存在该 Code），则更新以下字段
				DoUpdates: clause.AssignmentColumns([]string{
					"name",
					"img_url",
					"sort",
					"game_type_id",
					"updated_at",
					// 注意: 不更新 views, status (防止覆盖运营手动下架或热度数据)
				}),
			}).CreateInBatches(&gamesToUpsert, 100).Error

			if err != nil {
				if p.ID == debugGameProviderID {
					log.Printf("[Sync][Provider %d] DB upsert failed: %v\n", p.ID, err)
				}
				log.Printf("[Sync] DB Error for provider %d: %v\n", p.ID, err)
			} else {
				if p.ID == debugGameProviderID {
					log.Printf("[Sync][Provider %d] DB upsert succeeded, synced %d games\n", p.ID, len(gamesToUpsert))
				}
				log.Printf("[Sync] Provider %d: Synced %d games.\n", p.ID, len(gamesToUpsert))
			}
		}
	}
	log.Println("[Sync] Synchronization finished.")
	return nil
}

func m7ProviderConfig() provider.M7Config {
	cfg := helpers.GetCfgInstance().Conf.GamePlatforms.M7
	return provider.M7Config{
		BaseURL:   cfg.BaseURL,
		ClientID:  cfg.ClientID,
		ClientKey: cfg.ClientKey,
		AgentID:   cfg.AgentID,
		Currency:  cfg.Currency,
		ReturnURL: cfg.ReturnURL,
	}
}

func m7ppProviderConfig() provider.M7PPConfig {
	cfg := helpers.GetCfgInstance().Conf.GamePlatforms.M7PP
	return provider.M7PPConfig{
		BaseURL: cfg.BaseURL, APIKey: cfg.APIKey, APISecret: cfg.APISecret,
		VendorCode: cfg.VendorCode, Currency: cfg.Currency, ReturnURL: cfg.ReturnURL,
	}
}

func normalizeM7ImageURL(baseURL string, urls ...string) string {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.ddhh.work"
	}

	for _, raw := range urls {
		imageURL := strings.TrimSpace(raw)
		if imageURL == "" {
			continue
		}

		// Some M7 game-list responses wrap the real CDN host as a path below
		// the API host, for example:
		//   https://api.ddhh.work/cdn.example.com/games/cover.png
		// Browsers request that invalid API path instead of the CDN. Strip the
		// API prefix when the first path segment is itself a hostname.
		basePrefix := baseURL + "/"
		if strings.HasPrefix(strings.ToLower(imageURL), strings.ToLower(basePrefix)) {
			remainder := strings.TrimPrefix(imageURL, basePrefix)
			firstSegment := strings.SplitN(remainder, "/", 2)[0]
			if strings.Contains(firstSegment, ".") {
				return "https://" + remainder
			}
		}

		lowerURL := strings.ToLower(imageURL)
		switch {
		case strings.HasPrefix(lowerURL, "https://"):
			return imageURL
		case strings.HasPrefix(lowerURL, "http://"):
			return "https://" + strings.TrimPrefix(imageURL, "http://")
		case strings.HasPrefix(imageURL, "//"):
			return "https:" + imageURL
		case strings.HasPrefix(imageURL, "/"):
			return baseURL + imageURL
		case strings.Contains(strings.SplitN(imageURL, "/", 2)[0], "."):
			return "https://" + imageURL
		default:
			return baseURL + "/" + imageURL
		}
	}

	return ""
}

func gameTypeIDByProviderCode(code string, fallback uint) uint {
	normalized := strings.ToUpper(strings.TrimSpace(code))
	wantedCode := "SLOT"
	switch {
	case strings.Contains(normalized, "LIVE") || strings.Contains(normalized, "CASINO"):
		wantedCode = "CASINO"
	case strings.Contains(normalized, "SPORT"):
		wantedCode = "SPORTBOOK"
	case strings.Contains(normalized, "FISH"):
		wantedCode = "FISHING"
	}

	var gameType dtos.GameType
	if err := models.GetInstance().DbInstance.
		Where("code = ? AND status = ?", wantedCode, 1).
		First(&gameType).Error; err == nil {
		return gameType.ID
	}

	return fallback
}

func SyncM7Games() error {
	log.Println("[Sync][M7] Starting game synchronization...")

	m7Config := m7ProviderConfig()
	client, err := provider.NewM7Client(m7Config)
	if err != nil {
		log.Printf("[Sync][M7] skipped: %v\n", err)
		return nil
	}

	var providers []dtos.GameProvider
	if err := models.GetInstance().DbInstance.
		Where("platform_code = ? AND status = ?", m7GamePlatformCode, 1).
		Find(&providers).Error; err != nil {
		log.Printf("[Sync][M7] Failed to fetch providers: %v\n", err)
		return err
	}
	if len(providers) == 0 {
		log.Println("[Sync][M7] No enabled providers found.")
		return nil
	}

	defaultTypeID := uint(0)
	var defaultType dtos.GameType
	if err := models.GetInstance().DbInstance.
		Where("status = ?", 1).
		Order("CASE WHEN code = 'OTHER' THEN 0 ELSE 1 END, id ASC").
		First(&defaultType).Error; err == nil {
		defaultTypeID = defaultType.ID
	}

	externalGames, err := client.FetchGameList("")
	if err != nil {
		log.Printf("[Sync][M7] Error fetching full game list: %v\n", err)
		return err
	}
	if len(externalGames) == 0 {
		log.Println("[Sync][M7] Full game list returned 0 games.")
		return nil
	}
	log.Printf("[Sync][M7] Full game list received %d games.\n", len(externalGames))

	gamesByProviderCode := make(map[string][]provider.M7GameItem)
	for _, item := range externalGames {
		providerCode := strings.ToUpper(strings.TrimSpace(item.ThirdPartyType))
		if providerCode == "" {
			continue
		}
		gamesByProviderCode[providerCode] = append(gamesByProviderCode[providerCode], item)
	}

	for _, p := range providers {
		providerCode := strings.ToUpper(strings.TrimSpace(p.Code))
		providerGames := gamesByProviderCode[providerCode]
		if len(providerGames) == 0 {
			log.Printf("[Sync][M7] Provider %s returned 0 games in full list\n", p.Code)
			continue
		}

		targetTypeID := gameTypeIDByProviderCode(p.Code, defaultTypeID)
		gamesToUpsert := make([]dtos.Game, 0, len(providerGames))
		for index, item := range providerGames {
			if item.ID <= 0 {
				continue
			}
			imageURL := normalizeM7ImageURL(m7Config.BaseURL, item.Image2, item.Image1)
			namePayload := map[string]string{
				"EN": item.GameNameEn,
				"CN": item.GameNameCh,
				"ZH": item.GameNameTw,
			}
			if strings.TrimSpace(namePayload["EN"]) == "" {
				namePayload["EN"] = item.GameNo
			}
			nameJSONBytes, _ := json.Marshal(namePayload)

			status := 0
			if strings.EqualFold(strings.TrimSpace(item.Status), "OPEN") {
				status = 1
			}

			gamesToUpsert = append(gamesToUpsert, dtos.Game{
				PlatformCode: m7GamePlatformCode,
				ProviderID:   p.ID,
				GameTypeID:   targetTypeID,
				GameCode:     strconv.FormatInt(item.ID, 10),
				Name:         datatypes.JSON(nameJSONBytes),
				ImgUrl:       imageURL,
				Sort:         len(providerGames) - index,
				Status:       status,
			})
		}

		if len(gamesToUpsert) == 0 {
			continue
		}

		if err := models.GetInstance().DbInstance.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&dtos.Game{}).
				Where("platform_code = ? AND provider_id = ?", m7GamePlatformCode, p.ID).
				Updates(map[string]any{
					"status":     0,
					"updated_at": time.Now(),
				}).Error; err != nil {
				return err
			}

			return tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "platform_code"},
					{Name: "provider_id"},
					{Name: "game_code"},
				},
				DoUpdates: clause.AssignmentColumns([]string{
					"name",
					"img_url",
					"sort",
					"status",
					"game_type_id",
					"updated_at",
				}),
			}).CreateInBatches(&gamesToUpsert, 100).Error
		}); err != nil {
			log.Printf("[Sync][M7] DB Error for provider %s: %v\n", p.Code, err)
			continue
		}

		log.Printf("[Sync][M7] Provider %s: Synced %d games.\n", p.Code, len(gamesToUpsert))
	}

	log.Println("[Sync][M7] Synchronization finished.")
	return nil
}

func SyncM7PPGames() error {
	log.Println("[Sync][M7PP] Starting game synchronization...")
	cfg := m7ppProviderConfig()
	client, err := provider.NewM7PPClient(cfg)
	if err != nil {
		log.Printf("[Sync][M7PP] skipped: %v\n", err)
		return nil
	}

	var gameProvider dtos.GameProvider
	if err := models.GetInstance().DbInstance.
		Where("platform_code = ? AND code = ?", m7ppGamePlatformCode, strings.ToUpper(cfg.VendorCode)).
		Order("id ASC").
		First(&gameProvider).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("load M7PP provider %s: %w", cfg.VendorCode, err)
		}
		apiConfig, _ := json.Marshal(map[string]string{
			"wallet_mode": "transfer",
			"api_url":     strings.TrimRight(cfg.BaseURL, "/"),
		})
		gameProvider = dtos.GameProvider{
			PlatformCode: m7ppGamePlatformCode,
			Code:         strings.ToUpper(cfg.VendorCode),
			Name:         "PP",
			Status:       1,
			TypeId:       2,
			ApiConfig:    datatypes.JSON(apiConfig),
		}
		if err := models.GetInstance().DbInstance.Create(&gameProvider).Error; err != nil {
			return fmt.Errorf("create M7PP provider %s: %w", cfg.VendorCode, err)
		}
		log.Printf("[Sync][M7PP] created missing provider %s\n", cfg.VendorCode)
	}
	// Preserve an administrator's manual enable/disable setting. Synchronizing
	// games must never silently reopen a provider on service restart.

	liveProviderCode := strings.ToUpper(cfg.VendorCode) + "_LIVE"
	casinoTypeID := gameTypeIDByProviderCode("CASINO", uint(gameProvider.TypeId))
	var liveGameProvider dtos.GameProvider
	if err := models.GetInstance().DbInstance.
		Where("platform_code = ? AND code = ?", m7ppGamePlatformCode, liveProviderCode).
		Order("id ASC").
		First(&liveGameProvider).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("load M7PP live provider %s: %w", liveProviderCode, err)
		}
		apiConfig, _ := json.Marshal(map[string]string{
			"wallet_mode": "transfer",
			"api_url":     strings.TrimRight(cfg.BaseURL, "/"),
			"vendor_code": strings.ToUpper(cfg.VendorCode),
		})
		liveGameProvider = dtos.GameProvider{
			PlatformCode: m7ppGamePlatformCode,
			Code:         liveProviderCode,
			Name:         "PP真人",
			Status:       1,
			TypeId:       int(casinoTypeID),
			ApiConfig:    datatypes.JSON(apiConfig),
		}
		if err := models.GetInstance().DbInstance.Create(&liveGameProvider).Error; err != nil {
			return fmt.Errorf("create M7PP live provider %s: %w", liveProviderCode, err)
		}
	} else if liveGameProvider.TypeId != int(casinoTypeID) || liveGameProvider.Name != "PP真人" {
		if err := models.GetInstance().DbInstance.Model(&liveGameProvider).Updates(map[string]any{
			"type_id": casinoTypeID, "name": "PP真人",
		}).Error; err != nil {
			return fmt.Errorf("enable M7PP live provider %s: %w", liveProviderCode, err)
		}
	}

	externalGames, err := client.FetchGames()
	if err != nil {
		return fmt.Errorf("sync M7PP games: %w", err)
	}
	gamesToUpsert := make([]dtos.Game, 0, len(externalGames))
	for index, item := range externalGames {
		if item.GameCode == "" {
			continue
		}
		name := item.GameName
		if name == "" {
			name = item.GameCode
		}
		nameJSON, _ := json.Marshal(map[string]string{"EN": name, "CN": name, "ZH": name})
		imageURL := item.ImageSquare
		if imageURL == "" {
			imageURL = item.ImageLandscape
		}
		targetProvider := gameProvider
		categoryCode := strings.ToUpper(strings.TrimSpace(item.CategoryCode))
		if strings.Contains(categoryCode, "CASINO") || strings.Contains(categoryCode, "LIVE") {
			targetProvider = liveGameProvider
		}
		gamesToUpsert = append(gamesToUpsert, dtos.Game{
			PlatformCode: m7ppGamePlatformCode,
			ProviderID:   targetProvider.ID,
			GameTypeID:   gameTypeIDByProviderCode(item.CategoryCode, uint(targetProvider.TypeId)),
			GameCode:     item.GameCode, Name: datatypes.JSON(nameJSON), ImgUrl: imageURL,
			Sort: len(externalGames) - index, Status: 1,
		})
	}
	if len(gamesToUpsert) == 0 {
		return fmt.Errorf("M7PP game list returned no usable games")
	}

	err = models.GetInstance().DbInstance.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&dtos.Game{}).
			Where("platform_code = ?", m7ppGamePlatformCode).
			Update("status", 0).Error; err != nil {
			return err
		}
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "platform_code"}, {Name: "provider_id"}, {Name: "game_code"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "img_url", "sort", "status", "game_type_id", "updated_at"}),
		}).CreateInBatches(&gamesToUpsert, 100).Error
	})
	if err != nil {
		return err
	}
	log.Printf("[Sync][M7PP] Synchronized %d games.\n", len(gamesToUpsert))
	return nil
}

// SyncM7PPGamesNow forces a BOAN game refresh and reports the persisted row
// count. It is JWT-protected by the route group and intended for deployment
// verification when the scheduled sync has not run yet.
func SyncM7PPGamesNow(c *fiber.Ctx) error {
	if !QueryUserIsAdmin(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"code":    fiber.StatusForbidden,
			"message": "administrator access required",
		})
	}
	if err := SyncM7PPGames(); err != nil {
		log.Printf("[Sync][M7PP] manual synchronization failed: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    fiber.StatusInternalServerError,
			"message": "M7PP game synchronization failed",
			"error":   err.Error(),
		})
	}

	var providerCount int64
	var gameCount int64
	db := models.GetInstance().DbInstance
	if err := db.Model(&dtos.GameProvider{}).
		Where("platform_code = ? AND code = ?", m7ppGamePlatformCode, strings.ToUpper(m7ppProviderConfig().VendorCode)).
		Count(&providerCount).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "failed to count M7PP providers", "error": err.Error()})
	}
	if err := db.Model(&dtos.Game{}).
		Where("platform_code = ?", m7ppGamePlatformCode).
		Count(&gameCount).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "failed to count M7PP games", "error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "M7PP games synchronized",
		"data": fiber.Map{
			"platform_code":  m7ppGamePlatformCode,
			"vendor_code":    strings.ToUpper(m7ppProviderConfig().VendorCode),
			"provider_count": providerCount,
			"game_count":     gameCount,
		},
	})
}

func SyncAllGames() error {
	var firstErr error
	if err := SyncHedocGames(); err != nil {
		firstErr = err
	}
	if err := SyncM7Games(); err != nil && firstErr == nil {
		firstErr = err
	}
	if err := SyncM7PPGames(); err != nil && firstErr == nil {
		firstErr = err
	}
	return firstErr
}

func SyncAllGamesIfStale(maxAge time.Duration) error {
	if maxAge <= 0 {
		maxAge = gameListSyncInterval
	}

	for {
		gameListSyncState.mu.Lock()
		if !gameListSyncState.lastSuccess.IsZero() && time.Since(gameListSyncState.lastSuccess) < maxAge {
			gameListSyncState.mu.Unlock()
			return nil
		}

		if gameListSyncState.running {
			done := gameListSyncState.done
			gameListSyncState.mu.Unlock()
			<-done
			continue
		}

		done := make(chan struct{})
		gameListSyncState.running = true
		gameListSyncState.done = done
		gameListSyncState.mu.Unlock()

		err := SyncAllGames()

		gameListSyncState.mu.Lock()
		if err == nil {
			gameListSyncState.lastSuccess = time.Now()
		}
		gameListSyncState.running = false
		gameListSyncState.done = nil
		close(done)
		gameListSyncState.mu.Unlock()

		return err
	}
}

// 同步三方游戏
func SyncGames(c *fiber.Ctx) error {
	client := provider.NewHedocClient()
	games, err := client.GetGameList()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// 批量插入或更新到本地 games 表
	for _, g := range games {
		var gameTypeID uint = 1 // 需根据 g.Type 映射本地分类ID

		models.GetInstance().DbInstance.Where(dtos.Game{PlatformCode: defaultGamePlatformCode, GameCode: g.GameCode}).
			Assign(dtos.Game{
				// Name:       g.GameName,
				PlatformCode: defaultGamePlatformCode,
				ImgUrl:       g.ImgUrl,
				ProviderID:   1, // 假设 Hedoc ID=1
				GameTypeID:   gameTypeID,
			}).
			FirstOrCreate(&dtos.Game{})
	}

	return c.JSON(fiber.Map{"success": true, "count": len(games)})
}

// 获取三方的游戏
func ListGames(c *fiber.Ctx) error {
	client := provider.NewHedocClient()
	games, err := client.GetGameList()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "count": len(games), "games": games})
}

// 请求进入游戏
func findLaunchGame(req requests.OpenGamePayload) (dtos.Game, error) {
	platformCode := normalizeGamePlatformCode(req.PlatformCode)
	var game dtos.Game
	query := models.GetInstance().DbInstance.
		Model(&dtos.Game{}).
		Joins("JOIN game_providers ON game_providers.id = games.provider_id AND game_providers.platform_code = games.platform_code").
		Where("games.status = ? AND game_providers.status = ?", 1, 1).
		Preload("Provider")

	if req.GameID > 0 {
		query = query.Where("games.id = ?", req.GameID)
	} else {
		query = query.Where("games.platform_code = ? AND games.game_code = ?", platformCode, req.GameCode)
	}

	err := query.First(&game).Error
	return game, err
}

func getLocalBalance(userID int) float64 {
	var user dtos.User
	if err := models.GetInstance().DbInstance.Select("balance").First(&user, userID).Error; err != nil {
		log.Printf("[LaunchGame] failed to load local balance for user %d: %v\n", userID, err)
		return 0
	}
	return user.Balance
}

func launchM7Game(c *fiber.Ctx, userID int, thirdPartyLoginID string, game dtos.Game, req requests.OpenGamePayload, desktopReward *dtos.ActivityClaimRecordResponse) error {
	localBalance := 0.0
	gameBalance := 0.0
	entryBillNo := ""
	entryAmountU := 0.0
	client, err := provider.NewM7Client(m7ProviderConfig())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	if err := client.CreatePlayer(thirdPartyLoginID); err != nil {
		if provider.IsM7IPNotAllowedError(err) {
			log.Printf("[LaunchGame][M7] provider region/IP blocked for user %d login %s: %v\n", userID, thirdPartyLoginID, err)
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"code":    fiber.StatusServiceUnavailable,
				"message": m7ProviderUnavailableMessage,
			})
		}
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}

	err = models.GetInstance().DbInstance.Transaction(func(tx *gorm.DB) error {
		var user dtos.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userID).Error; err != nil {
			return err
		}
		if user.Balance > 0 {
			usdAmount := user.Balance
			billNo := fmt.Sprintf("M7-AUTO-IN-%d-%d", time.Now().UnixNano(), userID)
			user.Balance = 0
			trans := dtos.Transaction{
				UserID:        uint64(userID),
				Type:          3,
				Amount:        -usdAmount,
				BeforeBalance: usdAmount,
				AfterBalance:  0,
				ReferenceID:   billNo,
				Remark:        "M7 auto transfer in (USD)",
			}
			if err := tx.Save(&user).Error; err != nil {
				return err
			}
			if err := tx.Create(&trans).Error; err != nil {
				return err
			}
			updatedBalance, err := client.Transfer(thirdPartyLoginID, usdAmount, billNo)
			gameBalance = updatedBalance
			if err != nil {
				return fmt.Errorf("transfer to M7 failed: %v", err)
			}
			entryBillNo = billNo
			entryAmountU = usdAmount
		}
		localBalance = user.Balance
		return nil
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	recordAddDesktopGameEntryIfNeeded(c, uint64(userID), m7GamePlatformCode, entryBillNo, entryAmountU)

	url, err := client.EntryGame(thirdPartyLoginID, game.GameCode, req.IsMobile, req.Language)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}

	models.GetInstance().DbInstance.Model(&game).UpdateColumn("views", gorm.Expr("views + ?", 1))
	if latestBalance, balanceErr := client.Balance(thirdPartyLoginID); balanceErr != nil {
		log.Printf("[LaunchGame][M7] failed to refresh third-party balance for user %d: %v\n", userID, balanceErr)
	} else {
		gameBalance = latestBalance
	}

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data": buildLaunchResponseData(url, m7GamePlatformCode, game, fiber.Map{
			"local_balance":          localBalance,
			"game_balance":           gameBalance,
			"desktop_reward_claim":   desktopReward,
			"desktop_reward_checked": desktopReward != nil,
		}),
	})
}

func launchM7PPGame(c *fiber.Ctx, userID int, username string, game dtos.Game, req requests.OpenGamePayload, desktopReward *dtos.ActivityClaimRecordResponse) error {
	client, err := provider.NewM7PPClient(m7ppProviderConfig())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	localBalance := 0.0
	gameBalance := 0.0
	entryBillNo := ""
	entryAmount := 0.0
	providerDepositBillNo := ""
	providerDepositAmount := 0.0
	platform := "web"
	if req.IsMobile {
		platform = "H5"
	}
	// BOAN may initialize the transfer wallet as part of /game/url. Create the
	// session before the first deposit so a new player is not rejected by
	// /cash/deposit with SC_INTERNAL_ERROR.
	url, err := client.LaunchGame(username, game.GameCode, req.Language, platform, QueryClientIP(c))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	err = models.GetInstance().DbInstance.Transaction(func(tx *gorm.DB) error {
		var user dtos.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userID).Error; err != nil {
			return err
		}
		if user.Balance > 0 {
			amount := math.Floor(user.Balance*100) / 100
			billNo := fmt.Sprintf("M7PP-AUTO-IN-%d-%d", time.Now().UnixNano(), userID)
			updatedBalance, err := client.Deposit(username, amount, billNo)
			if err != nil {
				return fmt.Errorf("transfer to M7PP failed: %w", err)
			}
			providerDepositBillNo = billNo
			providerDepositAmount = amount
			trans := dtos.Transaction{UserID: uint64(userID), Type: 3, Amount: -amount,
				BeforeBalance: user.Balance, AfterBalance: user.Balance - amount,
				ReferenceID: billNo, Remark: "M7PP auto transfer in (USD)"}
			user.Balance -= amount
			if err := tx.Create(&trans).Error; err != nil {
				return err
			}
			if err := tx.Save(&user).Error; err != nil {
				return err
			}
			gameBalance, entryBillNo, entryAmount = updatedBalance, billNo, amount
		}
		localBalance = user.Balance
		return nil
	})
	if err != nil {
		if providerDepositAmount > 0 {
			compensationBillNo := fmt.Sprintf("%s-COMP-%d", providerDepositBillNo, time.Now().UnixNano())
			if _, compensationErr := client.Withdraw(username, providerDepositAmount, compensationBillNo); compensationErr != nil {
				log.Printf("[LaunchGame][M7PP] CRITICAL deposit compensation failed user=%d login=%s deposit_bill=%s compensation_bill=%s amount=%.2f db_error=%v compensation_error=%v", userID, username, providerDepositBillNo, compensationBillNo, providerDepositAmount, err, compensationErr)
				return c.Status(500).JSON(fiber.Map{
					"code": 500, "message": "M7PP transfer reconciliation required",
					"reference_id": providerDepositBillNo,
				})
			}
			log.Printf("[LaunchGame][M7PP] compensated provider deposit after local transaction failure user=%d login=%s deposit_bill=%s compensation_bill=%s amount=%.2f error=%v", userID, username, providerDepositBillNo, compensationBillNo, providerDepositAmount, err)
		}
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	recordAddDesktopGameEntryIfNeeded(c, uint64(userID), m7ppGamePlatformCode, entryBillNo, entryAmount)
	models.GetInstance().DbInstance.Model(&game).UpdateColumn("views", gorm.Expr("views + ?", 1))
	if latest, balanceErr := client.Balance(username); balanceErr == nil {
		gameBalance = latest
	}
	return c.JSON(fiber.Map{"code": 0, "message": "success", "data": buildLaunchResponseData(url, m7ppGamePlatformCode, game, fiber.Map{
		"local_balance": localBalance, "game_balance": gameBalance,
		"desktop_reward_claim": desktopReward, "desktop_reward_checked": desktopReward != nil,
	})})
}

func LaunchGame(c *fiber.Ctx) error {
	var req requests.OpenGamePayload
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "Invalid params"})
	}
	if err := SyncAllGamesIfStale(gameListSyncInterval); err != nil {
		log.Printf("[LaunchGame] game sync failed before launch: %v\n", err)
	}
	userId := QueryUserIdFromJwt(c)
	thirdPartyLoginId := fmt.Sprintf("%su%d", helpers.GetCfgInstance().Conf.Prefix, userId)
	// 获得用户专属的第三方密码
	thirdPartyPassword := common.GenerateGamePassword(uint64(userId))
	// 1. 根据 GameCode 查找游戏信息 (获取 ProviderID)
	game, err := findLaunchGame(req)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "Game not found"})
	}
	if normalizeGamePlatformCode(game.PlatformCode) == m7GamePlatformCode {
		return launchM7Game(c, userId, thirdPartyLoginId, game, req, nil)
	}
	if normalizeGamePlatformCode(game.PlatformCode) == m7ppGamePlatformCode {
		return launchM7PPGame(c, userId, thirdPartyLoginId, game, req, nil)
	}

	client := provider.NewHedocClient()

	// 2. 确保用户已注册
	// _ = client.RegisterUser(thirdPartyLoginId, "Game@123456")

	// 3. 请求链接
	url, err := client.OpenGame(thirdPartyLoginId, thirdPartyPassword, int(game.Provider.ID), game.GameCode, req.IsMobile, req.Language,
		helpers.GetCfgInstance().Conf.Agentapi,
		helpers.GetCfgInstance().Conf.Agentid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}

	// 4. (可选) 增加热度
	models.GetInstance().DbInstance.Model(&game).UpdateColumn("views", gorm.Expr("views + ?", 1))

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data":    buildLaunchResponseData(url, normalizeGamePlatformCode(game.PlatformCode), game, nil),
	})
}

// 请求进入游戏 - 钱包自动带入
/*

智能免转 (Smart Auto-Transfer)
核心逻辑：
进入游戏时：后端检测用户在本地的余额。如果有钱，自动发起 TransferIn 请求，将本地余额全部转入该三方厂商，然后返回游戏链接。
退出/切换时：由于 Web 端很难捕获准确的“退出”事件，我们在前端首页/个人中心增加一个 “一键回收” (Recycle / Refresh) 按钮。
进阶逻辑（可选）：在“进入游戏”接口中，先检查该用户上次玩的是哪个厂商，先把那个厂商的钱取出来（归集），再转入当前要玩的厂商。

进游戏时：如果本地没钱，我们假设钱已经在三方了（或者用户就是没钱），直接放行进游戏。
如果用户在三方有钱，进游戏后能直接玩；如果两边都没钱，用户进游戏后余额为0，逻辑也是对的。

回收时：每次回收都是实时的去问三方：“你那里有这个人的钱吗？”。如果有，就全部取回来。这解决了“掉单”或“数据不一致”的问题。

用户操作简便：
用户充值到平台 -> 点击任意游戏 -> 自动转入。
用户玩完 -> 点击“一键回收” -> 自动转出到平台 -> 提现。

*/
func getThirdPartyBalanceSnapshot(client *provider.HedoClient, loginID string) (float64, error) {
	resp, err := client.GetBalance(helpers.GetCfgInstance().Conf.Agentid, loginID, helpers.GetCfgInstance().Conf.Agentapi)
	if err != nil {
		return 0, err
	}

	return resp.Balance, nil
}

func shouldGrantDesktopLaunchReward(launchSource string) bool {
	return strings.EqualFold(strings.TrimSpace(launchSource), "desktop_app")
}

func claimDesktopLaunchRewardIfNeeded(c *fiber.Ctx, userID uint64, launchSource string) *dtos.ActivityClaimRecordResponse {
	if !shouldGrantDesktopLaunchReward(launchSource) {
		return nil
	}

	resp, err := services.GetActivityService().ClaimAddDesktop(c.Context(), userID, QueryClientIP(c))
	if err != nil {
		log.Printf("[DesktopLaunchReward] failed to claim add desktop reward for user %d: %v\n", userID, err)
		return &dtos.ActivityClaimRecordResponse{
			Success:      false,
			ActivityType: "add_desktop",
			Message:      err.Error(),
		}
	}

	return resp
}

func recordAddDesktopGameEntryIfNeeded(c *fiber.Ctx, userID uint64, platformCode, billNo string, entryAmountU float64) {
	if err := services.GetActivityService().RecordAddDesktopGameEntry(
		c.Context(),
		userID,
		platformCode,
		billNo,
		entryAmountU,
		time.Now().UTC(),
	); err != nil {
		log.Printf("[AddDesktopInsurance] failed to record game entry user=%d platform=%s bill=%s amount=%.2f: %v\n",
			userID, platformCode, billNo, entryAmountU, err)
	}
}

func settleAddDesktopInsuranceIfNeeded(c *fiber.Ctx, userID uint64, platformCode string, exitAmountU float64) *dtos.AddDesktopInsuranceSettlementResponse {
	resp, err := services.GetActivityService().SettleAddDesktopInsurance(c.Context(), userID, platformCode, exitAmountU)
	if err != nil {
		log.Printf("[AddDesktopInsurance] failed to settle user=%d platform=%s exit=%.2f: %v\n", userID, platformCode, exitAmountU, err)
		return nil
	}
	return resp
}

func LaunchGame2(c *fiber.Ctx) error {
	localBalance := 0.0
	gameBalance := 0.0
	entryBillNo := ""
	entryAmountU := 0.0
	var req requests.OpenGamePayload
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "Invalid params"})
	}
	if err := SyncAllGamesIfStale(gameListSyncInterval); err != nil {
		log.Printf("[LaunchGame2] game sync failed before launch: %v\n", err)
	}
	userId := QueryUserIdFromJwt(c)
	thirdPartyLoginId := fmt.Sprintf("%su%d", helpers.GetCfgInstance().Conf.Prefix, userId)
	thirdPartyPassword := common.GenerateGamePassword(uint64(userId))
	game, err := findLaunchGame(req)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "Game not found"})
	}
	desktopReward := claimDesktopLaunchRewardIfNeeded(c, uint64(userId), req.LaunchSource)
	if normalizeGamePlatformCode(game.PlatformCode) == m7GamePlatformCode {
		return launchM7Game(c, userId, thirdPartyLoginId, game, req, desktopReward)
	}
	if normalizeGamePlatformCode(game.PlatformCode) == m7ppGamePlatformCode {
		return launchM7PPGame(c, userId, thirdPartyLoginId, game, req, desktopReward)
	}
	client := provider.NewHedocClient()
	_ = client.RegisterUser(helpers.GetCfgInstance().Conf.Agentid, thirdPartyLoginId, thirdPartyPassword, helpers.GetCfgInstance().Conf.Agentapi)
	err = models.GetInstance().DbInstance.Transaction(func(tx *gorm.DB) error {
		var user dtos.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userId).Error; err != nil {
			return err
		}
		if user.Balance > 0 {
			usdAmount := user.Balance
			targetAmount := usdAmount
			billNo := fmt.Sprintf("AUTO-IN-%d-%d", time.Now().UnixNano(), userId)
			user.Balance = 0
			trans := dtos.Transaction{
				UserID:        uint64(userId),
				Type:          3,
				Amount:        -usdAmount,
				BeforeBalance: usdAmount,
				AfterBalance:  0,
				ReferenceID:   billNo,
				Remark:        "auto transfer in (USD)",
			}
			if err := tx.Save(&user).Error; err != nil {
				return err
			}
			if err := tx.Create(&trans).Error; err != nil {
				return err
			}
			updatedBalance, err := client.UpdateBalance(helpers.GetCfgInstance().Conf.Agentid,
				thirdPartyLoginId, targetAmount, 1, billNo, helpers.GetCfgInstance().Conf.Agentapi)
			gameBalance = updatedBalance
			if err != nil {
				return fmt.Errorf("transfer to provider failed: %v", err)
			}
			entryBillNo = billNo
			entryAmountU = usdAmount
		}
		localBalance = user.Balance
		return nil
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	recordAddDesktopGameEntryIfNeeded(c, uint64(userId), defaultGamePlatformCode, entryBillNo, entryAmountU)
	url, err := client.OpenGame(thirdPartyLoginId, thirdPartyPassword, int(game.Provider.ID), game.GameCode, req.IsMobile, req.Language,
		helpers.GetCfgInstance().Conf.Agentapi,
		helpers.GetCfgInstance().Conf.Agentid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	models.GetInstance().DbInstance.Model(&game).UpdateColumn("views", gorm.Expr("views + ?", 1))
	if latestBalance, balanceErr := getThirdPartyBalanceSnapshot(client, thirdPartyLoginId); balanceErr != nil {
		log.Printf("[LaunchGame2] failed to refresh third-party balance for user %d: %v\n", userId, balanceErr)
	} else {
		gameBalance = latestBalance
	}
	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data": buildLaunchResponseData(url, defaultGamePlatformCode, game, fiber.Map{
			"local_balance":          localBalance,
			"game_balance":           gameBalance,
			"desktop_reward_claim":   desktopReward,
			"desktop_reward_checked": desktopReward != nil,
		}),
	})
}

// RecycleBalance 一键回收资金 (从三方统一钱包转回本地)
func RecycleBalance(c *fiber.Ctx) error {
	type RecycleBalanceRequest struct {
		PlatformCode     string `json:"platform_code"`
		ForceSummarySync bool   `json:"force_summary_sync"`
	}

	req := RecycleBalanceRequest{}
	_ = c.BodyParser(&req)
	platformCode := normalizeGamePlatformCode(req.PlatformCode)
	if platformCode != defaultGamePlatformCode && platformCode != m7GamePlatformCode && platformCode != m7ppGamePlatformCode {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "Unsupported game platform"})
	}

	userId := QueryUserIdFromJwt(c)
	thirdPartyLoginId := fmt.Sprintf("%su%d", helpers.GetCfgInstance().Conf.Prefix, userId)
	var recycledAmount float64 = 0
	var currentLocalBalance float64 = 0
	var gameBalance float64 = 0
	var gameSummarySyncError string
	var gameSummarySyncResult *playerTransactionSyncResult
	var gameSummarySyncPending bool
	var insuranceSettlement *dtos.AddDesktopInsuranceSettlementResponse
	var insuranceExitAmount float64 = 0
	m7BalanceNotEnoughErr := errors.New("m7 balance not enough")

	recycleToLocal := func(amount float64, billNo, remark string, transferOut func() error) error {
		if amount <= 0 {
			return nil
		}
		return models.GetInstance().DbInstance.Transaction(func(tx *gorm.DB) error {
			var user dtos.User
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userId).Error; err != nil {
				return err
			}
			if err := transferOut(); err != nil {
				return err
			}
			trans := dtos.Transaction{
				UserID:        uint64(userId),
				Type:          3,
				Amount:        amount,
				BeforeBalance: user.Balance,
				AfterBalance:  user.Balance + amount,
				ReferenceID:   billNo,
				Remark:        remark,
			}
			user.Balance += amount
			if err := tx.Create(&trans).Error; err != nil {
				return err
			}
			if err := tx.Save(&user).Error; err != nil {
				return err
			}
			recycledAmount += amount
			currentLocalBalance = user.Balance
			return nil
		})
	}
	returnM7ProviderUnavailable := func(source string, err error) error {
		log.Printf("[RecycleBalance][M7] provider region/IP blocked during %s for user %d login %s: %v\n", source, userId, thirdPartyLoginId, err)
		var user dtos.User
		if dbErr := models.GetInstance().DbInstance.Select("balance").First(&user, userId).Error; dbErr != nil {
			return c.Status(500).JSON(fiber.Map{"code": 500, "message": "failed to load user info"})
		}
		currentLocalBalance = user.Balance
		return c.JSON(fiber.Map{
			"code":    0,
			"message": m7ProviderUnavailableMessage,
			"data": fiber.Map{
				"recycled_amount":      0,
				"current_balance":      currentLocalBalance,
				"game_balance":         gameBalance,
				"platform_code":        platformCode,
				"summary_sync":         gameSummarySyncResult,
				"summary_error":        gameSummarySyncError,
				"summary_pending":      false,
				"provider_unavailable": true,
				"provider_error_code":  "9020",
				"insurance":            insuranceSettlement,
			},
		})
	}

	if platformCode == defaultGamePlatformCode {
		client := provider.NewHedocClient()
		resp, err := client.GetBalance(helpers.GetCfgInstance().Conf.Agentid, thirdPartyLoginId, helpers.GetCfgInstance().Conf.Agentapi)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"code": 500, "message": "failed to query provider balance"})
		}
		gameBalance = resp.Balance
		insuranceExitAmount = resp.Balance

		if resp.Balance > 0 {
			targetAmount := resp.Balance
			usdAmount := targetAmount
			billNo := fmt.Sprintf("AUTO-OUT-%d-%d", time.Now().UnixNano(), userId)
			err = recycleToLocal(usdAmount, billNo, "one-click recycle (USD)", func() error {
				updatedBalance, err := client.UpdateBalance(helpers.GetCfgInstance().Conf.Agentid,
					thirdPartyLoginId, targetAmount, 2, billNo, helpers.GetCfgInstance().Conf.Agentapi)
				if err != nil {
					return fmt.Errorf("recycle from provider failed: %v", err)
				}
				gameBalance = updatedBalance
				return nil
			})
			if err != nil {
				return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
			}
		}
	}

	if platformCode == m7GamePlatformCode {
		m7Client, err := provider.NewM7Client(m7ProviderConfig())
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
		}
		m7Balance, err := m7Client.Balance(thirdPartyLoginId)
		if err != nil {
			log.Printf("[RecycleBalance][M7] failed to query balance for user %d login %s: %v\n", userId, thirdPartyLoginId, err)
			if provider.IsM7IPNotAllowedError(err) {
				return returnM7ProviderUnavailable("balance", err)
			}
			if provider.IsM7PlayerNotFoundError(err) {
				m7Balance = 0
			} else {
				return c.Status(500).JSON(fiber.Map{"code": 500, "message": "failed to query M7 balance"})
			}
		}
		gameBalance = m7Balance
		insuranceExitAmount = m7Balance
		if m7Balance > 0 {
			targetAmount := math.Floor(m7Balance*100) / 100
			if targetAmount <= 0 {
				gameBalance = 0
			}
			billNo := fmt.Sprintf("M7-AUTO-OUT-%d-%d", time.Now().UnixNano(), userId)
			if targetAmount > 0 {
				err = recycleToLocal(targetAmount, billNo, "M7 one-click recycle (USD)", func() error {
					updatedBalance, err := m7Client.Transfer(thirdPartyLoginId, -targetAmount, billNo)
					if err != nil {
						if provider.IsM7BalanceNotEnoughError(err) {
							if latestBalance, balanceErr := m7Client.Balance(thirdPartyLoginId); balanceErr == nil {
								gameBalance = latestBalance
							}
							log.Printf("[RecycleBalance][M7] skip transfer out for user %d login %s amount %.2f bill %s: provider balance not enough: %v\n", userId, thirdPartyLoginId, targetAmount, billNo, err)
							return m7BalanceNotEnoughErr
						}
						log.Printf("[RecycleBalance][M7] failed to transfer out for user %d login %s amount %.2f bill %s: %v\n", userId, thirdPartyLoginId, targetAmount, billNo, err)
						return fmt.Errorf("recycle from M7 failed: %v", err)
					}
					gameBalance = updatedBalance
					return nil
				})
				if err != nil {
					if errors.Is(err, m7BalanceNotEnoughErr) {
						recycledAmount = 0
					} else if provider.IsM7IPNotAllowedError(err) {
						return returnM7ProviderUnavailable("transfer", err)
					} else {
						return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
					}
				}
			}
		}
	}

	if platformCode == m7ppGamePlatformCode {
		m7ppClient, err := provider.NewM7PPClient(m7ppProviderConfig())
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
		}
		balance, err := m7ppClient.Balance(thirdPartyLoginId)
		if err != nil {
			if provider.IsM7PPWalletNotInitializedError(err) {
				log.Printf("[RecycleBalance][M7PP] treating uninitialized wallet as zero balance for user %d login %s: %v\n", userId, thirdPartyLoginId, err)
				balance = 0
			} else {
				log.Printf("[RecycleBalance][M7PP] failed to query balance for user %d login %s: %v\n", userId, thirdPartyLoginId, err)
				return c.Status(500).JSON(fiber.Map{"code": 500, "message": "failed to query M7PP balance"})
			}
		}
		gameBalance = balance
		insuranceExitAmount = balance
		amount := math.Floor(balance*100) / 100
		if amount > 0 {
			billNo := fmt.Sprintf("M7PP-AUTO-OUT-%d-%d", time.Now().UnixNano(), userId)
			err = recycleToLocal(amount, billNo, "M7PP one-click recycle (USD)", func() error {
				updatedBalance, err := m7ppClient.Withdraw(thirdPartyLoginId, amount, billNo)
				if err != nil {
					return fmt.Errorf("recycle from M7PP failed: %w", err)
				}
				gameBalance = updatedBalance
				return nil
			})
			if err != nil {
				return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
			}
		}
	}

	if recycledAmount <= 0 {
		var user dtos.User
		if err := models.GetInstance().DbInstance.Select("balance").First(&user, userId).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"code": 500, "message": "failed to load user info"})
		}
		currentLocalBalance = user.Balance
	}

	insuranceSettlement = settleAddDesktopInsuranceIfNeeded(c, uint64(userId), platformCode, insuranceExitAmount)
	if insuranceSettlement != nil && insuranceSettlement.Triggered {
		currentLocalBalance = insuranceSettlement.CurrentBalance
	}

	if platformCode == m7GamePlatformCode {
		gameSummarySyncPending = true
		TriggerTodayPlayerDailyGameSummarySyncIfStale(uint64(userId), 30*time.Second, req.ForceSummarySync || recycledAmount > 0)
		if req.ForceSummarySync || recycledAmount > 0 {
			TriggerTodayPlayerDailyGameSummarySyncAfter(uint64(userId), 20*time.Second)
		}
	} else {
		TriggerTodayPlayerDailyGameSummarySyncIfStale(uint64(userId), 30*time.Second, req.ForceSummarySync || recycledAmount > 0)
	}
	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data": fiber.Map{
			"recycled_amount": recycledAmount,
			"current_balance": currentLocalBalance,
			"game_balance":    gameBalance,
			"platform_code":   platformCode,
			"summary_sync":    gameSummarySyncResult,
			"summary_error":   gameSummarySyncError,
			"summary_pending": gameSummarySyncPending,
			"insurance":       insuranceSettlement,
		},
	})
}

// 资金转动
func TransferToProvider(c *fiber.Ctx) error {
	var req requests.TransferRequest
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(400)
	}

	userId := QueryUserIdFromJwt(c)
	// username := QueryUserNameFromJwt(c)

	tx := models.GetInstance().DbInstance.Begin() // 开启事务

	// 1. 锁定用户并在本地处理资金
	var user dtos.User
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userId).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"msg": "系统繁忙"})
	}

	// 创建订单号
	refId := fmt.Sprintf("TR-%d-%d", time.Now().UnixNano(), userId)

	// 记录账变 (Transaction 表)
	trans := dtos.Transaction{
		UserID:      uint64(userId),
		Amount:      req.Amount, // 注意正负
		ReferenceID: refId,
		Type:        3, // 3=上下分
	}

	if req.Type == 1 {
		// 转入游戏：本地扣钱
		if user.Balance < req.Amount {
			tx.Rollback()
			return c.Status(400).JSON(fiber.Map{"msg": "余额不足"})
		}
		trans.Amount = -req.Amount
		trans.Remark = "转入游戏平台"
		user.Balance -= req.Amount
	} else {
		// 转出游戏：本地加钱 (注意：实际应该先查第三方余额够不够，这里简化直接调API)
		trans.Amount = req.Amount
		trans.Remark = "从游戏平台转出"
		user.Balance += req.Amount
	}

	// 更新本地数据库
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Create(&trans).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 2. 调用第三方 API
	client := provider.NewHedocClient()
	// 如果是网络错误，这里其实比较危险。生产环境通常是：
	// 先提交"处理中"状态，异步去调三方，或者在这里调，超时则回滚。
	// 这里演示同步调用，失败回滚模式：
	err := client.Transfer(userId, req.Amount, req.Type, refId)

	if err != nil {
		tx.Rollback() // API失败，回滚本地资金
		return c.Status(500).JSON(fiber.Map{"success": false, "msg": "转账失败: " + err.Error()})
	}

	// 3. 成功，提交事务
	tx.Commit()

	return c.JSON(fiber.Map{
		"success": true,
		"balance": user.Balance,
		"msg":     "转账成功",
	})
}

func UpdateGamePassword(c *fiber.Ctx) error {
	type Req struct {
		NewPassword string `json:"new_password"`
	}
	var req Req
	c.BodyParser(&req)
	userId := QueryUserIdFromJwt(c)

	client := provider.NewHedocClient()
	if err := client.ChangePassword(userId, req.NewPassword, helpers.GetCfgInstance().Conf.Agentapi); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "msg": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true})
}

// Deposit 本地充值 (简化版：直接加钱，实际应为回调处理)
func Deposit(c *fiber.Ctx) error {
	var req requests.DepositRequest
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(400)
	}

	userId := QueryUserIdFromJwt(c)

	err := models.GetInstance().DbInstance.Transaction(func(tx *gorm.DB) error {
		var user dtos.User
		// 锁行
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userId).Error; err != nil {
			return err
		}

		// 加钱
		newBalance := user.Balance + req.Amount

		// 记录流水
		log := dtos.Transaction{
			UserID:        uint64(userId),
			Type:          1, // 1=充值
			Amount:        req.Amount,
			BeforeBalance: user.Balance,
			AfterBalance:  newBalance,
			Remark:        "在线充值: " + req.Channel,
			ReferenceID:   fmt.Sprintf("DEP-%d", time.Now().Unix()),
		}

		if err := tx.Create(&log).Error; err != nil {
			return err
		}
		if err := tx.Model(&user).Update("balance", newBalance).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "msg": "充值失败"})
	}

	// [新增] 异步记录充值统计
	go func() {
		RecordDailyStats(uint64(userId), 0, 0, req.Amount, 0, 0)
	}()

	return c.JSON(fiber.Map{"success": true, "msg": "充值成功"})
}

// Withdraw 提现 (本地扣钱，通常先冻结或扣除，等待后台人工审核)
func Withdraw(c *fiber.Ctx) error {
	var req requests.WithdrawRequest
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(400)
	}

	userId := QueryUserIdFromJwt(c)

	err := models.GetInstance().DbInstance.Transaction(func(tx *gorm.DB) error {
		var user dtos.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userId).Error; err != nil {
			return err
		}

		if user.Balance < req.Amount {
			return fmt.Errorf("余额不足")
		}

		newBalance := user.Balance - req.Amount

		// 记录流水
		log := dtos.Transaction{
			UserID:        uint64(userId),
			Type:          2,           // 2=提现
			Amount:        -req.Amount, // 负数
			BeforeBalance: user.Balance,
			AfterBalance:  newBalance,
			Remark:        "提现申请: " + req.BankInfo,
			ReferenceID:   fmt.Sprintf("WIT-%d", time.Now().Unix()),
		}

		if err := tx.Create(&log).Error; err != nil {
			return err
		}
		if err := tx.Model(&user).Update("balance", newBalance).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "msg": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "msg": "提现申请已提交"})
}

// GetGameList 获取游戏列表接口
// Method: POST
// Content-Type: application/json
func GetGameList(c *fiber.Ctx) error {
	if err := SyncAllGamesIfStale(gameListSyncInterval); err != nil {
		log.Printf("[Sync] Game list sync before GetGameList failed: %v\n", err)
	}
	// 1. 初始化默认参数
	req := requests.GameListClientRequest{
		Page:     1,
		PageSize: 20,
	}

	// 2. 解析 JSON Body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    400,
			"message": "Invalid JSON format",
			"error":   err.Error(),
		})
	}
	platformCode := normalizeGamePlatformCode(req.PlatformCode)

	// 3. 参数边界修正
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100 // 限制最大页大小，防止性能问题
	}

	// 4. 构建基础查询
	// 使用 Joins 关联 game_types 和 game_providers 表，以便利用 Code 字段进行筛选
	query := models.GetInstance().DbInstance.Model(&dtos.Game{}).
		Joins("LEFT JOIN game_types ON game_types.id = games.game_type_id").
		Joins("LEFT JOIN game_providers ON game_providers.id = games.provider_id AND game_providers.platform_code = games.platform_code").
		Where("games.platform_code = ?", platformCode).
		Where("games.status = ?", 1).         // 仅查询上架状态的游戏
		Where("game_types.status = ?", 1).    //
		Where("game_providers.status = ?", 1) //

	// 5. 应用筛选条件
	// 5.1 游戏类型筛选 (排除空字符串和 "ALL")
	if req.GameTypeCode != "" && req.GameTypeCode != "ALL" {
		query = query.Where("game_types.code = ?", req.GameTypeCode)
	}

	// 5.2 厂商筛选 (排除空字符串和 "ALL")
	if req.ProviderCode != "" && req.ProviderCode != "ALL" {
		query = query.Where("game_providers.code = ?", req.ProviderCode)
	}

	// 5.3 游戏名称模糊搜索
	if req.GameName != "" {
		// 因为 name 字段是 JSON 类型 (例如 `{"CN":"麻将","EN":"Mahjong"}`)
		// 使用 LIKE %keyword% 可以同时匹配中文或英文名称
		query = query.Where("games.name LIKE ?", "%"+req.GameName+"%")
	}

	// 6. 获取符合条件的总记录数 (Total Count)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": "Database query error",
		})
	}

	// 7. 执行最终查询 (分页 + 排序 + 预加载详情)
	var games []dtos.Game
	offset := (req.Page - 1) * req.PageSize
	orderClause := "games.sort DESC, games.views DESC"
	if req.ProviderCode == "" || req.ProviderCode == "ALL" {
		orderClause = pgProviderOrderSQL + ", " + orderClause
	}

	err := query.
		Preload("Provider"). // 预加载 Provider 详情 (用于返回厂商名)
		Preload("GameType"). // 预加载 GameType 详情 (用于返回分类名)
		Order(orderClause).  // 排序: PG 优先, 权重优先, 其次热度
		Limit(req.PageSize).
		Offset(offset).
		Find(&games).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": "Failed to fetch games",
		})
	}

	if platformCode == defaultGamePlatformCode {
		for index := range games {
			games[index].ImgUrl = normalizeHedocGameImageURL(games[index].ImgUrl, helpers.GetCfgInstance().Conf.Agentapi)
		}
	}

	// 8. 计算总页数
	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(req.PageSize)))
	}

	// 9. 返回标准格式响应
	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data": fiber.Map{
			"list":        games,
			"total":       total,
			"page":        req.Page,
			"page_size":   req.PageSize,
			"total_pages": totalPages,
		},
	})
}

// GetProviderList 获取可用游戏厂商列表 (用于前端下拉筛选)
// Method: POST
// Content-Type: application/json
func GetGameOptions(c *fiber.Ctx) error {
	type GameOptionsRequest struct {
		PlatformCode string `json:"platform_code"`
	}

	req := GameOptionsRequest{}
	_ = c.BodyParser(&req)
	platformCode := normalizeGamePlatformCode(req.PlatformCode)

	// 定义返回结构
	type GameOptionsData struct {
		Providers []dtos.GameProvider `json:"providers"`
		GameTypes []dtos.GameType     `json:"game_types"`
	}

	var providers []dtos.GameProvider
	var gameTypes []dtos.GameType

	// 1. 查询可用厂商 (Status=1)
	// 只查询前端需要的字段，减少传输量
	if err := models.GetInstance().DbInstance.Model(&dtos.GameProvider{}).
		Select("id, platform_code, code, name, status,type_id").
		Where("platform_code = ? AND status = ?", platformCode, 1).
		Order(pgProviderOptionsOrderSQL + ", id ASC").
		Find(&providers).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "Failed to fetch providers"})
	}

	// 2. 查询可用分类 (Status=1)
	if err := models.GetInstance().DbInstance.Model(&dtos.GameType{}).
		Select("id, code, name, sort, status").
		Where("status = ?", 1).
		Order("sort ASC").
		Find(&gameTypes).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "Failed to fetch game types"})
	}

	// 3. 返回组合数据
	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data": GameOptionsData{
			Providers: providers,
			GameTypes: gameTypes,
		},
	})
}

// TransferToGame 资金划转 (转入/转出三方游戏)
func TransferToGame(c *fiber.Ctx) error {
	type TransferReq struct {
		Amount float64 `json:"amount"` // 始终为正数
		Type   int     `json:"type"`   // 1:转入游戏(充值) 2:转出游戏(提现)
	}

	var req TransferReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "Invalid params"})
	}

	userId := QueryUserIdFromJwt(c)

	// 生成唯一订单号
	billNo := fmt.Sprintf("TR-%d-%d", time.Now().UnixNano(), userId)
	var latestBalance float64
	var localBalance float64

	// 1. 开启事务处理本地资金
	err := models.GetInstance().DbInstance.Transaction(func(tx *gorm.DB) error {
		var user dtos.User
		// 锁行
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userId).Error; err != nil {
			return err
		}

		thirdPartyLoginId := fmt.Sprintf("%su%d", helpers.GetCfgInstance().Conf.Prefix, user.ID)

		// 记录本地账变
		trans := dtos.Transaction{
			UserID:        uint64(userId),
			Type:          3, // 3: 上下分
			ReferenceID:   billNo,
			BeforeBalance: user.Balance,
		}

		if req.Type == 1 {
			// 转入游戏 (本地扣款)
			if user.Balance < req.Amount {
				return fmt.Errorf("余额不足")
			}
			user.Balance -= req.Amount
			trans.Amount = -req.Amount
			trans.Remark = "转入游戏"
		} else {
			// 转出游戏 (本地加款)
			// 注意：理论上应先查三方余额，这里简化直接转
			user.Balance += req.Amount
			trans.Amount = req.Amount
			trans.Remark = "游戏转出"
		}
		trans.AfterBalance = user.Balance

		if err := tx.Save(&user).Error; err != nil {
			return err
		}
		if err := tx.Create(&trans).Error; err != nil {
			return err
		}

		localBalance = user.Balance

		// 2. 调用第三方接口 (在事务内调用有风险，但为了简单保持一致性，失败则回滚)
		// 生产环境建议：本地先扣款->提交事务->异步调三方->失败则冲正
		client := provider.NewHedocClient()

		// 假设文档定义: 1=Deposit(User In), 2=Withdraw(User Out)
		balance, err := client.UpdateBalance(helpers.GetCfgInstance().Conf.Agentid,
			thirdPartyLoginId, req.Amount, req.Type, billNo, helpers.GetCfgInstance().Conf.Agentapi)
		if err != nil {
			return err // 触发事务回滚，本地钱退回
		}

		latestBalance = balance
		return nil
	})

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"code":    500,
			"message": "转账失败: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data": fiber.Map{
			"local_balance": localBalance,
			"game_balance":  latestBalance,
			"bill_no":       billNo,
		},
	})
}

// OpenLobbyHandler 获取游戏大厅链接
func OpenLobbyHandler(c *fiber.Ctx) error {
	var req requests.OpenLobbyPayload
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "Invalid params"})
	}

	userId := QueryUserIdFromJwt(c)
	thirdPartyLoginId := fmt.Sprintf("%su%d", helpers.GetCfgInstance().Conf.Prefix, userId)
	// 获得用户专属的第三方密码
	thirdPartyPassword := common.GenerateGamePassword(uint64(userId))

	// 1. 获取厂商信息 (需要 ID)
	var gameprovider dtos.GameProvider
	platformCode := normalizeGamePlatformCode(req.PlatformCode)
	if err := models.GetInstance().DbInstance.
		Where("platform_code = ? AND code = ? AND status = ?", platformCode, req.ProviderCode, 1).
		First(&gameprovider).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "Unknown provider"})
	}
	if platformCode == m7GamePlatformCode || platformCode == m7ppGamePlatformCode {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": platformCode + " lobby is not supported, please open a specific game"})
	}

	client := provider.NewHedocClient()

	// 2. 确保用户已在第三方注册 (静默注册)
	// 使用固定密码或查表获取
	_ = client.RegisterUser(helpers.GetCfgInstance().Conf.Agentid, thirdPartyLoginId, thirdPartyPassword, helpers.GetCfgInstance().Conf.Agentapi)

	// 3. 请求链接
	url, err := client.OpenLobby(helpers.GetCfgInstance().Conf.Agentid,
		thirdPartyLoginId, thirdPartyPassword, req.IsMobile, req.Language, int(gameprovider.ID), helpers.GetCfgInstance().Conf.Agentapi)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}

	return c.JSON(fiber.Map{"code": 0, "message": "success", "data": fiber.Map{"url": url}})
}

// GetUserInfo 获取当前登录用户信息
func GetUserInfo(c *fiber.Ctx) error {
	// 1. 从 JWT 中获取 UserID (由中间件设置)
	userId := QueryUserIdFromJwt(c)
	if userId <= 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    401,
			"message": "Unauthorized",
		})
	}

	if _, err := services.RecalculateAndPersistUserVIPLevel(c.Context(), uint64(userId)); err != nil {
		log.Printf("[GetUserInfo] failed to recalculate vip level for user %d: %v\n", userId, err)
	}

	// 2. 查询数据库
	var user dtos.User
	// 使用 Select 指定字段，避免查出不必要的数据
	if err := models.GetInstance().DbInstance.Model(&dtos.User{}).
		Select("id, username, invite_code, balance, vip_level, level, total_deposit").
		First(&user, userId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code":    401,
				"message": "Unauthorized",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": "Failed to load user info",
		})
	}

	remainingWager, withdrawableBalance, err := services.GetActivityService().GetWithdrawWagerStatus(c.Context(), uint64(userId), user.Balance)
	if err != nil {
		log.Printf("[GetUserInfo] failed to load withdrawable balance for user %d: %v\n", userId, err)
		remainingWager = 0
		withdrawableBalance = user.Balance
	}
	depositWagerMultiplier, rewardWagerMultiplier := services.GetActivityService().GetActivityWagerMultipliers(c.Context())

	// 3. 返回数据
	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data": fiber.Map{
			"uid":                      user.ID,         // 用户ID
			"username":                 user.Username,   // 用户名
			"invite_code":              user.InviteCode, // 邀请码
			"balance":                  user.Balance,    // 当前本地余额
			"vip_level":                user.VipLevel,   // VIP等级
			"level":                    user.Level,      // 代理层级 (如果前端需要展示)
			"total_deposit":            user.TotalDeposit,
			"remaining_wager":          remainingWager,
			"withdrawable_balance":     withdrawableBalance,
			"deposit_wager_multiplier": depositWagerMultiplier,
			"reward_wager_multiplier":  rewardWagerMultiplier,
		},
	})
}

// GetInviteInfo 获取用户邀请信息
func GetInviteInfo(c *fiber.Ctx) error {
	userId := QueryUserIdFromJwt(c)

	// 定义返回结构
	type InviteData struct {
		InviteCode  string  `json:"invite_code"`
		ShareLink   string  `json:"share_link"`   // 完整的注册分享链接
		DirectCount int64   `json:"direct_count"` // 直推人数 (一级下线)
		TeamCount   int64   `json:"team_count"`   // 团队总人数 (所有下线，可选)
		TotalEarn   float64 `json:"total_earn"`   // 总佣金
	}

	var user dtos.User
	// 1. 获取当前用户的邀请码和 Path
	if err := models.GetInstance().DbInstance.Select("id, invite_code, path").First(&user, userId).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "User not found"})
	}

	// 2. 统计直推人数 (ParentID = 当前用户ID)
	var directCount int64
	if err := models.GetInstance().DbInstance.Model(&dtos.User{}).
		Where("parent_id = ?", userId).
		Count(&directCount).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "Count error"})
	}

	// 3. (可选) 统计团队总人数 (利用 Path 左前缀匹配)
	// 逻辑: 查找 path 以 "user.Path + user.ID + /" 开头的所有用户
	// 例如: 我是 "1/5/", 下级必须包含 "1/5/%"
	var teamCount int64
	// 构造当前用户的层级路径前缀 (注意: 根用户的 path 为空字符串时的处理)
	// 假设 user.ID = 5, user.Path = "1/" -> prefix = "1/5/"
	// 假设 user.ID = 1, user.Path = ""   -> prefix = "1/"

	/*
	   注意：如果你的系统数据量非常大（千万级），实时 Count Like 查询可能会略慢，
	   建议后续将 team_count 做成字段存入 users 表异步更新。
	   但在百万级数据下，path 有索引，Count 依然很快。
	*/

	// 简单拼接逻辑
	pathPrefix := user.Path + fmt.Sprintf("%d/", user.ID)
	if user.Path == "" {
		pathPrefix = fmt.Sprintf("%d/", user.ID)
	}

	models.GetInstance().DbInstance.Model(&dtos.User{}).
		Where("path LIKE ?", pathPrefix+"%").
		Count(&teamCount)

	// 4. 拼接分享链接
	shareLink := fmt.Sprintf("https://www.your-game-site.com/register?code=%s", user.InviteCode)

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data": InviteData{
			InviteCode:  user.InviteCode,
			ShareLink:   shareLink,
			DirectCount: directCount,
			TeamCount:   teamCount,
			TotalEarn:   7.7, // 暂时
		},
	})
}

// GetBanners 获取首页轮播图
// Method: GET
func GetBanners(c *fiber.Ctx) error {
	var banners []dtos.Banner

	// 查询逻辑：
	// 1. status = 1 (启用)
	// 2. 按 sort 倒序排列 (权重越大越靠前)，其次按 id 倒序
	err := models.GetInstance().DbInstance.Model(&dtos.Banner{}).
		Where("status = ?", 1).
		Order("sort DESC, id DESC").
		Find(&banners).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"code":    500,
			"message": "Failed to fetch banners",
		})
	}

	// 这里的 banners 序列化为 JSON 时：
	// 1. 如果 JumpLink 为 "", 字段消失
	// 2. 如果 Buttons 为 NULL, 字段消失
	// 完美符合前端多态需求
	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data":    banners,
	})
}

// GetAllLiveGames 获取所有真人游戏 (Type=2) - 针对大数据量的优化版
func GetAllGames(c *fiber.Ctx) error {
	// 目标类型 ID (根据需求固定为 2，即真人视讯)
	targetTypeID := 2

	// 定义接收结果的切片
	var games []responses.GameLiteResponse

	// ---------------------------------------------------------
	// 优化点 1: 字段裁剪 (Select)
	// 只查询前端渲染列表必须的字段，不查 created_at, sort, views 等
	// ---------------------------------------------------------
	err := models.GetInstance().DbInstance.Table("games").
		Select("id, platform_code, game_code as code, name, img_url as img, provider_id as pid").
		Joins("LEFT JOIN game_providers ON game_providers.id = games.provider_id AND game_providers.platform_code = games.platform_code").
		Where("games.platform_code = ?", defaultGamePlatformCode).
		Where("game_type_id = ?", targetTypeID).
		Where("games.status = ?", 1).          // 只查上架的
		Where("game_providers.status = ?", 1). // 厂商也必须开启
		// ---------------------------------------------------------
		// 优化点 2: 移除 Order By (如果数据量极大)
		// 排序是非常消耗数据库内存的操作。如果前端可以自己排，或者对顺序不敏感，
		// 去掉 Order By 可以显著提升查询速度。如果必须排，确保有索引。
		// ---------------------------------------------------------
		// Order("sort DESC").
		Scan(&games).Error // 使用 Scan 映射到精简结构体

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "Query failed"})
	}

	// ---------------------------------------------------------
	// 优化点 3: HTTP 缓存控制 (Client Cache)
	// 告诉浏览器/客户端：这个列表在 5 分钟内不会变，不要重复请求。
	// ---------------------------------------------------------
	c.Set("Cache-Control", "public, max-age=300")

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data":    games,
		// count: len(games) // 可选，返回总数
	})
}

// GetFeaturedGames 获取特色游戏榜单 (热门 + 推荐)
func GetFeaturedGames(c *fiber.Ctx) error {
	var (
		hotList []dtos.Game
		recList []dtos.Game
		errHot  error
		errRec  error
	)

	// 使用 WaitGroup 并发查询
	var wg sync.WaitGroup
	wg.Add(2)

	// 1. 热门游戏 (Top 20 Views)
	go func() {
		defer wg.Done()
		errHot = models.GetInstance().DbInstance.Model(&dtos.Game{}).
			Joins("LEFT JOIN game_providers ON game_providers.id = games.provider_id AND game_providers.platform_code = games.platform_code").
			Where("games.platform_code = ?", defaultGamePlatformCode).
			Where("games.status = ?", 1).
			Where("game_providers.status = ?", 1).
			Preload("Provider"). // 仅加载厂商，通常榜单不需要分类详情
			Order("views DESC"). // 走 idx_views 索引
			Limit(20).
			Find(&hotList).Error
	}()

	// 2. 推荐游戏 (Top 20 Sort)
	go func() {
		defer wg.Done()
		errRec = models.GetInstance().DbInstance.Model(&dtos.Game{}).
			Joins("LEFT JOIN game_providers ON game_providers.id = games.provider_id AND game_providers.platform_code = games.platform_code").
			Where("games.platform_code = ?", defaultGamePlatformCode).
			Where("games.status = ?", 1).
			Where("game_providers.status = ?", 1).
			Preload("Provider").
			Order("sort DESC"). // 走 idx_sort_views 索引
			Limit(20).
			Find(&recList).Error
	}()

	wg.Wait()

	if errHot != nil || errRec != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "Failed to fetch featured games"})
	}

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data": fiber.Map{
			"hot_games":       hotList,
			"recommend_games": recList,
		},
	})
}

// OpenLobbyHandler2 获取游戏大厅链接+自动带入
func OpenLobbyHandler2(c *fiber.Ctx) error {
	localBalance := 0.0
	gameBalance := 0.0
	entryBillNo := ""
	entryAmountU := 0.0
	var req requests.OpenLobbyPayload
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "Invalid params"})
	}
	userId := QueryUserIdFromJwt(c)
	thirdPartyLoginId := fmt.Sprintf("%su%d", helpers.GetCfgInstance().Conf.Prefix, userId)
	thirdPartyPassword := common.GenerateGamePassword(uint64(userId))
	var gameprovider dtos.GameProvider
	platformCode := normalizeGamePlatformCode(req.PlatformCode)
	if err := models.GetInstance().DbInstance.
		Where("platform_code = ? AND code = ? AND status = ?", platformCode, req.ProviderCode, 1).
		First(&gameprovider).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "Unknown provider"})
	}
	if platformCode == m7GamePlatformCode || platformCode == m7ppGamePlatformCode {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": platformCode + " lobby is not supported, please open a specific game"})
	}
	desktopReward := claimDesktopLaunchRewardIfNeeded(c, uint64(userId), req.LaunchSource)
	client := provider.NewHedocClient()
	_ = client.RegisterUser(helpers.GetCfgInstance().Conf.Agentid, thirdPartyLoginId, thirdPartyPassword, helpers.GetCfgInstance().Conf.Agentapi)
	err := models.GetInstance().DbInstance.Transaction(func(tx *gorm.DB) error {
		var user dtos.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userId).Error; err != nil {
			return err
		}
		if user.Balance > 0 {
			usdAmount := user.Balance
			targetAmount := usdAmount
			billNo := fmt.Sprintf("AUTO-IN-%d-%d", time.Now().UnixNano(), userId)
			user.Balance = 0
			trans := dtos.Transaction{
				UserID:        uint64(userId),
				Type:          3,
				Amount:        -usdAmount,
				BeforeBalance: usdAmount,
				AfterBalance:  0,
				ReferenceID:   billNo,
				Remark:        "auto transfer in (USD)",
			}
			if err := tx.Save(&user).Error; err != nil {
				return err
			}
			if err := tx.Create(&trans).Error; err != nil {
				return err
			}
			balance, err := client.UpdateBalance(helpers.GetCfgInstance().Conf.Agentid,
				thirdPartyLoginId, targetAmount, 1, billNo, helpers.GetCfgInstance().Conf.Agentapi)
			gameBalance = balance
			if err != nil {
				return fmt.Errorf("transfer to provider failed: %v", err)
			}
			entryBillNo = billNo
			entryAmountU = usdAmount
		}
		localBalance = user.Balance
		return nil
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	recordAddDesktopGameEntryIfNeeded(c, uint64(userId), defaultGamePlatformCode, entryBillNo, entryAmountU)
	url, err := client.OpenLobby(helpers.GetCfgInstance().Conf.Agentid,
		thirdPartyLoginId, thirdPartyPassword, req.IsMobile, req.Language, int(gameprovider.ID), helpers.GetCfgInstance().Conf.Agentapi)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	if latestBalance, balanceErr := getThirdPartyBalanceSnapshot(client, thirdPartyLoginId); balanceErr != nil {
		log.Printf("[OpenLobbyHandler2] failed to refresh third-party balance for user %d: %v\n", userId, balanceErr)
	} else {
		gameBalance = latestBalance
	}
	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data": fiber.Map{
			"url":                    url,
			"local_balance":          localBalance,
			"game_balance":           gameBalance,
			"desktop_reward_claim":   desktopReward,
			"desktop_reward_checked": desktopReward != nil,
		},
	})
}

// ToggleFavorite 添加或移除收藏
// Method: POST
func ToggleFavorite(c *fiber.Ctx) error {
	var req requests.FavoriteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "Invalid JSON"})
	}

	userId := QueryUserIdFromJwt(c)

	// 1. 检查游戏是否存在
	var gameCount int64
	models.GetInstance().DbInstance.Model(&dtos.Game{}).Where("id = ?", req.GameID).Count(&gameCount)
	if gameCount == 0 {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "Game not found"})
	}

	// 2. 根据 IsFavorite 执行不同逻辑
	if req.IsFavorite {
		// --- 添加收藏 ---
		// 使用 FirstOrCreate 防止重复插入报错
		fav := dtos.UserFavorite{
			UserID: uint64(userId),
			GameID: req.GameID,
		}
		if err := models.GetInstance().DbInstance.FirstOrCreate(&fav, dtos.UserFavorite{UserID: uint64(userId), GameID: req.GameID}).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"code": 500, "message": "Failed to add favorite"})
		}
	} else {
		// --- 取消收藏 ---
		// 直接删除记录
		if err := models.GetInstance().DbInstance.Where("user_id = ? AND game_id = ?", userId, req.GameID).Delete(&dtos.UserFavorite{}).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"code": 500, "message": "Failed to remove favorite"})
		}
	}

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
	})
}

// GetFavoriteList 获取我的收藏列表
// Method: POST
// Content-Type: application/json
func GetFavoriteList(c *fiber.Ctx) error {
	userId := QueryUserIdFromJwt(c)
	if userId <= 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": 401, "message": "Unauthorized"})
	}

	db := models.GetInstance().DbInstance
	if !db.Migrator().HasTable(&dtos.UserFavorite{}) {
		return c.JSON(fiber.Map{
			"code":    0,
			"message": "success",
			"data": fiber.Map{
				"list":        []dtos.Game{},
				"total":       0,
				"page":        1,
				"page_size":   20,
				"total_pages": 0,
			},
		})
	}

	// 1. 解析参数
	req := requests.FavoriteListRequest{
		Page:     1,
		PageSize: 20,
	}
	if err := c.BodyParser(&req); err != nil {
		// 如果解析失败，通常说明没有传 Body，使用默认值即可，不一定要报错
		// 或者返回 400: return c.Status(400).JSON(fiber.Map{"code": 400, "message": "Invalid JSON"})
	}

	// 2. 参数修正
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	var games []dtos.Game
	var total int64
	offset := (req.Page - 1) * req.PageSize

	// 3. 构建查询
	// 使用 Join 关联 user_favorites 表
	query := db.Model(&dtos.Game{}).
		Joins("JOIN user_favorites ON user_favorites.game_id = games.id").
		Joins("JOIN game_providers ON game_providers.id = games.provider_id AND game_providers.platform_code = games.platform_code").
		Where("user_favorites.user_id = ?", userId).
		Where("games.status = ?", 1).
		Where("game_providers.status = ?", 1) // 依然要判断游戏和厂商是否上架

	// 4. 获取总数
	if err := query.Count(&total).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "Database error"})
	}

	// 5. 获取列表详情
	err := query.
		Preload("Provider").
		Preload("GameType").
		Order("user_favorites.created_at DESC"). // 按收藏时间倒序
		Limit(req.PageSize).
		Offset(offset).
		Find(&games).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "Query failed"})
	}

	// 6. 计算总页数
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(req.PageSize) - 1) / int64(req.PageSize))
	}

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data": fiber.Map{
			"list":        games,
			"total":       total,
			"page":        req.Page,
			"page_size":   req.PageSize,
			"total_pages": totalPages,
		},
	})
}

// GetReorderGameList 获取游戏特定列表 (热门/最新/推荐)
// Method: POST
func GetReorderGameList(c *fiber.Ctx) error {
	// 1. 初始化参数
	req := requests.TolsGameListRequest{
		Page:     1,
		PageSize: 20,
		SortType: "RECOMMEND", // 默认按推荐排序
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "Invalid JSON"})
	}
	platformCode := normalizeGamePlatformCode(req.PlatformCode)

	// 参数修正
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	if req.PageSize > 50 {
		req.PageSize = 50
	} // 限制单页数量

	// 2. 构建基础查询
	// 核心条件: 必须是上架游戏 (games.status=1) 且 分类是电子 (game_types.code='SLOT')
	query := models.GetInstance().DbInstance.Model(&dtos.Game{}).
		Joins("LEFT JOIN game_types ON game_types.id = games.game_type_id").
		Joins("LEFT JOIN game_providers ON game_providers.id = games.provider_id AND game_providers.platform_code = games.platform_code").
		Where("games.platform_code = ?", platformCode).
		Where("games.status = ?", 1).
		Where("game_types.code = ?", "SLOT"). // 强制只查 Slot
		Where("game_types.status = ?", 1).
		Where("game_providers.status = ?", 1) // 厂商也必须是启用状态

	// 3. 根据 SortType 应用不同的排序逻辑
	switch req.SortType {
	case "POPULAR":
		// 按热度倒序 (Views)
		query = query.Order("games.views DESC")
	case "NEW":
		// 按创建时间倒序 (Newest)
		query = query.Order("games.created_at DESC")
	case "RECOMMEND":
		// 按权重倒序 (Recommend)
		query = query.Order("games.sort DESC")
	default:
		// 默认兜底: 权重倒序
		query = query.Order("games.sort DESC")
	}

	// 4. 二级排序 (防止主排序字段相同时乱序)
	query = query.Order("games.id DESC")

	// 5. 获取总数 (Count)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "Database error"})
	}

	// 6. 获取列表 (Find)
	var games []dtos.Game
	offset := (req.Page - 1) * req.PageSize

	err := query.
		Preload("Provider"). // 加载厂商信息
		// Preload("GameType"). // Slot列表通常不需要返回Type信息，可省略以提升性能
		Limit(req.PageSize).
		Offset(offset).
		Find(&games).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "Query failed"})
	}

	// 7. 计算总页数
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(req.PageSize) - 1) / int64(req.PageSize))
	}

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data": fiber.Map{
			"list":        games,
			"total":       total,
			"page":        req.Page,
			"page_size":   req.PageSize,
			"total_pages": totalPages,
			"sort_type":   req.SortType, // 返回当前使用的排序类型
		},
	})
}
