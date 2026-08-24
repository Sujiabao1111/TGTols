package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gogogo/models"
	"gogogo/models/dtos"
	"hash/fnv"
	"log"
	"math"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ActivityService 活动服务
type ActivityService struct {
	db *gorm.DB
}

const (
	bettingRankInterval        = 5 * time.Second
	bettingRankMinIncrement    = int64(100)
	bettingRankMaxIncrement    = int64(300)
	addDesktopInsuranceType    = "add_desktop_insurance"
	addDesktopInsuranceTrigger = 0.5
	addDesktopInsuranceRate    = 0.2
	weeklySpinBaseThresholdU   = 100.0
	weeklySpinExtraThresholdU  = 1000.0
	weeklySpinPrizePoolU       = 20000.0
	dailyChallengeBaseBetU     = 50.0
	dailyChallengeHighBetU     = 500.0
	dailyChallengeBaseRewardU  = 10.0
	dailyChallengeHighRewardU  = 100.0
	weeklyChallengeBaseBetU    = 1000.0
	weeklyChallengeHighBetU    = 10000.0
	weeklyChallengeBaseRewardU = 200.0
	weeklyChallengeHighRewardU = 2000.0
	newUserRechargeMinDeposit  = 5.0
	defaultDepositWagerMulti   = 2.0
	defaultRewardWagerMulti    = 20.0
	activityWagerConfigKey     = "default"
)

var newUserRechargeRates = []float64{1, 1.75, 2, 2.5}

const (
	depositWagerSourceType = "deposit"
)

func calculateWagerRequired(rewardAmount, multiplier, fallback float64) float64 {
	if rewardAmount > 0 && multiplier > 0 {
		return roundCurrency(rewardAmount * multiplier)
	}

	if fallback <= 0 {
		return 0
	}

	return roundCurrency(fallback)
}

type activityWagerConfigValues struct {
	DepositWagerMultiplier float64
	RewardWagerMultiplier  float64
}

func (s *ActivityService) GetActivityWagerMultipliers(ctx context.Context) (float64, float64) {
	config := s.getActivityWagerConfig(ctx, s.db)
	return config.DepositWagerMultiplier, config.RewardWagerMultiplier
}

func normalizeActivityWagerConfig(config dtos.ActivityWagerConfig) activityWagerConfigValues {
	values := activityWagerConfigValues{
		DepositWagerMultiplier: config.DepositWagerMultiplier,
		RewardWagerMultiplier:  config.RewardWagerMultiplier,
	}
	if values.DepositWagerMultiplier <= 0 {
		values.DepositWagerMultiplier = defaultDepositWagerMulti
	}
	if values.RewardWagerMultiplier <= 0 {
		values.RewardWagerMultiplier = defaultRewardWagerMulti
	}
	return values
}

func (s *ActivityService) getActivityWagerConfig(ctx context.Context, db *gorm.DB) activityWagerConfigValues {
	if db == nil {
		db = s.db
	}
	if db == nil {
		return activityWagerConfigValues{
			DepositWagerMultiplier: defaultDepositWagerMulti,
			RewardWagerMultiplier:  defaultRewardWagerMulti,
		}
	}
	if ctx == nil {
		ctx = context.Background()
	}

	var config dtos.ActivityWagerConfig
	err := db.WithContext(ctx).Where("config_key = ?", activityWagerConfigKey).Take(&config).Error
	if err != nil {
		return activityWagerConfigValues{
			DepositWagerMultiplier: defaultDepositWagerMulti,
			RewardWagerMultiplier:  defaultRewardWagerMulti,
		}
	}
	return normalizeActivityWagerConfig(config)
}

func singleRewardSourceType(activityType string) string {
	switch activityType {
	case "world_cup":
		return "world_cup_reward"
	default:
		return activityType
	}
}

func singleRewardReferenceID(activityType string, userID uint64) string {
	return fmt.Sprintf("%s:%d", singleRewardSourceType(activityType), userID)
}

func (s *ActivityService) upsertWagerLockTx(
	tx *gorm.DB,
	userID uint64,
	sourceType string,
	referenceID string,
	amountU float64,
	multiplier float64,
	grantedAt time.Time,
) error {
	if tx == nil || userID == 0 || strings.TrimSpace(referenceID) == "" || amountU <= 0 || multiplier <= 0 {
		return nil
	}

	lock := dtos.UserWagerLock{
		UserID:          userID,
		SourceType:      sourceType,
		ReferenceID:     referenceID,
		RewardAmountU:   roundCurrency(amountU),
		WagerMultiplier: multiplier,
		WagerRequired:   roundCurrency(amountU * multiplier),
		GrantedAt:       grantedAt,
	}

	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "reference_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"user_id":          lock.UserID,
			"source_type":      lock.SourceType,
			"reward_amount_u":  lock.RewardAmountU,
			"wager_multiplier": lock.WagerMultiplier,
			"wager_required":   lock.WagerRequired,
			"granted_at":       lock.GrantedAt,
			"updated_at":       time.Now(),
		}),
	}).Create(&lock).Error
}

func (s *ActivityService) GrantDepositWagerLockTx(tx *gorm.DB, userID uint64, referenceID string, amountU float64, grantedAt time.Time) error {
	ctx := context.Background()
	if tx != nil && tx.Statement != nil && tx.Statement.Context != nil {
		ctx = tx.Statement.Context
	}
	config := s.getActivityWagerConfig(ctx, tx)
	return s.upsertWagerLockTx(tx, userID, depositWagerSourceType, referenceID, amountU, config.DepositWagerMultiplier, grantedAt)
}

func (s *ActivityService) GrantRewardWagerLockTx(tx *gorm.DB, userID uint64, sourceType string, referenceID string, amountU float64, grantedAt time.Time) error {
	ctx := context.Background()
	if tx != nil && tx.Statement != nil && tx.Statement.Context != nil {
		ctx = tx.Statement.Context
	}
	config := s.getActivityWagerConfig(ctx, tx)
	return s.upsertWagerLockTx(tx, userID, sourceType, referenceID, amountU, config.RewardWagerMultiplier, grantedAt)
}

func normalizeActivityDisplayCurrency(currency string) string {
	switch strings.ToUpper(strings.TrimSpace(currency)) {
	case "PHP":
		return "PHP"
	case "IDR":
		return "IDR"
	default:
		return ""
	}
}

func getActivityDisplayRate(currency string) float64 {
	normalized := normalizeActivityDisplayCurrency(currency)
	if normalized == "" {
		normalized = "IDR"
	}

	rate, err := models.GetInstance().GetExchangeRate(normalized)
	if err == nil && rate > 0 {
		return rate
	}

	if normalized == "PHP" {
		return 56
	}

	return 17450
}

func convertUToActivityDisplay(amountU float64, currency string) float64 {
	return roundCurrency(amountU * getActivityDisplayRate(currency))
}

func convertUWithRate(amountU float64, rate float64) float64 {
	if rate <= 0 {
		return 0
	}

	return roundCurrency(amountU * rate)
}

func selectActivityDisplayAmount(currency string, idrAmount float64, phpAmount float64) float64 {
	if normalizeActivityDisplayCurrency(currency) == "PHP" {
		return phpAmount
	}

	return idrAmount
}

func activityDisplayCurrencyFromDstCode(dstCode string) string {
	normalized := strings.ToUpper(strings.TrimSpace(dstCode))
	switch {
	case normalized == "":
		return ""
	case strings.HasPrefix(normalized, "GCASH"),
		strings.HasPrefix(normalized, "BDO"),
		strings.HasPrefix(normalized, "BPI"),
		strings.HasPrefix(normalized, "METROBANK"),
		strings.HasPrefix(normalized, "LANDBANK"),
		strings.HasPrefix(normalized, "PNB"),
		strings.HasPrefix(normalized, "SECURITY"),
		strings.HasPrefix(normalized, "UNIONBANK"),
		strings.HasPrefix(normalized, "CHINABANK"),
		strings.HasPrefix(normalized, "RCBC"),
		strings.HasPrefix(normalized, "EASTWEST"):
		return "PHP"
	case strings.HasPrefix(normalized, "USDT"), normalized == "PAYPAL":
		return ""
	default:
		return "IDR"
	}
}

func (s *ActivityService) resolveActivityDisplayCurrency(ctx context.Context, userID uint64, requested string) string {
	if normalized := normalizeActivityDisplayCurrency(requested); normalized != "" {
		return normalized
	}

	type orderCurrencyRow struct {
		DstCode string `gorm:"column:dst_code"`
	}

	var latestPayment orderCurrencyRow
	if err := s.db.WithContext(ctx).
		Model(&dtos.PaymentOrder{}).
		Select("dst_code").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(1).
		Scan(&latestPayment).Error; err == nil {
		if currency := activityDisplayCurrencyFromDstCode(latestPayment.DstCode); currency != "" {
			return currency
		}
	}

	var latestWithdraw orderCurrencyRow
	if err := s.db.WithContext(ctx).
		Model(&dtos.WithdrawOrder{}).
		Select("dst_code").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(1).
		Scan(&latestWithdraw).Error; err == nil {
		if currency := activityDisplayCurrencyFromDstCode(latestWithdraw.DstCode); currency != "" {
			return currency
		}
	}

	return "IDR"
}

func activityTimeLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("GMT+8", 8*60*60)
	}

	return loc
}

func formatActivityTimeUTC(value time.Time) string {
	return value.UTC().Format("2006-01-02 15:04:05")
}

// NewActivityService 创建活动服务实例
func NewActivityService(db *gorm.DB) *ActivityService {
	return &ActivityService{db: db}
}

// ============================================
// 储值返利活动 (Recharge Rebate)
// ============================================

// GetRechargeRebateStatus 获取用户储值返利活动状态
func (s *ActivityService) GetRechargeRebateStatus(ctx context.Context, userID uint64, requestedCurrency string) (interface{}, error) {
	now := time.Now()
	displayCurrency := s.resolveActivityDisplayCurrency(ctx, userID, requestedCurrency)

	// 1. 获取活动配置（包括未开始的）
	var activity dtos.Activity
	err := s.db.Where("type = ? AND start_time <= ? AND end_time >= ?",
		"recharge_rebate", now, now).First(&activity).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 查找是否有即将开始的活动
			var upcomingActivity dtos.Activity
			if err := s.db.Where("type = ? AND start_time > ?",
				"recharge_rebate", now).Order("start_time ASC").First(&upcomingActivity).Error; err == nil {
				// 返回即将开始的活动信息和倒计时
				return s.buildActivityCountdown(&upcomingActivity, false), nil
			}
			// 查找是否有已结束的活动用于展示
			var lastActivity dtos.Activity
			if err := s.db.Where("type = ?", "recharge_rebate").Order("end_time DESC").First(&lastActivity).Error; err == nil {
				return s.buildActivityCountdown(&lastActivity, false), nil
			}
			return nil, errors.New("储值返利活动未配置")
		}
		return nil, err
	}

	// 活动存在但可能未启用（status=0）
	if activity.Status != 1 {
		return s.buildActivityCountdown(&activity, false), nil
	}

	// 2. 解析活动配置
	var baseConfig dtos.ActivityBaseConfig
	if err := json.Unmarshal(activity.Config, &baseConfig); err != nil {
		return nil, err
	}

	// 3. 获取奖励档次配置（先尝试 tier_rewards，兼容 day_rewards）
	var activityConfig dtos.ActivityConfig
	if err := s.db.Where("activity_id = ? AND config_key = ?", activity.ID, "tier_rewards").First(&activityConfig).Error; err != nil {
		if err := s.db.Where("activity_id = ? AND config_key = ?", activity.ID, "day_rewards").First(&activityConfig).Error; err != nil {
			return nil, err
		}
	}

	var tierRewards []dtos.DayRewardConfig
	if err := json.Unmarshal(activityConfig.ConfigValue, &tierRewards); err != nil {
		return nil, err
	}

	// 4. 获取用户在活动期间已完成的充值次数（status >= 2 的记录）
	var completedCount int64
	if err := s.db.Model(&dtos.UserActivityProgress{}).
		Where("user_id = ? AND activity_type = ? AND status >= ?",
			userID, "recharge_rebate", 2).
		Count(&completedCount).Error; err != nil {
		return nil, err
	}

	// 5. 构建进度列表
	var tiers []dtos.RechargeRebateDayInfo
	currentTier := int(completedCount) + 1 // 下一次充值将享受第几档返利

	for _, tierConfig := range tierRewards {
		tierInfo := dtos.RechargeRebateDayInfo{
			Day:         tierConfig.Day,
			Rate:        tierConfig.Rate,
			Label:       tierConfig.Label,
			Description: tierConfig.Description,
			Status:      "upcoming",
			MinDeposit:  convertUToActivityDisplay(baseConfig.MinDeposit, displayCurrency),
		}

		// 根据已完成次数确定每档的状态
		if tierConfig.Day < int(completedCount) {
			tierInfo.Status = "claimed"
		} else if tierConfig.Day == int(completedCount) {
			// 最后一笔完成的，可能已领取或未领取
			var lastProgress dtos.UserActivityProgress
			s.db.Where("user_id = ? AND activity_type = ? AND day_number = ?",
				userID, "recharge_rebate", tierConfig.Day).
				Order("created_at DESC").First(&lastProgress)
			if lastProgress.Status == 3 {
				tierInfo.Status = "claimed"
			} else {
				tierInfo.Status = "completed" // 已完成充值，待领取
			}
		} else if tierConfig.Day == currentTier {
			tierInfo.Status = "active" // 下一次充值享受的档次
		}

		tiers = append(tiers, tierInfo)
	}

	return &dtos.RechargeRebateStatusResponse{
		ActivityID:     activity.ID,
		ActivityName:   activity.Name,
		StartTime:      activity.StartTime.Format("2006-01-02"),
		EndTime:        activity.EndTime.Format("2006-01-02"),
		CurrentTier:    currentTier,
		CompletedCount: int(completedCount),
		MinDeposit:     convertUToActivityDisplay(baseConfig.MinDeposit, displayCurrency),
		Currency:       displayCurrency,
		Tiers:          tiers,
	}, nil
}

// ProcessRechargeRebate 处理储值返利
func (s *ActivityService) ProcessRechargeRebate(ctx context.Context, userID uint64, depositAmount float64) error {
	// 1. 获取活动配置
	var activity dtos.Activity
	if err := s.db.Where("type = ? AND status = ? AND start_time <= ? AND end_time >= ?",
		"recharge_rebate", 1, time.Now(), time.Now()).First(&activity).Error; err != nil {
		return nil // 活动未开启，不处理
	}

	var baseConfig dtos.ActivityBaseConfig
	if err := json.Unmarshal(activity.Config, &baseConfig); err != nil {
		return err
	}

	// 检查最低充值要求
	if depositAmount < baseConfig.MinDeposit {
		return nil
	}

	// 2. 获取奖励档次配置（支持 tier_rewards 和 day_rewards）
	var activityConfig dtos.ActivityConfig
	if err := s.db.Where("activity_id = ? AND config_key = ?", activity.ID, "tier_rewards").First(&activityConfig).Error; err != nil {
		if err := s.db.Where("activity_id = ? AND config_key = ?", activity.ID, "day_rewards").First(&activityConfig).Error; err != nil {
			return err
		}
	}

	var tierRewards []dtos.DayRewardConfig
	if err := json.Unmarshal(activityConfig.ConfigValue, &tierRewards); err != nil {
		return err
	}

	// 3. 查询用户已完成的充值次数
	var completedCount int64
	if err := s.db.Model(&dtos.UserActivityProgress{}).
		Where("user_id = ? AND activity_type = ? AND status >= ?",
			userID, "recharge_rebate", 2).
		Count(&completedCount).Error; err != nil {
		return err
	}

	// 4. 确定当前是第几次充值
	tierNumber := int(completedCount) + 1
	if tierNumber < 1 || tierNumber > len(tierRewards) {
		return nil // 已经超过4次充值，不再享受返利
	}

	// 5. 计算返利
	tierConfig := tierRewards[tierNumber-1]
	rewardAmount := depositAmount * tierConfig.Rate

	// 6. 创建进度记录
	now := time.Now()
	progressDate, _ := time.Parse("2006-01-02", now.Format("2006-01-02"))

	progressData := dtos.RechargeRebateProgressData{
		ActivityID:    uint64(activity.ID),
		DayNumber:     tierNumber,
		DepositAmount: depositAmount,
		RewardAmount:  rewardAmount,
		RewardRate:    tierConfig.Rate,
		MinDeposit:    baseConfig.MinDeposit,
		Claimed:       false,
		ActivityStart: activity.StartTime.Format("2006-01-02"),
	}

	progressDataJSON, _ := json.Marshal(progressData)

	// 创建新记录（每次符合条件的充值都创建一条记录）
	progress := dtos.UserActivityProgress{
		UserID:       userID,
		ActivityType: "recharge_rebate",
		DayNumber:    tierNumber,
		ProgressData: progressDataJSON,
		Status:       2, // 已完成
		ProgressDate: progressDate,
	}
	return s.db.Create(&progress).Error
}

// ClaimRechargeRebate 领取储值返利
// 根据 tierNumber（第几次充值）来领取对应的返利
func (s *ActivityService) ClaimRechargeRebate(ctx context.Context, userID uint64, tierNumber int) (*dtos.ClaimResponse, error) {
	// 1. 验证活动
	var activity dtos.Activity
	if err := s.db.Where("type = ? AND status = ?", "recharge_rebate", 1).First(&activity).Error; err != nil {
		return nil, errors.New("活动未开启")
	}

	// 2. 获取指定第几次充值的进度记录（status=2 已完成但未领取）
	var progress dtos.UserActivityProgress
	if err := s.db.Where("user_id = ? AND activity_type = ? AND day_number = ? AND status = ?",
		userID, "recharge_rebate", tierNumber, 2).
		Order("created_at DESC").First(&progress).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("未达到领取条件或已经领取")
		}
		return nil, err
	}

	// 3. 解析进度数据
	var progressData dtos.RechargeRebateProgressData
	if err := json.Unmarshal(progress.ProgressData, &progressData); err != nil {
		return nil, err
	}

	now := time.Now()
	progress.Status = 3
	wagerConfig := s.getActivityWagerConfig(ctx, s.db)
	progressData.Claimed = true
	progressData.WagerRequired = roundCurrency(progressData.RewardAmount * wagerConfig.RewardWagerMultiplier)
	progressData.WagerCompleted = 0
	progressData.WagerUnlocked = false
	progressData.RewardGrantedAt = now.Format("2006-01-02 15:04:05")
	progress.ProgressData, _ = json.Marshal(progressData)

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user dtos.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userID).Error; err != nil {
			return err
		}

		beforeBalance := user.Balance
		afterBalance := roundCurrency(beforeBalance + progressData.RewardAmount)

		if err := tx.Save(&progress).Error; err != nil {
			return err
		}
		if err := tx.Model(&user).Update("balance", gorm.Expr("balance + ?", progressData.RewardAmount)).Error; err != nil {
			return err
		}

		transaction := dtos.Transaction{
			UserID:        userID,
			Type:          6,
			Amount:        progressData.RewardAmount,
			BeforeBalance: beforeBalance,
			AfterBalance:  afterBalance,
			ReferenceID:   fmt.Sprintf("recharge-rebate-%d", progress.ID),
			Remark:        fmt.Sprintf("Recharge rebate bonus tier %d", tierNumber),
			CreatedAt:     now,
		}
		return tx.Create(&transaction).Error
	}); err != nil {
		return nil, err
	}

	return &dtos.ClaimResponse{
		Success:      true,
		RewardType:   "balance",
		RewardAmount: progressData.RewardAmount,
		Message:      fmt.Sprintf("成功领取 $%.2f 返利", progressData.RewardAmount),
	}, nil
}

// ============================================
// 轮盘活动 (Coupon Wheel)
// ============================================

// GetWheelStatus 获取轮盘活动状态
func (s *ActivityService) GetWheelStatus(ctx context.Context, userID uint64) (interface{}, error) {
	now := time.Now()

	// 1. 获取活动配置（包括未开始的）
	var activity dtos.Activity
	err := s.db.Where("type = ? AND start_time <= ? AND end_time >= ?",
		"coupon_wheel", now, now).First(&activity).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 查找是否有即将开始的活动
			var upcomingActivity dtos.Activity
			if err := s.db.Where("type = ? AND start_time > ?",
				"coupon_wheel", now).Order("start_time ASC").First(&upcomingActivity).Error; err == nil {
				// 返回即将开始的活动信息和倒计时
				return s.buildActivityCountdown(&upcomingActivity, false), nil
			}
			// 查找是否有已结束的活动用于展示
			var lastActivity dtos.Activity
			if err := s.db.Where("type = ?", "coupon_wheel").Order("end_time DESC").First(&lastActivity).Error; err == nil {
				return s.buildActivityCountdown(&lastActivity, false), nil
			}
			return nil, errors.New("轮盘活动未配置")
		}
		return nil, err
	}

	// 活动存在但可能未启用（status=0）
	if activity.Status != 1 {
		return s.buildActivityCountdown(&activity, false), nil
	}

	// 2. 解析基础配置
	var baseConfig dtos.ActivityBaseConfig
	if err := json.Unmarshal(activity.Config, &baseConfig); err != nil {
		return nil, err
	}

	// 3. 获取轮盘奖项配置
	var activityConfig dtos.ActivityConfig
	if err := s.db.Where("activity_id = ? AND config_key = ?", activity.ID, "wheel_rewards").First(&activityConfig).Error; err != nil {
		return nil, err
	}

	var wheelRewards []dtos.WheelRewardConfig
	if err := json.Unmarshal(activityConfig.ConfigValue, &wheelRewards); err != nil {
		return nil, err
	}

	// 4. 转换奖项列表
	var items []dtos.WheelItem
	for _, reward := range wheelRewards {
		items = append(items, dtos.WheelItem{
			Value:       reward.Value,
			Label:       reward.Label,
			Probability: reward.Probability,
		})
	}

	// 5. 检查今日是否已抽奖
	today := time.Now().Format("2006-01-02")
	progressDate, _ := time.Parse("2006-01-02", today)

	var progress dtos.UserActivityProgress
	err = s.db.Where("user_id = ? AND activity_type = ? AND DATE(progress_date) = ?",
		userID, "coupon_wheel", today).First(&progress).Error

	canSpin := errors.Is(err, gorm.ErrRecordNotFound)

	var currentCoupon *dtos.WheelCurrentCoupon
	if err == nil {
		// 解析进度数据
		var progressData dtos.WheelProgressData
		if err := json.Unmarshal(progress.ProgressData, &progressData); err == nil {
			currentCoupon = &dtos.WheelCurrentCoupon{
				CouponCode:           progressData.CouponCode,
				CouponValue:          progressData.CouponValue,
				MinDepositToActivate: progressData.MinDepositToActivate,
				Activated:            progressData.Activated,
				ActivatedAt:          progressData.ActivatedAt,
				ExpiresAt:            progressDate.AddDate(0, 0, 1).Format("2006-01-02 15:04:05"),
			}
		}
	}

	return &dtos.WheelStatusResponse{
		ActivityID:    activity.ID,
		ActivityName:  activity.Name,
		MinDeposit:    baseConfig.MinDeposit,
		CanSpin:       canSpin,
		DailyLimit:    baseConfig.DailyLimit,
		Items:         items,
		CurrentCoupon: currentCoupon,
	}, nil
}

// SpinWheel 轮盘抽奖
func (s *ActivityService) SpinWheel(ctx context.Context, userID uint64) (*dtos.SpinResponse, error) {
	// 1. 获取活动配置
	var activity dtos.Activity
	if err := s.db.Where("type = ? AND status = ?", "coupon_wheel", 1).First(&activity).Error; err != nil {
		return nil, errors.New("活动未开启")
	}

	var baseConfig dtos.ActivityBaseConfig
	if err := json.Unmarshal(activity.Config, &baseConfig); err != nil {
		return nil, err
	}

	// 2. 检查今日是否已抽奖（使用数据库事务避免竞态条件）
	today := time.Now().Format("2006-01-02")
	progressDate, _ := time.Parse("2006-01-02", today)

	// 先检查是否已存在记录
	var existingProgress dtos.UserActivityProgress
	err := s.db.Where("user_id = ? AND activity_type = ? AND DATE(progress_date) = ?",
		userID, "coupon_wheel", today).First(&existingProgress).Error
	if err == nil {
		return nil, errors.New("今日已抽奖")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 3. 获取轮盘奖项配置
	var activityConfig dtos.ActivityConfig
	if err := s.db.Where("activity_id = ? AND config_key = ?", activity.ID, "wheel_rewards").First(&activityConfig).Error; err != nil {
		return nil, err
	}

	var wheelRewards []dtos.WheelRewardConfig
	if err := json.Unmarshal(activityConfig.ConfigValue, &wheelRewards); err != nil {
		return nil, err
	}

	// 4. 根据概率抽奖
	selectedReward := s.selectByProbability(wheelRewards)

	// 5. 生成优惠券码
	couponCode := s.generateCouponCode()

	// 6. 使用事务创建优惠券和进度记录
	tx := s.db.Begin()

	// 创建优惠券
	coupon := dtos.UserCoupon{
		UserID:       userID,
		CouponCode:   couponCode,
		ActivityType: "coupon_wheel",
		CouponValue:  selectedReward.Value,
		MinDeposit:   baseConfig.MinDeposit,
		Status:       0, // 未激活
		ValidStart:   time.Now(),
		ValidEnd:     time.Now().AddDate(0, 0, 1), // 有效期1天
	}

	if err := tx.Create(&coupon).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 7. 创建进度记录
	progressData := dtos.WheelProgressData{
		CouponCode:           couponCode,
		CouponValue:          selectedReward.Value,
		MinDepositToActivate: baseConfig.MinDeposit,
		SpunAt:               time.Now().Format("2006-01-02 15:04:05"),
		Activated:            false,
	}

	progressDataJSON, _ := json.Marshal(progressData)

	progress := dtos.UserActivityProgress{
		UserID:       userID,
		ActivityType: "coupon_wheel",
		DayNumber:    1,
		ProgressData: progressDataJSON,
		Status:       1, // 进行中
		ProgressDate: progressDate,
	}

	if err := tx.Create(&progress).Error; err != nil {
		tx.Rollback()
		// 检查是否是重复键错误
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, errors.New("今日已抽奖")
		}
		return nil, err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &dtos.SpinResponse{
		Success:    true,
		Value:      selectedReward.Value,
		Label:      selectedReward.Label,
		CouponCode: couponCode,
		MinDeposit: baseConfig.MinDeposit,
		ValidUntil: coupon.ValidEnd.Format("2006-01-02 15:04:05"),
	}, nil
}

// ============================================
// 输返活动 (Loss Rebate)
// ============================================

// GetLossRebateStatus 获取输返活动状态
func (s *ActivityService) GetLossRebateStatus(ctx context.Context, userID uint64) (interface{}, error) {
	now := time.Now()

	// 1. 获取活动配置（包括未开始的）
	var activity dtos.Activity
	err := s.db.Where("type = ? AND start_time <= ? AND end_time >= ?",
		"loss_rebate", now, now).First(&activity).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 查找是否有即将开始的活动
			var upcomingActivity dtos.Activity
			if err := s.db.Where("type = ? AND start_time > ?",
				"loss_rebate", now).Order("start_time ASC").First(&upcomingActivity).Error; err == nil {
				// 返回即将开始的活动信息和倒计时
				return s.buildActivityCountdown(&upcomingActivity, false), nil
			}
			// 查找是否有已结束的活动用于展示
			var lastActivity dtos.Activity
			if err := s.db.Where("type = ?", "loss_rebate").Order("end_time DESC").First(&lastActivity).Error; err == nil {
				return s.buildActivityCountdown(&lastActivity, false), nil
			}
			return nil, errors.New("输返活动未配置")
		}
		return nil, err
	}

	// 活动存在但可能未启用（status=0）
	if activity.Status != 1 {
		return s.buildActivityCountdown(&activity, false), nil
	}

	// 2. 解析配置
	var baseConfig dtos.ActivityBaseConfig
	if err := json.Unmarshal(activity.Config, &baseConfig); err != nil {
		return nil, err
	}

	// 3. 获取今日进度
	today := time.Now().Format("2006-01-02")

	var progress dtos.UserActivityProgress
	err = s.db.Where("user_id = ? AND activity_type = ? AND DATE(progress_date) = ?",
		userID, "loss_rebate", today).First(&progress).Error

	var stats dtos.LossRebateStats
	if err == nil {
		var progressData dtos.LossRebateProgressData
		if err := json.Unmarshal(progress.ProgressData, &progressData); err == nil {
			stats = dtos.LossRebateStats{
				BetAmount:        progressData.BetAmount,
				WinAmount:        progressData.WinAmount,
				NetLoss:          progressData.NetLoss,
				RebateRate:       progressData.RebateRate,
				RebateAmount:     progressData.RebateAmount,
				MinLossThreshold: progressData.MinLossThreshold,
				Claimed:          progressData.Claimed,
			}
		}
	} else {
		// 默认统计
		stats = dtos.LossRebateStats{
			RebateRate:       baseConfig.RebateRate,
			MinLossThreshold: baseConfig.MinLoss,
			NetLoss:          0,
			RebateAmount:     0,
			Claimed:          false,
		}
	}

	return &dtos.LossRebateStatusResponse{
		ActivityID:       activity.ID,
		ActivityName:     activity.Name,
		StartTime:        activity.StartTime.Format("2006-01-02"),
		EndTime:          activity.EndTime.Format("2006-01-02"),
		RebateRate:       baseConfig.RebateRate,
		MinLossThreshold: baseConfig.MinLoss,
		Currency:         baseConfig.Currency,
		Stats:            stats,
		CanClaim:         stats.NetLoss >= baseConfig.MinLoss && !stats.Claimed,
	}, nil
}

// UpdateLossRebateStats 更新输返统计（下注时调用）
func (s *ActivityService) UpdateLossRebateStats(ctx context.Context, userID uint64, betAmount, winAmount float64) error {
	// 1. 获取活动配置
	var activity dtos.Activity
	if err := s.db.Where("type = ? AND status = ? AND start_time <= ? AND end_time >= ?",
		"loss_rebate", 1, time.Now(), time.Now()).First(&activity).Error; err != nil {
		return nil // 活动未开启
	}

	var baseConfig dtos.ActivityBaseConfig
	if err := json.Unmarshal(activity.Config, &baseConfig); err != nil {
		return err
	}

	// 2. 获取或创建今日进度
	today := time.Now().Format("2006-01-02")
	progressDate, _ := time.Parse("2006-01-02", today)

	var progress dtos.UserActivityProgress
	err := s.db.Where("user_id = ? AND activity_type = ? AND DATE(progress_date) = ?",
		userID, "loss_rebate", today).First(&progress).Error

	var progressData dtos.LossRebateProgressData

	if err == nil {
		// 解析现有数据
		if err := json.Unmarshal(progress.ProgressData, &progressData); err != nil {
			return err
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// 初始化新数据
		progressData = dtos.LossRebateProgressData{
			ActivityID:       uint64(activity.ID),
			RebateRate:       baseConfig.RebateRate,
			MinLossThreshold: baseConfig.MinLoss,
			Claimed:          false,
		}
		progress = dtos.UserActivityProgress{
			UserID:       userID,
			ActivityType: "loss_rebate",
			DayNumber:    1,
			ProgressDate: progressDate,
			Status:       1, // 进行中
		}
	} else {
		return err
	}

	// 3. 更新统计数据
	progressData.BetAmount += betAmount
	progressData.WinAmount += winAmount
	progressData.NetLoss = progressData.BetAmount - progressData.WinAmount

	// 计算返利金额
	if progressData.NetLoss > 0 {
		progressData.RebateAmount = progressData.NetLoss * progressData.RebateRate
	}

	progressData.CalculatedAt = time.Now().Format("2006-01-02 15:04:05")

	// 4. 保存
	progressDataJSON, _ := json.Marshal(progressData)
	progress.ProgressData = progressDataJSON

	if progress.ID == 0 {
		return s.db.Create(&progress).Error
	}
	return s.db.Save(&progress).Error
}

// ClaimLossRebate 领取输返奖励
func (s *ActivityService) ClaimLossRebate(ctx context.Context, userID uint64) (*dtos.ClaimResponse, error) {
	// 1. 获取活动配置
	var activity dtos.Activity
	if err := s.db.Where("type = ? AND status = ?", "loss_rebate", 1).First(&activity).Error; err != nil {
		return nil, errors.New("活动未开启")
	}

	var baseConfig dtos.ActivityBaseConfig
	if err := json.Unmarshal(activity.Config, &baseConfig); err != nil {
		return nil, err
	}

	// 2. 获取今日进度
	today := time.Now().Format("2006-01-02")

	var progress dtos.UserActivityProgress
	if err := s.db.Where("user_id = ? AND activity_type = ? AND DATE(progress_date) = ?",
		userID, "loss_rebate", today).First(&progress).Error; err != nil {
		return nil, errors.New("今日无亏损记录")
	}

	// 3. 解析进度数据
	var progressData dtos.LossRebateProgressData
	if err := json.Unmarshal(progress.ProgressData, &progressData); err != nil {
		return nil, err
	}

	if progressData.Claimed {
		return nil, errors.New("已经领取过奖励")
	}

	if progressData.NetLoss < baseConfig.MinLoss {
		return nil, fmt.Errorf("亏损未达到最低门槛 $%.2f", baseConfig.MinLoss)
	}

	// 4. 更新状态
	now := time.Now().Format("2006-01-02 15:04:05")
	progressData.Claimed = true
	progressData.ClaimedAt = &now
	progress.Status = 3 // 已领取

	progressDataJSON, _ := json.Marshal(progressData)
	progress.ProgressData = progressDataJSON

	if err := s.db.Save(&progress).Error; err != nil {
		return nil, err
	}

	// 5. 发放奖励
	// TODO: 调用钱包服务发放奖励

	return &dtos.ClaimResponse{
		Success:      true,
		RewardType:   "balance",
		RewardAmount: progressData.RebateAmount,
		Message:      fmt.Sprintf("成功领取 $%.2f 输返奖励", progressData.RebateAmount),
	}, nil
}

// ============================================
// 辅助方法
// ============================================

// getCurrentCycleStart 获取当前周期开始时间
func (s *ActivityService) GetAddDesktopStatus(ctx context.Context, userID uint64, clientIP string) (*dtos.ActivityClaimStatusResponse, error) {
	claimedAt, claimed, err := s.getAddDesktopClaimedAt(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !claimed {
		claimedAt, claimed, err = s.getAddDesktopClaimedAtByIP(ctx, userID, clientIP)
		if err != nil {
			return nil, err
		}
	}

	if !claimed {
		return &dtos.ActivityClaimStatusResponse{
			ActivityType: "add_desktop",
			Claimed:      false,
		}, nil
	}

	return &dtos.ActivityClaimStatusResponse{
		ActivityType: "add_desktop",
		Claimed:      true,
		ClaimedAt:    s.formatClaimedAtPtr(claimedAt),
	}, nil
}

func (s *ActivityService) ClaimAddDesktop(ctx context.Context, userID uint64, clientIP string) (*dtos.ActivityClaimRecordResponse, error) {
	claimedAt, claimed, err := s.getAddDesktopClaimedAt(ctx, userID)
	if err != nil {
		return nil, err
	}
	if claimed {
		return s.buildAddDesktopInsuranceClaimResponse(claimedAt, true, 0), nil
	}

	claimedAt, claimed, err = s.getAddDesktopClaimedAtByIP(ctx, userID, clientIP)
	if err != nil {
		return nil, err
	}
	if claimed {
		return s.buildSingleRewardIPBlockedResponse("add_desktop", claimedAt), nil
	}

	return s.claimAddDesktopInsurance(ctx, userID, clientIP)
}

func (s *ActivityService) claimAddDesktopInsurance(ctx context.Context, userID uint64, clientIP string) (*dtos.ActivityClaimRecordResponse, error) {
	now := time.Now().UTC()
	normalizedClaimIP := strings.TrimSpace(clientIP)
	var couponID uint64

	txErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if normalizedClaimIP != "" {
			var existingIPClaim dtos.AddDesktopRewardClaim
			result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("claim_ip = ? AND user_id <> ?", normalizedClaimIP, userID).
				Order("claimed_at ASC, id ASC").
				Limit(1).
				Find(&existingIPClaim)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected > 0 {
				return gorm.ErrDuplicatedKey
			}
		}

		claim := dtos.AddDesktopRewardClaim{
			UserID:    userID,
			ClaimedAt: now,
		}
		if normalizedClaimIP != "" {
			claim.ClaimIP = &normalizedClaimIP
		}
		if err := tx.Create(&claim).Error; err != nil {
			if strings.Contains(err.Error(), "Duplicate entry") {
				return gorm.ErrDuplicatedKey
			}
			return err
		}

		coupon := dtos.UserCoupon{
			UserID:       userID,
			CouponCode:   s.generateAddDesktopInsuranceCouponCode(userID),
			ActivityType: addDesktopInsuranceType,
			CouponValue:  addDesktopInsuranceRate,
			MinDeposit:   0,
			Status:       1,
			ValidStart:   now,
			ValidEnd:     now.AddDate(10, 0, 0),
			ActivatedAt:  &now,
		}
		if err := tx.Create(&coupon).Error; err != nil {
			return err
		}
		couponID = coupon.ID
		return nil
	})
	if txErr != nil {
		if errors.Is(txErr, gorm.ErrDuplicatedKey) {
			claimedAt, claimed, err := s.getAddDesktopClaimedAt(ctx, userID)
			if err != nil {
				return nil, err
			}
			if claimed {
				return s.buildAddDesktopInsuranceClaimResponse(claimedAt, true, 0), nil
			}
			if normalizedClaimIP != "" {
				claimedAt, claimed, err = s.getAddDesktopClaimedAtByIP(ctx, userID, normalizedClaimIP)
				if err != nil {
					return nil, err
				}
				if claimed {
					return s.buildSingleRewardIPBlockedResponse("add_desktop", claimedAt), nil
				}
			}
		}
		return nil, txErr
	}

	return s.buildAddDesktopInsuranceClaimResponse(now, false, couponID), nil
}

func (s *ActivityService) RecordAddDesktopGameEntry(ctx context.Context, userID uint64, platformCode, referenceID string, entryAmountU float64, enteredAt time.Time) error {
	referenceID = strings.TrimSpace(referenceID)
	if userID == 0 || referenceID == "" || entryAmountU <= 0 {
		return nil
	}
	if !s.hasAvailableAddDesktopInsuranceCoupon(ctx, userID) {
		return nil
	}
	if enteredAt.IsZero() {
		enteredAt = time.Now().UTC()
	}

	entry := dtos.AddDesktopGameEntry{
		UserID:       userID,
		PlatformCode: strings.ToUpper(strings.TrimSpace(platformCode)),
		ReferenceID:  referenceID,
		EntryAmountU: roundCurrency(entryAmountU),
		Status:       0,
		EnteredAt:    enteredAt.UTC(),
	}
	if entry.PlatformCode == "" {
		entry.PlatformCode = "HEDOC"
	}

	err := s.db.WithContext(ctx).Create(&entry).Error
	if err != nil && strings.Contains(err.Error(), "Duplicate entry") {
		return nil
	}
	return err
}

func (s *ActivityService) hasAvailableAddDesktopInsuranceCoupon(ctx context.Context, userID uint64) bool {
	var count int64
	err := s.db.WithContext(ctx).Model(&dtos.UserCoupon{}).
		Where("user_id = ? AND activity_type = ? AND status = ? AND valid_start <= ? AND valid_end >= ?",
			userID, addDesktopInsuranceType, 1, time.Now(), time.Now()).
		Count(&count).Error
	return err == nil && count > 0
}

func (s *ActivityService) SettleAddDesktopInsurance(ctx context.Context, userID uint64, platformCode string, exitAmountU float64) (*dtos.AddDesktopInsuranceSettlementResponse, error) {
	if userID == 0 || exitAmountU < 0 {
		return nil, nil
	}

	now := time.Now().UTC()
	resp := &dtos.AddDesktopInsuranceSettlementResponse{
		Triggered:   false,
		ExitAmountU: roundCurrency(exitAmountU),
	}

	txErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var entry dtos.AddDesktopGameEntry
		query := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND status = ?", userID, 0)
		if strings.TrimSpace(platformCode) != "" {
			query = query.Where("platform_code = ?", strings.ToUpper(strings.TrimSpace(platformCode)))
		}

		result := query.Order("entered_at DESC, id DESC").Limit(1).Find(&entry)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}

		entry.ExitAmountU = roundCurrency(exitAmountU)
		entry.LossAmountU = roundCurrency(entry.EntryAmountU - entry.ExitAmountU)
		entry.Status = 1
		entry.SettledAt = &now

		resp.EntryAmountU = entry.EntryAmountU
		resp.ExitAmountU = entry.ExitAmountU
		resp.LossAmountU = entry.LossAmountU

		if entry.EntryAmountU <= 0 || entry.LossAmountU < roundCurrency(entry.EntryAmountU*addDesktopInsuranceTrigger) {
			return tx.Save(&entry).Error
		}

		var coupon dtos.UserCoupon
		couponResult := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND activity_type = ? AND status = ? AND valid_start <= ? AND valid_end >= ?",
				userID, addDesktopInsuranceType, 1, now, now).
			Order("created_at ASC, id ASC").
			Limit(1).
			Find(&coupon)
		if couponResult.Error != nil {
			return couponResult.Error
		}
		if couponResult.RowsAffected == 0 {
			return tx.Save(&entry).Error
		}

		compensation := roundCurrency(entry.LossAmountU * addDesktopInsuranceRate)
		if compensation <= 0 {
			return tx.Save(&entry).Error
		}

		var user dtos.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userID).Error; err != nil {
			return err
		}

		beforeBalance := user.Balance
		afterBalance := roundCurrency(beforeBalance + compensation)
		if err := tx.Model(&user).Update("balance", afterBalance).Error; err != nil {
			return err
		}

		coupon.Status = 2
		coupon.UsedAt = &now
		if err := tx.Save(&coupon).Error; err != nil {
			return err
		}

		entry.Status = 2
		entry.CouponID = &coupon.ID
		entry.CompensationAmountU = compensation
		if err := tx.Save(&entry).Error; err != nil {
			return err
		}

		transaction := dtos.Transaction{
			UserID:        userID,
			Type:          6,
			Amount:        compensation,
			BeforeBalance: beforeBalance,
			AfterBalance:  afterBalance,
			ReferenceID:   fmt.Sprintf("add-desktop-insurance-%d", entry.ID),
			Remark:        "Add desktop insurance compensation",
			CreatedAt:     now,
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		if err := s.GrantRewardWagerLockTx(
			tx,
			userID,
			addDesktopInsuranceType,
			fmt.Sprintf("add-desktop-insurance-%d", entry.ID),
			compensation,
			now,
		); err != nil {
			return err
		}

		resp.Triggered = true
		resp.CompensationAmountU = compensation
		resp.CurrentBalance = afterBalance
		resp.CouponCode = coupon.CouponCode
		resp.Message = fmt.Sprintf("保障券已触发，亏损 %.2fU，补偿 %.2fU 已到账", entry.LossAmountU, compensation)
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}
	if resp.EntryAmountU <= 0 && resp.LossAmountU <= 0 && !resp.Triggered {
		return nil, nil
	}
	return resp, nil
}

func (s *ActivityService) GetWorldCupStatus(ctx context.Context, userID uint64) (*dtos.ActivityClaimStatusResponse, error) {
	var claim dtos.WorldCupRewardClaim
	return s.getSingleRewardClaimStatus(ctx, "world_cup", userID, &claim)
}

func (s *ActivityService) ClaimWorldCup(ctx context.Context, userID uint64) (*dtos.ActivityClaimRecordResponse, error) {
	var claim dtos.WorldCupRewardClaim
	return s.claimSingleReward(ctx, "world_cup", userID, &claim)
}

func (s *ActivityService) GetVIPMonthlyBonusStatus(ctx context.Context, userID uint64) (*dtos.VIPMonthlyBonusStatusResponse, error) {
	if _, err := RecalculateAndPersistUserVIPLevel(ctx, userID); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var user dtos.User
	if err := s.db.WithContext(ctx).
		Select("id, vip_level, balance").
		First(&user, userID).Error; err != nil {
		return nil, err
	}

	claimMonth := currentVIPMonthlyBonusMonth()
	rewardAmount := getVIPMonthlyBonusRewardAmount(user.VipLevel)

	resp := &dtos.VIPMonthlyBonusStatusResponse{
		ActivityType:    "vip_monthly_bonus",
		CurrentVIPLevel: user.VipLevel,
		ClaimMonth:      claimMonth,
		RewardAmount:    rewardAmount,
		CanClaim:        user.VipLevel >= 2 && rewardAmount > 0,
		Claimed:         false,
		Balance:         user.Balance,
	}

	var claim dtos.VIPMonthlyBonusClaim
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND claim_month = ?", userID, claimMonth).
		First(&claim).Error
	if err == nil {
		resp.Claimed = true
		resp.CanClaim = false
		resp.RewardAmount = claim.RewardAmount
		resp.ClaimedAt = s.formatClaimedAtPtr(claim.ClaimedAt)
		return resp, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return resp, nil
}

func (s *ActivityService) ClaimVIPMonthlyBonus(ctx context.Context, userID uint64) (*dtos.ActivityClaimRecordResponse, error) {
	if _, err := RecalculateAndPersistUserVIPLevel(ctx, userID); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	claimMonth := currentVIPMonthlyBonusMonth()
	now := time.Now().UTC()
	var response *dtos.ActivityClaimRecordResponse

	txErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user dtos.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("id, vip_level, balance").
			First(&user, userID).Error; err != nil {
			return err
		}

		rewardAmount := getVIPMonthlyBonusRewardAmount(user.VipLevel)
		if user.VipLevel < 2 || rewardAmount <= 0 {
			return errors.New("vip level not eligible for monthly bonus")
		}

		var existingClaim dtos.VIPMonthlyBonusClaim
		if err := tx.
			Where("user_id = ? AND claim_month = ?", userID, claimMonth).
			First(&existingClaim).Error; err == nil {
			response = s.buildSingleRewardClaimResponse("vip_monthly_bonus", existingClaim.ClaimedAt, true, existingClaim.RewardAmount, user.Balance)
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		beforeBalance := user.Balance
		afterBalance := roundCurrency(beforeBalance + rewardAmount)

		if err := tx.Model(&user).Update("balance", afterBalance).Error; err != nil {
			return err
		}

		transaction := dtos.Transaction{
			UserID:        userID,
			Type:          6,
			Amount:        rewardAmount,
			BeforeBalance: beforeBalance,
			AfterBalance:  afterBalance,
			ReferenceID:   fmt.Sprintf("vip-monthly-bonus-%s", claimMonth),
			Remark:        fmt.Sprintf("VIP monthly bonus %s", claimMonth),
			CreatedAt:     now,
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		claim := dtos.VIPMonthlyBonusClaim{
			UserID:       userID,
			ClaimMonth:   claimMonth,
			VIPLevel:     user.VipLevel,
			RewardAmount: rewardAmount,
			ClaimedAt:    now,
		}
		if err := tx.Create(&claim).Error; err != nil {
			return err
		}

		response = s.buildSingleRewardClaimResponse("vip_monthly_bonus", now, false, rewardAmount, afterBalance)
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	return response, nil
}

func (s *ActivityService) getAddDesktopClaimedAt(ctx context.Context, userID uint64) (time.Time, bool, error) {
	var claim dtos.AddDesktopRewardClaim
	claimResult := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Limit(1).
		Find(&claim)
	if claimResult.Error == nil && claimResult.RowsAffected > 0 {
		return claim.ClaimedAt, true, nil
	} else if claimResult.Error != nil {
		return time.Time{}, false, claimResult.Error
	}

	var wagerLock dtos.UserWagerLock
	wagerResult := s.db.WithContext(ctx).
		Where("user_id = ? AND source_type = ?", userID, "desktop_reward").
		Order("granted_at DESC, created_at DESC, id DESC").
		Limit(1).
		Find(&wagerLock)
	if wagerResult.Error == nil && wagerResult.RowsAffected > 0 {
		claimedAt := wagerLock.GrantedAt
		if claimedAt.IsZero() {
			claimedAt = wagerLock.CreatedAt
		}
		return claimedAt, true, nil
	} else if wagerResult.Error != nil {
		return time.Time{}, false, wagerResult.Error
	}

	return time.Time{}, false, nil
}

func (s *ActivityService) getAddDesktopClaimedAtByIP(ctx context.Context, userID uint64, clientIP string) (time.Time, bool, error) {
	clientIP = strings.TrimSpace(clientIP)
	if clientIP == "" {
		return time.Time{}, false, nil
	}

	var claim dtos.AddDesktopRewardClaim
	result := s.db.WithContext(ctx).
		Where("claim_ip = ? AND user_id <> ?", clientIP, userID).
		Order("claimed_at ASC, id ASC").
		Limit(1).
		Find(&claim)
	if result.Error != nil {
		return time.Time{}, false, result.Error
	}
	if result.RowsAffected == 0 {
		return time.Time{}, false, nil
	}

	return claim.ClaimedAt, true, nil
}

func (s *ActivityService) GetBettingRankValue(ctx context.Context) (*dtos.BettingRankValueResponse, error) {
	now := time.Now()
	weekStart := getWeekStart(now)
	completedIntervals := int(now.Sub(weekStart) / bettingRankInterval)

	var currentValue int64
	for slotIndex := 0; slotIndex < completedIntervals; slotIndex++ {
		currentValue += getBettingRankIncrement(weekStart, slotIndex)
	}

	return &dtos.BettingRankValueResponse{
		CurrentValue:       currentValue,
		WeekStart:          weekStart.Format("2006-01-02 15:04:05"),
		NextResetAt:        weekStart.AddDate(0, 0, 7).Format("2006-01-02 15:04:05"),
		RefreshedAt:        now.Format("2006-01-02 15:04:05"),
		CompletedIntervals: completedIntervals,
	}, nil
}

type weeklySpinPrizeTierConfig struct {
	Label       string
	WinnerCount int
	PrizeRate   float64
}

type weeklySpinTicketRecord struct {
	Progress dtos.UserActivityProgress
	Data     dtos.WeeklySpinWheelTicketProgressData
}

func planWeeklySpinWheelTodayTicketReconciliation(
	todayRows []weeklySpinTicketRecord,
	expectedCount int,
) ([]weeklySpinTicketRecord, []weeklySpinTicketRecord) {
	if expectedCount < 0 {
		expectedCount = 0
	}

	sortedRows := append([]weeklySpinTicketRecord(nil), todayRows...)
	sort.Slice(sortedRows, func(i, j int) bool {
		if sortedRows[i].Progress.DayNumber != sortedRows[j].Progress.DayNumber {
			return sortedRows[i].Progress.DayNumber < sortedRows[j].Progress.DayNumber
		}
		if sortedRows[i].Data.DailyTicketIndex != sortedRows[j].Data.DailyTicketIndex {
			return sortedRows[i].Data.DailyTicketIndex < sortedRows[j].Data.DailyTicketIndex
		}
		if !sortedRows[i].Progress.CreatedAt.Equal(sortedRows[j].Progress.CreatedAt) {
			return sortedRows[i].Progress.CreatedAt.Before(sortedRows[j].Progress.CreatedAt)
		}
		return sortedRows[i].Progress.ID < sortedRows[j].Progress.ID
	})

	if expectedCount > len(sortedRows) {
		expectedCount = len(sortedRows)
	}

	keepRows := append([]weeklySpinTicketRecord(nil), sortedRows[:expectedCount]...)
	deleteRows := append([]weeklySpinTicketRecord(nil), sortedRows[expectedCount:]...)

	for index := range keepRows {
		desiredIndex := index + 1
		keepRows[index].Progress.DayNumber = desiredIndex
		keepRows[index].Data.DailyTicketIndex = desiredIndex
		keepRows[index].Data.IsBaseTicket = index == 0
	}

	return keepRows, deleteRows
}

func (s *ActivityService) reconcileWeeklySpinWheelTodayRowsTx(
	tx *gorm.DB,
	todayRows []weeklySpinTicketRecord,
	expectedCount int,
) (int, bool, error) {
	keepRows, deleteRows := planWeeklySpinWheelTodayTicketReconciliation(todayRows, expectedCount)

	if len(deleteRows) > 0 {
		deleteIDs := make([]uint64, 0, len(deleteRows))
		for _, row := range deleteRows {
			deleteIDs = append(deleteIDs, row.Progress.ID)
		}

		if err := tx.Where("id IN ?", deleteIDs).Delete(&dtos.UserActivityProgress{}).Error; err != nil {
			return 0, false, err
		}
	}

	for _, row := range keepRows {
		progressDataJSON, err := json.Marshal(row.Data)
		if err != nil {
			return 0, false, err
		}

		row.Progress.ProgressData = progressDataJSON
		if err := tx.Save(&row.Progress).Error; err != nil {
			return 0, false, err
		}
	}

	return len(keepRows), len(keepRows) > 0, nil
}

func (s *ActivityService) GetWeeklySpinWheelStatus(ctx context.Context, userID uint64, requestedCurrency string) (*dtos.WeeklySpinWheelStatusResponse, error) {
	nowUTC := time.Now().UTC()
	currentCycleStart, currentCycleEnd := getWeeklySpinWheelCycleBoundsUTC8(nowUTC)
	displayCurrency := s.resolveActivityDisplayCurrency(ctx, userID, requestedCurrency)

	latestDraw, err := s.ensureLatestWeeklySpinWheelDraw(ctx, currentCycleStart)
	if err != nil {
		return nil, err
	}

	tickets, todayBetAmountU, err := s.syncWeeklySpinWheelTickets(ctx, userID, nowUTC, currentCycleStart, currentCycleEnd)
	if err != nil {
		return nil, err
	}

	sort.Slice(tickets, func(i, j int) bool {
		return tickets[i].Progress.CreatedAt.After(tickets[j].Progress.CreatedAt)
	})

	idrRate, err := models.GetInstance().GetExchangeRate("IDR")
	if err != nil || idrRate <= 0 {
		idrRate = 17450
	}

	phpRate, err := models.GetInstance().GetExchangeRate("PHP")
	if err != nil || phpRate <= 0 {
		phpRate = 56
	}

	todayKey := nowUTC.In(activityTimeLocation()).Format("2006-01-02")
	respTickets := make([]dtos.WeeklySpinWheelTicket, 0, len(tickets))
	for _, ticket := range tickets {
		ticketDay := ticket.Progress.ProgressDate.In(activityTimeLocation()).Format("2006-01-02")
		respTickets = append(respTickets, dtos.WeeklySpinWheelTicket{
			TicketCode:       ticket.Data.TicketCode,
			AwardedAt:        ticket.Data.AwardedAt,
			BetAmountU:       ticket.Data.BetAmountU,
			IsDailyQualified: ticket.Data.IsBaseTicket,
			IsToday:          ticketDay == todayKey,
			IsWinner:         ticket.Data.IsWinner,
			PrizeTier:        ticket.Data.PrizeTier,
			PrizeRate:        ticket.Data.PrizeRate,
			PrizeAmountU:     ticket.Data.PrizeAmountU,
		})
	}

	todayBetAmountIDR := convertUWithRate(todayBetAmountU, idrRate)
	todayBetAmountPHP := convertUWithRate(todayBetAmountU, phpRate)
	baseTicketThreshold := selectActivityDisplayAmount(displayCurrency, roundCurrency(weeklySpinBaseThresholdU*idrRate), roundCurrency(weeklySpinBaseThresholdU*phpRate))
	extraTicketThreshold := selectActivityDisplayAmount(displayCurrency, roundCurrency(weeklySpinExtraThresholdU*idrRate), roundCurrency(weeklySpinExtraThresholdU*phpRate))
	todayBetAmount := selectActivityDisplayAmount(displayCurrency, todayBetAmountIDR, todayBetAmountPHP)
	baseTicketQualified := todayBetAmount >= baseTicketThreshold
	extraTicketsToday := calculateWeeklySpinWheelTicketCount(todayBetAmountU)
	if extraTicketsToday > 0 {
		extraTicketsToday--
	}
	totalPrizePool := selectActivityDisplayAmount(displayCurrency, roundCurrency(weeklySpinPrizePoolU*idrRate), roundCurrency(weeklySpinPrizePoolU*phpRate))

	return &dtos.WeeklySpinWheelStatusResponse{
		ActivityType:          "weekly_spin_wheel",
		ServerTimeUTC:         formatActivityTimeUTC(nowUTC),
		CurrentCycleStart:     formatActivityTimeUTC(currentCycleStart),
		CurrentCycleEnd:       formatActivityTimeUTC(currentCycleEnd),
		NextDailyResetAt:      formatActivityTimeUTC(getNextUTC8Midnight(nowUTC)),
		NextWeeklyDrawAt:      formatActivityTimeUTC(currentCycleEnd),
		DisplayCurrency:       displayCurrency,
		TodayBetAmountU:       todayBetAmountU,
		TodayBetAmountIDR:     todayBetAmountIDR,
		TodayBetAmountPHP:     todayBetAmountPHP,
		TodayBetAmount:        todayBetAmount,
		BaseTicketThresholdU:  weeklySpinBaseThresholdU,
		ExtraTicketThresholdU: weeklySpinExtraThresholdU,
		BaseTicketThreshold:   baseTicketThreshold,
		ExtraTicketThreshold:  extraTicketThreshold,
		BaseTicketQualified:   baseTicketQualified,
		ExtraTicketsToday:     extraTicketsToday,
		CurrentCycleTickets:   len(respTickets),
		TotalPrizePoolU:       weeklySpinPrizePoolU,
		TotalPrizePool:        totalPrizePool,
		TotalPrizePoolIDR:     roundCurrency(weeklySpinPrizePoolU * idrRate),
		TotalPrizePoolPHP:     roundCurrency(weeklySpinPrizePoolU * phpRate),
		HundredUIDR:           roundCurrency(weeklySpinBaseThresholdU * idrRate),
		HundredUPHP:           roundCurrency(weeklySpinBaseThresholdU * phpRate),
		ThousandUIDR:          roundCurrency(weeklySpinExtraThresholdU * idrRate),
		ThousandUPHP:          roundCurrency(weeklySpinExtraThresholdU * phpRate),
		Tickets:               respTickets,
		LatestDraw:            latestDraw,
	}, nil
}

// EnsureWeeklySpinWheelTickets 依据当前UTC日的打码量，自动补齐当前用户应得的奖券。
// 这个方法可以在第三方打码汇总同步完成后主动调用，避免必须进入活动页才发券。
func (s *ActivityService) EnsureWeeklySpinWheelTickets(ctx context.Context, userID uint64) error {
	nowUTC := time.Now().UTC()
	currentCycleStart, currentCycleEnd := getWeeklySpinWheelCycleBoundsUTC8(nowUTC)
	_, _, err := s.syncWeeklySpinWheelTickets(ctx, userID, nowUTC, currentCycleStart, currentCycleEnd)
	return err
}

func (s *ActivityService) ensureLatestWeeklySpinWheelDraw(ctx context.Context, currentCycleStart time.Time) (*dtos.WeeklySpinWheelLatestDraw, error) {
	previousCycleStart := currentCycleStart.AddDate(0, 0, -7)
	previousCycleEnd := currentCycleStart
	return s.finalizeWeeklySpinWheelCycle(ctx, previousCycleStart, previousCycleEnd)
}

func (s *ActivityService) syncWeeklySpinWheelTickets(
	ctx context.Context,
	userID uint64,
	nowUTC time.Time,
	currentCycleStart time.Time,
	currentCycleEnd time.Time,
) ([]weeklySpinTicketRecord, float64, error) {
	loc := activityTimeLocation()
	todayKey := nowUTC.In(loc).Format("2006-01-02")
	todayBetAmountU, err := s.getWeeklySpinWheelTodayBetU(ctx, userID, todayKey)
	if err != nil {
		return nil, 0, err
	}

	expectedTodayTickets := calculateWeeklySpinWheelTicketCount(todayBetAmountU)
	cycleStartText := formatActivityTimeUTC(currentCycleStart)
	cycleEndText := formatActivityTimeUTC(currentCycleEnd)
	progressDate, _ := time.ParseInLocation("2006-01-02", todayKey, loc)

	currentCycleTickets := make([]weeklySpinTicketRecord, 0)

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user dtos.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userID).Error; err != nil {
			return err
		}

		userRows, err := s.loadWeeklySpinWheelUserRowsTx(tx, userID)
		if err != nil {
			return err
		}

		todayRows := make([]weeklySpinTicketRecord, 0)
		for _, row := range userRows {
			if row.Data.CycleStart != cycleStartText {
				continue
			}

			currentCycleTickets = append(currentCycleTickets, row)
			if row.Progress.ProgressDate.In(loc).Format("2006-01-02") != todayKey {
				continue
			}

			todayRows = append(todayRows, row)
		}

		todayTicketCount, baseTicketExists, err := s.reconcileWeeklySpinWheelTodayRowsTx(tx, todayRows, expectedTodayTickets)
		if err != nil {
			return err
		}

		missingTickets := expectedTodayTickets - todayTicketCount
		if missingTickets <= 0 {
			return nil
		}

		cycleRows, err := s.loadWeeklySpinWheelCycleRowsTx(tx, currentCycleStart, currentCycleEnd)
		if err != nil {
			return err
		}

		usedCodes := make(map[string]struct{}, len(cycleRows))
		for _, row := range cycleRows {
			usedCodes[row.Data.TicketCode] = struct{}{}
		}

		for index := 0; index < missingTickets; index++ {
			isBaseTicket := !baseTicketExists && index == 0
			ticketData := dtos.WeeklySpinWheelTicketProgressData{
				TicketCode:       generateWeeklySpinWheelTicketCode(usedCodes),
				AwardedAt:        formatActivityTimeUTC(nowUTC),
				BetAmountU:       todayBetAmountU,
				DailyTicketIndex: todayTicketCount + index + 1,
				IsBaseTicket:     isBaseTicket,
				CycleStart:       cycleStartText,
				CycleEnd:         cycleEndText,
			}

			ticketDataJSON, err := json.Marshal(ticketData)
			if err != nil {
				return err
			}

			progress := dtos.UserActivityProgress{
				UserID:       userID,
				ActivityType: "weekly_spin_wheel_ticket",
				DayNumber:    ticketData.DailyTicketIndex,
				ProgressData: ticketDataJSON,
				Status:       1,
				ProgressDate: progressDate,
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "user_id"},
					{Name: "activity_type"},
					{Name: "progress_date"},
					{Name: "day_number"},
				},
				DoNothing: true,
			}).Create(&progress).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	userRows, err := s.loadWeeklySpinWheelUserRowsTx(s.db.WithContext(ctx), userID)
	if err != nil {
		return nil, 0, err
	}

	currentCycleTickets = currentCycleTickets[:0]
	for _, row := range userRows {
		if row.Data.CycleStart != cycleStartText {
			continue
		}
		currentCycleTickets = append(currentCycleTickets, row)
	}

	return currentCycleTickets, todayBetAmountU, nil
}

func (s *ActivityService) getWeeklySpinWheelTodayBetU(ctx context.Context, userID uint64, todayKey string) (float64, error) {
	var stat dtos.UserGameTransactionStat
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND period_type = ? AND period_key = ?", userID, "daily", todayKey).
		First(&stat).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return stat.Turnover, nil
}

func (s *ActivityService) finalizeWeeklySpinWheelCycle(
	ctx context.Context,
	cycleStart time.Time,
	cycleEnd time.Time,
) (*dtos.WeeklySpinWheelLatestDraw, error) {
	cycleRows, err := s.loadWeeklySpinWheelCycleRows(ctx, cycleStart, cycleEnd)
	if err != nil {
		return nil, err
	}
	if len(cycleRows) == 0 {
		return nil, nil
	}

	allResolved := true
	for _, row := range cycleRows {
		if row.Data.DrawnAt == nil {
			allResolved = false
			break
		}
	}

	if allResolved {
		return buildWeeklySpinWheelLatestDraw(cycleRows, cycleStart, cycleEnd), nil
	}

	drawTime := formatActivityTimeUTC(time.Now().UTC())
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		rows, err := s.loadWeeklySpinWheelCycleRowsTx(tx, cycleStart, cycleEnd)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			cycleRows = rows
			return nil
		}

		stillPending := false
		for _, row := range rows {
			if row.Data.DrawnAt == nil {
				stillPending = true
				break
			}
		}
		if !stillPending {
			cycleRows = rows
			return nil
		}

		// Display-only draw:
		// pick exactly one random ticket code as the "winner" for UI display,
		// but do not issue any real balance reward to players.
		for index := range rows {
			rows[index].Data.IsWinner = false
			rows[index].Data.PrizeTier = ""
			rows[index].Data.PrizeRate = 0
			rows[index].Data.PrizeAmountU = 0
		}

		displayTier := getWeeklySpinWheelPrizeTiers()[0]
		winnerIndex := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(rows))
		rows[winnerIndex].Data.IsWinner = true
		rows[winnerIndex].Data.PrizeTier = displayTier.Label
		rows[winnerIndex].Data.PrizeRate = displayTier.PrizeRate
		rows[winnerIndex].Data.PrizeAmountU = roundCurrency(weeklySpinPrizePoolU * displayTier.PrizeRate)

		for index := range rows {
			rows[index].Data.DrawnAt = &drawTime
			if !rows[index].Data.IsWinner {
				rows[index].Data.InvalidatedAt = &drawTime
			}

			progressDataJSON, err := json.Marshal(rows[index].Data)
			if err != nil {
				return err
			}

			rows[index].Progress.ProgressData = progressDataJSON
			rows[index].Progress.Status = 2
			if err := tx.Save(&rows[index].Progress).Error; err != nil {
				return err
			}
		}

		cycleRows = rows
		return nil
	})
	if err != nil {
		return nil, err
	}

	return buildWeeklySpinWheelLatestDraw(cycleRows, cycleStart, cycleEnd), nil
}

func (s *ActivityService) loadWeeklySpinWheelUserRowsTx(tx *gorm.DB, userID uint64) ([]weeklySpinTicketRecord, error) {
	var rows []dtos.UserActivityProgress
	if err := tx.
		Where("user_id = ? AND activity_type = ?", userID, "weekly_spin_wheel_ticket").
		Order("progress_date ASC, created_at ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	return buildWeeklySpinWheelTicketRecords(rows, "")
}

func (s *ActivityService) loadWeeklySpinWheelCycleRows(ctx context.Context, cycleStart time.Time, cycleEnd time.Time) ([]weeklySpinTicketRecord, error) {
	return s.loadWeeklySpinWheelCycleRowsTx(s.db.WithContext(ctx), cycleStart, cycleEnd)
}

func (s *ActivityService) loadWeeklySpinWheelCycleRowsTx(tx *gorm.DB, cycleStart time.Time, cycleEnd time.Time) ([]weeklySpinTicketRecord, error) {
	loc := activityTimeLocation()
	var rows []dtos.UserActivityProgress
	if err := tx.
		Where("activity_type = ? AND progress_date >= ? AND progress_date <= ?",
			"weekly_spin_wheel_ticket",
			cycleStart.In(loc).Format("2006-01-02"),
			cycleEnd.In(loc).Format("2006-01-02"),
		).
		Order("created_at ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	return buildWeeklySpinWheelTicketRecords(rows, formatActivityTimeUTC(cycleStart))
}

func buildWeeklySpinWheelTicketRecords(rows []dtos.UserActivityProgress, cycleStartFilter string) ([]weeklySpinTicketRecord, error) {
	records := make([]weeklySpinTicketRecord, 0, len(rows))
	for _, row := range rows {
		var progressData dtos.WeeklySpinWheelTicketProgressData
		if err := json.Unmarshal(row.ProgressData, &progressData); err != nil {
			continue
		}
		if cycleStartFilter != "" && progressData.CycleStart != cycleStartFilter {
			continue
		}

		records = append(records, weeklySpinTicketRecord{
			Progress: row,
			Data:     progressData,
		})
	}

	return records, nil
}

func buildWeeklySpinWheelLatestDraw(
	records []weeklySpinTicketRecord,
	cycleStart time.Time,
	cycleEnd time.Time,
) *dtos.WeeklySpinWheelLatestDraw {
	winners := make([]dtos.WeeklySpinWheelDrawWinner, 0)
	drawnAt := ""
	for _, record := range records {
		if record.Data.DrawnAt != nil && *record.Data.DrawnAt > drawnAt {
			drawnAt = *record.Data.DrawnAt
		}
		if !record.Data.IsWinner {
			continue
		}

		winners = append(winners, dtos.WeeklySpinWheelDrawWinner{
			TicketCode:   record.Data.TicketCode,
			PrizeTier:    record.Data.PrizeTier,
			PrizeRate:    record.Data.PrizeRate,
			PrizeAmountU: record.Data.PrizeAmountU,
		})
	}

	sort.Slice(winners, func(i, j int) bool {
		leftOrder := getWeeklySpinWheelPrizeTierOrder(winners[i].PrizeTier)
		rightOrder := getWeeklySpinWheelPrizeTierOrder(winners[j].PrizeTier)
		if leftOrder != rightOrder {
			return leftOrder < rightOrder
		}
		return winners[i].TicketCode < winners[j].TicketCode
	})

	return &dtos.WeeklySpinWheelLatestDraw{
		CycleStart:   formatActivityTimeUTC(cycleStart),
		CycleEnd:     formatActivityTimeUTC(cycleEnd),
		DrawnAt:      drawnAt,
		TotalTickets: len(records),
		Winners:      winners,
	}
}

func getWeeklySpinWheelPrizeTiers() []weeklySpinPrizeTierConfig {
	return []weeklySpinPrizeTierConfig{
		{Label: "1st", WinnerCount: 1, PrizeRate: 0.30},
		{Label: "2nd", WinnerCount: 2, PrizeRate: 0.20},
		{Label: "3rd", WinnerCount: 3, PrizeRate: 0.15},
		{Label: "4-10", WinnerCount: 7, PrizeRate: 0.10},
		{Label: "11-30", WinnerCount: 20, PrizeRate: 0.15},
		{Label: "31-100", WinnerCount: 70, PrizeRate: 0.10},
	}
}

func getWeeklySpinWheelPrizeTierOrder(prizeTier string) int {
	for index, tier := range getWeeklySpinWheelPrizeTiers() {
		if tier.Label == prizeTier {
			return index
		}
	}
	return len(getWeeklySpinWheelPrizeTiers()) + 1
}

func calculateWeeklySpinWheelTicketCount(todayBetAmountU float64) int {
	if todayBetAmountU < weeklySpinBaseThresholdU {
		return 0
	}

	return 1 + int(math.Floor((todayBetAmountU-weeklySpinBaseThresholdU)/weeklySpinExtraThresholdU))
}

func getWeeklySpinWheelCycleBoundsUTC8(nowUTC time.Time) (time.Time, time.Time) {
	loc := activityTimeLocation()
	current := nowUTC.In(loc)
	weekday := int(current.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	weekStart := time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, loc).
		AddDate(0, 0, -(weekday - 1))

	return weekStart, weekStart.AddDate(0, 0, 7)
}

func getNextUTC8Midnight(nowUTC time.Time) time.Time {
	loc := activityTimeLocation()
	current := nowUTC.In(loc)
	return time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)
}

func generateWeeklySpinWheelTicketCode(usedCodes map[string]struct{}) string {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	for attempts := 0; attempts < 32; attempts++ {
		ticketCode := fmt.Sprintf("%010d", random.Int63n(10000000000))
		if _, exists := usedCodes[ticketCode]; exists {
			continue
		}

		usedCodes[ticketCode] = struct{}{}
		return ticketCode
	}

	fallback := fmt.Sprintf("%010d", time.Now().UnixNano()%10000000000)
	usedCodes[fallback] = struct{}{}
	return fallback
}

type dailyWeeklyChallengeTemplate struct {
	CycleType string
	TaskIndex int
	BetU      float64
	RewardU   float64
}

func getDailyWeeklyChallengeTemplates() []dailyWeeklyChallengeTemplate {
	return []dailyWeeklyChallengeTemplate{
		{CycleType: "daily", TaskIndex: 1, BetU: dailyChallengeBaseBetU, RewardU: dailyChallengeBaseRewardU},
		{CycleType: "daily", TaskIndex: 2, BetU: dailyChallengeHighBetU, RewardU: dailyChallengeHighRewardU},
		{CycleType: "weekly", TaskIndex: 1, BetU: weeklyChallengeBaseBetU, RewardU: weeklyChallengeBaseRewardU},
		{CycleType: "weekly", TaskIndex: 2, BetU: weeklyChallengeHighBetU, RewardU: weeklyChallengeHighRewardU},
	}
}

func (s *ActivityService) GetDailyWeeklyChallengeStatus(ctx context.Context, userID uint64, requestedCurrency string) (*dtos.DailyWeeklyChallengeStatusResponse, error) {
	nowUTC := time.Now().UTC()
	dailyStart, dailyEnd := getCurrentUTC8DayBounds(nowUTC)
	weeklyStart, weeklyEnd := getCurrentUTC8WeekBounds(nowUTC)
	displayCurrency := s.resolveActivityDisplayCurrency(ctx, userID, requestedCurrency)

	if err := s.ensureDailyWeeklyChallengeSnapshots(ctx, userID, dailyStart, dailyEnd, weeklyStart, weeklyEnd); err != nil {
		return nil, err
	}

	dailyBetU, err := s.getChallengeBetSumU(ctx, userID, dailyStart, dailyEnd)
	if err != nil {
		return nil, err
	}

	weeklyBetU, err := s.getChallengeBetSumU(ctx, userID, weeklyStart, weeklyEnd)
	if err != nil {
		return nil, err
	}

	rows, err := s.loadDailyWeeklyChallengeProgress(ctx, userID, dailyStart, weeklyStart)
	if err != nil {
		return nil, err
	}

	dailyTasks := make([]dtos.DailyWeeklyChallengeTask, 0, 2)
	weeklyTasks := make([]dtos.DailyWeeklyChallengeTask, 0, 2)
	for _, template := range getDailyWeeklyChallengeTemplates() {
		periodStart := dailyStart
		currentBetU := dailyBetU
		if template.CycleType == "weekly" {
			periodStart = weeklyStart
			currentBetU = weeklyBetU
		}

		row, exists := rows[buildDailyWeeklyChallengeRowKey(template.CycleType, template.TaskIndex, periodStart)]
		if !exists {
			continue
		}

		idrRate := row.IDRRate
		if idrRate <= 0 {
			idrRate = getActivityDisplayRate("IDR")
		}
		phpRate := row.PHPRate
		if phpRate <= 0 {
			phpRate = getActivityDisplayRate("PHP")
		}

		currentBetIDR := convertUWithRate(currentBetU, idrRate)
		currentBetPHP := convertUWithRate(currentBetU, phpRate)
		requiredBet := selectActivityDisplayAmount(displayCurrency, row.RequiredBetIDR, row.RequiredBetPHP)
		rewardAmount := selectActivityDisplayAmount(displayCurrency, row.RewardAmountIDR, row.RewardAmountPHP)
		currentBet := selectActivityDisplayAmount(displayCurrency, currentBetIDR, currentBetPHP)

		progressRatio := 0.0
		if requiredBet > 0 {
			progressRatio = currentBet / requiredBet
			if progressRatio > 1 {
				progressRatio = 1
			}
		}

		task := dtos.DailyWeeklyChallengeTask{
			CycleType:       row.CycleType,
			TaskIndex:       row.TaskIndex,
			RequiredBetU:    row.RequiredBetU,
			RewardAmountU:   row.RewardAmountU,
			RequiredBetIDR:  row.RequiredBetIDR,
			RewardAmountIDR: row.RewardAmountIDR,
			RequiredBetPHP:  row.RequiredBetPHP,
			RewardAmountPHP: row.RewardAmountPHP,
			CurrentBetU:     currentBetU,
			CurrentBetIDR:   currentBetIDR,
			CurrentBetPHP:   currentBetPHP,
			DisplayCurrency: displayCurrency,
			RequiredBet:     requiredBet,
			RewardAmount:    rewardAmount,
			CurrentBet:      currentBet,
			ProgressRatio:   progressRatio,
			Claimed:         row.Claimed,
			Claimable:       !row.Claimed && currentBet >= requiredBet,
			PeriodStartUTC:  formatActivityTimeUTC(row.PeriodStart),
			PeriodEndUTC:    formatActivityTimeUTC(row.PeriodEnd),
		}
		if row.ClaimedAt != nil {
			task.ClaimedAt = formatActivityTimeUTC(*row.ClaimedAt)
		}

		if template.CycleType == "daily" {
			dailyTasks = append(dailyTasks, task)
		} else {
			weeklyTasks = append(weeklyTasks, task)
		}
	}

	return &dtos.DailyWeeklyChallengeStatusResponse{
		ActivityType:      "daily_weekly_challenge",
		ServerTimeUTC:     formatActivityTimeUTC(nowUTC),
		DailyPeriodStart:  formatActivityTimeUTC(dailyStart),
		DailyPeriodEnd:    formatActivityTimeUTC(dailyEnd),
		WeeklyPeriodStart: formatActivityTimeUTC(weeklyStart),
		WeeklyPeriodEnd:   formatActivityTimeUTC(weeklyEnd),
		NextDailyResetAt:  formatActivityTimeUTC(dailyEnd),
		NextWeeklyResetAt: formatActivityTimeUTC(weeklyEnd),
		DailyTasks:        dailyTasks,
		WeeklyTasks:       weeklyTasks,
	}, nil
}

func (s *ActivityService) ClaimDailyWeeklyChallenge(
	ctx context.Context,
	userID uint64,
	req dtos.DailyWeeklyChallengeClaimRequest,
) (*dtos.DailyWeeklyChallengeClaimResponse, error) {
	cycleType := strings.ToLower(strings.TrimSpace(req.CycleType))
	if cycleType != "daily" && cycleType != "weekly" {
		return nil, errors.New("invalid cycle type")
	}
	if req.TaskIndex < 1 || req.TaskIndex > 2 {
		return nil, errors.New("invalid task index")
	}

	nowUTC := time.Now().UTC()
	dailyStart, dailyEnd := getCurrentUTC8DayBounds(nowUTC)
	weeklyStart, weeklyEnd := getCurrentUTC8WeekBounds(nowUTC)
	if err := s.ensureDailyWeeklyChallengeSnapshots(ctx, userID, dailyStart, dailyEnd, weeklyStart, weeklyEnd); err != nil {
		return nil, err
	}

	periodStart := dailyStart
	periodEnd := dailyEnd
	if cycleType == "weekly" {
		periodStart = weeklyStart
		periodEnd = weeklyEnd
	}

	currentBetU, err := s.getChallengeBetSumU(ctx, userID, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}

	var response *dtos.DailyWeeklyChallengeClaimResponse
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var progress dtos.DailyWeeklyChallengeProgress
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND cycle_type = ? AND task_index = ? AND period_start = ?",
				userID, cycleType, req.TaskIndex, periodStart).
			First(&progress).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("challenge task not found")
			}
			return err
		}

		if progress.Claimed {
			claimedAt := time.Now().UTC()
			if progress.ClaimedAt != nil {
				claimedAt = progress.ClaimedAt.UTC()
			}
			var user dtos.User
			if err := tx.First(&user, userID).Error; err != nil {
				return err
			}

			response = &dtos.DailyWeeklyChallengeClaimResponse{
				Success:        true,
				ActivityType:   "daily_weekly_challenge",
				CycleType:      cycleType,
				TaskIndex:      req.TaskIndex,
				Claimed:        true,
				AlreadyClaimed: true,
				ClaimedAt:      formatActivityTimeUTC(claimedAt),
				RewardAmountU:  progress.RewardAmountU,
				Balance:        user.Balance,
				Message:        "reward already claimed",
			}
			return nil
		}

		if currentBetU < progress.RequiredBetU {
			return errors.New("challenge target not reached yet")
		}

		var user dtos.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userID).Error; err != nil {
			return err
		}

		now := time.Now().UTC()
		progress.Claimed = true
		progress.ClaimedAt = &now
		if err := tx.Save(&progress).Error; err != nil {
			return err
		}

		beforeBalance := user.Balance
		afterBalance := roundCurrency(beforeBalance + progress.RewardAmountU)
		if err := tx.Model(&user).Update("balance", gorm.Expr("balance + ?", progress.RewardAmountU)).Error; err != nil {
			return err
		}

		referenceID := buildDailyWeeklyChallengeReferenceID(userID, cycleType, req.TaskIndex, periodStart)

		transaction := dtos.Transaction{
			UserID:        userID,
			Type:          6,
			Amount:        progress.RewardAmountU,
			BeforeBalance: beforeBalance,
			AfterBalance:  afterBalance,
			ReferenceID:   referenceID,
			Remark:        fmt.Sprintf("Daily weekly challenge %s task %d", cycleType, req.TaskIndex),
			CreatedAt:     now,
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		if err := s.GrantRewardWagerLockTx(tx, userID, "daily_weekly_challenge", referenceID, progress.RewardAmountU, now); err != nil {
			return err
		}

		response = &dtos.DailyWeeklyChallengeClaimResponse{
			Success:        true,
			ActivityType:   "daily_weekly_challenge",
			CycleType:      cycleType,
			TaskIndex:      req.TaskIndex,
			Claimed:        true,
			AlreadyClaimed: false,
			ClaimedAt:      formatActivityTimeUTC(now),
			RewardAmountU:  progress.RewardAmountU,
			Balance:        afterBalance,
			Message:        "reward claimed successfully",
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *ActivityService) ensureDailyWeeklyChallengeSnapshots(
	ctx context.Context,
	userID uint64,
	dailyStart time.Time,
	dailyEnd time.Time,
	weeklyStart time.Time,
	weeklyEnd time.Time,
) error {
	idrRate, err := models.GetInstance().GetExchangeRate("IDR")
	if err != nil || idrRate <= 0 {
		idrRate = 17450
	}
	phpRate, err := models.GetInstance().GetExchangeRate("PHP")
	if err != nil || phpRate <= 0 {
		phpRate = 56
	}

	rows := make([]dtos.DailyWeeklyChallengeProgress, 0, 4)
	for _, template := range getDailyWeeklyChallengeTemplates() {
		periodStart := dailyStart
		periodEnd := dailyEnd
		if template.CycleType == "weekly" {
			periodStart = weeklyStart
			periodEnd = weeklyEnd
		}

		rows = append(rows, dtos.DailyWeeklyChallengeProgress{
			UserID:          userID,
			CycleType:       template.CycleType,
			TaskIndex:       template.TaskIndex,
			PeriodStart:     periodStart,
			PeriodEnd:       periodEnd,
			RequiredBetU:    template.BetU,
			RewardAmountU:   template.RewardU,
			RequiredBetIDR:  roundCurrency(template.BetU * idrRate),
			RewardAmountIDR: roundCurrency(template.RewardU * idrRate),
			RequiredBetPHP:  roundCurrency(template.BetU * phpRate),
			RewardAmountPHP: roundCurrency(template.RewardU * phpRate),
			IDRRate:         idrRate,
			PHPRate:         phpRate,
		})
	}

	return s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error
}

func (s *ActivityService) loadDailyWeeklyChallengeProgress(
	ctx context.Context,
	userID uint64,
	dailyStart time.Time,
	weeklyStart time.Time,
) (map[string]dtos.DailyWeeklyChallengeProgress, error) {
	var rows []dtos.DailyWeeklyChallengeProgress
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND period_start IN ?", userID, []time.Time{dailyStart, weeklyStart}).
		Order("cycle_type ASC, task_index ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	result := make(map[string]dtos.DailyWeeklyChallengeProgress, len(rows))
	for _, row := range rows {
		result[buildDailyWeeklyChallengeRowKey(row.CycleType, row.TaskIndex, row.PeriodStart)] = row
	}
	return result, nil
}

func buildDailyWeeklyChallengeRowKey(cycleType string, taskIndex int, periodStart time.Time) string {
	return fmt.Sprintf("%s:%d:%s", cycleType, taskIndex, formatActivityTimeUTC(periodStart))
}

func (s *ActivityService) getChallengeBetSumU(ctx context.Context, userID uint64, periodStart time.Time, periodEnd time.Time) (float64, error) {
	periodType := "weekly"
	if periodEnd.Sub(periodStart) <= 24*time.Hour {
		periodType = "daily"
	}

	var stat dtos.UserGameTransactionStat
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND period_type = ? AND period_key = ?",
			userID,
			periodType,
			periodStart.In(activityTimeLocation()).Format("2006-01-02"),
		).
		First(&stat).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return stat.Turnover, nil
}

func getCurrentUTC8DayBounds(nowUTC time.Time) (time.Time, time.Time) {
	loc := activityTimeLocation()
	current := nowUTC.In(loc)
	start := time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, loc)
	return start, start.AddDate(0, 0, 1)
}

func getCurrentUTC8WeekBounds(nowUTC time.Time) (time.Time, time.Time) {
	loc := activityTimeLocation()
	current := nowUTC.In(loc)
	weekday := int(current.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	start := time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, loc).
		AddDate(0, 0, -(weekday - 1))
	return start, start.AddDate(0, 0, 7)
}

type sevenDayTopupQualifiedDay struct {
	Day         string
	QualifiedAt time.Time
}

type sevenDayTopupInviteSuccessRow struct {
	FirstPaidAt time.Time `gorm:"column:first_paid_at"`
}

type sevenDayTopupDailyAccumulator struct {
	IDRQualified  bool
	PHPQualified  bool
	USDTQualified bool
	IDRTotal      float64
	PHTotal       float64
	USDTTotal     float64
}

func (s *ActivityService) GetSevenDayTopupStatus(ctx context.Context, userID uint64, currency string) (*dtos.SevenDayTopupStatusResponse, error) {
	normalizedCurrency := normalizeSevenDayTopupCurrency(currency)
	dailyThreshold := s.getSevenDayTopupThreshold(normalizedCurrency)

	qualifyingDays, err := s.getSevenDayTopupQualifiedDays(ctx, userID, normalizedCurrency, dailyThreshold)
	if err != nil {
		return nil, err
	}

	inviteSuccessTimes, err := s.getSevenDayTopupInviteSuccessTimes(ctx, userID)
	if err != nil {
		return nil, err
	}

	directInviteCount, err := s.getSevenDayTopupDirectInviteCount(ctx, userID)
	if err != nil {
		return nil, err
	}

	progressDates := make([]string, 0, 7)
	usedMakeup := false
	completedRound := false
	var pendingMakeup *dtos.SevenDayTopupPendingMakeup
	lastQualifiedDay := ""
	todayKey := time.Now().In(time.Local).Format("2006-01-02")

	addProgressDay := func(day string, isMakeup bool) {
		if len(progressDates) == 7 {
			progressDates = []string{}
			usedMakeup = false
			completedRound = false
		}

		progressDates = append(progressDates, day)
		if isMakeup {
			usedMakeup = true
		}
		completedRound = len(progressDates) == 7
	}

	restartRound := func(day string) {
		progressDates = []string{}
		usedMakeup = false
		completedRound = false
		addProgressDay(day, false)
		lastQualifiedDay = day
	}

	for _, entry := range qualifyingDays {
		currentDay := entry.Day

		if completedRound {
			restartRound(currentDay)
			continue
		}

		if len(progressDates) == 0 {
			addProgressDay(currentDay, false)
			lastQualifiedDay = currentDay
			continue
		}

		diff := sevenDayTopupDayDiff(lastQualifiedDay, currentDay)

		if diff == 1 {
			addProgressDay(currentDay, false)
			lastQualifiedDay = currentDay
			continue
		}

		if diff == 2 && !usedMakeup {
			if sevenDayTopupHasInviteSuccessInWindow(inviteSuccessTimes, entry.QualifiedAt, sevenDayTopupEndOfDay(entry.QualifiedAt)) {
				addProgressDay(sevenDayTopupAddDays(lastQualifiedDay, 1), true)
				addProgressDay(currentDay, false)
				lastQualifiedDay = currentDay
				continue
			}

			if todayKey > currentDay {
				restartRound(currentDay)
				continue
			}

			pendingMakeup = &dtos.SevenDayTopupPendingMakeup{
				MissedDate:   sevenDayTopupAddDays(lastQualifiedDay, 1),
				RechargeDate: currentDay,
			}
			break
		}

		restartRound(currentDay)
	}

	progressCount := len(progressDates)
	if completedRound {
		progressCount = 7
	}

	return &dtos.SevenDayTopupStatusResponse{
		ActivityType:      "seven_day_topup",
		Currency:          normalizedCurrency,
		DailyThreshold:    dailyThreshold,
		ProgressDates:     progressDates,
		ProgressCount:     progressCount,
		UsedMakeup:        usedMakeup,
		CompletedRound:    completedRound,
		PendingMakeup:     pendingMakeup,
		DirectInviteCount: directInviteCount,
	}, nil
}

func (s *ActivityService) GetNewUserRechargeStatus(ctx context.Context, userID uint64, requestedCurrency string) (*dtos.NewUserRechargeStatusResponse, error) {
	tiers, err := s.loadNewUserRechargeProgress(ctx, userID)
	if err != nil {
		return nil, err
	}
	displayCurrency := s.resolveActivityDisplayCurrency(ctx, userID, requestedCurrency)

	remainingWager, allUnlocked := s.applyNewUserRechargeWagerProgress(ctx, userID, tiers)

	progressCount := len(tiers)
	nextTier := progressCount + 1
	if nextTier > len(newUserRechargeRates) {
		nextTier = len(newUserRechargeRates)
	}

	latestRewardAmount := 0.0
	if progressCount > 0 {
		latestRewardAmount = tiers[progressCount-1].RewardAmount
	}

	respTiers := make([]dtos.NewUserRechargeTierInfo, 0, len(newUserRechargeRates))
	for index, rate := range newUserRechargeRates {
		day := index + 1
		tier := dtos.NewUserRechargeTierInfo{
			Day:             day,
			Rate:            rate,
			Status:          "locked",
			MinDeposit:      convertUToActivityDisplay(newUserRechargeMinDeposit, displayCurrency),
			MinDepositU:     newUserRechargeMinDeposit,
			WagerRequired:   0,
			WagerCompleted:  0,
			WagerRequiredU:  0,
			WagerCompletedU: 0,
		}

		if day <= progressCount {
			record := tiers[day-1]
			tier.Status = "unlocked"
			tier.DepositAmount = convertUToActivityDisplay(record.DepositAmount, displayCurrency)
			tier.RewardAmount = convertUToActivityDisplay(record.RewardAmount, displayCurrency)
			tier.WagerRequired = convertUToActivityDisplay(record.WagerRequired, displayCurrency)
			tier.WagerCompleted = convertUToActivityDisplay(record.WagerCompleted, displayCurrency)
			tier.DepositAmountU = record.DepositAmount
			tier.RewardAmountU = record.RewardAmount
			tier.WagerRequiredU = record.WagerRequired
			tier.WagerCompletedU = record.WagerCompleted
			tier.WagerUnlocked = record.WagerUnlocked
			tier.RewardGrantedAt = record.RewardGrantedAt
		} else if day == progressCount+1 && progressCount < len(newUserRechargeRates) {
			tier.Status = "current"
		}

		respTiers = append(respTiers, tier)
	}

	return &dtos.NewUserRechargeStatusResponse{
		ActivityType:        "new_user_recharge",
		MinDeposit:          convertUToActivityDisplay(newUserRechargeMinDeposit, displayCurrency),
		WagerMultiplier:     s.getActivityWagerConfig(ctx, s.db).RewardWagerMultiplier,
		Currency:            displayCurrency,
		ProgressCount:       progressCount,
		NextTier:            nextTier,
		Hidden:              progressCount >= len(newUserRechargeRates),
		AllWagerUnlocked:    allUnlocked,
		RemainingWager:      convertUToActivityDisplay(remainingWager, displayCurrency),
		RemainingWagerU:     remainingWager,
		LatestRewardAmount:  convertUToActivityDisplay(latestRewardAmount, displayCurrency),
		LatestRewardAmountU: latestRewardAmount,
		Tiers:               respTiers,
	}, nil
}

func (s *ActivityService) ProcessNewUserRecharge(ctx context.Context, userID uint64, depositAmount float64) error {
	if depositAmount < newUserRechargeMinDeposit {
		return nil
	}

	now := time.Now()
	progressDate, _ := time.Parse("2006-01-02", now.Format("2006-01-02"))

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user dtos.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userID).Error; err != nil {
			return err
		}

		var existingCount int64
		if err := tx.Model(&dtos.UserActivityProgress{}).
			Where("user_id = ? AND activity_type = ?", userID, "new_user_recharge").
			Count(&existingCount).Error; err != nil {
			return err
		}

		dayNumber := int(existingCount) + 1
		if dayNumber > len(newUserRechargeRates) {
			return nil
		}

		wagerConfig := s.getActivityWagerConfig(ctx, tx)
		rewardRate := newUserRechargeRates[dayNumber-1]
		rewardAmount := roundCurrency(depositAmount * rewardRate)
		wagerRequired := roundCurrency(rewardAmount * wagerConfig.RewardWagerMultiplier)

		progressData := dtos.NewUserRechargeProgressData{
			ActivityID:      0,
			DayNumber:       dayNumber,
			DepositAmount:   depositAmount,
			RewardAmount:    rewardAmount,
			RewardRate:      rewardRate,
			MinDeposit:      newUserRechargeMinDeposit,
			WagerMultiplier: wagerConfig.RewardWagerMultiplier,
			WagerRequired:   wagerRequired,
			WagerCompleted:  0,
			WagerUnlocked:   false,
			RewardGrantedAt: now.Format("2006-01-02 15:04:05"),
		}

		progressDataJSON, err := json.Marshal(progressData)
		if err != nil {
			return err
		}

		progress := dtos.UserActivityProgress{
			UserID:       userID,
			ActivityType: "new_user_recharge",
			DayNumber:    dayNumber,
			ProgressData: progressDataJSON,
			Status:       3,
			ProgressDate: progressDate,
		}
		if err := tx.Create(&progress).Error; err != nil {
			return err
		}

		beforeBalance := user.Balance
		afterBalance := roundCurrency(beforeBalance + rewardAmount)

		if err := tx.Model(&user).Update("balance", gorm.Expr("balance + ?", rewardAmount)).Error; err != nil {
			return err
		}

		transaction := dtos.Transaction{
			UserID:        userID,
			Type:          6,
			Amount:        rewardAmount,
			BeforeBalance: beforeBalance,
			AfterBalance:  afterBalance,
			ReferenceID:   fmt.Sprintf("new-user-recharge-%d", progress.ID),
			Remark:        fmt.Sprintf("New user recharge bonus tier %d", dayNumber),
			CreatedAt:     now,
		}
		return tx.Create(&transaction).Error
	})
}

func (s *ActivityService) GetNewUserRechargeWithdrawBlock(ctx context.Context, userID uint64) (float64, error) {
	return s.GetRechargeBonusWithdrawBlock(ctx, userID)
}

func (s *ActivityService) GetRechargeBonusWithdrawBlock(ctx context.Context, userID uint64) (float64, error) {
	locks, err := s.loadRechargeBonusWagerLocks(ctx, userID)
	if err != nil {
		return 0, err
	}

	remainingWager, _, err := s.applyRechargeBonusWagerProgress(ctx, userID, locks)
	if err != nil {
		return 0, err
	}
	return remainingWager, nil
}

func (s *ActivityService) GetWithdrawWagerBlock(ctx context.Context, userID uint64) (float64, error) {
	remainingWager, _, err := s.GetWithdrawWagerStatus(ctx, userID, 0)
	return remainingWager, err
}

func (s *ActivityService) GetWithdrawWagerStatus(ctx context.Context, userID uint64, balance float64) (float64, float64, error) {
	locks, err := s.loadWithdrawWagerLocks(ctx, userID)
	if err != nil {
		return 0, 0, err
	}

	remainingWager, _, err := s.applyRechargeBonusWagerProgress(ctx, userID, locks)
	if err != nil {
		return 0, 0, err
	}

	lockedAmount := 0.0
	for _, lock := range locks {
		if lock.WagerUnlocked {
			continue
		}
		lockedAmount += lock.LockedAmountU
	}

	withdrawableBalance := roundCurrency(balance - lockedAmount)
	if withdrawableBalance < 0 {
		withdrawableBalance = 0
	}

	return remainingWager, withdrawableBalance, nil
}

func (s *ActivityService) loadWithdrawWagerLocks(ctx context.Context, userID uint64) ([]activityWagerLockRecord, error) {
	activityLocks, err := s.loadRechargeBonusWagerLocks(ctx, userID)
	if err != nil {
		return nil, err
	}

	depositLocks, err := s.loadDepositWagerLocks(ctx, userID)
	if err != nil {
		return nil, err
	}

	manualRewardLocks, err := s.loadManualRewardWagerLocks(ctx, userID)
	if err != nil {
		return nil, err
	}

	locks := make([]activityWagerLockRecord, 0, len(activityLocks)+len(depositLocks)+len(manualRewardLocks))
	locks = append(locks, activityLocks...)
	locks = append(locks, depositLocks...)
	locks = append(locks, manualRewardLocks...)
	return locks, nil
}

func (s *ActivityService) loadDepositWagerLocks(ctx context.Context, userID uint64) ([]activityWagerLockRecord, error) {
	var rows []dtos.UserWagerLock
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND source_type = ?", userID, depositWagerSourceType).
		Order("granted_at ASC, created_at ASC, id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	wagerConfig := s.getActivityWagerConfig(ctx, s.db)
	locks := make([]activityWagerLockRecord, 0, len(rows))
	for _, row := range rows {
		wagerRequired := calculateWagerRequired(row.RewardAmountU, wagerConfig.DepositWagerMultiplier, row.WagerRequired)
		if wagerRequired <= 0 {
			continue
		}

		grantedAt := row.GrantedAt
		if grantedAt.IsZero() {
			grantedAt = row.CreatedAt
		}
		if grantedAt.IsZero() {
			grantedAt = time.Now()
		}

		locks = append(locks, activityWagerLockRecord{
			RecordKey:       fmt.Sprintf("deposit:%d", row.ID),
			SourceActivity:  row.SourceType,
			LockedAmountU:   row.RewardAmountU,
			WagerRequired:   wagerRequired,
			WagerCompleted:  0,
			WagerUnlocked:   false,
			RewardGrantedAt: grantedAt.Format("2006-01-02 15:04:05"),
			ProgressDate:    grantedAt.Format("2006-01-02"),
		})
	}

	return locks, nil
}

func (s *ActivityService) loadRechargeBonusWagerLocks(ctx context.Context, userID uint64) ([]activityWagerLockRecord, error) {
	newUserLocks, err := s.loadNewUserRechargeWagerLocks(ctx, userID)
	if err != nil {
		return nil, err
	}

	rechargeRebateLocks, err := s.loadRechargeRebateWagerLocks(ctx, userID)
	if err != nil {
		return nil, err
	}

	locks := make([]activityWagerLockRecord, 0, len(newUserLocks)+len(rechargeRebateLocks))
	locks = append(locks, newUserLocks...)
	locks = append(locks, rechargeRebateLocks...)
	return locks, nil
}

func (s *ActivityService) loadManualRewardWagerLocks(ctx context.Context, userID uint64) ([]activityWagerLockRecord, error) {
	var rows []dtos.UserWagerLock
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND source_type <> ?", userID, depositWagerSourceType).
		Order("granted_at ASC, created_at ASC, id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	wagerConfig := s.getActivityWagerConfig(ctx, s.db)
	locks := make([]activityWagerLockRecord, 0, len(rows))
	for _, row := range rows {
		wagerRequired := calculateWagerRequired(row.RewardAmountU, wagerConfig.RewardWagerMultiplier, row.WagerRequired)
		if wagerRequired <= 0 {
			continue
		}

		grantedAt := row.GrantedAt
		if grantedAt.IsZero() {
			grantedAt = row.CreatedAt
		}
		if grantedAt.IsZero() {
			grantedAt = time.Now()
		}

		locks = append(locks, activityWagerLockRecord{
			RecordKey:       fmt.Sprintf("manual_reward:%d", row.ID),
			SourceActivity:  row.SourceType,
			LockedAmountU:   row.RewardAmountU,
			WagerRequired:   wagerRequired,
			WagerCompleted:  0,
			WagerUnlocked:   false,
			RewardGrantedAt: grantedAt.Format("2006-01-02 15:04:05"),
			ProgressDate:    grantedAt.Format("2006-01-02"),
		})
	}

	return locks, nil
}

func (s *ActivityService) loadNewUserRechargeWagerLocks(ctx context.Context, userID uint64) ([]activityWagerLockRecord, error) {
	tiers, err := s.loadNewUserRechargeProgress(ctx, userID)
	if err != nil {
		return nil, err
	}

	locks := make([]activityWagerLockRecord, 0, len(tiers))
	for _, tier := range tiers {
		locks = append(locks, activityWagerLockRecord{
			RecordKey:       tier.RecordKey,
			SourceActivity:  "new_user_recharge",
			LockedAmountU:   tier.RewardAmount,
			WagerRequired:   tier.WagerRequired,
			WagerCompleted:  tier.WagerCompleted,
			WagerUnlocked:   tier.WagerUnlocked,
			RewardGrantedAt: tier.RewardGrantedAt,
			ProgressDate:    tier.ProgressDate,
		})
	}

	return locks, nil
}

func (s *ActivityService) loadRechargeRebateWagerLocks(ctx context.Context, userID uint64) ([]activityWagerLockRecord, error) {
	var rows []dtos.UserActivityProgress
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND activity_type = ? AND status = ?", userID, "recharge_rebate", 3).
		Order("progress_date ASC, created_at ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	wagerConfig := s.getActivityWagerConfig(ctx, s.db)
	locks := make([]activityWagerLockRecord, 0, len(rows))
	for _, row := range rows {
		var progressData dtos.RechargeRebateProgressData
		if err := json.Unmarshal(row.ProgressData, &progressData); err != nil {
			continue
		}
		wagerRequired := calculateWagerRequired(progressData.RewardAmount, wagerConfig.RewardWagerMultiplier, progressData.WagerRequired)
		if wagerRequired <= 0 {
			continue
		}

		rewardGrantedAt := progressData.RewardGrantedAt
		if rewardGrantedAt == "" {
			rewardGrantedAt = row.UpdatedAt.Format("2006-01-02 15:04:05")
		}

		locks = append(locks, activityWagerLockRecord{
			RecordKey:       fmt.Sprintf("recharge_rebate:%d:%s", progressData.DayNumber, row.ProgressDate.Format("2006-01-02")),
			SourceActivity:  "recharge_rebate",
			LockedAmountU:   progressData.RewardAmount,
			WagerRequired:   wagerRequired,
			WagerCompleted:  progressData.WagerCompleted,
			WagerUnlocked:   progressData.WagerUnlocked,
			RewardGrantedAt: rewardGrantedAt,
			ProgressDate:    row.ProgressDate.Format("2006-01-02"),
		})
	}

	return locks, nil
}

func normalizeSevenDayTopupCurrency(currency string) string {
	if strings.EqualFold(currency, "PHP") {
		return "PHP"
	}

	return "IDR"
}

func (s *ActivityService) getSevenDayTopupThreshold(currency string) float64 {
	rate, err := models.GetInstance().GetExchangeRate(currency)
	if err != nil || rate <= 0 {
		if currency == "PHP" {
			rate = 56
		} else {
			rate = 16000
		}
	}

	threshold := rate * 10
	if currency == "PHP" {
		return math.Round(threshold*100) / 100
	}

	return math.Round(threshold)
}

func (s *ActivityService) getSevenDayTopupQualifiedDays(ctx context.Context, userID uint64, currency string, dailyThreshold float64) ([]sevenDayTopupQualifiedDay, error) {
	var orders []dtos.PaymentOrder
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, 1).
		Order("COALESCE(paid_at, created_at) ASC").
		Find(&orders).Error; err != nil {
		return nil, err
	}

	entries := make([]sevenDayTopupQualifiedDay, 0)
	dailyAccumulators := make(map[string]*sevenDayTopupDailyAccumulator)

	for _, order := range orders {
		orderTime := order.CreatedAt
		if order.PaidAt != nil {
			orderTime = *order.PaidAt
		}
		orderTime = orderTime.In(time.Local)

		dayKey := orderTime.Format("2006-01-02")
		accumulator := dailyAccumulators[dayKey]
		if accumulator == nil {
			accumulator = &sevenDayTopupDailyAccumulator{}
			dailyAccumulators[dayKey] = accumulator
		}

		paymentCurrency := classifySevenDayTopupPaymentCurrency(order.DstCode)
		qualifiedNow := false

		if currency == "PHP" {
			if paymentCurrency != "PHP" || accumulator.PHPQualified {
				continue
			}

			accumulator.PHTotal += order.Amount
			qualifiedNow = accumulator.PHTotal >= dailyThreshold
			if qualifiedNow {
				accumulator.PHPQualified = true
			}
		} else {
			switch paymentCurrency {
			case "IDR":
				if accumulator.IDRQualified {
					continue
				}
				accumulator.IDRTotal += order.Amount
				qualifiedNow = accumulator.IDRTotal >= dailyThreshold
				if qualifiedNow {
					accumulator.IDRQualified = true
				}
			case "USDT":
				if accumulator.USDTQualified {
					continue
				}
				accumulator.USDTTotal += order.Amount
				qualifiedNow = accumulator.USDTTotal >= 10
				if qualifiedNow {
					accumulator.USDTQualified = true
				}
			default:
				continue
			}
		}

		if qualifiedNow {
			entries = append(entries, sevenDayTopupQualifiedDay{
				Day:         dayKey,
				QualifiedAt: orderTime,
			})
		}
	}

	return entries, nil
}

func (s *ActivityService) getSevenDayTopupInviteSuccessTimes(ctx context.Context, userID uint64) ([]time.Time, error) {
	var rows []sevenDayTopupInviteSuccessRow
	if err := s.db.WithContext(ctx).
		Table("users").
		Select("MIN(COALESCE(payment_orders.paid_at, payment_orders.created_at)) AS first_paid_at").
		Joins("JOIN payment_orders ON payment_orders.user_id = users.id AND payment_orders.status = 1").
		Where("users.parent_id = ?", userID).
		Group("users.id").
		Order("first_paid_at ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	times := make([]time.Time, 0, len(rows))
	for _, row := range rows {
		times = append(times, row.FirstPaidAt.In(time.Local))
	}

	return times, nil
}

func (s *ActivityService) getSevenDayTopupDirectInviteCount(ctx context.Context, userID uint64) (int64, error) {
	var directInviteCount int64
	if err := s.db.WithContext(ctx).
		Model(&dtos.User{}).
		Where("parent_id = ?", userID).
		Count(&directInviteCount).Error; err != nil {
		return 0, err
	}

	return directInviteCount, nil
}

func classifySevenDayTopupPaymentCurrency(dstCode string) string {
	normalizedCode := strings.ToUpper(strings.TrimSpace(dstCode))
	switch {
	case strings.HasPrefix(normalizedCode, "USDT"):
		return "USDT"
	case normalizedCode == "GCASH" || normalizedCode == "GCASH_QR" || normalizedCode == "GCASH_APP":
		return "PHP"
	default:
		return "IDR"
	}
}

func sevenDayTopupDayDiff(startDay string, endDay string) int {
	start, err := time.ParseInLocation("2006-01-02", startDay, time.Local)
	if err != nil {
		return 0
	}

	end, err := time.ParseInLocation("2006-01-02", endDay, time.Local)
	if err != nil {
		return 0
	}

	return int(end.Sub(start).Hours() / 24)
}

func sevenDayTopupAddDays(day string, amount int) string {
	date, err := time.ParseInLocation("2006-01-02", day, time.Local)
	if err != nil {
		return day
	}

	return date.AddDate(0, 0, amount).Format("2006-01-02")
}

func sevenDayTopupEndOfDay(value time.Time) time.Time {
	year, month, day := value.In(time.Local).Date()
	return time.Date(year, month, day, 23, 59, 59, int(time.Second-time.Nanosecond), time.Local)
}

func sevenDayTopupHasInviteSuccessInWindow(successTimes []time.Time, start time.Time, end time.Time) bool {
	windowStart := start.In(time.Local)
	windowEnd := end.In(time.Local)

	for _, successTime := range successTimes {
		current := successTime.In(time.Local)
		if current.Before(windowStart) {
			continue
		}
		if current.After(windowEnd) {
			return false
		}
		return true
	}

	return false
}

type newUserRechargeProgressRecord struct {
	RecordKey       string
	DayNumber       int
	DepositAmount   float64
	RewardAmount    float64
	WagerRequired   float64
	WagerCompleted  float64
	WagerUnlocked   bool
	RewardGrantedAt string
	ProgressDate    string
}

type activityWagerLockRecord struct {
	RecordKey       string
	SourceActivity  string
	LockedAmountU   float64
	WagerRequired   float64
	WagerCompleted  float64
	WagerUnlocked   bool
	RewardGrantedAt string
	ProgressDate    string
}

func (s *ActivityService) loadNewUserRechargeProgress(ctx context.Context, userID uint64) ([]newUserRechargeProgressRecord, error) {
	var rows []dtos.UserActivityProgress
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND activity_type = ?", userID, "new_user_recharge").
		Order("day_number ASC, created_at ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	records := make([]newUserRechargeProgressRecord, 0, len(rows))
	for _, row := range rows {
		var progressData dtos.NewUserRechargeProgressData
		if err := json.Unmarshal(row.ProgressData, &progressData); err != nil {
			continue
		}
		wagerConfig := s.getActivityWagerConfig(ctx, s.db)
		wagerRequired := calculateWagerRequired(progressData.RewardAmount, wagerConfig.RewardWagerMultiplier, progressData.WagerRequired)

		rewardGrantedAt := progressData.RewardGrantedAt
		if rewardGrantedAt == "" {
			rewardGrantedAt = row.CreatedAt.Format("2006-01-02 15:04:05")
		}

		records = append(records, newUserRechargeProgressRecord{
			RecordKey:       fmt.Sprintf("new_user_recharge:%d:%s", progressData.DayNumber, row.ProgressDate.Format("2006-01-02")),
			DayNumber:       progressData.DayNumber,
			DepositAmount:   progressData.DepositAmount,
			RewardAmount:    progressData.RewardAmount,
			WagerRequired:   wagerRequired,
			WagerCompleted:  progressData.WagerCompleted,
			WagerUnlocked:   progressData.WagerUnlocked,
			RewardGrantedAt: rewardGrantedAt,
			ProgressDate:    row.ProgressDate.Format("2006-01-02"),
		})
	}

	return records, nil
}

func (s *ActivityService) applyNewUserRechargeWagerProgress(ctx context.Context, userID uint64, tiers []newUserRechargeProgressRecord) (float64, bool) {
	if len(tiers) == 0 {
		return 0, true
	}

	locks, err := s.loadRechargeBonusWagerLocks(ctx, userID)
	if err != nil {
		return 0, false
	}

	_, updates, err := s.applyRechargeBonusWagerProgress(ctx, userID, locks)
	if err != nil {
		return 0, false
	}

	totalRemaining := 0.0
	allUnlocked := true
	for index := range tiers {
		if updated, exists := updates[tiers[index].RecordKey]; exists {
			tiers[index].WagerCompleted = updated.WagerCompleted
			tiers[index].WagerUnlocked = updated.WagerUnlocked
		}

		remaining := roundCurrency(tiers[index].WagerRequired - tiers[index].WagerCompleted)
		if remaining <= 0 {
			tiers[index].WagerCompleted = tiers[index].WagerRequired
			tiers[index].WagerUnlocked = true
			continue
		}

		allUnlocked = false
		totalRemaining += remaining
	}

	return roundCurrency(totalRemaining), allUnlocked
}

func (s *ActivityService) applyRechargeBonusWagerProgress(ctx context.Context, userID uint64, locks []activityWagerLockRecord) (float64, map[string]activityWagerLockRecord, error) {
	if len(locks) == 0 {
		return 0, map[string]activityWagerLockRecord{}, nil
	}

	sort.Slice(locks, func(i, j int) bool {
		if locks[i].ProgressDate != locks[j].ProgressDate {
			return locks[i].ProgressDate < locks[j].ProgressDate
		}
		if locks[i].RewardGrantedAt != locks[j].RewardGrantedAt {
			return locks[i].RewardGrantedAt < locks[j].RewardGrantedAt
		}
		return locks[i].RecordKey < locks[j].RecordKey
	})

	startDate := locks[0].ProgressDate
	var stats []dtos.DailyUserStats
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND stat_date >= ?", userID, startDate).
		Order("stat_date ASC").
		Find(&stats).Error; err != nil {
		return 0, nil, err
	}

	for index := range locks {
		locks[index].WagerCompleted = 0
		locks[index].WagerUnlocked = false
	}

	for _, stat := range stats {
		availableBet := stat.BetAmount
		if availableBet <= 0 {
			continue
		}

		for index := range locks {
			if locks[index].ProgressDate > stat.StatDate {
				continue
			}

			remaining := locks[index].WagerRequired - locks[index].WagerCompleted
			if remaining <= 0 {
				continue
			}

			usedBet := math.Min(availableBet, remaining)
			locks[index].WagerCompleted = roundCurrency(locks[index].WagerCompleted + usedBet)
			availableBet = roundCurrency(availableBet - usedBet)

			if availableBet <= 0 {
				break
			}
		}
	}

	totalRemaining := 0.0
	updates := make(map[string]activityWagerLockRecord, len(locks))
	for index := range locks {
		remaining := roundCurrency(locks[index].WagerRequired - locks[index].WagerCompleted)
		if remaining <= 0 {
			locks[index].WagerCompleted = locks[index].WagerRequired
			locks[index].WagerUnlocked = true
		} else {
			totalRemaining += remaining
		}

		updates[locks[index].RecordKey] = locks[index]
	}

	return roundCurrency(totalRemaining), updates, nil
}

func (s *ActivityService) getSingleRewardClaimStatus(ctx context.Context, activityType string, userID uint64, model interface{}) (*dtos.ActivityClaimStatusResponse, error) {
	result := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Limit(1).
		Find(model)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return &dtos.ActivityClaimStatusResponse{
			ActivityType: activityType,
			Claimed:      false,
		}, nil
	}

	claimedAt := s.extractClaimedAt(model)
	return &dtos.ActivityClaimStatusResponse{
		ActivityType: activityType,
		Claimed:      true,
		ClaimedAt:    s.formatClaimedAtPtr(claimedAt),
	}, nil
}

func (s *ActivityService) claimSingleReward(ctx context.Context, activityType string, userID uint64, model interface{}, claimIP ...string) (*dtos.ActivityClaimRecordResponse, error) {
	if activityType == "add_desktop" {
		return nil, errors.New("add desktop now grants an insurance coupon")
	}

	result := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Limit(1).
		Find(model)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected > 0 {
		claimedAt := s.extractClaimedAt(model)
		return s.buildSingleRewardClaimResponse(activityType, claimedAt, true), nil
	}

	now := time.Now()
	normalizedClaimIP := ""
	if len(claimIP) > 0 {
		normalizedClaimIP = strings.TrimSpace(claimIP[0])
	}
	switch claim := model.(type) {
	case *dtos.AddDesktopRewardClaim:
		claim.UserID = userID
		if normalizedClaimIP != "" {
			claim.ClaimIP = &normalizedClaimIP
		}
		claim.ClaimedAt = now
	case *dtos.WorldCupRewardClaim:
		claim.UserID = userID
		claim.ClaimedAt = now
	default:
		return nil, errors.New("unsupported reward claim model")
	}

	rewardAmount, remark, rewardErr := s.getSingleRewardConfig(activityType)
	if rewardErr != nil {
		return nil, rewardErr
	}

	var currentBalance float64
	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		var user dtos.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userID).Error; err != nil {
			return err
		}

		if err := tx.Create(model).Error; err != nil {
			if strings.Contains(err.Error(), "Duplicate entry") {
				return gorm.ErrDuplicatedKey
			}
			return err
		}

		beforeBalance := user.Balance
		afterBalance := beforeBalance + rewardAmount

		if err := tx.Model(&user).Update("balance", afterBalance).Error; err != nil {
			return err
		}

		transaction := dtos.Transaction{
			UserID:        userID,
			Type:          6,
			Amount:        rewardAmount,
			BeforeBalance: beforeBalance,
			AfterBalance:  afterBalance,
			ReferenceID:   activityType,
			Remark:        remark,
			CreatedAt:     now,
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		if err := s.GrantRewardWagerLockTx(
			tx,
			userID,
			singleRewardSourceType(activityType),
			singleRewardReferenceID(activityType, userID),
			rewardAmount,
			now,
		); err != nil {
			return err
		}

		currentBalance = afterBalance
		return nil
	})
	if txErr != nil {
		if errors.Is(txErr, gorm.ErrDuplicatedKey) {
			refetchResult := s.db.WithContext(ctx).
				Where("user_id = ?", userID).
				Limit(1).
				Find(model)
			if refetchResult.Error != nil {
				return nil, refetchResult.Error
			}
			if refetchResult.RowsAffected == 0 {
				return nil, gorm.ErrRecordNotFound
			}
			claimedAt := s.extractClaimedAt(model)
			return s.buildSingleRewardClaimResponse(activityType, claimedAt, true), nil
		}
		return nil, txErr
	}

	return s.buildSingleRewardClaimResponse(activityType, now, false, rewardAmount, currentBalance), nil
}

func (s *ActivityService) buildSingleRewardClaimResponse(activityType string, claimedAt time.Time, alreadyClaimed bool, rewardAmount ...float64) *dtos.ActivityClaimRecordResponse {
	message := "reward claimed successfully"
	if alreadyClaimed {
		message = "reward already claimed"
	}

	resp := &dtos.ActivityClaimRecordResponse{
		Success:        true,
		ActivityType:   activityType,
		Claimed:        true,
		AlreadyClaimed: alreadyClaimed,
		ClaimedAt:      claimedAt.Format("2006-01-02 15:04:05"),
		Message:        message,
	}

	if len(rewardAmount) > 0 {
		resp.RewardAmount = rewardAmount[0]
	}
	if len(rewardAmount) > 1 {
		resp.Balance = rewardAmount[1]
	}

	return resp
}

func (s *ActivityService) buildSingleRewardIPBlockedResponse(activityType string, claimedAt time.Time) *dtos.ActivityClaimRecordResponse {
	resp := s.buildSingleRewardClaimResponse(activityType, claimedAt, true)
	resp.Claimed = false
	resp.Message = "desktop insurance coupon already claimed from this IP"
	return resp
}

func (s *ActivityService) buildAddDesktopInsuranceClaimResponse(claimedAt time.Time, alreadyClaimed bool, couponID uint64) *dtos.ActivityClaimRecordResponse {
	message := "保障券领取成功"
	if alreadyClaimed {
		message = "保障券已领取"
	}

	return &dtos.ActivityClaimRecordResponse{
		Success:        true,
		ActivityType:   "add_desktop",
		Claimed:        true,
		AlreadyClaimed: alreadyClaimed,
		ClaimedAt:      claimedAt.Format("2006-01-02 15:04:05"),
		RewardType:     "insurance_coupon",
		CouponID:       couponID,
		Message:        message,
	}
}

func (s *ActivityService) extractClaimedAt(model interface{}) time.Time {
	switch claim := model.(type) {
	case *dtos.AddDesktopRewardClaim:
		return claim.ClaimedAt
	case *dtos.WorldCupRewardClaim:
		return claim.ClaimedAt
	default:
		return time.Time{}
	}
}

func (s *ActivityService) formatClaimedAtPtr(claimedAt time.Time) *string {
	if claimedAt.IsZero() {
		return nil
	}

	formatted := claimedAt.Format("2006-01-02 15:04:05")
	return &formatted
}

func (s *ActivityService) getSingleRewardConfig(activityType string) (float64, string, error) {
	switch activityType {
	case "world_cup":
		return 1, "World Cup activity reward", nil
	default:
		return 0, "", errors.New("unsupported reward activity type")
	}
}

func (s *ActivityService) getCurrentCycleStart(cycleDays int, activityStart time.Time) time.Time {
	now := time.Now()
	daysSinceStart := int(now.Sub(activityStart).Hours() / 24)
	cycleNumber := daysSinceStart / cycleDays
	return activityStart.AddDate(0, 0, cycleNumber*cycleDays)
}

// getTodayDeposit 获取用户今日充值金额
// buildActivityCountdown 构建带倒计时的活动信息
func (s *ActivityService) buildActivityCountdown(activity *dtos.Activity, isActive bool) *dtos.ActivityWithCountdown {
	now := time.Now()
	countdown := int64(0)
	countdownText := ""
	status := 0 // 未开始

	if isActive {
		status = 1
		countdown = 0
		countdownText = "进行中"
	} else if activity.StartTime.After(now) {
		// 未开始
		countdown = int64(activity.StartTime.Sub(now).Seconds())
		days := int(activity.StartTime.Sub(now).Hours() / 24)
		hours := int(activity.StartTime.Sub(now).Hours()) % 24
		minutes := int(activity.StartTime.Sub(now).Minutes()) % 60
		countdownText = fmt.Sprintf("%d天%d小时%d分", days, hours, minutes)
	} else if activity.EndTime.Before(now) {
		// 已结束
		status = 2
		countdown = -1
		countdownText = "已结束"
	}

	return &dtos.ActivityWithCountdown{
		ActivityStatus: dtos.ActivityStatus{
			ActivityID:   activity.ID,
			ActivityName: activity.Name,
			Type:         activity.Type,
			Status:       status,
			StartTime:    activity.StartTime.Format("2006-01-02 15:04:05"),
			EndTime:      activity.EndTime.Format("2006-01-02 15:04:05"),
			Description:  activity.Description,
		},
		Countdown:     countdown,
		CountdownText: countdownText,
		IsActive:      isActive,
		Config:        s.parseActivityConfig(activity),
	}
}

// parseActivityConfig 解析活动配置
func (s *ActivityService) parseActivityConfig(activity *dtos.Activity) interface{} {
	var config map[string]interface{}
	if err := json.Unmarshal(activity.Config, &config); err != nil {
		return nil
	}
	return config
}

func roundCurrency(value float64) float64 {
	return math.Round(value*100) / 100
}

func getWeekStart(now time.Time) time.Time {
	year, month, day := now.Date()
	startOfToday := time.Date(year, month, day, 0, 0, 0, 0, now.Location())

	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	return startOfToday.AddDate(0, 0, -(weekday - 1))
}

func getBettingRankIncrement(weekStart time.Time, slotIndex int) int64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(fmt.Sprintf("%d:%d", weekStart.Unix(), slotIndex)))

	rangeSize := uint64(bettingRankMaxIncrement - bettingRankMinIncrement + 1)
	return bettingRankMinIncrement + int64(hasher.Sum64()%rangeSize)
}

func (s *ActivityService) getTodayDeposit(userID uint64) float64 {
	today := time.Now().Format("2006-01-02")
	startOfDay, _ := time.Parse("2006-01-02", today)
	endOfDay := startOfDay.AddDate(0, 0, 1).Add(-time.Second)

	var totalDeposit float64
	s.db.Model(&dtos.Transaction{}).
		Where("user_id = ? AND type = ? AND created_at >= ? AND created_at <= ?",
			userID, 1, startOfDay, endOfDay).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalDeposit)

	return totalDeposit
}

// selectByProbability 根据概率选择奖励
func (s *ActivityService) selectByProbability(rewards []dtos.WheelRewardConfig) dtos.WheelRewardConfig {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	random := r.Float64()

	var cumulative float64
	for _, reward := range rewards {
		cumulative += reward.Probability
		if random <= cumulative {
			return reward
		}
	}

	// 默认返回第一个
	return rewards[0]
}

// generateCouponCode 生成优惠券码
func (s *ActivityService) generateCouponCode() string {
	prefix := "CP"
	timestamp := time.Now().Unix()
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomNum := random.Intn(10000)
	return fmt.Sprintf("%s%d%04d", prefix, timestamp, randomNum)
}

func (s *ActivityService) generateAddDesktopInsuranceCouponCode(userID uint64) string {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("ADI%d%d%04d", userID, time.Now().Unix(), random.Intn(10000))
}

func buildDailyWeeklyChallengeReferenceID(userID uint64, cycleType string, taskIndex int, periodStart time.Time) string {
	return fmt.Sprintf(
		"daily-weekly-challenge-%d-%s-%d-%s",
		userID,
		cycleType,
		taskIndex,
		periodStart.In(activityTimeLocation()).Format("20060102150405"),
	)
}

func currentVIPMonthlyBonusMonth() string {
	return time.Now().In(activityTimeLocation()).Format("2006-01")
}

func getVIPMonthlyBonusRewardAmount(vipLevel int) float64 {
	if vipLevel < 2 {
		return 0
	}

	for _, rule := range GetVIPTierRules() {
		if rule.Level == vipLevel {
			return roundCurrency(rule.MinDepositU * 0.1)
		}
	}

	return 0
}

// ============================================
// 单例模式
// ============================================

var (
	activityService     *ActivityService
	activityServiceOnce sync.Once
)

// GetActivityService 获取活动服务单例
func GetActivityService() *ActivityService {
	activityServiceOnce.Do(func() {
		db := models.GetInstance().DbInstance
		if err := ensureActivityTables(db); err != nil {
			log.Printf("[ActivityService] ensure activity tables failed: %v\n", err)
		}
		activityService = NewActivityService(db)
	})
	return activityService
}

func ensureActivityTables(db *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}

	return db.AutoMigrate(
		&dtos.Activity{},
		&dtos.ActivityConfig{},
		&dtos.ActivityWagerConfig{},
		&dtos.WithdrawFeeConfig{},
		&dtos.UserActivityProgress{},
		&dtos.DailyWeeklyChallengeProgress{},
		&dtos.UserGameTransactionStat{},
		&dtos.UserCoupon{},
		&dtos.AddDesktopRewardClaim{},
		&dtos.AddDesktopGameEntry{},
		&dtos.WorldCupRewardClaim{},
		&dtos.VIPMonthlyBonusClaim{},
		&dtos.UserWagerLock{},
	)
}
