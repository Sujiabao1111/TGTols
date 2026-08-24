package controllers

import (
	"github.com/gofiber/fiber/v2"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func HealthChecker(ctx *fiber.Ctx) error {
	return ctx.SendString("ok")
}

func InitRouteTables(r *fiber.App) {
	// 测试
	r.Get("/jsonecho", JsonEchoHandler)
	r.Post("/posttodo", JsonPostEchoHandler)
	// 健康检查
	r.Get("/healthCheck", HealthChecker)
	// 业务
	r.Post("/appWakeup", DowakeUp)
	r.Get("/getAppcfgs", GetAppCfgs)

	r.Post("/getTokenTest", GetJwtForTest)

	r.Post("/api/login", LoginHandler)
	r.Post("/api/regtest", Registertest)
	r.Post("/api/reg", Register)

	// test 发送邮件
	r.Post("/mail/send_6942342fe37f068e3f56594e", SendMail)

	// r.Post("/games/sync", SyncGames) // 管理员同步三方游戏到数据库

	r.Post("/games/list", GetGameList)          // 根据条件列出游戏
	r.Post("/games/featured", GetFeaturedGames) // 特色榜单游戏hot,recommend
	r.Post("/games/options", GetGameOptions)    // 根据条件列出厂商,类型
	r.Get("/games/banner", GetBanners)          // 获取banner
	r.Post("/games/all", GetAllGames)           // 获取所有游戏-慎重

	r.Get("/games/lobby3", OpenLobbyHandler3) // 用户获取lobby html
	r.Get("/games/launch3", LaunchGame3)      // 用户获取game html

	// 大富翁游戏专用接口 (Monopoly Integration)
	r.Post("/api/auth/quick-login", QuickLoginForMonopoly)             // 快速登录（同服务器免密）
	r.Get("/api/games/random-for-monopoly", GetRandomGamesForMonopoly) // 随机游戏列表
	r.Get("/games/launch-monopoly", LaunchMonopolyGame)                // 直达游戏（GET，WebView直接打开）
}

func InitRouteTablesWithJWT(r *fiber.App) {
	// 测试
	r.Get("/jwtecho", JsonEchoHandlerJWT)

	// 邮件
	r.Get("/mails", GetUserMails)               // 列表
	r.Post("/mails/:id/claim", ClaimMailReward) // 领取

	// 游戏相关
	r.Post("/games/info", GetUserInfo)            // 用户信息
	r.Post("/games/inviteinfo", GetInviteInfo)    // 用户邀请信息
	r.Post("/games/launch", LaunchGame)           // 用户进入游戏
	r.Post("/games/launch2", LaunchGame2)         // 用户进入游戏-自动带入
	r.Post("/games/recyle", RecycleBalance)       // 回收资金-退出时自动带出
	r.Post("/games/password", UpdateGamePassword) // 修改游戏密码
	r.Post("/games/lobby", OpenLobbyHandler)      // 用户获取lobby链接
	r.Post("/games/lobby2", OpenLobbyHandler2)    // 用户获取lobby链接

	// 钱包相关
	// r.Post("/wallet/transfer", TransferToProvider) // 转入/转出第三方
	r.Post("/wallet/transfer", TransferToGame) // 转入/转出第三方
	r.Post("/wallet/deposit", Deposit)         // 本地充值
	r.Post("/wallet/withdraw", Withdraw)       // 本地提现
}
