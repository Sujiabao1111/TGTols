package controllers

import (
	"fmt"
	"gogogo/helpers"
	cooo "gogogo/services/dtos"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func JsonEchoHandler(c *fiber.Ctx) error {
	// return c.Status(fiber.StatusOK).JSON(fiber.Map{
	// 	"success": true,
	// 	"message": fmt.Sprintf("You are at the endpoint 😉 version:%s", helpers.Version),
	// 	"Ip":      c.IP(),
	// })
	return c.Status(fiber.StatusOK).JSON(&cooo.RespJsonEchoBack{
		Success: true,
		Message: fmt.Sprintf("You are at the endpoint 😉 version:%s", helpers.Version),
		Ip:      c.IP(),
	})
}

func JsonPostEchoHandler(c *fiber.Ctx) error {
	var body cooo.Todo
	err := c.BodyParser(&body)
	if err != nil {
		return helpers.ErrorMsgReturn(c, helpers.ErrorParseErrorStr, helpers.ErrorParseError)
	} else {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"data": fiber.Map{
				"todo": body,
			},
		})
	}
}

func DowakeUp(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "You are at the endpoint 😉",
	})
}

func GetAppCfgs(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "You are at the endpoint 😉",
	})
}

func GetJwtForTest(c *fiber.Ctx) error {
	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = "John Doe"
	claims["admin"] = true
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(helpers.GetCfgInstance().Conf.Jwt))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "You are at the endpoint 😉",
		"token":   t,
	})
}
