package helpers

import (
	"github.com/gofiber/fiber/v2"
)

// fiber返回错误函数
func ErrorMsgReturn(c *fiber.Ctx, msg string, code int) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"success": false,
		"message": msg,
		"code":    code,
	})
}
