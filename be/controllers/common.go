package controllers

import (
	"net"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gogogo/models"
	"gogogo/models/dtos"
	"gogogo/services"
)

func QueryRequestDomain(c *fiber.Ctx) string {
	for _, raw := range []string{
		c.Get("Origin"),
		c.Get("Referer"),
		c.Get("X-Forwarded-Host"),
		c.Get("X-Original-Host"),
		c.Hostname(),
	} {
		if domain := normalizeRequestDomain(raw); domain != "" {
			return domain
		}
	}
	return ""
}

func NormalizeRequestDomain(raw string) string {
	return normalizeRequestDomain(raw)
}

func normalizeRequestDomain(raw string) string {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return ""
	}

	parseTarget := raw
	if !strings.Contains(parseTarget, "://") {
		parseTarget = "//" + parseTarget
	}
	parsed, err := url.Parse(parseTarget)
	if err != nil {
		return ""
	}

	hostname := strings.TrimSuffix(strings.TrimSpace(parsed.Hostname()), ".")
	hostname = strings.TrimPrefix(hostname, "www.")
	if hostname == "" || strings.ContainsAny(hostname, " /\\") {
		return ""
	}
	return hostname
}

func QueryUserIdFromJwt(c *fiber.Ctx) int {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	uid, ok := claims["uid"].(float64)
	if ok {
		return int(uid)
	}
	return -1
}

func QueryUserNameFromJwt(c *fiber.Ctx) string {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	username, ok := claims["username"].(string)
	if ok {
		return username
	}
	return ""
}

func QueryUserIsAdmin(c *fiber.Ctx) bool {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)

	if isAdmin, ok := claims["admin"].(bool); ok && isAdmin {
		return true
	}

	if username, ok := claims["username"].(string); ok && username == "admin_root" {
		return true
	}

	uid, ok := claims["uid"].(float64)
	if !ok {
		return false
	}

	db := models.GetInstance().DbInstance
	if db == nil {
		return false
	}

	var userRecord dtos.User
	if err := db.Select("id, username, vip_level").First(&userRecord, uint64(uid)).Error; err != nil {
		return false
	}

	return userRecord.Username == "admin_root" || userRecord.VipLevel >= 99
}

func QueryClientIP(c *fiber.Ctx) string {
	var firstValid string

	for _, header := range []string{
		"CF-Connecting-IP",
		"True-Client-IP",
		"X-Forwarded-For",
		"X-Original-Forwarded-For",
		"X-Real-IP",
		"X-Client-IP",
		"X-Cluster-Client-IP",
		"Forwarded",
	} {
		for _, candidate := range extractClientIPCandidates(c.Get(header)) {
			if isPublicIP(candidate) {
				return candidate
			}
			if firstValid == "" {
				firstValid = candidate
			}
		}
	}

	if firstValid != "" {
		return firstValid
	}

	return normalizeSingleClientIP(c.IP())
}

func buildFBEventContext(c *fiber.Ctx) services.EventContext {
	fbc := c.Cookies("_fbc")
	if fbc == "" {
		fbc = c.Get("X-FB-Click-Id")
	}
	fbp := c.Cookies("_fbp")
	if fbp == "" {
		fbp = c.Get("X-FB-Browser-Id")
	}

	return services.EventContext{
		ClientIP:  QueryClientIP(c),
		UserAgent: c.Get("User-Agent"),
		FBC:       fbc,
		FBP:       fbp,
		SourceURL: c.Get("Referer"),
	}
}

func extractClientIPCandidates(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	candidates := make([]string, 0, len(parts))
	for _, part := range parts {
		if ip := normalizeSingleClientIP(part); ip != "" {
			candidates = append(candidates, ip)
		}
	}
	return candidates
}

func normalizeSingleClientIP(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	raw = strings.Trim(raw, "\"'")
	if len(raw) >= 4 && strings.EqualFold(raw[:4], "for=") {
		raw = strings.TrimSpace(raw[4:])
		raw = strings.Trim(raw, "\"'")
	}
	raw = strings.TrimPrefix(raw, "[")
	raw = strings.TrimSuffix(raw, "]")

	if ip := net.ParseIP(raw); ip != nil {
		return ip.String()
	}

	if host, _, err := net.SplitHostPort(raw); err == nil {
		host = strings.Trim(host, "[]")
		if ip := net.ParseIP(host); ip != nil {
			return ip.String()
		}
	}

	return ""
}

func isPublicIP(raw string) bool {
	ip := net.ParseIP(strings.TrimSpace(raw))
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() {
		return false
	}
	return true
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}
