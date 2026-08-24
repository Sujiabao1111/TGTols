
// 自动生成模板Banners
package example
import (
	"time"
	"gorm.io/datatypes"
)

// banners表 结构体  Banners
type Banners struct {
  Id  *int `json:"id" form:"id" gorm:"primarykey;column:id;size:10;"`  //id字段
  Title  *string `json:"title" form:"title" gorm:"comment:对应前端 key, 如 hero.slide1.title;column:title;size:255;"`  //对应前端 key, 如 hero.slide1.title
  Subtitle  *string `json:"subtitle" form:"subtitle" gorm:"comment:副标题;column:subtitle;size:255;"`  //副标题
  Tag  *string `json:"tag" form:"tag" gorm:"comment:标签;column:tag;size:100;"`  //标签
  Image  *string `json:"image" form:"image" gorm:"comment:图片URL;column:image;size:500;"`  //图片URL
  Color  *string `json:"color" form:"color" gorm:"comment:Tailwind 渐变色类名;column:color;size:255;"`  //Tailwind 渐变色类名
  JumpLink  *string `json:"jumpLink" form:"jumpLink" gorm:"comment:整图跳转链接 (形式2);column:jump_link;size:255;"`  //整图跳转链接 (形式2)
  Buttons  datatypes.JSON `json:"buttons" form:"buttons" gorm:"comment:按钮配置数组 (形式1);column:buttons;" swaggertype:"object"`  //按钮配置数组 (形式1)
  Sort  *int `json:"sort" form:"sort" gorm:"comment:排序;column:sort;size:10;"`  //排序
  Status  *bool `json:"status" form:"status" gorm:"comment:1:启用 0:禁用;column:status;"`  //1:启用 0:禁用
  CreatedAt  *time.Time `json:"createdAt" form:"createdAt" gorm:"column:created_at;"`  //createdAt字段
  UpdatedAt  *time.Time `json:"updatedAt" form:"updatedAt" gorm:"column:updated_at;"`  //updatedAt字段
}


// TableName banners表 Banners自定义表名 banners
func (Banners) TableName() string {
    return "banners"
}





