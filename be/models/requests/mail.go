package requests

import "gogogo/models/dtos"

// SendMailRequest 发送邮件请求
type SendMailRequest struct {
	UserIDs     []uint64              `json:"user_ids"` // 接收者ID列表
	Title       string                `json:"title"`
	Content     string                `json:"content"`
	Attachments []dtos.MailAttachment `json:"attachments"` // 附件列表
}
