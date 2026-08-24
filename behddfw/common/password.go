package common

import (
	"crypto/md5"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword 对密码进行加密
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash 验证密码是否正确
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateJWT 生成 JWT Token
func GenerateJWT(userId uint64, username string, jwtsecret string) (string, error) {
	claims := jwt.MapClaims{
		"uid":      userId,
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // 24小时过期
		"iss":      "game-center",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtsecret))
}

func GenerateUniqueCode() string {
	// uuid.New().String() 生成格式如: "f47ac10b-58cc-4372-a567-0e02b2c3d479" (36位)
	// 去掉 "-" 后正好是 32 位
	raw := uuid.New().String()
	return strings.ReplaceAll(raw, "-", "")
}

// generateGamePassword 根据用户ID生成固定的第三方游戏密码
// 逻辑: MD5(UserID + 固定盐值) -> 取前16位或32位
func GenerateGamePassword(userID uint64) string {
	salt := "My_Secret_Game_Salt_2025" // 请修改为复杂的密钥
	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%d%s", userID, salt))
	return fmt.Sprintf("%x", h.Sum(nil)) // 返回 32位 hex 字符串
}
