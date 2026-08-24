package controllers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// ExchangeRateController 汇率控制器
type ExchangeRateController struct{}

// NewExchangeRateController 创建汇率控制器
func NewExchangeRateController() *ExchangeRateController {
	return &ExchangeRateController{}
}

// GetExchangeRates 获取所有汇率
// GET /api/exchange-rates
func (c *ExchangeRateController) GetExchangeRates(ctx *fiber.Ctx) error {
	service := GetExchangeRateService()
	if service == nil {
		return ctx.Status(500).JSON(fiber.Map{
			"error": "Exchange rate service not initialized",
		})
	}

	rates, err := service.GetRates(ctx.Context())
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"error": "Failed to get exchange rates",
		})
	}

	return ctx.JSON(fiber.Map{
		"base":  "IDR",
		"rates": rates,
	})
}

// ConvertCurrency 转换货币金额
// POST /api/exchange-rates/convert
// Request: { "amount": 100000, "from": "IDR", "to": "CNY" }
func (c *ExchangeRateController) ConvertCurrency(ctx *fiber.Ctx) error {
	service := GetExchangeRateService()
	if service == nil {
		return ctx.Status(500).JSON(fiber.Map{
			"error": "Exchange rate service not initialized",
		})
	}

	type Request struct {
		Amount float64 `json:"amount"`
		From   string  `json:"from"`
		To     string  `json:"to"`
	}

	var req Request
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Amount <= 0 {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Amount must be greater than 0",
		})
	}

	if req.From == "" {
		req.From = "IDR"
	}
	if req.To == "" {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Target currency (to) is required",
		})
	}

	result, err := service.Convert(ctx.Context(), req.Amount, req.From, req.To)
	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// 获取汇率
	rate, _ := service.GetRate(ctx.Context(), req.To)

	return ctx.JSON(fiber.Map{
		"amount": req.Amount,
		"from":   req.From,
		"to":     req.To,
		"rate":   rate,
		"result": result,
	})
}

// ConvertFromIDR 从IDR转换到目标货币（简化接口）
// GET /api/exchange-rates/convert/:target?amount=100000
func (c *ExchangeRateController) ConvertFromIDR(ctx *fiber.Ctx) error {
	service := GetExchangeRateService()
	if service == nil {
		return ctx.Status(500).JSON(fiber.Map{
			"error": "Exchange rate service not initialized",
		})
	}

	target := ctx.Params("target")
	if target == "" {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Target currency is required",
		})
	}

	amount := ctx.QueryFloat("amount", 0)
	if amount <= 0 {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Invalid amount",
		})
	}

	result, err := service.ConvertFromIDR(ctx.Context(), amount, target)
	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// 获取汇率
	rate, _ := service.GetRate(ctx.Context(), target)

	return ctx.JSON(fiber.Map{
		"amount_idr":     amount,
		"target":         target,
		"rate":           rate,
		"result":         result,
		"formatted":      formatCurrency(result, target),
	})
}

// formatCurrency 格式化货币显示
func formatCurrency(amount float64, currency string) string {
	switch currency {
	case "USD":
		return "$" + formatNumber(amount, 2)
	case "CNY":
		return "¥" + formatNumber(amount, 2)
	case "RUB":
		return "₽" + formatNumber(amount, 2)
	case "EUR":
		return "€" + formatNumber(amount, 2)
	default:
		return formatNumber(amount, 2)
	}
}

// formatNumber 格式化数字
func formatNumber(n float64, decimals int) string {
	format := "%." + string(rune('0'+decimals)) + "f"
	return fmt.Sprintf(format, n)
}
