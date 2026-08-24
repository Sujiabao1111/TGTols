
// 自动生成模板AgentRebateConfigs
package example
import (
)

// agentRebateConfigs表 结构体  AgentRebateConfigs
type AgentRebateConfigs struct {
  Level  *int `json:"level" form:"level" gorm:"primarykey;comment:层级距离 (1表示直属上级, 2表示上上级...);column:level;size:10;"`  //层级距离 (1表示直属上级, 2表示上上级...)
  Rate  *float64 `json:"rate" form:"rate" gorm:"comment:返点比例 (如 0.0050 表示 0.5%);column:rate;size:5;"`  //返点比例 (如 0.0050 表示 0.5%)
}


// TableName agentRebateConfigs表 AgentRebateConfigs自定义表名 agent_rebate_configs
func (AgentRebateConfigs) TableName() string {
    return "agent_rebate_configs"
}





