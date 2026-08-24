package common

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net/url"
	"sort"
	"strings"
	"time"
)

// GenerateSign builds the legacy gateway MD5 signature in uppercase.
func GenerateSign(params map[string]string, secret string) string {
	var keys []string
	for k, v := range params {
		if k == "sign" || isSignEmptyValue(v) {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
	}

	signStr := strings.Join(parts, "&") + "&key=" + secret
	hash := md5.Sum([]byte(signStr))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

// VerifySign verifies the legacy gateway signature.
func VerifySign(params map[string]string, secret string, sign string) bool {
	expectedSign := GenerateSign(params, secret)
	return strings.EqualFold(expectedSign, sign)
}

func isSignEmptyValue(v string) bool {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "", "0", "null", "false":
		return true
	default:
		return false
	}
}

// BuildFormData builds a query-style string for debugging or form posts.
func BuildFormData(params map[string]string) string {
	var parts []string
	for k, v := range params {
		if !isSignEmptyValue(v) {
			parts = append(parts, fmt.Sprintf("%s=%s", k, url.QueryEscape(v)))
		}
	}
	return strings.Join(parts, "&")
}

// StructToMap is a placeholder for future struct-tag-based mapping.
func StructToMap(data interface{}) map[string]string {
	return nil
}

func GenerateOrderID(userID uint64) string {
	timestamp := TimeFormatYYYYMMDDHHMMSS()
	random := RandomNumber(1000, 9999)
	userPart := userID % 10000
	return fmt.Sprintf("ORD%s%04d%04d", timestamp, random, userPart)
}

func GenerateWithdrawOrderID(userID uint64) string {
	timestamp := TimeFormatYYYYMMDDHHMMSS()
	random := RandomNumber(1000, 9999)
	userPart := userID % 10000
	return fmt.Sprintf("WDR%s%04d%04d", timestamp, random, userPart)
}

func TimeFormatYYYYMMDDHHMMSS() string {
	return TimeNow().Format("20060102150405")
}

func TimeNow() time.Time {
	return time.Now()
}

func RandomNumber(min, max int) int {
	return rand.Intn(max-min) + min
}
