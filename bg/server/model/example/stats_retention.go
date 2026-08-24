package example

import (
	"time"
)

type StatsRetention struct {
	Id               *int       `json:"id" form:"id" gorm:"primarykey;column:id;size:20;"`
	StatDate         *time.Time `json:"statDate" form:"statDate" gorm:"comment:stat date;type:date;column:stat_date;"`
	NewUsers         *int       `json:"newUsers" form:"newUsers" gorm:"comment:new users;column:new_users;size:10;"`
	DepositUsers     *int       `json:"depositUsers" form:"depositUsers" gorm:"comment:deposit users;column:deposit_users;size:10;"`
	DepositAmount    *float64   `json:"depositAmount" form:"depositAmount" gorm:"comment:deposit amount;column:deposit_amount;size:15;"`
	GameWinloseUsers *int       `json:"gameWinloseUsers" form:"gameWinloseUsers" gorm:"comment:game winlose users;column:game_winlose_users;size:10;"`
	Retention1       *int       `json:"retention1" form:"retention1" gorm:"comment:retention day 1;column:retention_1;size:10;"`
	Retention3       *int       `json:"retention3" form:"retention3" gorm:"comment:retention day 3;column:retention_3;size:10;"`
	Retention7       *int       `json:"retention7" form:"retention7" gorm:"comment:retention day 7;column:retention_7;size:10;"`
	Retention30      *int       `json:"retention30" form:"retention30" gorm:"comment:retention day 30;column:retention_30;size:10;"`
	CreatedAt        *time.Time `json:"createdAt" form:"createdAt" gorm:"column:created_at;"`
	UpdatedAt        *time.Time `json:"updatedAt" form:"updatedAt" gorm:"column:updated_at;"`
}

func (StatsRetention) TableName() string {
	return "stats_retention"
}
