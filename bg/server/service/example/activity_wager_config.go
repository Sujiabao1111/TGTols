package example

import (
	"context"
	"errors"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/example"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ActivityWagerConfigService struct{}

const (
	defaultActivityWagerConfigKey = "default"
	defaultDepositWagerMultiplier = 2.0
	defaultRewardWagerMultiplier  = 20.0
)

func normalizeActivityWagerConfig(config example.ActivityWagerConfig) example.ActivityWagerConfig {
	if config.ConfigKey == "" {
		config.ConfigKey = defaultActivityWagerConfigKey
	}
	if config.DepositWagerMultiplier <= 0 {
		config.DepositWagerMultiplier = defaultDepositWagerMultiplier
	}
	if config.RewardWagerMultiplier <= 0 {
		config.RewardWagerMultiplier = defaultRewardWagerMultiplier
	}
	return config
}

func getActivityWagerConfig(ctx context.Context, db *gorm.DB) example.ActivityWagerConfig {
	if db == nil {
		db = global.GVA_DB
	}
	if db == nil {
		return normalizeActivityWagerConfig(example.ActivityWagerConfig{})
	}
	if ctx == nil {
		ctx = context.Background()
	}

	var config example.ActivityWagerConfig
	err := db.WithContext(ctx).Where("config_key = ?", defaultActivityWagerConfigKey).Take(&config).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		config = normalizeActivityWagerConfig(config)
		_ = db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "config_key"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"deposit_wager_multiplier": config.DepositWagerMultiplier,
				"reward_wager_multiplier":  config.RewardWagerMultiplier,
			}),
		}).Create(&config).Error
		return config
	}
	if err != nil {
		return normalizeActivityWagerConfig(example.ActivityWagerConfig{})
	}
	return normalizeActivityWagerConfig(config)
}

func (service *ActivityWagerConfigService) GetActivityWagerConfig(ctx context.Context) (example.ActivityWagerConfig, error) {
	config := getActivityWagerConfig(ctx, global.GVA_DB)
	return config, nil
}

func (service *ActivityWagerConfigService) UpdateActivityWagerConfig(ctx context.Context, depositMultiplier float64, rewardMultiplier float64) (example.ActivityWagerConfig, error) {
	if depositMultiplier <= 0 || rewardMultiplier <= 0 {
		return example.ActivityWagerConfig{}, errors.New("wager multipliers must be greater than 0")
	}

	config := example.ActivityWagerConfig{
		ConfigKey:              defaultActivityWagerConfigKey,
		DepositWagerMultiplier: depositMultiplier,
		RewardWagerMultiplier:  rewardMultiplier,
	}

	err := global.GVA_DB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "config_key"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"deposit_wager_multiplier": config.DepositWagerMultiplier,
			"reward_wager_multiplier":  config.RewardWagerMultiplier,
		}),
	}).Create(&config).Error
	if err != nil {
		return example.ActivityWagerConfig{}, err
	}

	return getActivityWagerConfig(ctx, global.GVA_DB), nil
}
