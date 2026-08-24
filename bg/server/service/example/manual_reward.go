package example

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	model "github.com/flipped-aurora/gin-vue-admin/server/model/example"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ManualRewardService struct{}

const desktopInsuranceCouponType = "add_desktop_insurance"
const desktopInsuranceCouponValue = 0.2

var errDesktopRewardAlreadyGranted = errors.New("\u8be5\u73a9\u5bb6ID\u5df2\u53d1\u653e\u8fc7\u684c\u9762\u8865\u507f\u5238")

type rewardUser struct {
	ID          uint64  `gorm:"column:id"`
	Username    string  `gorm:"column:username"`
	Balance     float64 `gorm:"column:balance"`
	RegisterIP  string  `gorm:"column:register_ip"`
	LastLoginIP string  `gorm:"column:last_login_ip"`
}

type rewardOrder struct {
	DstCode string `gorm:"column:dst_code"`
}

func ensureDesktopRewardClaimIPColumn(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	if !manualRewardColumnExists(db, "desktop_reward_claims", "claim_ip") {
		if err := db.Exec("ALTER TABLE desktop_reward_claims ADD COLUMN claim_ip VARCHAR(64) NULL DEFAULT NULL").Error; err != nil {
			return err
		}
	}
	if !manualRewardIndexExists(db, "desktop_reward_claims", "uk_desktop_reward_claim_ip") {
		return db.Exec("ALTER TABLE desktop_reward_claims ADD UNIQUE KEY uk_desktop_reward_claim_ip (claim_ip)").Error
	}
	return nil
}

func manualRewardUserSelectSQL(db *gorm.DB) string {
	parts := []string{"id", "username", "balance"}
	if manualRewardColumnExists(db, "users", "register_ip") {
		parts = append(parts, "COALESCE(register_ip, '') AS register_ip")
	} else {
		parts = append(parts, "'' AS register_ip")
	}
	if manualRewardColumnExists(db, "users", "last_login_ip") {
		parts = append(parts, "COALESCE(last_login_ip, '') AS last_login_ip")
	} else {
		parts = append(parts, "'' AS last_login_ip")
	}
	return strings.Join(parts, ", ")
}

func manualRewardColumnExists(db *gorm.DB, tableName string, columnName string) bool {
	if db == nil {
		return false
	}

	var count int64
	err := db.Table("INFORMATION_SCHEMA.COLUMNS").
		Where("TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?", tableName, columnName).
		Count(&count).Error
	return err == nil && count > 0
}

func manualRewardTableExists(db *gorm.DB, tableName string) bool {
	if db == nil {
		return false
	}
	return db.Migrator().HasTable(tableName)
}

func manualRewardIndexExists(db *gorm.DB, tableName string, indexName string) bool {
	if db == nil {
		return false
	}

	var count int64
	err := db.Table("INFORMATION_SCHEMA.STATISTICS").
		Where("TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND INDEX_NAME = ?", tableName, indexName).
		Count(&count).Error
	return err == nil && count > 0
}

func manualRewardUserIPParts(db *gorm.DB, userAlias string) []string {
	prefix := ""
	if strings.TrimSpace(userAlias) != "" {
		prefix = strings.TrimSpace(userAlias) + "."
	}

	parts := make([]string, 0, 2)
	if manualRewardColumnExists(db, "users", "register_ip") {
		parts = append(parts, "NULLIF("+prefix+"register_ip, '')")
	}
	if manualRewardColumnExists(db, "users", "last_login_ip") {
		parts = append(parts, "NULLIF("+prefix+"last_login_ip, '')")
	}
	return parts
}

func manualRewardClaimIPSQL(db *gorm.DB, claimTableName string, claimAlias string, userAlias string) string {
	parts := make([]string, 0, 3)
	claimPrefix := strings.TrimSpace(claimAlias)
	if claimPrefix != "" {
		claimPrefix += "."
	}
	if manualRewardColumnExists(db, claimTableName, "claim_ip") {
		parts = append(parts, "NULLIF("+claimPrefix+"claim_ip, '')")
	}
	parts = append(parts, manualRewardUserIPParts(db, userAlias)...)
	if len(parts) == 0 {
		return "''"
	}
	parts = append(parts, "''")
	return "COALESCE(" + strings.Join(parts, ", ") + ")"
}

func firstNonEmptyManualRewardString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func normalizeManualRewardIP(value string) string {
	return strings.TrimSpace(value)
}

func (manualRewardService *ManualRewardService) GrantDesktopReward(ctx context.Context, userID uint64) (*model.ManualRewardResult, error) {
	return manualRewardService.grantDesktopInsuranceCoupon(ctx, userID)
}

func (manualRewardService *ManualRewardService) GrantReward(ctx context.Context, userID uint64, rewardAmountU float64) (*model.ManualRewardResult, error) {
	return manualRewardService.grantReward(ctx, userID, rewardAmountU, "manual_reward")
}

func (manualRewardService *ManualRewardService) grantDesktopInsuranceCoupon(ctx context.Context, userID uint64) (*model.ManualRewardResult, error) {
	_ = ensureDesktopRewardClaimIPColumn(global.GVA_DB)

	result := &model.ManualRewardResult{
		UserID:         userID,
		RewardType:     "insurance_coupon",
		RewardCategory: desktopInsuranceCouponType,
	}

	err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user rewardUser
		if err := tx.Table("users").
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Select(manualRewardUserSelectSQL(tx)).
			Where("id = ?", userID).
			Take(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("user %d not found", userID)
			}
			return err
		}

		claimIP := normalizeManualRewardIP(firstNonEmptyManualRewardString(user.RegisterIP, user.LastLoginIP))
		alreadyGranted, err := hasDesktopRewardGrant(tx.WithContext(ctx), userID)
		if err != nil {
			return err
		}
		if alreadyGranted {
			return errDesktopRewardAlreadyGranted
		}
		if claimIP != "" {
			ipAlreadyGranted, err := hasDesktopRewardGrantByIP(tx.WithContext(ctx), userID, claimIP)
			if err != nil {
				return err
			}
			if ipAlreadyGranted {
				return fmt.Errorf("\u8be5IP\u5df2\u53d1\u653e\u8fc7\u684c\u9762\u8865\u507f\u5238")
			}
		}

		now := time.Now()
		claim := &model.DesktopRewardClaim{
			UserID:    userID,
			ClaimedAt: now,
		}
		if claimIP != "" {
			claim.ClaimIP = &claimIP
		}
		if err := tx.Create(claim).Error; err != nil {
			return err
		}

		couponCode := buildDesktopInsuranceCouponCode(userID)
		coupon := map[string]interface{}{
			"user_id":       userID,
			"coupon_code":   couponCode,
			"activity_type": desktopInsuranceCouponType,
			"coupon_value":  desktopInsuranceCouponValue,
			"min_deposit":   0,
			"status":        1,
			"valid_start":   now,
			"valid_end":     now.AddDate(10, 0, 0),
			"activated_at":  now,
			"created_at":    now,
			"updated_at":    now,
		}
		if err := tx.Table("user_coupons").Create(coupon).Error; err != nil {
			return err
		}

		var couponID uint64
		_ = tx.Table("user_coupons").Select("id").Where("coupon_code = ?", couponCode).Scan(&couponID).Error

		result.Username = user.Username
		result.CouponID = couponID
		result.CouponCode = couponCode
		result.ReferenceID = couponCode
		result.BeforeBalance = user.Balance
		result.AfterBalance = user.Balance
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (manualRewardService *ManualRewardService) grantReward(ctx context.Context, userID uint64, rewardAmountU float64, rewardCategory string) (*model.ManualRewardResult, error) {
	if rewardCategory == "desktop_reward" {
		return nil, errors.New("desktop reward now grants an insurance coupon")
	}
	if rewardAmountU <= 0 {
		return nil, errors.New("reward amount must be greater than 0")
	}

	result := &model.ManualRewardResult{
		UserID:         userID,
		RewardAmountU:  rewardAmountU,
		RewardCategory: rewardCategory,
	}

	err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		wagerConfig := getActivityWagerConfig(ctx, tx)
		result.WagerMultiplier = wagerConfig.RewardWagerMultiplier
		result.WagerRequired = roundRewardAmount(rewardAmountU * wagerConfig.RewardWagerMultiplier)

		var user rewardUser
		if err := tx.Table("users").
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Select(manualRewardUserSelectSQL(tx)).
			Where("id = ?", userID).
			Take(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("user %d not found", userID)
			}
			return err
		}

		currency := inferRewardCurrency(tx.WithContext(ctx), userID)
		rate, err := getRewardExchangeRate(tx.WithContext(ctx), currency)
		if err != nil {
			return err
		}

		now := time.Now()
		beforeBalance := user.Balance
		afterBalance := beforeBalance + rewardAmountU
		referenceID := buildRewardReferenceID(rewardCategory, userID)

		if err := tx.Table("users").
			Where("id = ?", userID).
			Updates(map[string]interface{}{
				"balance":    afterBalance,
				"updated_at": now,
			}).Error; err != nil {
			return err
		}

		localAmount := roundRewardAmount(rewardAmountU * rate)
		remark := buildRewardRemark(rewardCategory, rewardAmountU, currency, localAmount)

		if err := tx.Table("transactions").Create(map[string]interface{}{
			"user_id":        userID,
			"type":           6,
			"amount":         rewardAmountU,
			"before_balance": beforeBalance,
			"after_balance":  afterBalance,
			"reference_id":   referenceID,
			"remark":         remark,
			"created_at":     now,
		}).Error; err != nil {
			return err
		}

		if err := tx.Table("user_wager_locks").Create(map[string]interface{}{
			"user_id":          userID,
			"source_type":      rewardCategory,
			"reference_id":     referenceID,
			"reward_amount_u":  rewardAmountU,
			"wager_multiplier": result.WagerMultiplier,
			"wager_required":   result.WagerRequired,
			"granted_at":       now,
			"created_at":       now,
			"updated_at":       now,
		}).Error; err != nil {
			return err
		}

		result.Username = user.Username
		result.Currency = currency
		result.ExchangeRate = rate
		result.LocalAmount = localAmount
		result.BeforeBalance = beforeBalance
		result.AfterBalance = afterBalance
		result.ReferenceID = referenceID
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func inferRewardCurrency(db *gorm.DB, userID uint64) string {
	var order rewardOrder
	if err := db.Table("payment_orders").
		Select("dst_code").
		Where("user_id = ? AND status = 1", userID).
		Order("paid_at DESC").
		Order("created_at DESC").
		Take(&order).Error; err == nil {
		return normalizeExamplePaymentCurrency(order.DstCode)
	}

	if err := db.Table("withdraw_orders").
		Select("dst_code").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Take(&order).Error; err == nil {
		return normalizeExamplePaymentCurrency(order.DstCode)
	}

	return "IDR"
}

func hasDesktopRewardGrant(db *gorm.DB, userID uint64) (bool, error) {
	var count int64
	if err := db.Model(&model.DesktopRewardClaim{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}

	if manualRewardTableExists(db, "add_desktop_reward_claims") {
		if err := db.Table("add_desktop_reward_claims").
			Where("user_id = ?", userID).
			Count(&count).Error; err != nil {
			return false, err
		}
		if count > 0 {
			return true, nil
		}
	}

	if err := db.Table("user_wager_locks").
		Where("user_id = ? AND source_type = ?", userID, "desktop_reward").
		Count(&count).Error; err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}

	if manualRewardTableExists(db, "user_coupons") {
		if err := db.Table("user_coupons").
			Where("user_id = ? AND activity_type = ?", userID, desktopInsuranceCouponType).
			Count(&count).Error; err != nil {
			return false, err
		}
		if count > 0 {
			return true, nil
		}
	}

	return false, nil
}

func hasDesktopRewardGrantByIP(db *gorm.DB, userID uint64, claimIP string) (bool, error) {
	claimIP = normalizeManualRewardIP(claimIP)
	if claimIP == "" {
		return false, nil
	}

	var count int64
	manualExpr := manualRewardClaimIPSQL(db, "desktop_reward_claims", "drc", "u")
	if err := db.Table("desktop_reward_claims drc").
		Joins("JOIN users u ON u.id = drc.user_id").
		Where("drc.user_id <> ? AND "+manualExpr+" = ?", userID, claimIP).
		Count(&count).Error; err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}

	if !manualRewardTableExists(db, "add_desktop_reward_claims") {
		return false, nil
	}

	addDesktopExpr := manualRewardClaimIPSQL(db, "add_desktop_reward_claims", "adc", "u")
	if err := db.Table("add_desktop_reward_claims adc").
		Joins("JOIN users u ON u.id = adc.user_id").
		Where("adc.user_id <> ? AND "+addDesktopExpr+" = ?", userID, claimIP).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func getRewardExchangeRate(db *gorm.DB, currency string) (float64, error) {
	rate, err := getExampleExchangeRate(db, currency)
	if err != nil {
		return 0, fmt.Errorf("exchange rate for %s load failed: %w", currency, err)
	}
	return rate, nil
}

func buildRewardReferenceID(rewardCategory string, userID uint64) string {
	return fmt.Sprintf("%s_%d_%d", rewardCategory, userID, time.Now().UnixNano())
}

func buildDesktopInsuranceCouponCode(userID uint64) string {
	return fmt.Sprintf("ADI%d%d", userID, time.Now().UnixNano())
}

func buildRewardRemark(rewardCategory string, rewardAmountU float64, currency string, localAmount float64) string {
	switch rewardCategory {
	default:
		return fmt.Sprintf("Manual reward %.2fU (%s %.2f)", rewardAmountU, currency, localAmount)
	}
}

func roundRewardAmount(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}
