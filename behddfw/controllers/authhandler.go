package controllers

import (
	"fmt"
	"gogogo/common"
	"gogogo/helpers"
	"gogogo/models"
	"gogogo/models/dtos"
	"gogogo/models/requests"
	"gogogo/models/responses"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func LoginHandler(c *fiber.Ctx) error {
	var req requests.LoginRequest

	// 1. 解析请求体
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.LoginResponse{
			Success: false,
			Message: "无效的请求数据",
		})
	}

	// 2. 简单的参数校验
	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(responses.LoginResponse{
			Success: false,
			Message: "用户名和密码不能为空",
		})
	}

	var user dtos.TagUser
	userid, err := strconv.Atoi(req.Username)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.LoginResponse{
			Success: false,
			Message: "用户id错误",
		})
	}
	// 3. 查询用户
	result := models.GetInstance().DbInstance.Where("id = ?", userid).First(&user)
	if result.Error != nil {
		// 为了安全，通常不提示“用户不存在”，而是提示“账号或密码错误”
		return c.Status(fiber.StatusUnauthorized).JSON(responses.LoginResponse{
			Success: false,
			Message: "账号或密码错误",
		})
	}

	// // 4. 校验密码
	if req.Password != user.Acc {
		return c.Status(fiber.StatusUnauthorized).JSON(responses.LoginResponse{
			Success: false,
			Message: "账号或密码错误",
		})
	}
	// if !common.CheckPasswordHash(req.Password, user.Password) {
	// 	return c.Status(fiber.StatusUnauthorized).JSON(responses.LoginResponse{
	// 		Success: false,
	// 		Message: "账号或密码错误",
	// 	})
	// }

	// // 5. 校验用户状态 (是否被封禁)
	// if user.Status == 0 {
	// 	return c.Status(fiber.StatusForbidden).JSON(responses.LoginResponse{
	// 		Success: false,
	// 		Message: "账号已被禁用，请联系客服",
	// 	})
	// }

	// 6. 生成 JWT
	token, err := common.GenerateJWT(uint64(user.ID), user.Acc, helpers.GetCfgInstance().Conf.Jwt)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.LoginResponse{
			Success: false,
			Message: "登录失败，系统错误",
		})
	}

	// // 异步记录今日登录活跃
	// go func() {
	// 	// 仅增加 login_count，金额传 0
	// 	RecordDailyStats(user.ID, 0, 0, 0, 0, 1)
	// }()

	// 7. 返回成功 (success=true)
	return c.JSON(responses.LoginResponse{
		Success: true,
		Message: "登录成功",
		Token:   token,
		Data: fiber.Map{
			"user_id":   user.ID,
			"username":  user.Nick,
			"vip_level": 1,
			"balance":   user.FarmCoin,
		},
	})
}

// Register 注册接口 (简单版)
func Registertest(c *fiber.Ctx) error {
	var req requests.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(400)
	}

	// 加密密码
	hashedPwd, _ := common.HashPassword(req.Password)

	newUser := dtos.User{
		Username: req.Username,
		Password: hashedPwd, // 存入数据库的是哈希值
		Status:   1,
		// ... 其他默认值
	}

	if err := models.GetInstance().DbInstance.Create(&newUser).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "用户名已存在或注册失败",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "注册成功",
	})
}

// Register 用户注册接口
func Register(c *fiber.Ctx) error {
	var req requests.RegisterRequest

	// 1. 解析参数
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "请求参数格式错误",
		})
	}

	// 2. 基础校验
	if len(req.Username) < 4 || len(req.Password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "用户名至少4位，密码至少6位",
		})
	}

	// 3. 检查用户名是否已存在
	var count int64
	models.GetInstance().DbInstance.Model(&dtos.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "用户名已被注册",
		})
	}

	// 4. 处理分销关系 (查找上级)
	var parent dtos.User
	hasParent := false

	// 优先匹配邀请码
	if req.InviteCode != "" {
		if err := models.GetInstance().DbInstance.Where("invite_code = ?", req.InviteCode).First(&parent).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"message": "无效的邀请码", // 明确告知邀请码错误，而不是默默注册为无上级
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "系统繁忙"})
		}
		hasParent = true
	} else if req.ParentID > 0 {
		// 其次匹配 ParentID
		if err := models.GetInstance().DbInstance.First(&parent, req.ParentID).Error; err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "无效的推荐人ID",
			})
		}
		hasParent = true
	}

	// 5. 准备新用户数据
	hashedPwd, _ := common.HashPassword(req.Password)

	newUser := dtos.User{
		Username:   req.Username,
		Password:   hashedPwd,
		Status:     1,
		Balance:    0,
		VipLevel:   0,
		InviteCode: common.GenerateUniqueCode(), // 生成32位唯一邀请码
	}

	// 6. 设置层级与路径 (核心逻辑)
	if hasParent {
		newUser.ParentID = parent.ID
		newUser.Level = parent.Level + 1

		// 拼接路径: 父级路径 + 父级ID + "/"
		// 逻辑: 如果父级是顶层(Path=""), 新用户Path="5/" (假设父ID=5)
		// 如果父级是二级(Path="5/"), 新用户Path="5/12/" (假设父ID=12)
		pathPrefix := parent.Path
		if pathPrefix == "" {
			// 父级是根节点，之前没有路径
			// 注意：这里路径逻辑可以根据你的偏好定，通常建议顶层用户的Path为空字符串或者特定标识
			// 如果parent.ID是1，新Path就是 "1/"
			newUser.Path = fmt.Sprintf("%d/", parent.ID)
		} else {
			// 父级已有路径，追加父级ID
			newUser.Path = fmt.Sprintf("%s%d/", parent.Path, parent.ID)
		}
	} else {
		// 无上级，设为顶层
		newUser.ParentID = 0
		newUser.Level = 1
		newUser.Path = ""
	}

	// 7. 写入数据库
	// 这里存在极小概率 InviteCode 重复，生产环境可以加个简单的重试循环
	if err := models.GetInstance().DbInstance.Create(&newUser).Error; err != nil {
		// 检查是否是邀请码冲突 (MySQL Error 1062)
		// 简单处理：如果是系统错误返回给前端
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "注册失败: " + err.Error(),
		})
	}

	token, err := common.GenerateJWT(newUser.ID, newUser.Username, helpers.GetCfgInstance().Conf.Jwt)
	if err != nil {
		// 极端情况：用户创建成功但Token生成失败，这种情况下通常确保如果在生成 Token 步骤失败（虽然极少），或者后续逻辑出错时能回滚，
		// 不过对于让前端引导用户手动登录
		return c.JSON(fiber.Map{
			"success": true,
			"message": "注册成功，但自动登录失败，请手动登录",
			"token":   "",
		})
	}

	// 异步记录今日登录活跃
	go func() {
		// 仅增加 login_count，金额传 0
		RecordDailyStats(newUser.ID, 0, 0, 0, 0, 1)
	}()

	// 8. 成功响应
	return c.JSON(fiber.Map{
		"success": true,
		"message": "注册成功",
		"token":   token,
		"data": fiber.Map{
			"user_id":     newUser.ID,
			"username":    newUser.Username,
			"invite_code": newUser.InviteCode,
			"level":       newUser.Level,
		},
	})
}
