package services

import (
	"context"
	"encoding/json"
	"errors"
	"gogogo/models"
	"gogogo/models/dtos"
	"sync"
	"time"

	"gorm.io/gorm"
)

// CouponService 优惠券服务
type CouponService struct {
	db *gorm.DB
}

// NewCouponService 创建优惠券服务实例
func NewCouponService(db *gorm.DB) *CouponService {
	return &CouponService{db: db}
}

// GetUserCoupons 获取用户优惠券列表
func (s *CouponService) GetUserCoupons(ctx context.Context, userID uint64, status int) ([]dtos.UserCoupon, error) {
	var coupons []dtos.UserCoupon

	query := s.db.Where("user_id = ?", userID)

	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&coupons).Error; err != nil {
		return nil, err
	}

	// 检查并更新过期状态
	now := time.Now()
	for i := range coupons {
		if coupons[i].Status == 0 || coupons[i].Status == 1 {
			if coupons[i].ValidEnd.Before(now) {
				coupons[i].Status = 3 // 已过期
				s.db.Save(&coupons[i])
			}
		}
	}

	return coupons, nil
}

// GetCouponByCode 根据优惠券码获取优惠券
func (s *CouponService) GetCouponByCode(ctx context.Context, couponCode string) (*dtos.UserCoupon, error) {
	var coupon dtos.UserCoupon
	if err := s.db.Where("coupon_code = ?", couponCode).First(&coupon).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("优惠券不存在")
		}
		return nil, err
	}

	// 检查是否过期
	if coupon.Status == 0 || coupon.Status == 1 {
		if coupon.ValidEnd.Before(time.Now()) {
			coupon.Status = 3 // 已过期
			s.db.Save(&coupon)
		}
	}

	return &coupon, nil
}

// ActivateCoupon 激活优惠券
func (s *CouponService) ActivateCoupon(ctx context.Context, userID uint64, couponCode string, depositAmount float64) error {
	// 1. 获取优惠券
	coupon, err := s.GetCouponByCode(ctx, couponCode)
	if err != nil {
		return err
	}

	// 2. 验证所有权
	if coupon.UserID != userID {
		return errors.New("无权操作此优惠券")
	}

	// 3. 验证状态
	if coupon.Status != 0 {
		switch coupon.Status {
		case 1:
			return errors.New("优惠券已激活")
		case 2:
			return errors.New("优惠券已使用")
		case 3:
			return errors.New("优惠券已过期")
		default:
			return errors.New("优惠券状态异常")
		}
	}

	// 4. 验证最低充值要求
	if depositAmount < coupon.MinDeposit {
		return errors.New("充值金额未达到优惠券激活要求")
	}

	// 5. 激活优惠券
	now := time.Now()
	coupon.Status = 1 // 已激活
	coupon.ActivatedAt = &now

	if err := s.db.Save(coupon).Error; err != nil {
		return err
	}

	// 6. 更新轮盘活动进度（如果是轮盘优惠券）
	if coupon.ActivityType == "coupon_wheel" {
		s.updateWheelProgress(ctx, userID, couponCode)
	}

	return nil
}

// UseCoupon 使用优惠券
func (s *CouponService) UseCoupon(ctx context.Context, userID uint64, couponCode string) error {
	// 1. 获取优惠券
	coupon, err := s.GetCouponByCode(ctx, couponCode)
	if err != nil {
		return err
	}

	// 2. 验证所有权
	if coupon.UserID != userID {
		return errors.New("无权操作此优惠券")
	}

	// 3. 验证状态
	if coupon.Status != 1 {
		switch coupon.Status {
		case 0:
			return errors.New("优惠券未激活")
		case 2:
			return errors.New("优惠券已使用")
		case 3:
			return errors.New("优惠券已过期")
		default:
			return errors.New("优惠券状态异常")
		}
	}

	// 4. 使用优惠券
	now := time.Now()
	coupon.Status = 2 // 已使用
	coupon.UsedAt = &now

	return s.db.Save(coupon).Error
}

// updateWheelProgress 更新轮盘活动进度
func (s *CouponService) updateWheelProgress(ctx context.Context, userID uint64, couponCode string) error {
	today := time.Now().Format("2006-01-02")

	var progress dtos.UserActivityProgress
	if err := s.db.Where("user_id = ? AND activity_type = ? AND DATE(progress_date) = ?",
		userID, "coupon_wheel", today).First(&progress).Error; err != nil {
		return err
	}

	var progressData dtos.WheelProgressData
	if err := json.Unmarshal(progress.ProgressData, &progressData); err != nil {
		return err
	}

	if progressData.CouponCode == couponCode {
		now := time.Now().Format("2006-01-02 15:04:05")
		progressData.Activated = true
		progressData.ActivatedAt = &now

		progressDataJSON, _ := json.Marshal(progressData)
		progress.ProgressData = progressDataJSON

		return s.db.Save(&progress).Error
	}

	return nil
}

// GetAvailableCoupons 获取用户可用优惠券（已激活未使用且未过期）
func (s *CouponService) GetAvailableCoupons(ctx context.Context, userID uint64) ([]dtos.UserCoupon, error) {
	var coupons []dtos.UserCoupon

	now := time.Now()
	if err := s.db.Where("user_id = ? AND status = ? AND valid_end > ?",
		userID, 1, now).Order("coupon_value DESC").Find(&coupons).Error; err != nil {
		return nil, err
	}

	return coupons, nil
}

// CleanupExpiredCoupons 清理过期优惠券
func (s *CouponService) CleanupExpiredCoupons() error {
	now := time.Now()
	return s.db.Model(&dtos.UserCoupon{}).
		Where("status IN (?) AND valid_end < ?", []int{0, 1}, now).
		Update("status", 3).Error
}

// GetCouponStats 获取用户优惠券统计
func (s *CouponService) GetCouponStats(ctx context.Context, userID uint64) (*CouponStats, error) {
	var stats CouponStats
	stats.UserID = userID

	// 统计各状态数量
	var results []struct {
		Status int
		Count  int64
	}

	if err := s.db.Model(&dtos.UserCoupon{}).
		Select("status, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("status").
		Scan(&results).Error; err != nil {
		return nil, err
	}

	for _, r := range results {
		switch r.Status {
		case 0:
			stats.Unactivated = r.Count
		case 1:
			stats.Activated = r.Count
		case 2:
			stats.Used = r.Count
		case 3:
			stats.Expired = r.Count
		}
	}

	// 统计总优惠金额
	var totalValue float64
	s.db.Model(&dtos.UserCoupon{}).
		Where("user_id = ? AND status = ?", userID, 2).
		Select("COALESCE(SUM(coupon_value), 0)").
		Scan(&totalValue)
	stats.TotalValue = totalValue

	return &stats, nil
}

// CouponStats 优惠券统计
type CouponStats struct {
	UserID      uint64  `json:"user_id"`
	Unactivated int64   `json:"unactivated"`
	Activated   int64   `json:"activated"`
	Used        int64   `json:"used"`
	Expired     int64   `json:"expired"`
	TotalValue  float64 `json:"total_value"`
}

// ============================================
// 单例模式
// ============================================

var (
	couponService     *CouponService
	couponServiceOnce sync.Once
)

// GetCouponService 获取优惠券服务单例
func GetCouponService() *CouponService {
	couponServiceOnce.Do(func() {
		couponService = NewCouponService(models.GetInstance().DbInstance)
	})
	return couponService
}
