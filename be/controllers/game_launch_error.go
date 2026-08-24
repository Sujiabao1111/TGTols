package controllers

import (
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gogogo/models"
	"gogogo/models/dtos"
	"gogogo/models/requests"
)

const (
	GameLaunchErrorTypeLaunchRequestFailed = "launch_request_failed"
	GameLaunchErrorTypeIframeLoadTimeout   = "iframe_load_timeout"
)

func ReportGameLaunchError(c *fiber.Ctx) error {
	userID := QueryUserIdFromJwt(c)
	if userID <= 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    fiber.StatusUnauthorized,
			"message": "Unauthorized",
		})
	}

	var req requests.GameLaunchErrorReportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    fiber.StatusBadRequest,
			"message": "Invalid params",
		})
	}

	errorType := normalizeGameLaunchErrorType(req.ErrorType)
	if errorType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    fiber.StatusBadRequest,
			"message": "Invalid error type",
		})
	}

	record := dtos.UserGameLaunchError{
		UserID:       uint64(userID),
		Username:     truncateGameLaunchErrorText(QueryUserNameFromJwt(c), 100),
		GameCode:     truncateGameLaunchErrorText(req.GameCode, 100),
		GameName:     truncateGameLaunchErrorText(req.GameName, 255),
		ProviderCode: truncateGameLaunchErrorText(req.ProviderCode, 50),
		ProviderName: truncateGameLaunchErrorText(req.ProviderName, 100),
		IsLobby:      req.IsLobby,
		IsMobile:     req.IsMobile,
		Language:     truncateGameLaunchErrorText(req.Language, 20),
		ErrorType:    errorType,
		ErrorMessage: truncateGameLaunchErrorText(req.ErrorMessage, 4000),
		PageURL:      truncateGameLaunchErrorText(req.PageURL, 2000),
		GameURLHost:  truncateGameLaunchErrorText(extractGameLaunchErrorHost(req.GameURL), 255),
		ClientIP:     truncateGameLaunchErrorText(QueryClientIP(c), 64),
		UserAgent:    truncateGameLaunchErrorText(c.Get("User-Agent"), 2000),
	}

	if err := models.GetInstance().DbInstance.Create(&record).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    fiber.StatusInternalServerError,
			"message": "Failed to save launch error",
		})
	}

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
	})
}

func normalizeGameLaunchErrorType(raw string) string {
	switch strings.TrimSpace(raw) {
	case GameLaunchErrorTypeLaunchRequestFailed:
		return GameLaunchErrorTypeLaunchRequestFailed
	case GameLaunchErrorTypeIframeLoadTimeout:
		return GameLaunchErrorTypeIframeLoadTimeout
	default:
		return ""
	}
}

func extractGameLaunchErrorHost(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(parsed.Host)
}

func truncateGameLaunchErrorText(raw string, maxLen int) string {
	raw = strings.TrimSpace(raw)
	if maxLen <= 0 || len(raw) <= maxLen {
		return raw
	}
	return raw[:maxLen]
}
