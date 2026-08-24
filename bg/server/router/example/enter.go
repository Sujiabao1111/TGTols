package example

import (
	api "github.com/flipped-aurora/gin-vue-admin/server/api/v1"
)

type RouterGroup struct {
	CustomerRouter
	AttachmentCategoryRouter
	GamesRouter
	AgentRebateConfigsRouter
	BannersRouter
	GameTypesRouter
	GameProvidersRouter
	DailyUserStatsRouter
	StatsRetentionRouter
	InviteStatsRouter
	ManualRewardRouter
	ActivityWagerConfigRouter
	WithdrawFeeConfigRouter
	GameLaunchErrorRouter
	UserDataRouter
	RechargeQueryRouter
	UserGameRecordRouter
	WithdrawAuditRouter
	DataStatisticsRouter
	FileUploadAndDownloadRouter
}

var (
	exaCustomerApi              = api.ApiGroupApp.ExampleApiGroup.CustomerApi
	attachmentCategoryApi       = api.ApiGroupApp.ExampleApiGroup.AttachmentCategoryApi
	gamesApi                    = api.ApiGroupApp.ExampleApiGroup.GamesApi
	agentRebateConfigsApi       = api.ApiGroupApp.ExampleApiGroup.AgentRebateConfigsApi
	bannersApi                  = api.ApiGroupApp.ExampleApiGroup.BannersApi
	gameTypesApi                = api.ApiGroupApp.ExampleApiGroup.GameTypesApi
	gameProvidersApi            = api.ApiGroupApp.ExampleApiGroup.GameProvidersApi
	dailyUserStatsApi           = api.ApiGroupApp.ExampleApiGroup.DailyUserStatsApi
	statsRetentionApi           = api.ApiGroupApp.ExampleApiGroup.StatsRetentionApi
	inviteStatsApi              = api.ApiGroupApp.ExampleApiGroup.InviteStatsApi
	manualRewardApi             = api.ApiGroupApp.ExampleApiGroup.ManualRewardApi
	activityWagerConfigApi      = api.ApiGroupApp.ExampleApiGroup.ActivityWagerConfigApi
	withdrawFeeConfigApi        = api.ApiGroupApp.ExampleApiGroup.WithdrawFeeConfigApi
	gameLaunchErrorApi          = api.ApiGroupApp.ExampleApiGroup.GameLaunchErrorApi
	userDataApi                 = api.ApiGroupApp.ExampleApiGroup.UserDataApi
	rechargeQueryApi            = api.ApiGroupApp.ExampleApiGroup.RechargeQueryApi
	userGameRecordApi           = api.ApiGroupApp.ExampleApiGroup.UserGameRecordApi
	withdrawAuditApi            = api.ApiGroupApp.ExampleApiGroup.WithdrawAuditApi
	dataStatisticsApi           = api.ApiGroupApp.ExampleApiGroup.DataStatisticsApi
	exaFileUploadAndDownloadApi = api.ApiGroupApp.ExampleApiGroup.FileUploadAndDownloadApi
)
