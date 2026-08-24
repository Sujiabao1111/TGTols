package controllers

import "github.com/gofiber/fiber/v2"

func JsonEchoHandlerJWT(c *fiber.Ctx) error {
	uid := QueryUserIdFromJwt(c)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "You are at the endpoint with JWT 😉",
		"uid":     uid,
	})
}
