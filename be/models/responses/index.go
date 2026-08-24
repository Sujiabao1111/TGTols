package responses

import "github.com/gofiber/fiber/v2"

// LoginResponse 登录响应结构
type LoginResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Token   string    `json:"token,omitempty"`
	Data    fiber.Map `json:"data,omitempty"`
}
