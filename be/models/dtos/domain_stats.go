package dtos

import "time"

// DomainDailyStat stores aggregated visits for one domain and calendar day.
type DomainDailyStat struct {
	ID      uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Domain  string    `gorm:"size:255;not null;uniqueIndex:uk_domain_date" json:"domain"`
	StatDate string   `gorm:"type:date;not null;uniqueIndex:uk_domain_date" json:"stat_date"`
	Clicks  int64     `gorm:"not null;default:0" json:"clicks"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
