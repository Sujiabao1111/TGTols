package main

import (
	"log"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/core"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/initialize"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
)

//go:generate go env -w GO111MODULE=on
//go:generate go env -w GOPROXY=https://goproxy.cn,direct
//go:generate go mod tidy
//go:generate go mod download

// @Tag.Name        Base
// @Tag.Name        SysUser
// @Tag.Description 用户

// @title                       Gin-Vue-Admin Swagger API接口文档
// @version                     v2.8.2
// @description                 使用gin+vue进行极速开发的全栈开发基础平台
// @securityDefinitions.apikey  ApiKeyAuth
// @in                          header
// @name                        x-token
// @BasePath                    /
func main() {
	initializeSystem()
	core.RunServer()
}

func initializeSystem() {
	configureRuntimeTimezone()
	global.GVA_VP = core.Viper()
	initialize.OtherInit()
	global.GVA_LOG = core.Zap()
	zap.ReplaceGlobals(global.GVA_LOG)
	global.GVA_DB = initialize.Gorm()
	initialize.Timer()
	initialize.DBList()
	initialize.SetupHandlers()
	if global.GVA_DB != nil {
		initialize.RegisterTables()
		initialize.EnsureInviteManagementResources()
		initialize.EnsureRewardManagementResources()
		initialize.EnsureGameLaunchErrorResources()
		initialize.EnsureUserDataResources()
		initialize.EnsureRechargeQueryResources()
		initialize.EnsureUserGameRecordResources()
		initialize.EnsureWithdrawAuditResources()
		initialize.EnsureDataStatisticsResources()
		initialize.EnsureDomainStatsResources()
	}
}

func configureRuntimeTimezone() {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		log.Printf("[Setup] failed to load Asia/Shanghai timezone: %v", err)
		return
	}

	time.Local = loc
}
