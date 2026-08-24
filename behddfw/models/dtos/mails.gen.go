package dtos

import (
	"time"

	"gorm.io/datatypes"
)

// 定义附件的结构 (用于Go代码解析)
type MailAttachment struct {
	Type  string  `json:"type"`   // 奖励类型: balance(余额), vip_exp(经验), item(道具)
	Value float64 `json:"value"`  // 数量
	RefID uint    `json:"ref_id"` // 如果是道具，对应 item_id
}

// SysMail 信件内容
type SysMail struct {
	ID          uint64         `gorm:"primaryKey" json:"id"`
	Type        int            `gorm:"default:1" json:"type"` // 1:个人 2:广播
	Title       string         `gorm:"size:100" json:"title"`
	Content     string         `gorm:"type:text" json:"content"`
	Attachments datatypes.JSON `gorm:"type:json" json:"attachments"` // 存 MailAttachment 数组
	AdminID     uint64         `json:"admin_id"`
	CreatedAt   time.Time      `json:"created_at"`
}

// UserInbox 用户收件箱
type UserInbox struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	UserID     uint64    `gorm:"index:idx_user_status" json:"user_id"`
	SysMailID  uint64    `gorm:"index:idx_user_mail" json:"sys_mail_id"`
	IsRead     int       `gorm:"default:0" json:"is_read"`    // 0:未读 1:已读
	IsClaimed  int       `gorm:"default:0" json:"is_claimed"` // 0:未领 1:已领 2:无奖励
	ReceivedAt time.Time `json:"received_at"`

	// 关联查询
	Mail *SysMail `gorm:"foreignKey:SysMailID" json:"mail,omitempty"`
}

func (UserInbox) TableName() string {
	return "user_inbox"
}
