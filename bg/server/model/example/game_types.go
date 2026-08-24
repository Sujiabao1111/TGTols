
// 自动生成模板GameTypes
package example
import (
)

// gameTypes表 结构体  GameTypes
type GameTypes struct {
  Id  *int `json:"id" form:"id" gorm:"primarykey;column:id;size:10;"`  //id字段
  Name  *string `json:"name" form:"name" gorm:"comment:分类名称 (电子, 真人, 体育);column:name;size:50;"`  //分类名称 (电子, 真人, 体育)
  Sort  *int `json:"sort" form:"sort" gorm:"comment:排序权重;column:sort;size:10;"`  //排序权重
  Code  *string `json:"code" form:"code" gorm:"comment:代码字符串;column:code;size:32;"`  //代码字符串
  Status  *bool `json:"status" form:"status" gorm:"comment:1:启用 0:禁用;column:status;"`  //1:启用 0:禁用
}


// TableName gameTypes表 GameTypes自定义表名 game_types
func (GameTypes) TableName() string {
    return "game_types"
}





