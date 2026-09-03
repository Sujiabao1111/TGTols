package controllers

import (
	"bytes"
	stdjson "encoding/json"
	"fmt"
	"gogogo/helpers"
	"io"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// TestTokenPayCreateOrder 测试 TokenPay 创建订单接口
func TestTokenPayCreateOrder(c *fiber.Ctx) error {
	cfg := helpers.GetCfgInstance()
	if cfg == nil {
		return c.Status(500).JSON(fiber.Map{"error": "配置未加载"})
	}

	tokenPayCfg := cfg.Conf.Payment.TokenPay

	// 测试用户标识 (可以传用户ID或邮箱)
	testUserKey := "test_user_001"
	orderID := "TEST" + fmt.Sprintf("%d", c.Context().Time().Unix())

	// 获取币种参数，支持 EVM_ETH_USDT_ERC20 等新格式
	currency := c.Query("currency", "USDT_TRC20")
	// 获取金额参数，默认 10.00
	actualAmountNum := c.QueryFloat("amount", 10.00)
	// 格式化为两位小数字符串，确保签名一致性
	actualAmount := fmt.Sprintf("%.2f", actualAmountNum)

	// 构建请求数据 (按照 TokenPay 文档)
	testReq := map[string]interface{}{
		"OutOrderId":   orderID,                           // 外部订单号
		"OrderUserKey": testUserKey,                       // 支付用户标识 (必需!)
		"ActualAmount": actualAmount,                      // 法币金额 (保留两位小数字符串)
		"Currency":     currency,                          // 币种: USDT_TRC20, USDT_ERC20, EVM_ETH_USDT_ERC20 等
		"NotifyUrl":    tokenPayCfg.NotifyURL,             // 异步通知URL
		"RedirectUrl":  "https://gg.ppnet55.com/wallet", // 支付完成后跳转URL
	}

	// 生成签名
	signature := helpers.GenerateTokenPaySignature(testReq, tokenPayCfg.SecretKey)
	testReq["Signature"] = signature

	fmt.Println("[TokenPay Test] 请求数据:", testReq)
	fmt.Println("[TokenPay Test] BaseURL:", tokenPayCfg.BaseURL)
	fmt.Println("[TokenPay Test] SecretKey:", tokenPayCfg.SecretKey)

	// 发送请求到 TokenPay
	jsonBody, _ := stdjson.Marshal(testReq)
	apiURL := tokenPayCfg.BaseURL + "/CreateOrder"

	httpReq, err := http.NewRequest("POST", apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "创建请求失败: " + err.Error()})
	}

	httpReq.Header.Set("Content-Type", "application/json")
	// TokenPay 不需要 Authorization header，使用 Signature 验证

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "请求失败: " + err.Error()})
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	fmt.Println("[TokenPay Test] 状态码:", resp.StatusCode)
	fmt.Println("[TokenPay Test] 响应:", string(respBody))

	// 解析响应
	var result map[string]interface{}
	if err := stdjson.Unmarshal(respBody, &result); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":  "解析响应失败",
			"raw":    string(respBody),
			"status": resp.StatusCode,
		})
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"status":   resp.StatusCode,
		"request":  testReq,
		"response": result,
		"raw":      string(respBody),
	})
}

// TestTokenPayGetConfig 测试获取 TokenPay 配置
func TestTokenPayGetConfig(c *fiber.Ctx) error {
	cfg := helpers.GetCfgInstance()
	if cfg == nil {
		return c.Status(500).JSON(fiber.Map{"error": "配置未加载"})
	}

	tokenPayCfg := cfg.Conf.Payment.TokenPay

	return c.JSON(fiber.Map{
		"base_url":   tokenPayCfg.BaseURL,
		"secret_key": tokenPayCfg.SecretKey,
		"notify_url": tokenPayCfg.NotifyURL,
	})
}
