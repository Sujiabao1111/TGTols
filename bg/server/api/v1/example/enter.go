package example

import "github.com/flipped-aurora/gin-vue-admin/server/service"

type ApiGroup struct {
	CustomerApi
	AttachmentCategoryApi
	GamesApi
	AgentRebateConfigsApi
	BannersApi
	GameTypesApi
	GameProvidersApi
	DailyUserStatsApi
	StatsRetentionApi
	InviteStatsApi
	ManualRewardApi
	ActivityWagerConfigApi
	WithdrawFeeConfigApi
	GameLaunchErrorApi
	UserDataApi
	RechargeQueryApi
	UserGameRecordApi
	WithdrawAuditApi
	DataStatisticsApi
	FileUploadAndDownloadApi
}

var (
	customerService              = service.ServiceGroupApp.ExampleServiceGroup.CustomerService
	attachmentCategoryService    = service.ServiceGroupApp.ExampleServiceGroup.AttachmentCategoryService
	gamesService                 = service.ServiceGroupApp.ExampleServiceGroup.GamesService
	agentRebateConfigsService    = service.ServiceGroupApp.ExampleServiceGroup.AgentRebateConfigsService
	bannersService               = service.ServiceGroupApp.ExampleServiceGroup.BannersService
	gameTypesService             = service.ServiceGroupApp.ExampleServiceGroup.GameTypesService
	gameProvidersService         = service.ServiceGroupApp.ExampleServiceGroup.GameProvidersService
	dailyUserStatsService        = service.ServiceGroupApp.ExampleServiceGroup.DailyUserStatsService
	statsRetentionService        = service.ServiceGroupApp.ExampleServiceGroup.StatsRetentionService
	inviteStatsService           = service.ServiceGroupApp.ExampleServiceGroup.InviteStatsService
	manualRewardService          = service.ServiceGroupApp.ExampleServiceGroup.ManualRewardService
	activityWagerConfigService   = service.ServiceGroupApp.ExampleServiceGroup.ActivityWagerConfigService
	withdrawFeeConfigService     = service.ServiceGroupApp.ExampleServiceGroup.WithdrawFeeConfigService
	gameLaunchErrorService       = service.ServiceGroupApp.ExampleServiceGroup.GameLaunchErrorService
	userDataService              = service.ServiceGroupApp.ExampleServiceGroup.UserDataService
	rechargeQueryService         = service.ServiceGroupApp.ExampleServiceGroup.RechargeQueryService
	userGameRecordService        = service.ServiceGroupApp.ExampleServiceGroup.UserGameRecordService
	withdrawAuditService         = service.ServiceGroupApp.ExampleServiceGroup.WithdrawAuditService
	dataStatisticsService        = service.ServiceGroupApp.ExampleServiceGroup.DataStatisticsService
	fileUploadAndDownloadService = service.ServiceGroupApp.ExampleServiceGroup.FileUploadAndDownloadService
)
