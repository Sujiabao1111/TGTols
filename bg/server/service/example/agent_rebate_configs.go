
package example

import (
	"context"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/example"
    exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
)

type AgentRebateConfigsService struct {}
// CreateAgentRebateConfigs 创建agentRebateConfigs表记录
// Author [yourname](https://github.com/yourname)
func (agentRebateConfigsService *AgentRebateConfigsService) CreateAgentRebateConfigs(ctx context.Context, agentRebateConfigs *example.AgentRebateConfigs) (err error) {
	err = global.GVA_DB.Create(agentRebateConfigs).Error
	return err
}

// DeleteAgentRebateConfigs 删除agentRebateConfigs表记录
// Author [yourname](https://github.com/yourname)
func (agentRebateConfigsService *AgentRebateConfigsService)DeleteAgentRebateConfigs(ctx context.Context, level string) (err error) {
	err = global.GVA_DB.Delete(&example.AgentRebateConfigs{},"level = ?",level).Error
	return err
}

// DeleteAgentRebateConfigsByIds 批量删除agentRebateConfigs表记录
// Author [yourname](https://github.com/yourname)
func (agentRebateConfigsService *AgentRebateConfigsService)DeleteAgentRebateConfigsByIds(ctx context.Context, levels []string) (err error) {
	err = global.GVA_DB.Delete(&[]example.AgentRebateConfigs{},"level in ?",levels).Error
	return err
}

// UpdateAgentRebateConfigs 更新agentRebateConfigs表记录
// Author [yourname](https://github.com/yourname)
func (agentRebateConfigsService *AgentRebateConfigsService)UpdateAgentRebateConfigs(ctx context.Context, agentRebateConfigs example.AgentRebateConfigs) (err error) {
	err = global.GVA_DB.Model(&example.AgentRebateConfigs{}).Where("level = ?",agentRebateConfigs.Level).Updates(&agentRebateConfigs).Error
	return err
}

// GetAgentRebateConfigs 根据level获取agentRebateConfigs表记录
// Author [yourname](https://github.com/yourname)
func (agentRebateConfigsService *AgentRebateConfigsService)GetAgentRebateConfigs(ctx context.Context, level string) (agentRebateConfigs example.AgentRebateConfigs, err error) {
	err = global.GVA_DB.Where("level = ?", level).First(&agentRebateConfigs).Error
	return
}
// GetAgentRebateConfigsInfoList 分页获取agentRebateConfigs表记录
// Author [yourname](https://github.com/yourname)
func (agentRebateConfigsService *AgentRebateConfigsService)GetAgentRebateConfigsInfoList(ctx context.Context, info exampleReq.AgentRebateConfigsSearch) (list []example.AgentRebateConfigs, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
    // 创建db
	db := global.GVA_DB.Model(&example.AgentRebateConfigs{})
    var agentRebateConfigss []example.AgentRebateConfigs
    // 如果有条件搜索 下方会自动创建搜索语句
    
	err = db.Count(&total).Error
	if err!=nil {
    	return
    }

	if limit != 0 {
       db = db.Limit(limit).Offset(offset)
    }

	err = db.Find(&agentRebateConfigss).Error
	return  agentRebateConfigss, total, err
}
func (agentRebateConfigsService *AgentRebateConfigsService)GetAgentRebateConfigsPublic(ctx context.Context) {
    // 此方法为获取数据源定义的数据
    // 请自行实现
}
