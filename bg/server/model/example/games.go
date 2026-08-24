// 自动生成模板Games
package example

import (
	"gorm.io/datatypes"
	"time"
)

// games表 结构体  Games
type Games struct {
	Id           *int           `json:"id" form:"id" gorm:"primarykey;column:id;size:20;"`                                                  //id字段
	PlatformCode *string        `json:"platformCode" form:"platformCode" gorm:"comment:游戏平台代码;column:platform_code;size:32;"`               //游戏平台代码
	ProviderId   *int           `json:"providerId" form:"providerId" gorm:"comment:关联厂商ID (game_providers.id);column:provider_id;size:10;"` //关联厂商ID (game_providers.id)
	GameTypeId   *int           `json:"gameTypeId" form:"gameTypeId" gorm:"comment:关联分类ID (game_types.id);column:game_type_id;size:10;"`    //关联分类ID (game_types.id)
	GameCode     *string        `json:"gameCode" form:"gameCode" gorm:"comment:厂商侧的游戏ID;column:game_code;size:100;"`                        //厂商侧的游戏ID
	Name         datatypes.JSON `json:"name" form:"name" gorm:"comment:游戏名称(多语言);column:name;" swaggertype:"object"`                        //游戏名称(多语言)
	ImgUrl       *string        `json:"imgUrl" form:"imgUrl" gorm:"comment:封面图片URL;column:img_url;size:255;"`                               //封面图片URL
	Views        *int           `json:"views" form:"views" gorm:"comment:点击/热度;column:views;size:19;"`                                      //点击/热度
	Sort         *int           `json:"sort" form:"sort" gorm:"comment:排序, 越大越前;column:sort;size:10;"`                                      //排序, 越大越前
	Status       *bool          `json:"status" form:"status" gorm:"comment:1:上架 0:下架;column:status;"`                                       //1:上架 0:下架
	CreatedAt    *time.Time     `json:"createdAt" form:"createdAt" gorm:"column:created_at;"`                                               //createdAt字段
	UpdatedAt    *time.Time     `json:"updatedAt" form:"updatedAt" gorm:"column:updated_at;"`                                               //updatedAt字段
}

// TableName games表 Games自定义表名 games
func (Games) TableName() string {
	return "games"
}
