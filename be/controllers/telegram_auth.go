package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"gogogo/common"
	"gogogo/helpers"
	"gogogo/models"
	"gogogo/models/dtos"
	"gogogo/models/responses"
	"gogogo/services"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type telegramLoginRequest struct {
	InitData   string `json:"init_data"`
	StartParam string `json:"start_param"`
	SiteDomain string `json:"site_domain"`
}

func TelegramLoginHandler(c *fiber.Ctx) error {
	var req telegramLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.LoginResponse{Success: false, Message: "invalid request"})
	}
	tg, values, err := services.ValidateTelegramInitData(req.InitData)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(responses.LoginResponse{Success: false, Message: err.Error()})
	}
	if req.StartParam == "" {
		req.StartParam = values.Get("start_param")
	}
	db := models.GetInstance().DbInstance
	telegramID := strconv.FormatInt(tg.ID, 10)
	var user dtos.User
	err = db.Where("telegram_user_id = ?", telegramID).First(&user).Error
	created := false
	if err == gorm.ErrRecordNotFound {
		user, err = createTelegramUser(db, tg, telegramID, req.StartParam, QueryClientIP(c), firstNonEmptyString(NormalizeRequestDomain(req.SiteDomain), QueryRequestDomain(c)))
		created = err == nil
	}
	if err != nil {
		fmt.Printf("[TelegramLogin] user lookup/create failed: telegram_id=%s err=%v\n", telegramID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(responses.LoginResponse{Success: false, Message: "Telegram account login failed: " + err.Error()})
	}
	if user.Status == 0 {
		return c.Status(fiber.StatusForbidden).JSON(responses.LoginResponse{Success: false, Message: "account disabled"})
	}
	now := time.Now()
	_ = db.Model(&user).Updates(map[string]interface{}{"telegram_username": tg.Username, "telegram_first_name": tg.FirstName, "telegram_last_name": tg.LastName, "telegram_photo_url": tg.PhotoURL, "last_login_at": now, "last_login_ip": QueryClientIP(c), "updated_at": now}).Error
	token, err := common.GenerateJWT(user.ID, user.Username, helpers.GetCfgInstance().Conf.Jwt)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	go RecordDailyStats(user.ID, 0, 0, 0, 0, 1)
	TriggerTodayPlayerDailyGameSummarySync(user.ID)
	if created {
		services.GetFacebookPixelService().TrackCompleteRegistration(user.ID, user.ParentID, buildFBEventContext(c))
	}
	return c.JSON(responses.LoginResponse{Success: true, Message: "Telegram login successful", Token: token, Data: fiber.Map{"user_id": user.ID, "username": user.Username, "vip_level": user.VipLevel, "balance": user.Balance, "invite_code": user.InviteCode, "created": created}})
}

func createTelegramUser(db *gorm.DB, tg *services.TelegramWebAppUser, telegramID, startParam, ip, domain string) (dtos.User, error) {
	passwordBytes := make([]byte, 32)
	if _, err := rand.Read(passwordBytes); err != nil {
		return dtos.User{}, err
	}
	password, _ := common.HashPassword(hex.EncodeToString(passwordBytes))
	user := dtos.User{Username: "tg_" + telegramID, Password: password, Status: 1, InviteCode: common.GenerateUniqueCode(), Level: 1, RegisterIP: ip, RegisterDomain: domain, TelegramUserID: telegramID, TelegramUsername: tg.Username, TelegramFirstName: tg.FirstName, TelegramLastName: tg.LastName, TelegramPhotoURL: tg.PhotoURL}
	now := time.Now()
	user.TelegramBoundAt = &now
	inviteCode := strings.TrimPrefix(strings.TrimSpace(startParam), "ref_")
	if inviteCode != "" {
		var parent dtos.User
		if db.Where("invite_code = ?", inviteCode).First(&parent).Error == nil {
			user.ParentID = parent.ID
			user.Level = parent.Level + 1
			if parent.Path == "" {
				user.Path = fmt.Sprintf("%d/", parent.ID)
			} else {
				user.Path = fmt.Sprintf("%s%d/", parent.Path, parent.ID)
			}
		}
	}
	err := db.Create(&user).Error
	if err != nil {
		var existing dtos.User
		if db.Where("telegram_user_id = ?", telegramID).First(&existing).Error == nil {
			return existing, nil
		}
	}
	return user, err
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
