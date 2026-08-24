package controllers

import (
	stdjson "encoding/json"
	"fmt"
	"gogogo/common"
	"gogogo/helpers"
	"gogogo/models"
	"gogogo/models/dtos"
	"gogogo/models/requests"
	"gogogo/models/responses"
	"gogogo/provider"
	"log"
	"math"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const gameListSyncInterval = 15 * time.Minute
const debugGameProviderID uint = 23

var gameListSyncState = struct {
	mu          sync.Mutex
	running     bool
	done        chan struct{}
	lastSuccess time.Time
}{}

// SyncAllGames 执行同步任务
func SyncAllGames() error {
	log.Println("[Sync] Starting game synchronization...")

	// 1. 获取所有状态开启的厂商
	var providers []dtos.GameProvider
	if err := models.GetInstance().DbInstance.Where("status = ?", 1).Find(&providers).Error; err != nil {
		log.Printf("[Sync] Failed to fetch providers: %v\n", err)
		return err
	}

	var debugProvider dtos.GameProvider
	debugProviderFound := models.GetInstance().DbInstance.Select("id, code, name, status").First(&debugProvider, debugGameProviderID).Error == nil
	if debugProviderFound {
		log.Printf("[Sync][Provider %d] DB record found: code=%s name=%s status=%d\n", debugProvider.ID, debugProvider.Code, debugProvider.Name, debugProvider.Status)
	} else {
		log.Printf("[Sync][Provider %d] DB record not found\n", debugGameProviderID)
	}

	debugProviderIncluded := false
	for _, p := range providers {
		if p.ID == debugGameProviderID {
			debugProviderIncluded = true
			break
		}
	}
	if debugProviderIncluded {
		log.Printf("[Sync][Provider %d] Included in enabled providers list, will fetch games this round\n", debugGameProviderID)
	} else {
		log.Printf("[Sync][Provider %d] Not included in enabled providers list, current sync only fetches providers with status=1\n", debugGameProviderID)
	}

	// 2. 预加载所有游戏分类
	var gameTypes []dtos.GameType
	// 查询所有分类 (不管状态如何先查出来，后面在内存里判断)
	models.GetInstance().DbInstance.Find(&gameTypes)

	// Set: ID(uint) -> bool (是否存在)
	validTypeIds := make(map[uint]bool)
	// 兜底 ID (如果 API 返回的 Type 在本地找不到，归类到 Other)
	defaultTypeId := uint(0)

	for _, t := range gameTypes {
		// 只有 Status = 1 (启用) 的分类才被视为有效目标
		if t.Status == 1 {
			validTypeIds[t.ID] = true
			// 寻找默认分类 (例如 Code 为 OTHER 或 ID 为 1)
			if t.Code == "OTHER" || defaultTypeId == 0 {
				defaultTypeId = t.ID
			}
		}
	}

	client := provider.NewHedocClient()

	// 3. 遍历厂商进行同步
	for _, p := range providers {
		log.Printf("[Sync] Fetching games for provider: %s (ID: %d)\n", p.Name, p.ID)

		// 调用 API (传入 ID)
		externalGames, err := client.FetchGameList(int(p.ID), helpers.GetCfgInstance().Conf.Agentid, helpers.GetCfgInstance().Conf.Agentapi)
		if err != nil {
			if p.ID == debugGameProviderID {
				log.Printf("[Sync][Provider %d] FetchGameList failed: %v\n", p.ID, err)
			}
			log.Printf("[Sync] Error fetching provider %d: %v\n", p.ID, err)
			continue
		}
		if p.ID == debugGameProviderID {
			log.Printf("[Sync][Provider %d] FetchGameList succeeded, received %d games\n", p.ID, len(externalGames))
		}

		if len(externalGames) == 0 {
			if p.ID == debugGameProviderID {
				log.Printf("[Sync][Provider %d] Third-party returned 0 games\n", p.ID)
			}
			continue
		}

		var gamesToUpsert []dtos.Game

		// 4. 数据转换与整理
		for _, item := range externalGames {
			// 4.1 映射分类
			targetTypeId := uint(item.Type)
			// 校验 ID 是否有效，无效则使用兜底 ID
			if !validTypeIds[targetTypeId] {
				// 如果不想用兜底，而是直接跳过该游戏，可以使用 continue
				targetTypeId = defaultTypeId
			}

			// 4.2 处理多语言 Name (Map -> JSON)
			nameJsonBytes, _ := json.Marshal(item.Name)

			// 4.3 构造本地模型
			game := dtos.Game{
				ProviderID: p.ID,         // 使用本地循环的 ID
				GameTypeID: targetTypeId, // 映射后的分类 ID
				GameCode:   item.Code,    // 唯一标识
				Name:       datatypes.JSON(nameJsonBytes),
				ImgUrl:     item.GameIcon,
				Sort:       item.Order, // 使用 API 返回的排序
				Status:     1,          // 默认上架
			}
			gamesToUpsert = append(gamesToUpsert, game)
		}

		// 5. 批量 Upsert (存在则更新，不存在则插入)
		// 需要 game_code 上有唯一索引
		if len(gamesToUpsert) > 0 {
			err := models.GetInstance().DbInstance.Clauses(clause.OnConflict{
				// 修改关键点: 冲突检测基于 [provider_id, game_code] 组合
				Columns: []clause.Column{
					{Name: "provider_id"},
					{Name: "game_code"},
				},
				// 如果冲突（即该厂商下已存在该 Code），则更新以下字段
				DoUpdates: clause.AssignmentColumns([]string{
					"name",
					"img_url",
					"sort",
					"game_type_id",
					"updated_at",
					// 注意: 不更新 views, status (防止覆盖运营手动下架或热度数据)
				}),
			}).CreateInBatches(&gamesToUpsert, 100).Error

			if err != nil {
				if p.ID == debugGameProviderID {
					log.Printf("[Sync][Provider %d] DB upsert failed: %v\n", p.ID, err)
				}
				log.Printf("[Sync] DB Error for provider %d: %v\n", p.ID, err)
			} else {
				if p.ID == debugGameProviderID {
					log.Printf("[Sync][Provider %d] DB upsert succeeded, synced %d games\n", p.ID, len(gamesToUpsert))
				}
				log.Printf("[Sync] Provider %d: Synced %d games.\n", p.ID, len(gamesToUpsert))
			}
		}
	}
	log.Println("[Sync] Synchronization finished.")
	return nil
}

func SyncAllGamesIfStale(maxAge time.Duration) error {
	if maxAge <= 0 {
		maxAge = gameListSyncInterval
	}

	for {
		gameListSyncState.mu.Lock()
		if !gameListSyncState.lastSuccess.IsZero() && time.Since(gameListSyncState.lastSuccess) < maxAge {
			gameListSyncState.mu.Unlock()
			return nil
		}

		if gameListSyncState.running {
			done := gameListSyncState.done
			gameListSyncState.mu.Unlock()
			<-done
			continue
		}

		done := make(chan struct{})
		gameListSyncState.running = true
		gameListSyncState.done = done
		gameListSyncState.mu.Unlock()

		err := SyncAllGames()

		gameListSyncState.mu.Lock()
		if err == nil {
			gameListSyncState.lastSuccess = time.Now()
		}
		gameListSyncState.running = false
		gameListSyncState.done = nil
		close(done)
		gameListSyncState.mu.Unlock()

		return err
	}
}

// 同步三方游戏
func SyncGames(c *fiber.Ctx) error {
	client := provider.NewHedocClient()
	games, err := client.GetGameList()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// 批量插入或更新到本地 games 表
	for _, g := range games {
		var gameTypeID uint = 1 // 需根据 g.Type 映射本地分类ID

		models.GetInstance().DbInstance.Where(dtos.Game{GameCode: g.GameCode}).
			Assign(dtos.Game{
				// Name:       g.GameName,
				ImgUrl:     g.ImgUrl,
				ProviderID: 1, // 假设 Hedoc ID=1
				GameTypeID: gameTypeID,
			}).
			FirstOrCreate(&dtos.Game{})
	}

	return c.JSON(fiber.Map{"success": true, "count": len(games)})
}

// 获取三方的游戏
func ListGames(c *fiber.Ctx) error {
	client := provider.NewHedocClient()
	games, err := client.GetGameList()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "count": len(games), "games": games})
}

// 请求进入游戏
func LaunchGame(c *fiber.Ctx) error {
	var req requests.OpenGamePayload
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "Invalid params"})
	}
	userId := QueryUserIdFromJwt(c)
	thirdPartyLoginId := fmt.Sprintf("%su%d", helpers.GetCfgInstance().Conf.Prefix, userId)
	// 获得用户专属的第三方密码
	thirdPartyPassword := common.GenerateGamePassword(uint64(userId))
	// 1. 根据 GameCode 查找游戏信息 (获取 ProviderID)
	var game dtos.Game
	if err := models.GetInstance().DbInstance.Preload("Provider").Where("game_code = ?", req.GameCode).First(&game).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "Game not found"})
	}

	client := provider.NewHedocClient()

	// 2. 确保用户已注册
	// _ = client.RegisterUser(thirdPartyLoginId, "Game@123456")

	// 3. 请求链接
	url, err := client.OpenGame(thirdPartyLoginId, thirdPartyPassword, int(game.Provider.ID), req.GameCode, req.IsMobile, req.Language,
		helpers.GetCfgInstance().Conf.Agentapi,
		helpers.GetCfgInstance().Conf.Agentid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}

	// 4. (可选) 增加热度
	models.GetInstance().DbInstance.Model(&game).UpdateColumn("views", gorm.Expr("views + ?", 1))

	return c.JSON(fiber.Map{"code": 0, "message": "success", "data": fiber.Map{"url": url}})
}

// 请求进入游戏 - 钱包自动带入
/*

智能免转 (Smart Auto-Transfer)
核心逻辑：
进入游戏时：后端检测用户在本地的余额。如果有钱，自动发起 TransferIn 请求，将本地余额全部转入该三方厂商，然后返回游戏链接。
退出/切换时：由于 Web 端很难捕获准确的“退出”事件，我们在前端首页/个人中心增加一个 “一键回收” (Recycle / Refresh) 按钮。
进阶逻辑（可选）：在“进入游戏”接口中，先检查该用户上次玩的是哪个厂商，先把那个厂商的钱取出来（归集），再转入当前要玩的厂商。

进游戏时：如果本地没钱，我们假设钱已经在三方了（或者用户就是没钱），直接放行进游戏。
如果用户在三方有钱，进游戏后能直接玩；如果两边都没钱，用户进游戏后余额为0，逻辑也是对的。

回收时：每次回收都是实时的去问三方：“你那里有这个人的钱吗？”。如果有，就全部取回来。这解决了“掉单”或“数据不一致”的问题。

用户操作简便：
用户充值到平台 -> 点击任意游戏 -> 自动转入。
用户玩完 -> 点击“一键回收” -> 自动转出到平台 -> 提现。

*/
func LaunchGame2(c *fiber.Ctx) error {
	var req requests.OpenGamePayload
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "Invalid params"})
	}
	userId := QueryUserIdFromJwt(c)
	thirdPartyLoginId := fmt.Sprintf("%su%d", helpers.GetCfgInstance().Conf.Prefix, userId)
	// 获得用户专属的第三方密码
	thirdPartyPassword := common.GenerateGamePassword(uint64(userId))
	// 1. 根据 GameCode 查找游戏信息 (获取 ProviderID)
	var game dtos.Game
	if err := models.GetInstance().DbInstance.Preload("Provider").Where("game_code = ?", req.GameCode).First(&game).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "Game not found"})
	}

	client := provider.NewHedocClient()

	// 2. 确保用户已注册
	// _ = client.RegisterUser(thirdPartyLoginId, "Game@123456")
	_ = client.RegisterUser(helpers.GetCfgInstance().Conf.Agentid, thirdPartyLoginId, thirdPartyPassword, helpers.GetCfgInstance().Conf.Agentapi)

	// ==========================================
	// 核心：智能自动转入 (Smart Auto-In)
	// ==========================================
	// =================================================================
	// 步骤 A: 先查询用户本地余额 (为了知道要扣多少)
	// =================================================================

	// =================================================================
	// 步骤 B: 核心同步卡点 —— 请求 12661 扣款
	// =================================================================
	// 调用我们刚才写的同步函数
	// 如果这里报错，函数直接 return，后续的 OpenGame、三方转账统统不会发生
	// 这里先简单查一下12661余额，不用锁，因为真正的锁在后面或者在 12661 那边
	balance, err := SyncGameBalance(uint64(userId), 0, 1)
	if err != nil {
		// 阻断流程！返回错误给前端
		return c.Status(500).JSON(fiber.Map{
			"code":    503,
			"message": "资金同步失败，请重试: " + err.Error(),
		})
	}
	if balance > 0 {
		// =================================================================
		// 步骤 C: 12661 扣款成功后，处理 Web 端本地数据和三方转账
		// =================================================================
		// 此时可以认为钱已经从“游戏进程”扣除了，我们现在要在 Web 数据库记录流水并转入三方
		// 记录流水
		billNo := fmt.Sprintf("AUTO-IN-%d-%d", time.Now().UnixNano(), userId)
		trans := dtos.Transaction{
			UserID: uint64(userId), Type: 3, Amount: -float64(balance),
			ReferenceID: billNo, Remark: "同步扣款转入游戏",
		}
		if err := models.GetInstance().DbInstance.Create(&trans).Error; err != nil {
			return err
		}

		// 调用三方 (Hedoc) 加钱
		// 风险提示：如果这里失败了，12661 那边已经扣钱了。
		// 生产环境需要：如果这里失败，发起一个“补偿请求”给 12661 把钱加回去 (冲正)
		_, err = client.UpdateBalance(helpers.GetCfgInstance().Conf.Agentid,
			thirdPartyLoginId, float64(balance), 1, billNo, helpers.GetCfgInstance().Conf.Agentapi)
		if err != nil {
			SyncGameBalance(uint64(userId), -float64(balance), 2)
			return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
		}
	}

	// 3. 请求链接
	url, err := client.OpenGame(thirdPartyLoginId, thirdPartyPassword, int(game.Provider.ID), req.GameCode, req.IsMobile, req.Language,
		helpers.GetCfgInstance().Conf.Agentapi,
		helpers.GetCfgInstance().Conf.Agentid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}

	// 4. (可选) 增加热度
	models.GetInstance().DbInstance.Model(&game).UpdateColumn("views", gorm.Expr("views + ?", 1))

	return c.JSON(fiber.Map{"code": 0, "message": "success", "data": fiber.Map{"url": url}})
}

// RecycleBalance 一键回收资金 (从三方统一钱包转回本地)
func RecycleBalance(c *fiber.Ctx) error {
	userId := QueryUserIdFromJwt(c)
	thirdPartyLoginId := fmt.Sprintf("%su%d", helpers.GetCfgInstance().Conf.Prefix, userId)
	client := provider.NewHedocClient()

	var recycledAmount float64 = 0
	var currentLocalBalance float64 = 0 // 用于存储最终的本地余额
	// 1. 先查询三方余额 (这一步不需要事务，因为只读)
	resp, err := client.GetBalance(helpers.GetCfgInstance().Conf.Agentid, thirdPartyLoginId, helpers.GetCfgInstance().Conf.Agentapi)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "无法查询三方余额，请稍后重试"})
	}

	// 如果三方没有余额，直接返回成功，不做任何操作
	if resp.Balance <= 0 {
		var user dtos.TagUser
		if err := models.GetInstance().DbInstance.Select("farm_coin").First(&user, userId).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"code": 500, "message": "获取用户信息失败"})
		}
		currentLocalBalance = float64(user.FarmCoin)
		return c.JSON(fiber.Map{
			"code":    0,
			"message": "success",
			"data":    fiber.Map{"recycled_amount": 0, "current_balance": currentLocalBalance},
		})
	}

	// =================================================================
	// 核心同步卡点 —— 请求 12661 加回款
	// =================================================================
	var user dtos.TagUser
	// 这里先简单查一下，不用锁，因为真正的锁在后面或者在 12661 那边
	if err := models.GetInstance().DbInstance.First(&user, userId).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "用户异常"})
	}

	// A. 调用三方提款
	// 将查到的余额全部转出
	billNo := fmt.Sprintf("AUTO-OUT-%d-%d", time.Now().UnixNano(), userId)
	// Type=2: 转出 (Withdraw)
	// 注意：这里使用查询到的 resp.Balance 进行全额回收
	_, err = client.UpdateBalance(helpers.GetCfgInstance().Conf.Agentid,
		thirdPartyLoginId, resp.Balance, 2, billNo, helpers.GetCfgInstance().Conf.Agentapi)
	if err != nil {
		return fmt.Errorf("资金回收失败: %v", err)
	}

	// B. 本地加款
	trans := dtos.Transaction{
		UserID: uint64(userId), Type: 3, Amount: resp.Balance,
		BeforeBalance: float64(user.FarmCoin),
		AfterBalance:  float64(user.FarmCoin) + resp.Balance,
		ReferenceID:   billNo,
		Remark:        "一键回收资金(统一钱包)",
	}

	user.FarmCoin += int64(resp.Balance)

	if err := models.GetInstance().DbInstance.Create(&trans).Error; err != nil {
		return err
	}

	recycledAmount = resp.Balance
	// 1usd=15000idr,1usd=1000 farm_coin,1farm coin=15idr,
	// here is idr to farm_coin
	currentLocalBalance = float64(resp.Balance) / 15.0

	_, err = SyncGameBalance(uint64(userId), -float64(resp.Balance), 2)
	if err != nil {
		// 阻断流程！返回错误给前端
		return c.Status(500).JSON(fiber.Map{
			"code":    503,
			"message": "资金同步失败，请重试: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data": fiber.Map{
			"recycled_amount": recycledAmount,
			"current_balance": currentLocalBalance, // 用户当前在本地的总余额
		},
	})
}

// 资金转动
func TransferToProvider(c *fiber.Ctx) error {
	var req requests.TransferRequest
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(400)
	}

	userId := QueryUserIdFromJwt(c)
	// username := QueryUserNameFromJwt(c)

	tx := models.GetInstance().DbInstance.Begin() // 开启事务

	// 1. 锁定用户并在本地处理资金
	var user dtos.User
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userId).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"msg": "系统繁忙"})
	}

	// 创建订单号
	refId := fmt.Sprintf("TR-%d-%d", time.Now().UnixNano(), userId)

	// 记录账变 (Transaction 表)
	trans := dtos.Transaction{
		UserID:      uint64(userId),
		Amount:      req.Amount, // 注意正负
		ReferenceID: refId,
		Type:        3, // 3=上下分
	}

	if req.Type == 1 {
		// 转入游戏：本地扣钱
		if user.Balance < req.Amount {
			tx.Rollback()
			return c.Status(400).JSON(fiber.Map{"msg": "余额不足"})
		}
		trans.Amount = -req.Amount
		trans.Remark = "转入游戏平台"
		user.Balance -= req.Amount
	} else {
		// 转出游戏：本地加钱 (注意：实际应该先查第三方余额够不够，这里简化直接调API)
		trans.Amount = req.Amount
		trans.Remark = "从游戏平台转出"
		user.Balance += req.Amount
	}

	// 更新本地数据库
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Create(&trans).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 2. 调用第三方 API
	client := provider.NewHedocClient()
	// 如果是网络错误，这里其实比较危险。生产环境通常是：
	// 先提交"处理中"状态，异步去调三方，或者在这里调，超时则回滚。
	// 这里演示同步调用，失败回滚模式：
	err := client.Transfer(userId, req.Amount, req.Type, refId)

	if err != nil {
		tx.Rollback() // API失败，回滚本地资金
		return c.Status(500).JSON(fiber.Map{"success": false, "msg": "转账失败: " + err.Error()})
	}

	// 3. 成功，提交事务
	tx.Commit()

	return c.JSON(fiber.Map{
		"success": true,
		"balance": user.Balance,
		"msg":     "转账成功",
	})
}

func UpdateGamePassword(c *fiber.Ctx) error {
	type Req struct {
		NewPassword string `json:"new_password"`
	}
	var req Req
	c.BodyParser(&req)
	userId := QueryUserIdFromJwt(c)

	client := provider.NewHedocClient()
	if err := client.ChangePassword(userId, req.NewPassword, helpers.GetCfgInstance().Conf.Agentapi); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "msg": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true})
}

// Deposit 本地充值 (简化版：直接加钱，实际应为回调处理)
func Deposit(c *fiber.Ctx) error {
	var req requests.DepositRequest
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(400)
	}

	userId := QueryUserIdFromJwt(c)

	err := models.GetInstance().DbInstance.Transaction(func(tx *gorm.DB) error {
		var user dtos.User
		// 锁行
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userId).Error; err != nil {
			return err
		}

		// 加钱
		newBalance := user.Balance + req.Amount

		// 记录流水
		log := dtos.Transaction{
			UserID:        uint64(userId),
			Type:          1, // 1=充值
			Amount:        req.Amount,
			BeforeBalance: user.Balance,
			AfterBalance:  newBalance,
			Remark:        "在线充值: " + req.Channel,
			ReferenceID:   fmt.Sprintf("DEP-%d", time.Now().Unix()),
		}

		if err := tx.Create(&log).Error; err != nil {
			return err
		}
		if err := tx.Model(&user).Update("balance", newBalance).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "msg": "充值失败"})
	}

	// [新增] 异步记录充值统计
	go func() {
		RecordDailyStats(uint64(userId), 0, 0, req.Amount, 0, 0)
	}()

	return c.JSON(fiber.Map{"success": true, "msg": "充值成功"})
}

// Withdraw 提现 (本地扣钱，通常先冻结或扣除，等待后台人工审核)
func Withdraw(c *fiber.Ctx) error {
	var req requests.WithdrawRequest
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(400)
	}

	userId := QueryUserIdFromJwt(c)

	err := models.GetInstance().DbInstance.Transaction(func(tx *gorm.DB) error {
		var user dtos.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userId).Error; err != nil {
			return err
		}

		if user.Balance < req.Amount {
			return fmt.Errorf("余额不足")
		}

		newBalance := user.Balance - req.Amount

		// 记录流水
		log := dtos.Transaction{
			UserID:        uint64(userId),
			Type:          2,           // 2=提现
			Amount:        -req.Amount, // 负数
			BeforeBalance: user.Balance,
			AfterBalance:  newBalance,
			Remark:        "提现申请: " + req.BankInfo,
			ReferenceID:   fmt.Sprintf("WIT-%d", time.Now().Unix()),
		}

		if err := tx.Create(&log).Error; err != nil {
			return err
		}
		if err := tx.Model(&user).Update("balance", newBalance).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "msg": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "msg": "提现申请已提交"})
}

// GetGameList 获取游戏列表接口
// Method: POST
// Content-Type: application/json
func GetGameList(c *fiber.Ctx) error {
	if err := SyncAllGamesIfStale(gameListSyncInterval); err != nil {
		log.Printf("[Sync] Game list sync before GetGameList failed: %v\n", err)
	}
	// 1. 初始化默认参数
	req := requests.GameListClientRequest{
		Page:     1,
		PageSize: 20,
	}

	// 2. 解析 JSON Body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    400,
			"message": "Invalid JSON format",
			"error":   err.Error(),
		})
	}

	// 3. 参数边界修正
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100 // 限制最大页大小，防止性能问题
	}

	// 4. 构建基础查询
	// 使用 Joins 关联 game_types 和 game_providers 表，以便利用 Code 字段进行筛选
	query := models.GetInstance().DbInstance.Model(&dtos.Game{}).
		Joins("LEFT JOIN game_types ON game_types.id = games.game_type_id").
		Joins("LEFT JOIN game_providers ON game_providers.id = games.provider_id").
		Where("games.status = ?", 1).         // 仅查询上架状态的游戏
		Where("game_types.status = ?", 1).    //
		Where("game_providers.status = ?", 1) //

	// 5. 应用筛选条件
	// 5.1 游戏类型筛选 (排除空字符串和 "ALL")
	if req.GameTypeCode != "" && req.GameTypeCode != "ALL" {
		query = query.Where("game_types.code = ?", req.GameTypeCode)
	}

	// 5.2 厂商筛选 (排除空字符串和 "ALL")
	if req.ProviderCode != "" && req.ProviderCode != "ALL" {
		query = query.Where("game_providers.code = ?", req.ProviderCode)
	}

	// 5.3 游戏名称模糊搜索
	if req.GameName != "" {
		// 因为 name 字段是 JSON 类型 (例如 `{"CN":"麻将","EN":"Mahjong"}`)
		// 使用 LIKE %keyword% 可以同时匹配中文或英文名称
		query = query.Where("games.name LIKE ?", "%"+req.GameName+"%")
	}

	// 6. 获取符合条件的总记录数 (Total Count)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": "Database query error",
		})
	}

	// 7. 执行最终查询 (分页 + 排序 + 预加载详情)
	var games []dtos.Game
	offset := (req.Page - 1) * req.PageSize

	err := query.
		Preload("Provider").                        // 预加载 Provider 详情 (用于返回厂商名)
		Preload("GameType").                        // 预加载 GameType 详情 (用于返回分类名)
		Order("games.sort DESC, games.views DESC"). // 排序: 权重优先, 其次热度
		Limit(req.PageSize).
		Offset(offset).
		Find(&games).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": "Failed to fetch games",
		})
	}

	// 8. 计算总页数
	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(req.PageSize)))
	}

	// 9. 返回标准格式响应
	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data": fiber.Map{
			"list":        games,
			"total":       total,
			"page":        req.Page,
			"page_size":   req.PageSize,
			"total_pages": totalPages,
		},
	})
}

// GetProviderList 获取可用游戏厂商列表 (用于前端下拉筛选)
// Method: POST
// Content-Type: application/json
func GetGameOptions(c *fiber.Ctx) error {
	// 定义返回结构
	type GameOptionsData struct {
		Providers []dtos.GameProvider `json:"providers"`
		GameTypes []dtos.GameType     `json:"game_types"`
	}

	var providers []dtos.GameProvider
	var gameTypes []dtos.GameType

	// 1. 查询可用厂商 (Status=1)
	// 只查询前端需要的字段，减少传输量
	if err := models.GetInstance().DbInstance.Model(&dtos.GameProvider{}).
		Select("id, code, name, status,type_id").
		Where("status = ?", 1).
		Order("id ASC").
		Find(&providers).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "Failed to fetch providers"})
	}

	// 2. 查询可用分类 (Status=1)
	if err := models.GetInstance().DbInstance.Model(&dtos.GameType{}).
		Select("id, code, name, sort, status").
		Where("status = ?", 1).
		Order("sort ASC").
		Find(&gameTypes).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "Failed to fetch game types"})
	}

	// 3. 返回组合数据
	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data": GameOptionsData{
			Providers: providers,
			GameTypes: gameTypes,
		},
	})
}

// TransferToGame 资金划转 (转入/转出三方游戏)
func TransferToGame(c *fiber.Ctx) error {
	type TransferReq struct {
		Amount float64 `json:"amount"` // 始终为正数
		Type   int     `json:"type"`   // 1:转入游戏(充值) 2:转出游戏(提现)
	}

	var req TransferReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "Invalid params"})
	}

	userId := QueryUserIdFromJwt(c)

	// 生成唯一订单号
	billNo := fmt.Sprintf("TR-%d-%d", time.Now().UnixNano(), userId)
	var latestBalance float64

	// 1. 开启事务处理本地资金
	err := models.GetInstance().DbInstance.Transaction(func(tx *gorm.DB) error {
		var user dtos.User
		// 锁行
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userId).Error; err != nil {
			return err
		}

		thirdPartyLoginId := fmt.Sprintf("%su%d", helpers.GetCfgInstance().Conf.Prefix, user.ID)

		// 记录本地账变
		trans := dtos.Transaction{
			UserID:        uint64(userId),
			Type:          3, // 3: 上下分
			ReferenceID:   billNo,
			BeforeBalance: user.Balance,
		}

		if req.Type == 1 {
			// 转入游戏 (本地扣款)
			if user.Balance < req.Amount {
				return fmt.Errorf("余额不足")
			}
			user.Balance -= req.Amount
			trans.Amount = -req.Amount
			trans.Remark = "转入游戏"
		} else {
			// 转出游戏 (本地加款)
			// 注意：理论上应先查三方余额，这里简化直接转
			user.Balance += req.Amount
			trans.Amount = req.Amount
			trans.Remark = "游戏转出"
		}
		trans.AfterBalance = user.Balance

		if err := tx.Save(&user).Error; err != nil {
			return err
		}
		if err := tx.Create(&trans).Error; err != nil {
			return err
		}

		// 2. 调用第三方接口 (在事务内调用有风险，但为了简单保持一致性，失败则回滚)
		// 生产环境建议：本地先扣款->提交事务->异步调三方->失败则冲正
		client := provider.NewHedocClient()

		// 假设文档定义: 1=Deposit(User In), 2=Withdraw(User Out)
		balance, err := client.UpdateBalance(helpers.GetCfgInstance().Conf.Agentid,
			thirdPartyLoginId, req.Amount, req.Type, billNo, helpers.GetCfgInstance().Conf.Agentapi)
		if err != nil {
			return err // 触发事务回滚，本地钱退回
		}

		latestBalance = balance
		return nil
	})

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"code":    500,
			"message": "转账失败: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data": fiber.Map{
			"local_balance": 0, // 这里应返回 user.Balance
			"game_balance":  latestBalance,
			"bill_no":       billNo,
		},
	})
}

// OpenLobbyHandler 获取游戏大厅链接
func OpenLobbyHandler(c *fiber.Ctx) error {
	var req requests.OpenLobbyPayload
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "Invalid params"})
	}

	userId := QueryUserIdFromJwt(c)
	thirdPartyLoginId := fmt.Sprintf("%su%d", helpers.GetCfgInstance().Conf.Prefix, userId)
	// 获得用户专属的第三方密码
	thirdPartyPassword := common.GenerateGamePassword(uint64(userId))

	// 1. 获取厂商信息 (需要 ID)
	var gameprovider dtos.GameProvider
	if err := models.GetInstance().DbInstance.Where("code = ?", req.ProviderCode).First(&gameprovider).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "Unknown provider"})
	}

	client := provider.NewHedocClient()

	// 2. 确保用户已在第三方注册 (静默注册)
	// 使用固定密码或查表获取
	_ = client.RegisterUser(helpers.GetCfgInstance().Conf.Agentid, thirdPartyLoginId, thirdPartyPassword, helpers.GetCfgInstance().Conf.Agentapi)

	// 3. 请求链接
	url, err := client.OpenLobby(helpers.GetCfgInstance().Conf.Agentid,
		thirdPartyLoginId, thirdPartyPassword, req.IsMobile, req.Language, int(gameprovider.ID), helpers.GetCfgInstance().Conf.Agentapi)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}

	return c.JSON(fiber.Map{"code": 0, "message": "success", "data": fiber.Map{"url": url}})
}

// GetUserInfo 获取当前登录用户信息
func GetUserInfo(c *fiber.Ctx) error {
	// 1. 从 JWT 中获取 UserID (由中间件设置)
	userId := QueryUserIdFromJwt(c)

	// 2. 查询数据库
	var user dtos.User
	// 使用 Select 指定字段，避免查出不必要的数据
	if err := models.GetInstance().DbInstance.Model(&dtos.User{}).
		Select("id, username, invite_code, balance, vip_level, level").
		First(&user, userId).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    404,
			"message": "用户不存在",
		})
	}

	// 3. 返回数据
	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data": fiber.Map{
			"uid":         user.ID,         // 用户ID
			"username":    user.Username,   // 用户名
			"invite_code": user.InviteCode, // 邀请码
			"balance":     user.Balance,    // 当前本地余额
			"vip_level":   user.VipLevel,   // VIP等级
			"level":       user.Level,      // 代理层级 (如果前端需要展示)
		},
	})
}

// GetInviteInfo 获取用户邀请信息
func GetInviteInfo(c *fiber.Ctx) error {
	userId := QueryUserIdFromJwt(c)

	// 定义返回结构
	type InviteData struct {
		InviteCode  string  `json:"invite_code"`
		ShareLink   string  `json:"share_link"`   // 完整的注册分享链接
		DirectCount int64   `json:"direct_count"` // 直推人数 (一级下线)
		TeamCount   int64   `json:"team_count"`   // 团队总人数 (所有下线，可选)
		TotalEarn   float64 `json:"total_earn"`   // 总佣金
	}

	var user dtos.User
	// 1. 获取当前用户的邀请码和 Path
	if err := models.GetInstance().DbInstance.Select("id, invite_code, path").First(&user, userId).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "User not found"})
	}

	// 2. 统计直推人数 (ParentID = 当前用户ID)
	var directCount int64
	if err := models.GetInstance().DbInstance.Model(&dtos.User{}).
		Where("parent_id = ?", userId).
		Count(&directCount).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "Count error"})
	}

	// 3. (可选) 统计团队总人数 (利用 Path 左前缀匹配)
	// 逻辑: 查找 path 以 "user.Path + user.ID + /" 开头的所有用户
	// 例如: 我是 "1/5/", 下级必须包含 "1/5/%"
	var teamCount int64
	// 构造当前用户的层级路径前缀 (注意: 根用户的 path 为空字符串时的处理)
	// 假设 user.ID = 5, user.Path = "1/" -> prefix = "1/5/"
	// 假设 user.ID = 1, user.Path = ""   -> prefix = "1/"

	/*
	   注意：如果你的系统数据量非常大（千万级），实时 Count Like 查询可能会略慢，
	   建议后续将 team_count 做成字段存入 users 表异步更新。
	   但在百万级数据下，path 有索引，Count 依然很快。
	*/

	// 简单拼接逻辑
	pathPrefix := user.Path + fmt.Sprintf("%d/", user.ID)
	if user.Path == "" {
		pathPrefix = fmt.Sprintf("%d/", user.ID)
	}

	models.GetInstance().DbInstance.Model(&dtos.User{}).
		Where("path LIKE ?", pathPrefix+"%").
		Count(&teamCount)

	// 4. 拼接分享链接
	shareLink := fmt.Sprintf("https://www.your-game-site.com/register?code=%s", user.InviteCode)

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data": InviteData{
			InviteCode:  user.InviteCode,
			ShareLink:   shareLink,
			DirectCount: directCount,
			TeamCount:   teamCount,
			TotalEarn:   7.7, // 暂时
		},
	})
}

// GetBanners 获取首页轮播图
// Method: GET
func GetBanners(c *fiber.Ctx) error {
	var banners []dtos.Banner

	// 查询逻辑：
	// 1. status = 1 (启用)
	// 2. 按 sort 倒序排列 (权重越大越靠前)，其次按 id 倒序
	err := models.GetInstance().DbInstance.Model(&dtos.Banner{}).
		Where("status = ?", 1).
		Order("sort DESC, id DESC").
		Find(&banners).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"code":    500,
			"message": "Failed to fetch banners",
		})
	}

	// 这里的 banners 序列化为 JSON 时：
	// 1. 如果 JumpLink 为 "", 字段消失
	// 2. 如果 Buttons 为 NULL, 字段消失
	// 完美符合前端多态需求
	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data":    banners,
	})
}

// GetAllLiveGames 获取所有真人游戏 (Type=2) - 针对大数据量的优化版
func GetAllGames(c *fiber.Ctx) error {
	// 目标类型 ID (根据需求固定为 2，即真人视讯)
	targetTypeID := 2

	// 定义接收结果的切片
	var games []responses.GameLiteResponse

	// ---------------------------------------------------------
	// 优化点 1: 字段裁剪 (Select)
	// 只查询前端渲染列表必须的字段，不查 created_at, sort, views 等
	// ---------------------------------------------------------
	err := models.GetInstance().DbInstance.Table("games").
		Select("id, game_code as code, name, img_url as img, provider_id as pid").
		Where("game_type_id = ?", targetTypeID).
		Where("status = ?", 1). // 只查上架的
		// ---------------------------------------------------------
		// 优化点 2: 移除 Order By (如果数据量极大)
		// 排序是非常消耗数据库内存的操作。如果前端可以自己排，或者对顺序不敏感，
		// 去掉 Order By 可以显著提升查询速度。如果必须排，确保有索引。
		// ---------------------------------------------------------
		// Order("sort DESC").
		Scan(&games).Error // 使用 Scan 映射到精简结构体

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "Query failed"})
	}

	// ---------------------------------------------------------
	// 优化点 3: HTTP 缓存控制 (Client Cache)
	// 告诉浏览器/客户端：这个列表在 5 分钟内不会变，不要重复请求。
	// ---------------------------------------------------------
	c.Set("Cache-Control", "public, max-age=300")

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data":    games,
		// count: len(games) // 可选，返回总数
	})
}

// GetFeaturedGames 获取特色游戏榜单 (热门 + 推荐)
func GetFeaturedGames(c *fiber.Ctx) error {
	var (
		hotList []dtos.Game
		recList []dtos.Game
		errHot  error
		errRec  error
	)

	// 使用 WaitGroup 并发查询
	var wg sync.WaitGroup
	wg.Add(2)

	// 1. 热门游戏 (Top 20 Views)
	go func() {
		defer wg.Done()
		errHot = models.GetInstance().DbInstance.Model(&dtos.Game{}).
			Where("status = ?", 1).
			Preload("Provider"). // 仅加载厂商，通常榜单不需要分类详情
			Order("views DESC"). // 走 idx_views 索引
			Limit(20).
			Find(&hotList).Error
	}()

	// 2. 推荐游戏 (Top 20 Sort)
	go func() {
		defer wg.Done()
		errRec = models.GetInstance().DbInstance.Model(&dtos.Game{}).
			Where("status = ?", 1).
			Preload("Provider").
			Order("sort DESC"). // 走 idx_sort_views 索引
			Limit(20).
			Find(&recList).Error
	}()

	wg.Wait()

	if errHot != nil || errRec != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "Failed to fetch featured games"})
	}

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data": fiber.Map{
			"hot_games":       hotList,
			"recommend_games": recList,
		},
	})
}

// OpenLobbyHandler2 获取游戏大厅链接+自动带入
func OpenLobbyHandler2(c *fiber.Ctx) error {
	var req requests.OpenLobbyPayload
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "Invalid params"})
	}

	userId := QueryUserIdFromJwt(c)
	thirdPartyLoginId := fmt.Sprintf("%su%d", helpers.GetCfgInstance().Conf.Prefix, userId)
	// 获得用户专属的第三方密码
	thirdPartyPassword := common.GenerateGamePassword(uint64(userId))

	// 1. 获取厂商信息 (需要 ID)
	var gameprovider dtos.GameProvider
	if err := models.GetInstance().DbInstance.Where("code = ?", req.ProviderCode).First(&gameprovider).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "Unknown provider"})
	}

	client := provider.NewHedocClient()

	// 2. 确保用户已在第三方注册 (静默注册)
	// 使用固定密码或查表获取
	_ = client.RegisterUser(helpers.GetCfgInstance().Conf.Agentid, thirdPartyLoginId, thirdPartyPassword, helpers.GetCfgInstance().Conf.Agentapi)

	// ==========================================
	// 核心：智能自动转入 (Smart Auto-In)
	// ==========================================
	// =================================================================
	// 步骤 A: 先查询用户本地余额 (为了知道要扣多少)
	// =================================================================

	// =================================================================
	// 步骤 B: 核心同步卡点 —— 请求 12661 扣款
	// =================================================================
	// 调用我们刚才写的同步函数
	// 如果这里报错，函数直接 return，后续的 OpenGame、三方转账统统不会发生
	// 这里先简单查一下12661余额，不用锁，因为真正的锁在后面或者在 12661 那边
	balance, err := SyncGameBalance(uint64(userId), 0, 1)
	if err != nil {
		// 阻断流程！返回错误给前端
		return c.Status(500).JSON(fiber.Map{
			"code":    503,
			"message": "资金同步失败，请重试: " + err.Error(),
		})
	}
	if balance > 0 {
		// =================================================================
		// 步骤 C: 12661 扣款成功后，处理 Web 端本地数据和三方转账
		// =================================================================
		// 此时可以认为钱已经从“游戏进程”扣除了，我们现在要在 Web 数据库记录流水并转入三方
		// 记录流水
		billNo := fmt.Sprintf("AUTO-IN-%d-%d", time.Now().UnixNano(), userId)
		trans := dtos.Transaction{
			UserID: uint64(userId), Type: 3, Amount: -float64(balance),
			ReferenceID: billNo, Remark: "同步扣款转入游戏",
		}
		if err := models.GetInstance().DbInstance.Create(&trans).Error; err != nil {
			return err
		}

		// 调用三方 (Hedoc) 加钱
		// 风险提示：如果这里失败了，12661 那边已经扣钱了。
		// 生产环境需要：如果这里失败，发起一个“补偿请求”给 12661 把钱加回去 (冲正)
		_, err = client.UpdateBalance(helpers.GetCfgInstance().Conf.Agentid,
			thirdPartyLoginId, float64(balance), 1, billNo, helpers.GetCfgInstance().Conf.Agentapi)
		if err != nil {
			SyncGameBalance(uint64(userId), -float64(balance), 2)
			return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
		}
	}

	// 3. 请求链接
	url, err := client.OpenLobby(helpers.GetCfgInstance().Conf.Agentid,
		thirdPartyLoginId, thirdPartyPassword, req.IsMobile, req.Language, int(gameprovider.ID), helpers.GetCfgInstance().Conf.Agentapi)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}

	return c.JSON(fiber.Map{"code": 0, "message": "success", "data": fiber.Map{"url": url}})
}

type GameProcessResponse struct {
	Code int `json:"code"` // 假设 0 或 200 表示成功
	// Message string `json:"message,omitempty"`
	Balance int64 `json:"balance"`
	// Data    interface{} `json:"data"` // 如果有返回数据可加
}

// SyncDeductGameBalance 同步请求扣减游戏进程余额
// 只有当该函数返回 nil 时，才代表扣款成功，允许继续后续逻辑
func SyncGameBalance(userID uint64, amount float64, direction int) (int64, error) {
	// 1. 配置
	targetURL := helpers.GetCfgInstance().Conf.Gamehallapi + "/pgs/game/syncbalance"

	// 请求参数
	body := map[string]interface{}{
		"uid":       userID,
		"amount":    amount,    // 扣减金额
		"direction": direction, // 1 in 2 out
	}

	// 2. 初始化 Client
	//以此处必须设置超时！因为是同步阻塞，如果对方死锁，不能让这边一直等
	client := resty.New().SetTimeout(3 * time.Second)

	var result GameProcessResponse

	// 3. 发送 POST 请求
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(targetURL)

	// 手动解析 JSON 响应（避免依赖 Content-Type 头）
	if err == nil && resp.StatusCode() == 200 {
		if err := stdjson.Unmarshal(resp.Body(), &result); err != nil {
			return 0, fmt.Errorf("解析响应失败: %v", err)
		}
	}

	fmt.Println("result:", result)

	// 4. 第一层关卡：网络错误处理
	if err != nil {
		return 0, fmt.Errorf("连接游戏进程失败: %v", err)
	}

	// 5. 第二层关卡：HTTP 状态码处理
	if resp.StatusCode() != 200 {
		return 0, fmt.Errorf("游戏进程HTTP异常: %d", resp.StatusCode())
	}

	// 6. 第三层关卡：业务逻辑错误处理
	// 假设 12661 返回 Code=0 表示成功，其他表示失败（如余额不足）
	if result.Code != 0 {
		return 0, fmt.Errorf("游戏进程扣款失败: [%d]", result.Code)
	}

	// 全部通过，允许放行
	return result.Balance, nil
}

// Helper function to render Error HTML
func renderErrorHTML(c *fiber.Ctx, errMsg string) error {
	c.Type("html", "utf-8")
	return c.SendString(fmt.Sprintf(common.ErrHtmlTemplate, errMsg))
}

func LaunchGame3(c *fiber.Ctx) error {
	var req requests.OpenGamePayload
	// 核心区别: 使用 QueryParser 解析 URL 参数
	if err := c.QueryParser(&req); err != nil {
		return renderErrorHTML(c, "Invalid JSON parameters")
	}
	// userId := QueryUserIdFromJwt(c)
	// 1. 获取并解析 Token
	tokenStr := c.Query("token")
	userId, err := ParseTokenManual(tokenStr, helpers.GetCfgInstance().Conf.Jwt)
	if err != nil {
		return renderErrorHTML(c, "Authentication failed: "+err.Error())
	}
	thirdPartyLoginId := fmt.Sprintf("%su%d", helpers.GetCfgInstance().Conf.Prefix, userId)
	// 获得用户专属的第三方密码
	thirdPartyPassword := common.GenerateGamePassword(uint64(userId))
	// 1. 根据 GameCode 查找游戏信息 (获取 ProviderID)
	var game dtos.Game
	if err := models.GetInstance().DbInstance.Preload("Provider").Where("game_code = ?", req.GameCode).First(&game).Error; err != nil {
		return renderErrorHTML(c, "Game not found or disabled")
	}

	client := provider.NewHedocClient()

	// 2. 确保用户已注册
	// _ = client.RegisterUser(thirdPartyLoginId, "Game@123456")
	_ = client.RegisterUser(helpers.GetCfgInstance().Conf.Agentid, thirdPartyLoginId, thirdPartyPassword, helpers.GetCfgInstance().Conf.Agentapi)

	// ==========================================
	// 核心：智能自动转入 (Smart Auto-In)
	// ==========================================
	// =================================================================
	// 步骤 A: 先查询用户本地余额 (为了知道要扣多少)
	// =================================================================

	// =================================================================
	// 步骤 B: 核心同步卡点 —— 请求 12661 扣款
	// =================================================================
	// 调用我们刚才写的同步函数
	// 这里先简单查一下12661余额，不用锁，因为真正的锁在后面或者在 12661 那边
	// 如果这里报错，函数直接 return，后续的 OpenGame、三方转账统统不会发生
	balance, err := SyncGameBalance(uint64(userId), 0, 1)
	if err != nil {
		// 阻断流程！返回错误给前端
		return renderErrorHTML(c, "资金同步失败，请重试: "+err.Error())
	}
	if balance > 0 {
		// =================================================================
		// 步骤 C: 12661 扣款成功后，处理 Web 端本地数据和三方转账
		// =================================================================
		// 此时可以认为钱已经从“游戏进程”扣除了，我们现在要在 Web 数据库记录流水并转入三方
		// 记录流水
		billNo := fmt.Sprintf("AUTO-IN-%d-%d", time.Now().UnixNano(), userId)
		trans := dtos.Transaction{
			UserID: uint64(userId), Type: 3, Amount: -float64(balance),
			ReferenceID: billNo, Remark: "同步扣款转入游戏",
		}
		if err := models.GetInstance().DbInstance.Create(&trans).Error; err != nil {
			return renderErrorHTML(c, err.Error())
		}

		// 调用三方 (Hedoc) 加钱
		// 风险提示：如果这里失败了，12661 那边已经扣钱了。
		// 生产环境需要：如果这里失败，发起一个“补偿请求”给 12661 把钱加回去 (冲正)
		_, err = client.UpdateBalance(helpers.GetCfgInstance().Conf.Agentid,
			thirdPartyLoginId, float64(balance), 1, billNo, helpers.GetCfgInstance().Conf.Agentapi)
		if err != nil {
			SyncGameBalance(uint64(userId), -float64(balance), 2)
			return renderErrorHTML(c, err.Error())
		}
	}

	// 3. 请求链接
	url, err := client.OpenGame(thirdPartyLoginId, thirdPartyPassword, int(game.Provider.ID), req.GameCode, req.IsMobile, req.Language,
		helpers.GetCfgInstance().Conf.Agentapi,
		helpers.GetCfgInstance().Conf.Agentid)
	if err != nil {
		return renderErrorHTML(c, err.Error())
	}

	// 4. (可选) 增加热度
	models.GetInstance().DbInstance.Model(&game).UpdateColumn("views", gorm.Expr("views + ?", 1))

	// --- 返回 HTML ---
	c.Type("html", "utf-8")
	// 填充模板: 1. 游戏名称(Title) 2. 游戏URL(Src)
	// 注意：game.Name 是 JSON，这里为了简单直接取 GameCode 或者处理后的名字
	return c.SendString(fmt.Sprintf(common.GameContainerTemplate, req.GameCode, url))
}

// OpenLobbyHandler3 获取游戏大厅链接html+自动带入
func OpenLobbyHandler3(c *fiber.Ctx) error {
	var req requests.OpenLobbyPayload
	// 核心区别: 使用 QueryParser 解析 URL 参数
	if err := c.QueryParser(&req); err != nil {
		return renderErrorHTML(c, "Invalid JSON parameters")
	}

	// userId := QueryUserIdFromJwt(c)
	// 1. 获取并解析 Token
	tokenStr := c.Query("token")
	userId, err := ParseTokenManual(tokenStr, helpers.GetCfgInstance().Conf.Jwt)
	if err != nil {
		return renderErrorHTML(c, "Authentication failed: "+err.Error())
	}
	thirdPartyLoginId := fmt.Sprintf("%su%d", helpers.GetCfgInstance().Conf.Prefix, userId)
	// 获得用户专属的第三方密码
	thirdPartyPassword := common.GenerateGamePassword(uint64(userId))

	// 1. 获取厂商信息 (需要 ID)
	var gameprovider dtos.GameProvider
	if err := models.GetInstance().DbInstance.Where("code = ?", req.ProviderCode).First(&gameprovider).Error; err != nil {
		return renderErrorHTML(c, "Unknown provider")
	}

	client := provider.NewHedocClient()

	// 2. 确保用户已在第三方注册 (静默注册)
	// 使用固定密码或查表获取
	_ = client.RegisterUser(helpers.GetCfgInstance().Conf.Agentid, thirdPartyLoginId, thirdPartyPassword, helpers.GetCfgInstance().Conf.Agentapi)

	// ==========================================
	// 核心：智能自动转入 (Smart Auto-In)
	// ==========================================
	// =================================================================
	// 步骤 A: 先查询用户本地余额 (为了知道要扣多少)
	// =================================================================
	// =================================================================
	// 步骤 B: 核心同步卡点 —— 请求 12661 扣款
	// =================================================================
	// 调用我们刚才写的同步函数
	// 如果这里报错，函数直接 return，后续的 OpenGame、三方转账统统不会发生
	// 这里先简单查一下12661余额，不用锁，因为真正的锁在后面或者在 12661 那边
	balance, err := SyncGameBalance(uint64(userId), 0, 1)
	if err != nil {
		// 阻断流程！返回错误给前端
		return renderErrorHTML(c, "资金同步失败，请重试: "+err.Error())
	}
	if balance > 0 {
		// =================================================================
		// 步骤 C: 12661 扣款成功后，处理 Web 端本地数据和三方转账
		// =================================================================
		// 此时可以认为钱已经从“游戏进程”扣除了，我们现在要在 Web 数据库记录流水并转入三方
		// 记录流水
		billNo := fmt.Sprintf("AUTO-IN-%d-%d", time.Now().UnixNano(), userId)
		trans := dtos.Transaction{
			UserID: uint64(userId), Type: 3, Amount: -float64(balance),
			ReferenceID: billNo, Remark: "同步扣款转入游戏",
		}
		if err := models.GetInstance().DbInstance.Create(&trans).Error; err != nil {
			return renderErrorHTML(c, err.Error())
		}

		// 调用三方 (Hedoc) 加钱
		// 风险提示：如果这里失败了，12661 那边已经扣钱了。
		// 生产环境需要：如果这里失败，发起一个“补偿请求”给 12661 把钱加回去 (冲正)
		_, err = client.UpdateBalance(helpers.GetCfgInstance().Conf.Agentid,
			thirdPartyLoginId, float64(balance), 1, billNo, helpers.GetCfgInstance().Conf.Agentapi)
		if err != nil {
			SyncGameBalance(uint64(userId), -float64(balance), 2)
			return renderErrorHTML(c, err.Error())
		}
	}

	// 3. 请求链接
	url, err := client.OpenLobby(helpers.GetCfgInstance().Conf.Agentid,
		thirdPartyLoginId, thirdPartyPassword, req.IsMobile, req.Language, int(gameprovider.ID), helpers.GetCfgInstance().Conf.Agentapi)
	if err != nil {
		return renderErrorHTML(c, err.Error())
	}

	// --- 返回 HTML ---
	c.Type("html", "utf-8")
	// 填充模板: 1. 游戏名称(Title) 2. 游戏URL(Src)
	// 注意：game.Name 是 JSON，这里为了简单直接取 GameCode 或者处理后的名字
	return c.SendString(fmt.Sprintf(common.GameContainerTemplate, req.ProviderCode, url))
}

// ==========================================
// 大富翁游戏接入专用接口 (Monopoly Integration)
// 追加日期: 2026-03-20
// 需求方: 12661游戏端
// 说明: 大富翁游戏格子展示和直达游戏功能
// ==========================================

// QuickLoginRequest 快速登录请求
type QuickLoginRequest struct {
	Username string `json:"username"`
}

// QuickLoginResponse 快速登录响应
type QuickLoginResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
}

// QuickLoginForMonopoly 快速登录接口（同服务器免密）
// 供12661游戏端调用，根据username直接生成JWT Token
// 注意: 此接口应限制仅内网/同服务器访问
func QuickLoginForMonopoly(c *fiber.Ctx) error {
	var req QuickLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"code":    400,
			"message": "Invalid params",
		})
	}

	if req.Username == "" {
		return c.Status(400).JSON(fiber.Map{
			"code":    400,
			"message": "username is required",
		})
	}

	// 查询用户是否存在
	var user dtos.TagUser
	err := models.GetInstance().DbInstance.Where("acc = ?", req.Username).First(&user).Error

	if err != nil {
		// 用户不存在，自动创建
		if err == gorm.ErrRecordNotFound {
			user = dtos.TagUser{
				Acc:      req.Username,
				Nick:     req.Username,
				FarmCoin: 0,
				// 其他字段使用默认值
			}
			if err := models.GetInstance().DbInstance.Create(&user).Error; err != nil {
				return c.Status(500).JSON(fiber.Map{
					"code":    500,
					"message": "Failed to create user: " + err.Error(),
				})
			}
		} else {
			return c.Status(500).JSON(fiber.Map{
				"code":    500,
				"message": "Database error: " + err.Error(),
			})
		}
	}

	// 生成JWT Token（5分钟有效期，仅用于单次游戏进入）
	token, err := common.GenerateJWT(uint64(user.ID), user.Acc, helpers.GetCfgInstance().Conf.Jwt)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"code":    500,
			"message": "Failed to generate token: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"code":       0,
		"message":    "success",
		"token":      token,
		"expires_in": 300, // 5分钟有效期（秒）
		"user_id":    user.ID,
	})
}

// GetRandomGamesForMonopoly 获取随机游戏列表（供大富翁格子展示）
// 返回随机5个游戏的信息，包含game_code、名称、图标等
func GetRandomGamesForMonopoly(c *fiber.Ctx) error {
	// 获取数量参数，默认5个
	count := c.QueryInt("count", 5)
	if count <= 0 || count > 20 {
		count = 5 // 限制最大20个
	}

	// 查询随机的游戏列表
	var games []dtos.Game
	err := models.GetInstance().DbInstance.
		Model(&dtos.Game{}).
		Where("status = ?", 1).
		Preload("Provider").
		Order("RAND()").
		Limit(count).
		Find(&games).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"code":    500,
			"message": "Failed to fetch games: " + err.Error(),
		})
	}

	// 构建返回数据
	type GameInfo struct {
		GameCode     string         `json:"game_code"`
		Name         datatypes.JSON `json:"name"`
		ImgUrl       string         `json:"img_url"`
		ProviderCode string         `json:"provider_code"`
		ProviderName string         `json:"provider_name"`
	}

	var gameList []GameInfo
	for _, game := range games {
		providerCode := ""
		providerName := ""
		if game.Provider != nil {
			providerCode = game.Provider.Code
			providerName = game.Provider.Name
		}

		gameList = append(gameList, GameInfo{
			GameCode:     game.GameCode,
			Name:         game.Name,
			ImgUrl:       game.ImgUrl,
			ProviderCode: providerCode,
			ProviderName: providerName,
		})
	}

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data": fiber.Map{
			"games": gameList,
		},
	})
}

// LaunchMonopolyRequest 直达游戏请求
type LaunchMonopolyRequest struct {
	Token     string `json:"token"`      // JWT Token（来自QuickLoginForMonopoly）
	GameCode  string `json:"game_code"`  // 游戏代码（来自GetRandomGamesForMonopoly）
	IsMobile  bool   `json:"is_mobile"`  // 是否移动端，默认false
	Language  string `json:"language"`   // 语言，默认"EN"（可选：CN, EN, ZH, TH等）
	ExtraData string `json:"extra_data"` // 额外数据，透传保留（可选）
}

// LaunchMonopolyGame 大富翁直达游戏接口
// 用户踩中格子后直接跳转进入游戏，跳过大厅和选游戏步骤
// 方法: GET
// 参数: Query String (token, game_code, is_mobile, language)
// 调用方: 12661游戏前端WebView直接打开
func LaunchMonopolyGame(c *fiber.Ctx) error {
	// 从Query参数获取
	tokenStr := c.Query("token")
	gameCode := c.Query("game_code")
	isMobile := c.QueryBool("is_mobile", false)
	language := c.Query("language", "EN")

	// 校验必填参数
	if tokenStr == "" || gameCode == "" {
		return renderErrorHTML(c, "Missing required parameters: token and game_code")
	}

	// 解析token获取userId（复用ParseTokenManual）
	userId, err := ParseTokenManual(tokenStr, helpers.GetCfgInstance().Conf.Jwt)
	if err != nil {
		return renderErrorHTML(c, "Authentication failed: "+err.Error())
	}

	thirdPartyLoginId := fmt.Sprintf("%su%d", helpers.GetCfgInstance().Conf.Prefix, userId)
	thirdPartyPassword := common.GenerateGamePassword(uint64(userId))

	// 1. 根据GameCode查找游戏信息
	var game dtos.Game
	if err := models.GetInstance().DbInstance.Preload("Provider").Where("game_code = ?", gameCode).First(&game).Error; err != nil {
		return renderErrorHTML(c, "Game not found or disabled")
	}

	client := provider.NewHedocClient()

	// 2. 确保用户已注册到Hedoc
	_ = client.RegisterUser(helpers.GetCfgInstance().Conf.Agentid, thirdPartyLoginId, thirdPartyPassword, helpers.GetCfgInstance().Conf.Agentapi)

	// ==========================================
	// 核心：同步扣款并转入游戏（复用LaunchGame3逻辑）
	// ==========================================

	// 步骤1: 查询12661余额
	balance, err := SyncGameBalance(uint64(userId), 0, 1)
	if err != nil {
		return renderErrorHTML(c, "资金同步失败，请重试: "+err.Error())
	}

	// 步骤2: 有余额则转入Hedoc
	if balance > 0 {
		// 记录流水
		billNo := fmt.Sprintf("MONO-IN-%d-%d", time.Now().UnixNano(), userId)
		trans := dtos.Transaction{
			UserID:      uint64(userId),
			Type:        3,
			Amount:      -float64(balance),
			ReferenceID: billNo,
			Remark:      "大富翁游戏自动转入",
		}
		if err := models.GetInstance().DbInstance.Create(&trans).Error; err != nil {
			return renderErrorHTML(c, "记录流水失败")
		}

		// 调用Hedoc加钱
		_, err = client.UpdateBalance(
			helpers.GetCfgInstance().Conf.Agentid,
			thirdPartyLoginId,
			float64(balance),
			1, // Type=1 转入
			billNo,
			helpers.GetCfgInstance().Conf.Agentapi,
		)
		if err != nil {
			// 冲正12661
			SyncGameBalance(uint64(userId), -float64(balance), 2)
			return renderErrorHTML(c, "资金转入游戏失败: "+err.Error())
		}
	}

	// 3. 请求游戏链接
	// 语言处理：默认EN，可根据请求参数
	if language == "" {
		language = "EN"
	}
	url, err := client.OpenGame(
		thirdPartyLoginId,
		thirdPartyPassword,
		int(game.Provider.ID),
		game.GameCode,
		isMobile,
		language,
		helpers.GetCfgInstance().Conf.Agentapi,
		helpers.GetCfgInstance().Conf.Agentid,
	)
	if err != nil {
		return renderErrorHTML(c, "获取游戏链接失败: "+err.Error())
	}

	// 4. 增加游戏热度
	models.GetInstance().DbInstance.Model(&game).UpdateColumn("views", gorm.Expr("views + ?", 1))

	// 5. 返回HTML页面
	c.Type("html", "utf-8")
	return c.SendString(fmt.Sprintf(common.GameContainerTemplate, game.Provider.Code, url))
}
