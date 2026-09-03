package controllers

import (
	"os"

	"github.com/gofiber/fiber/v2"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func isDevMode() bool {
	return os.Getenv("dev") == "1"
}

func HealthChecker(ctx *fiber.Ctx) error {
	return ctx.SendString("ok")
}

func InitRouteTables(r *fiber.App) {
	r.Get("/jsonecho", JsonEchoHandler)
	r.Post("/posttodo", JsonPostEchoHandler)
	r.Get("/healthCheck", HealthChecker)

	r.Post("/appWakeup", DowakeUp)
	r.Get("/getAppcfgs", GetAppCfgs)
	r.Post("/getTokenTest", GetJwtForTest)

	r.Post("/api/login", LoginHandler)
	r.Post("/api/auth/telegram", TelegramLoginHandler)
	r.Post("/api/regtest", Registertest)
	r.Post("/api/reg", Register)
	r.Post("/api/domain-click", RecordDomainClick)

	r.Post("/mail/send_6942342fe37f068e3f56594e", SendMail)

	paymentHandler := NewPaymentHandler()
	r.Post("/payments/notify", paymentHandler.HandlePaymentNotify)
	r.Post("/withdraw/notify", paymentHandler.HandleWithdrawNotify)
	r.Post("/withdraw/admin/approve", paymentHandler.AdminApproveWithdrawOrder)
	r.Get("/payments/methods", paymentHandler.GetPaymentMethods)
	r.Get("/withdraw/methods", paymentHandler.GetWithdrawMethods)
	r.Post("/payments/tokenpay/notify", paymentHandler.HandleTokenPayNotify)
	r.Post("/payments/telegram-stars/webhook", paymentHandler.HandleTelegramStarsWebhook)
	// Keep an /api-prefixed alias for deployments whose reverse proxy exposes
	// backend callbacks below the /api prefix.
	r.Post("/api/payments/telegram-stars/webhook", paymentHandler.HandleTelegramStarsWebhook)
	r.Post("/payments/telegram-stars/refund", paymentHandler.RefundTelegramStarsPayment)
	r.Get("/payments/telegram-stars/transactions", paymentHandler.GetTelegramStarsTransactions)

	if isDevMode() {
		r.Get("/test/tokenpay/config", TestTokenPayGetConfig)
		r.Get("/test/tokenpay/create", TestTokenPayCreateOrder)
	}

	exchangeRateController := NewExchangeRateController()
	r.Get("/exchange-rates", exchangeRateController.GetExchangeRates)
	r.Get("/exchange-rates/convert/:target", exchangeRateController.ConvertFromIDR)

	r.Post("/games/reorderlist", GetReorderGameList)
	r.Post("/games/list", GetGameList)
	r.Post("/games/featured", GetFeaturedGames)
	r.Post("/games/options", GetGameOptions)
	r.Get("/games/banner", GetBanners)
	r.Post("/games/all", GetAllGames)
	r.Post("/temp/games/transactions/backfill", TempBackfillGameTransactions)
	r.Post("/temp/games/transactions/backfill/legacy", TempBackfillLegacyGameTransactions)
	r.Post("/temp/games/transactions/m7/history", TempDebugM7GameHistory)
}

func InitRouteTablesWithJWT(r *fiber.App) {
	r.Get("/admin/domain-stats", GetDomainStats)
	r.Get("/jwtecho", JsonEchoHandlerJWT)

	r.Get("/mails", GetUserMails)
	r.Post("/mails/:id/claim", ClaimMailReward)

	r.Post("/games/info", GetUserInfo)
	r.Post("/games/inviteinfo", GetInviteInfo)
	r.Post("/games/launch", LaunchGame)
	r.Post("/games/launch2", LaunchGame2)
	r.Post("/games/launch-error/report", ReportGameLaunchError)
	r.Post("/games/recyle", RecycleBalance)
	r.Post("/games/password", UpdateGamePassword)
	r.Post("/games/lobby", OpenLobbyHandler)
	r.Post("/games/lobby2", OpenLobbyHandler2)
	r.Post("/games/transactions/sync", SyncMyGameTransactions)
	r.Post("/games/sync/m7pp", SyncM7PPGamesNow)
	r.Post("/games/transactions/sync/m7pp", SyncM7PPTransactionsNow)
	r.Post("/games/transactions/sync/all", SyncAllGameTransactions)
	r.Post("/games/transactions/backfill", BackfillMyGameTransactions)
	r.Post("/games/transactions/backfill/legacy", BackfillLegacyGameTransactions)
	r.Get("/games/transactions", GetMyGameTransactions)
	r.Get("/games/transactions/overview", GetMyGameTransactionOverview)
	r.Post("/games/dofav", ToggleFavorite)
	r.Post("/games/favlist", GetFavoriteList)

	r.Post("/wallet/transfer", TransferToGame)
	r.Post("/wallet/deposit", Deposit)
	r.Post("/wallet/withdraw", Withdraw)

	activityHandler := NewActivityHandler()
	r.Get("/activities/recharge-rebate", activityHandler.GetRechargeRebateStatus)
	r.Post("/activities/recharge-rebate/claim", activityHandler.ClaimRechargeRebate)
	r.Get("/activities/wheel", activityHandler.GetWheelStatus)
	r.Post("/activities/wheel/spin", activityHandler.SpinWheel)
	r.Get("/activities/loss-rebate", activityHandler.GetLossRebateStatus)
	r.Post("/activities/loss-rebate/claim", activityHandler.ClaimLossRebate)
	r.Get("/activities/add-desktop", activityHandler.GetAddDesktopStatus)
	r.Post("/activities/add-desktop/claim", activityHandler.ClaimAddDesktop)
	r.Get("/activities/world-cup", activityHandler.GetWorldCupStatus)
	r.Post("/activities/world-cup/claim", activityHandler.ClaimWorldCup)
	r.Get("/activities/vip-monthly-bonus", activityHandler.GetVIPMonthlyBonusStatus)
	r.Post("/activities/vip-monthly-bonus/claim", activityHandler.ClaimVIPMonthlyBonus)
	r.Get("/activities/betting-rank/value", activityHandler.GetBettingRankValue)
	r.Get("/activities/weekly-spin-wheel", activityHandler.GetWeeklySpinWheelStatus)
	r.Get("/activities/seven-day-topup", activityHandler.GetSevenDayTopupStatus)
	r.Get("/activities/new-user-recharge", activityHandler.GetNewUserRechargeStatus)
	r.Get("/activities/daily-weekly-challenge", activityHandler.GetDailyWeeklyChallengeStatus)
	r.Post("/activities/daily-weekly-challenge/claim", activityHandler.ClaimDailyWeeklyChallenge)

	r.Get("/coupons", activityHandler.GetUserCoupons)
	r.Post("/coupons/activate", activityHandler.ActivateCoupon)
	r.Get("/coupons/stats", activityHandler.GetCouponStats)

	paymentHandler := NewPaymentHandler()
	r.Post("/payments/order", paymentHandler.CreatePaymentOrder)
	r.Post("/payments/ton/order", paymentHandler.CreateTonOrder)
	r.Get("/payments/ton/rate", paymentHandler.GetTonRate)
	r.Post("/payments/ton/confirm", paymentHandler.ConfirmTonOrder)
	r.Get("/payments/order/:orderId", paymentHandler.GetPaymentStatus)
	r.Get("/payments/orders", paymentHandler.GetUserPayments)
	r.Get("/payments/voucher/config", paymentHandler.GetVoucherConfig)
	r.Post("/payments/voucher/redeem", paymentHandler.RedeemVoucher)
	r.Post("/payments/tokenpay/order", paymentHandler.CreateTokenPayOrder)
	r.Post("/payments/telegram-stars/order", paymentHandler.CreateTelegramStarsOrder)

	r.Post("/withdraw", paymentHandler.CreateWithdrawOrder)
}
