package controllers

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func QueryUserIdFromJwt(c *fiber.Ctx) int {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	uid, ok := claims["uid"].(float64)
	if ok {
		return int(uid)
	} else {
		return -1
	}
}

func QueryUserNameFromJwt(c *fiber.Ctx) string {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	u, ok := claims["username"].(string)
	if ok {
		return u
	} else {
		return ""
	}
}

// ParseTokenManual 手动解析 Token 字符串获取 UserID
func ParseTokenManual(tokenString, GlobalSecret string) (uint64, error) {
	if tokenString == "" {
		return 0, errors.New("token is required")
	}

	// 1. 解析 Token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(GlobalSecret), nil
	})

	if err != nil {
		return 0, fmt.Errorf("invalid token: %v", err)
	}

	// 2. 验证并获取 Claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// 3. 提取 user_id (注意：JWT 中的数字通常解析为 float64)
		if uidFloat, ok := claims["user_id"].(float64); ok {
			return uint64(uidFloat), nil
		}
		// 兼容 uid 字段名
		if uidFloat, ok := claims["uid"].(float64); ok {
			return uint64(uidFloat), nil
		}
		return 0, errors.New("user_id claim not found")
	}

	return 0, errors.New("invalid token claims")
}
