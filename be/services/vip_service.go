package services

import (
	"context"
	"gogogo/models"
	"gogogo/models/dtos"

	"gorm.io/gorm"
)

type vipTierRule struct {
	Level       int
	MinDepositU float64
	MinWagerU   float64
	RebateRate  float64
}

var vipTierRules = []vipTierRule{
	{Level: 1, MinDepositU: 100, MinWagerU: 1000, RebateRate: 0.05},
	{Level: 2, MinDepositU: 1000, MinWagerU: 10000, RebateRate: 0.15},
	{Level: 3, MinDepositU: 5000, MinWagerU: 50000, RebateRate: 0.30},
	{Level: 4, MinDepositU: 10000, MinWagerU: 100000, RebateRate: 0.45},
	{Level: 5, MinDepositU: 50000, MinWagerU: 500000, RebateRate: 0.60},
	{Level: 6, MinDepositU: 100000, MinWagerU: 1000000, RebateRate: 1.00},
}

func ResolveVIPLevel(totalDepositU float64, totalWagerU float64) int {
	level := 0
	for _, rule := range vipTierRules {
		if totalDepositU >= rule.MinDepositU && totalWagerU >= rule.MinWagerU {
			level = rule.Level
		}
	}

	return level
}

func GetVIPTierRules() []vipTierRule {
	cloned := make([]vipTierRule, len(vipTierRules))
	copy(cloned, vipTierRules)
	return cloned
}

func RecalculateAndPersistUserVIPLevel(ctx context.Context, userID uint64) (int, error) {
	db := models.GetInstance().DbInstance.WithContext(ctx)

	var user dtos.User
	if err := db.Select("id, vip_level, total_deposit").First(&user, userID).Error; err != nil {
		return 0, err
	}

	var totalStat dtos.UserGameTransactionStat
	totalWagerU := 0.0
	if err := db.
		Where("user_id = ? AND period_type = ? AND period_key = ?", userID, "total", "all").
		First(&totalStat).Error; err == nil {
		totalWagerU = totalStat.Turnover
		if totalWagerU <= 0 {
			totalWagerU = totalStat.Bet
		}
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	}

	nextLevel := ResolveVIPLevel(user.TotalDeposit, totalWagerU)
	if nextLevel != user.VipLevel {
		if err := db.Model(&dtos.User{}).Where("id = ?", userID).Update("vip_level", nextLevel).Error; err != nil {
			return 0, err
		}
	}

	return nextLevel, nil
}
