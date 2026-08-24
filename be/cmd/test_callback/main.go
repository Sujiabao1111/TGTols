package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

func main() {
	// 默认配置
	callbackURL := "http://localhost:3555/payments/notify"
	secretKey := "150kz9esh1s2f793abb8l0e4mjot3r2c"
	merchantID := "10036"

	// 默认参数
	orderID := "ORD2026020411433885740002"
	amount := "50000.00"
	status := "1"
	platOrderID := fmt.Sprintf("PLAT%d", time.Now().Unix())
	refCode := "0"
	refMsg := "SUCCESS"

	// 解析命令行参数
	// 用法: test_callback.exe [order_id] [amount] [status] [callback_url]
	if len(os.Args) > 1 {
		orderID = os.Args[1]
	}
	if len(os.Args) > 2 {
		amount = os.Args[2]
	}
	if len(os.Args) > 3 {
		status = os.Args[3]
		if status == "2" {
			refMsg = "FAILED"
		}
	}
	if len(os.Args) > 4 {
		callbackURL = os.Args[4]
	}

	// 从环境变量读取（优先级高于默认值，低于命令行参数）
	if v := os.Getenv("SECRET_KEY"); v != "" {
		secretKey = v
	}
	if v := os.Getenv("MERCHANT_ID"); v != "" {
		merchantID = v
	}
	if v := os.Getenv("PLAT_ORDER_ID"); v != "" {
		platOrderID = v
	}

	// 构建参数
	params := map[string]string{
		"merchant_id":   merchantID,
		"order_id":      orderID,
		"plat_order_id": platOrderID,
		"amount":        amount,
		"status":        status,
		"ref_code":      refCode,
		"ref_msg":       refMsg,
	}

	// 生成签名
	sign := generateSign(params, secretKey)
	params["sign"] = sign

	fmt.Println("========================================")
	fmt.Println("🧪 支付回调测试")
	fmt.Println("========================================")
	fmt.Printf("回调URL: %s\n", callbackURL)
	fmt.Println("\n请求参数:")
	for k, v := range params {
		fmt.Printf("  %s: %s\n", k, v)
	}
	fmt.Println("\n签名原串:")
	fmt.Printf("  %s\n", buildSignString(params, secretKey))
	fmt.Printf("\nMD5签名: %s\n", sign)
	fmt.Println("\n----------------------------------------")
	fmt.Println("📤 发送请求...")
	fmt.Println()

	// 发送请求
	resp, err := sendCallback(callbackURL, params)
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 响应结果: %s\n", resp)

	if resp == "OK" {
		fmt.Println("\n🎉 回调测试成功！")
	} else {
		fmt.Println("\n❌ 回调测试失败")
	}
}

// generateSign 生成MD5签名
func generateSign(params map[string]string, secret string) string {
	signStr := buildSignString(params, secret)
	hash := md5.Sum([]byte(signStr))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

// buildSignString 构建签名原串
func buildSignString(params map[string]string, secret string) string {
	// 获取所有key并排序
	var keys []string
	for k := range params {
		if k != "sign" && params[k] != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// 拼接参数
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
	}

	// 拼接成字符串并追加key
	return strings.Join(parts, "&") + "&key=" + secret
}

// sendCallback 发送回调请求
func sendCallback(callbackURL string, params map[string]string) (string, error) {
	data := url.Values{}
	for k, v := range params {
		data.Set(k, v)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.PostForm(callbackURL, data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 读取响应
	buf := make([]byte, 1024)
	n, err := resp.Body.Read(buf)
	if err != nil && err.Error() != "EOF" {
		return "", err
	}

	return string(buf[:n]), nil
}
