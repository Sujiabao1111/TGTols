package controllers

import (
	"fmt"
	"gogogo/helpers"
	"gogogo/models"
	cooo "gogogo/services/dtos"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func JsonEchoHandler(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(&cooo.RespJsonEchoBack{
		Success: true,
		Message: fmt.Sprintf("You are at the endpoint version:%s", helpers.Version),
		Ip:      c.IP(),
	})
}

func JsonPostEchoHandler(c *fiber.Ctx) error {
	var body cooo.Todo
	err := c.BodyParser(&body)
	if err != nil {
		return helpers.ErrorMsgReturn(c, helpers.ErrorParseErrorStr, helpers.ErrorParseError)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"todo": body,
		},
	})
}

func DowakeUp(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "ok",
	})
}

func GetAppCfgs(c *fiber.Ctx) error {
	type sysParamRow struct {
		Value string `gorm:"column:value"`
	}

	whatsAppURL := "https://wa.me/message/G4CV6NJKBOKAM1"
	var row sysParamRow
	if db := models.GetInstance().DbInstance; db != nil {
		if err := db.Table("sys_params").
			Select("value").
			Where("`key` = ?", "WHATSAPP_URL").
			Take(&row).Error; err == nil && row.Value != "" {
			whatsAppURL = row.Value
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "ok",
		"data": fiber.Map{
			"whatsapp_url": whatsAppURL,
		},
	})
}

func GetJwtForTest(c *fiber.Ctx) error {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = "John Doe"
	claims["admin"] = true
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	t, err := token.SignedString([]byte(helpers.GetCfgInstance().Conf.Jwt))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "ok",
		"token":   t,
	})
}
