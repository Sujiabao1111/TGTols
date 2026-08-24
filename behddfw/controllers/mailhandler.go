package controllers

import (
	"errors"
	"gogogo/models"
	"gogogo/models/dtos"
	"gogogo/models/requests"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 后端发送邮件
func SendMail(c *fiber.Ctx) error {
	var req requests.SendMailRequest
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(400)
	}

	if len(req.UserIDs) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "请选择收件人"})
	}

	return models.GetInstance().DbInstance.Transaction(func(tx *gorm.DB) error {
		// 1. 序列化附件
		attachJson, _ := json.Marshal(req.Attachments)

		// 2. 创建信件内容 (SysMail)
		mailContent := dtos.SysMail{
			Type:        1, // 个人信件
			Title:       req.Title,
			Content:     req.Content,
			Attachments: datatypes.JSON(attachJson),
			AdminID:     1, // 这里应获取当前登录管理员ID
		}
		if err := tx.Create(&mailContent).Error; err != nil {
			return err
		}

		// 3. 批量分发给用户 (UserInbox)
		var inboxItems []dtos.UserInbox

		// 判断初始领取状态: 如果没附件，状态直接设为 2 (无奖励)，否则 0 (未领取)
		initClaimStatus := 2
		if len(req.Attachments) > 0 {
			initClaimStatus = 0
		}

		for _, uid := range req.UserIDs {
			inboxItems = append(inboxItems, dtos.UserInbox{
				UserID:     uid,
				SysMailID:  mailContent.ID,
				IsRead:     0,
				IsClaimed:  initClaimStatus,
				ReceivedAt: time.Now(),
			})
		}

		// 批量插入
		if err := tx.Create(&inboxItems).Error; err != nil {
			return err
		}

		return nil
	})
}

// ClaimMailReward 用户领取邮件附件
func ClaimMailReward(c *fiber.Ctx) error {
	// 获取当前登录用户ID (从中间件JWT获取)
	userId := QueryUserIdFromJwt(c)
	mailId, _ := c.ParamsInt("id") // user_inbox 的 id

	var claimedRewards []dtos.MailAttachment

	// 开启事务
	err := models.GetInstance().DbInstance.Transaction(func(tx *gorm.DB) error {
		var inbox dtos.UserInbox

		// 1. 查询并锁定记录 (防止并发领取)
		// JOIN sys_mails 获取附件内容
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Mail").
			Where("id = ? AND user_id = ?", mailId, userId).
			First(&inbox).Error; err != nil {
			return errors.New("邮件不存在")
		}

		// 2. 校验状态
		if inbox.IsClaimed == 1 {
			return errors.New("奖励已领取")
		}
		if inbox.IsClaimed == 2 {
			return errors.New("该邮件无附件")
		}

		// 3. 解析附件
		var attachments []dtos.MailAttachment
		if err := json.Unmarshal(inbox.Mail.Attachments, &attachments); err != nil {
			return err
		}

		// 4. 发放奖励 (遍历附件列表)
		var user dtos.User
		if err := tx.First(&user, userId).Error; err != nil {
			return err
		}

		// 累加计算变动
		totalBalance := 0.0

		for _, item := range attachments {
			switch item.Type {
			case "balance": // 金币/余额
				totalBalance += item.Value

				// 记录账变日志
				tx.Create(&dtos.Transaction{
					UserID:        uint64(userId),
					Type:          6, // 假设6是邮件赠送
					Amount:        item.Value,
					BeforeBalance: user.Balance,
					AfterBalance:  user.Balance + totalBalance, // 注意这里逻辑，简单起见累加
					ReferenceID:   c.Params("id"),              // 关联邮件ID
					Remark:        "邮件附件领取: " + inbox.Mail.Title,
				})

			case "vip_exp": // VIP经验
				// logic to add vip exp...
			}
			claimedRewards = append(claimedRewards, item)
		}

		// 5. 更新用户余额
		if totalBalance > 0 {
			if err := tx.Model(&user).Update("balance", gorm.Expr("balance + ?", totalBalance)).Error; err != nil {
				return err
			}
		}

		// 6. 更新邮件状态为已领取
		if err := tx.Model(&inbox).Update("is_claimed", 1).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "领取成功",
		"data":    claimedRewards,
	})
}

// GetUserMails 获取用户邮件列表
func GetUserMails(c *fiber.Ctx) error {
	userId := QueryUserIdFromJwt(c)

	var mails []dtos.UserInbox
	// 预加载信件内容，按时间倒序
	models.GetInstance().DbInstance.Preload("Mail").
		Where("user_id = ?", userId).
		Order("id DESC").
		Limit(50).
		Find(&mails)

	return c.JSON(fiber.Map{"success": true, "data": mails})
}
