// 自动生成模板GameProviders
package example

import (
	"gorm.io/datatypes"
	"time"
)

// gameProviders表 结构体  GameProviders
type GameProviders struct {
	Id           *int           `json:"id" form:"id" gorm:"primarykey;column:id;size:10;"`                                                                        //id字段
	PlatformCode *string        `json:"platformCode" form:"platformCode" gorm:"comment:游戏平台代码;column:platform_code;size:32;"`                                     //游戏平台代码
	Code         *string        `json:"code" form:"code" gorm:"comment:厂商代码 (如 PG, AG);column:code;size:50;"`                                                     //厂商代码 (如 PG, AG)
	Name         *string        `json:"name" form:"name" gorm:"comment:厂商显示名称;column:name;size:100;"`                                                             //厂商显示名称
	ApiConfig    datatypes.JSON `json:"apiConfig" form:"apiConfig" gorm:"comment:JSON格式存储API Key, Secret, Endpoint等敏感信息;column:api_config;" swaggertype:"object"` //JSON格式存储API Key, Secret, Endpoint等敏感信息
	Status       *bool          `json:"status" form:"status" gorm:"comment:1:开启 0:维护;column:status;"`                                                             //1:开启 0:维护
	CreatedAt    *time.Time     `json:"createdAt" form:"createdAt" gorm:"column:created_at;"`                                                                     //createdAt字段
	TypeId       *int           `json:"typeId" form:"typeId" gorm:"comment:游戏类型Id;column:type_id;size:10;"`                                                       //游戏类型Id
}

// TableName gameProviders表 GameProviders自定义表名 game_providers
func (GameProviders) TableName() string {
	return "game_providers"
}
