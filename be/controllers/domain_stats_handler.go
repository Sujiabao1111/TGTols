package controllers

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gogogo/models"
	"gorm.io/gorm"
)

func RecordDomainClick(c *fiber.Ctx) error {
	domain := NormalizeRequestDomain(c.FormValue("domain"))
	if domain == "" { domain = QueryRequestDomain(c) }
	if domain == "" { return c.Status(400).JSON(fiber.Map{"message": "domain is required"}) }
	date := time.Now().Format("2006-01-02")
	db := models.GetInstance().DbInstance
	result := incrementDomainClick(db, domain, date)
	if result.Error != nil { return c.Status(500).JSON(fiber.Map{"message": "failed to record click"}) }
	return c.JSON(fiber.Map{"success": true})
}

func incrementDomainClick(db *gorm.DB, domain, date string) *gorm.DB {
	return db.Exec("INSERT INTO domain_daily_stats (domain, stat_date, clicks, created_at, updated_at) VALUES (?, ?, 1, NOW(), NOW()) ON DUPLICATE KEY UPDATE clicks = clicks + 1, updated_at = NOW()", domain, date)
}

func GetDomainStats(c *fiber.Ctx) error {
	if !QueryUserIsAdmin(c) { return c.Status(403).JSON(fiber.Map{"message": "administrator access required"}) }
	from, to := c.Query("from"), c.Query("to")
	if from == "" { from = to }; if to == "" { to = from }
	if from == "" { from = time.Now().Format("2006-01-02") }; if to == "" { to = from }
	domain := NormalizeRequestDomain(c.Query("domain"))
	var rows []struct { Domain string `json:"domain"`; StatDate string `json:"date"`; Clicks int64 `json:"clicks"`; Registrations int64 `json:"registrations"` }
	q := models.GetInstance().DbInstance.Table("domain_daily_stats s").Select("s.domain, s.stat_date, s.clicks, COUNT(u.id) registrations").Joins("LEFT JOIN users u ON u.register_domain = s.domain AND DATE(u.created_at) = s.stat_date").Where("s.stat_date BETWEEN ? AND ?", from, to).Group("s.domain, s.stat_date, s.clicks").Order("s.stat_date DESC, s.domain")
	if domain != "" { q = q.Where("s.domain = ?", domain) }
	if err := q.Scan(&rows).Error; err != nil { return c.Status(500).JSON(fiber.Map{"message": "failed to query domain stats"}) }
	return c.JSON(fiber.Map{"from": from, "to": to, "domain": strings.TrimSpace(domain), "items": rows})
}
