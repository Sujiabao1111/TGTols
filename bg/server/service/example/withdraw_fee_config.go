package example

import (
	"context"
	"errors"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/example"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WithdrawFeeConfigService struct{}

const (
	defaultWithdrawFeeConfigKey    = "default"
	defaultWithdrawPlatformFeeRate = 0.003
)

type withdrawFeeConfigValues struct {
	Enabled bool
	FeeRate float64
}

func normalizeWithdrawFeeConfig(config example.WithdrawFeeConfig) example.WithdrawFeeConfig {
	if config.ConfigKey == "" {
		config.ConfigKey = defaultWithdrawFeeConfigKey
	}
	if config.FeeRate < 0 || config.FeeRate > 1 {
		config.FeeRate = defaultWithdrawPlatformFeeRate
	}
	return config
}

func withdrawFeeValues(config example.WithdrawFeeConfig) withdrawFeeConfigValues {
	config = normalizeWithdrawFeeConfig(config)
	return withdrawFeeConfigValues{
		Enabled: config.Enabled,
		FeeRate: config.FeeRate,
	}
}

func getWithdrawFeeConfig(ctx context.Context, db *gorm.DB) example.WithdrawFeeConfig {
	if db == nil {
		db = global.GVA_DB
	}
	if db == nil {
		return normalizeWithdrawFeeConfig(example.WithdrawFeeConfig{
			Enabled: true,
			FeeRate: defaultWithdrawPlatformFeeRate,
		})
	}
	if ctx == nil {
		ctx = context.Background()
	}

	var config example.WithdrawFeeConfig
	err := db.WithContext(ctx).Where("config_key = ?", defaultWithdrawFeeConfigKey).Take(&config).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		config = normalizeWithdrawFeeConfig(example.WithdrawFeeConfig{
			Enabled: true,
			FeeRate: defaultWithdrawPlatformFeeRate,
		})
		_ = db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "config_key"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"enabled":  config.Enabled,
				"fee_rate": config.FeeRate,
			}),
		}).Table("withdraw_fee_configs").Create(map[string]interface{}{
			"config_key": config.ConfigKey,
			"enabled":    config.Enabled,
			"fee_rate":   config.FeeRate,
		}).Error
		return config
	}
	if err != nil {
		return normalizeWithdrawFeeConfig(example.WithdrawFeeConfig{
			Enabled: true,
			FeeRate: defaultWithdrawPlatformFeeRate,
		})
	}
	return normalizeWithdrawFeeConfig(config)
}

func (service *WithdrawFeeConfigService) GetWithdrawFeeConfig(ctx context.Context) (example.WithdrawFeeConfig, error) {
	return getWithdrawFeeConfig(ctx, global.GVA_DB), nil
}

func (service *WithdrawFeeConfigService) UpdateWithdrawFeeConfig(ctx context.Context, enabled bool, feeRate float64) (example.WithdrawFeeConfig, error) {
	if feeRate < 0 || feeRate > 1 {
		return example.WithdrawFeeConfig{}, errors.New("withdraw fee rate must be between 0 and 1")
	}

	config := example.WithdrawFeeConfig{
		ConfigKey: defaultWithdrawFeeConfigKey,
		Enabled:   enabled,
		FeeRate:   feeRate,
	}

	err := global.GVA_DB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "config_key"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"enabled":  config.Enabled,
			"fee_rate": config.FeeRate,
		}),
	}).Table("withdraw_fee_configs").Create(map[string]interface{}{
		"config_key": config.ConfigKey,
		"enabled":    config.Enabled,
		"fee_rate":   config.FeeRate,
	}).Error
	if err != nil {
		return example.WithdrawFeeConfig{}, err
	}

	return getWithdrawFeeConfig(ctx, global.GVA_DB), nil
}
