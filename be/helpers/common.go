package helpers

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// GenerateTokenPaySignature 生成 TokenPay 签名
// 按照ASCII排序后拼接参数，末尾拼接密钥，计算MD5
func GenerateTokenPaySignature(params map[string]interface{}, secretKey string) string {
	// 获取所有key并排序
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 按ASCII排序拼接参数
	var parts []string
	for _, k := range keys {
		v := params[k]
		// 忽略空值
		if v == nil {
			continue
		}
		switch val := v.(type) {
		case string:
			if val == "" {
				continue
			}
			parts = append(parts, fmt.Sprintf("%s=%s", k, val))
		case float64:
			parts = append(parts, fmt.Sprintf("%s=%s", k, formatFloat(val)))
		case int:
			parts = append(parts, fmt.Sprintf("%s=%d", k, val))
		case int64:
			parts = append(parts, fmt.Sprintf("%s=%d", k, val))
		default:
			str := fmt.Sprintf("%v", v)
			if str != "" {
				parts = append(parts, fmt.Sprintf("%s=%s", k, str))
			}
		}
	}

	// 拼接字符串并附加密钥
	signStr := strings.Join(parts, "&") + secretKey

	// 计算MD5
	hash := md5.Sum([]byte(signStr))
	return hex.EncodeToString(hash[:])
}

// formatFloat 格式化浮点数，去除末尾的0
func formatFloat(f float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", f), "0"), ".")
}

// VerifyTokenPayCallbackSignature 验证 TokenPay 回调签名
// 将除Signature字段外的所有字段，按照字母升序排序，忽略没有值的字段
// 按顺序拼接为key1=value1&key2=value2形式，然后在末尾拼接上异步通知密钥，计算MD5
func VerifyTokenPayCallbackSignature(body map[string]interface{}, signature string, secretKey string) bool {
	// 获取所有key并排序
	var keys []string
	for k := range body {
		// 跳过 Signature 字段
		if k == "Signature" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 按ASCII排序拼接参数
	var parts []string
	for _, k := range keys {
		v := body[k]
		// 忽略空值
		if v == nil {
			continue
		}
		switch val := v.(type) {
		case string:
			if val == "" {
				continue
			}
			parts = append(parts, fmt.Sprintf("%s=%s", k, val))
		case float64:
			// 如果是整数，不输出小数部分
			if val == float64(int64(val)) {
				parts = append(parts, fmt.Sprintf("%s=%d", k, int64(val)))
			} else {
				parts = append(parts, fmt.Sprintf("%s=%s", k, formatFloat(val)))
			}
		case int:
			parts = append(parts, fmt.Sprintf("%s=%d", k, val))
		case int64:
			parts = append(parts, fmt.Sprintf("%s=%d", k, val))
		default:
			str := fmt.Sprintf("%v", v)
			if str != "" && str != "<nil>" {
				parts = append(parts, fmt.Sprintf("%s=%s", k, str))
			}
		}
	}

	// 拼接字符串并附加密钥
	signStr := strings.Join(parts, "&") + secretKey

	// 计算MD5
	hash := md5.Sum([]byte(signStr))
	calculatedSign := hex.EncodeToString(hash[:])

	return calculatedSign == signature
}

// fiber返回错误函数
func ErrorMsgReturn(c *fiber.Ctx, msg string, code int) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"success": false,
		"message": msg,
		"code":    code,
	})
}
