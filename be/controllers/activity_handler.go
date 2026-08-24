package controllers

import (
	"strconv"

	"gogogo/models/dtos"
	"gogogo/services"

	"github.com/gofiber/fiber/v2"
)

type ActivityHandler struct {
	activityService *services.ActivityService
	couponService   *services.CouponService
}

func NewActivityHandler() *ActivityHandler {
	return &ActivityHandler{
		activityService: services.GetActivityService(),
		couponService:   services.GetCouponService(),
	}
}

func getUserIDFromJWT(c *fiber.Ctx) uint64 {
	uid := QueryUserIdFromJwt(c)
	if uid < 0 {
		return 0
	}
	return uint64(uid)
}

func (h *ActivityHandler) GetRechargeRebateStatus(c *fiber.Ctx) error {
	userID := getUserIDFromJWT(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{Error: "unauthorized"})
	}

	status, err := h.activityService.GetRechargeRebateStatus(c.Context(), userID, c.Query("currency"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(status)
}

func (h *ActivityHandler) ClaimRechargeRebate(c *fiber.Ctx) error {
	userID := getUserIDFromJWT(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{Error: "unauthorized"})
	}

	var req dtos.ClaimRechargeRebateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: "invalid request: " + err.Error()})
	}

	resp, err := h.activityService.ClaimRechargeRebate(c.Context(), userID, req.DayNumber)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(resp)
}

func (h *ActivityHandler) GetWheelStatus(c *fiber.Ctx) error {
	userID := getUserIDFromJWT(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{Error: "unauthorized"})
	}

	status, err := h.activityService.GetWheelStatus(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(status)
}

func (h *ActivityHandler) SpinWheel(c *fiber.Ctx) error {
	userID := getUserIDFromJWT(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{Error: "unauthorized"})
	}

	resp, err := h.activityService.SpinWheel(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(resp)
}

func (h *ActivityHandler) GetLossRebateStatus(c *fiber.Ctx) error {
	userID := getUserIDFromJWT(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{Error: "unauthorized"})
	}

	status, err := h.activityService.GetLossRebateStatus(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(status)
}

func (h *ActivityHandler) ClaimLossRebate(c *fiber.Ctx) error {
	userID := getUserIDFromJWT(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{Error: "unauthorized"})
	}

	resp, err := h.activityService.ClaimLossRebate(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(resp)
}

func (h *ActivityHandler) GetAddDesktopStatus(c *fiber.Ctx) error {
	userID := getUserIDFromJWT(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{Error: "unauthorized"})
	}

	status, err := h.activityService.GetAddDesktopStatus(c.Context(), userID, QueryClientIP(c))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(status)
}

func (h *ActivityHandler) ClaimAddDesktop(c *fiber.Ctx) error {
	userID := getUserIDFromJWT(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{Error: "unauthorized"})
	}

	resp, err := h.activityService.ClaimAddDesktop(c.Context(), userID, QueryClientIP(c))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(resp)
}

func (h *ActivityHandler) GetWorldCupStatus(c *fiber.Ctx) error {
	userID := getUserIDFromJWT(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{Error: "unauthorized"})
	}

	status, err := h.activityService.GetWorldCupStatus(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(status)
}

func (h *ActivityHandler) ClaimWorldCup(c *fiber.Ctx) error {
	userID := getUserIDFromJWT(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{Error: "unauthorized"})
	}

	resp, err := h.activityService.ClaimWorldCup(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(resp)
}

func (h *ActivityHandler) GetVIPMonthlyBonusStatus(c *fiber.Ctx) error {
	userID := getUserIDFromJWT(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{Error: "unauthorized"})
	}

	status, err := h.activityService.GetVIPMonthlyBonusStatus(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(status)
}

func (h *ActivityHandler) ClaimVIPMonthlyBonus(c *fiber.Ctx) error {
	userID := getUserIDFromJWT(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{Error: "unauthorized"})
	}

	resp, err := h.activityService.ClaimVIPMonthlyBonus(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(resp)
}

func (h *ActivityHandler) GetBettingRankValue(c *fiber.Ctx) error {
	userID := getUserIDFromJWT(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{Error: "unauthorized"})
	}

	value, err := h.activityService.GetBettingRankValue(c.Context())
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(value)
}

func (h *ActivityHandler) GetWeeklySpinWheelStatus(c *fiber.Ctx) error {
	userID := getUserIDFromJWT(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{Error: "unauthorized"})
	}

	status, err := h.activityService.GetWeeklySpinWheelStatus(c.Context(), userID, c.Query("currency"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(status)
}

func (h *ActivityHandler) GetSevenDayTopupStatus(c *fiber.Ctx) error {
	userID := getUserIDFromJWT(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{Error: "unauthorized"})
	}

	currency := c.Query("currency")
	status, err := h.activityService.GetSevenDayTopupStatus(c.Context(), userID, currency)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(status)
}

func (h *ActivityHandler) GetNewUserRechargeStatus(c *fiber.Ctx) error {
	userID := getUserIDFromJWT(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{Error: "unauthorized"})
	}

	status, err := h.activityService.GetNewUserRechargeStatus(c.Context(), userID, c.Query("currency"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(status)
}

func (h *ActivityHandler) GetDailyWeeklyChallengeStatus(c *fiber.Ctx) error {
	userID := getUserIDFromJWT(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{Error: "unauthorized"})
	}

	status, err := h.activityService.GetDailyWeeklyChallengeStatus(c.Context(), userID, c.Query("currency"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(status)
}

func (h *ActivityHandler) ClaimDailyWeeklyChallenge(c *fiber.Ctx) error {
	userID := getUserIDFromJWT(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{Error: "unauthorized"})
	}

	var req dtos.DailyWeeklyChallengeClaimRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: "invalid request: " + err.Error()})
	}

	resp, err := h.activityService.ClaimDailyWeeklyChallenge(c.Context(), userID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(resp)
}

func (h *ActivityHandler) GetUserCoupons(c *fiber.Ctx) error {
	userID := getUserIDFromJWT(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{Error: "unauthorized"})
	}

	status := -1
	if statusStr := c.Query("status"); statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			status = s
		}
	}

	coupons, err := h.couponService.GetUserCoupons(c.Context(), userID, status)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(coupons)
}

func (h *ActivityHandler) ActivateCoupon(c *fiber.Ctx) error {
	userID := getUserIDFromJWT(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{Error: "unauthorized"})
	}

	var req dtos.ActivateCouponRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: "invalid request: " + err.Error()})
	}

	if err := h.couponService.ActivateCoupon(c.Context(), userID, req.CouponCode, req.DepositAmount); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(dtos.SuccessResponse{Message: "coupon activated successfully"})
}

func (h *ActivityHandler) GetCouponStats(c *fiber.Ctx) error {
	userID := getUserIDFromJWT(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{Error: "unauthorized"})
	}

	stats, err := h.couponService.GetCouponStats(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(stats)
}
