# 后端API设计文档 - 活动功能

## 概述

本文档描述活动功能相关的后端API设计，基于Go + Fiber框架实现。

## 路由规划

在 `be/controllers/router.go` 中新增以下路由：

```go
func InitRouteTablesWithJWT(r *fiber.App) {
    // ... 现有路由 ...

    // 活动相关API（需要JWT认证）
    r.Get("/activities", GetActivities)                          // 获取活动列表
    r.Get("/activities/:type/progress", GetActivityProgress)     // 获取活动进度
    r.Post("/activities/:type/claim", ClaimActivityReward)       // 领取活动奖励

    // 轮盘活动API
    r.Post("/wheel/spin", SpinWheel)                             // 抽奖
    r.Get("/wheel/status", GetWheelStatus)                       // 获取轮盘状态

    // Coupon相关API
    r.Get("/coupons", GetUserCoupons)                            // 获取用户优惠券列表
    r.Post("/coupons/:code/activate", ActivateCoupon)            // 激活优惠券
    r.Post("/coupons/:code/use", UseCoupon)                      // 使用优惠券
}
```

## API详细设计

### 1. 获取活动列表

**请求信息：**
- **Method**: GET
- **Path**: `/activities`
- **Auth**: 需要JWT

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "type": "recharge_rebate",
      "name": "储值返利活动",
      "description": "活动期间连续4天充值返利",
      "config": {
        "min_deposit": 3,
        "currency": "USD",
        "repeatable": true,
        "cycle_days": 4,
        "days": [
          {"day": 1, "rate": 0.50, "label": "+50%"},
          {"day": 2, "rate": 0.75, "label": "+75%"},
          {"day": 3, "rate": 1.00, "label": "+100%"},
          {"day": 4, "rate": 1.50, "label": "+150%"}
        ]
      },
      "start_time": "2025-02-01T00:00:00Z",
      "end_time": "2025-02-28T23:59:59Z",
      "status": 1
    },
    {
      "id": 2,
      "type": "coupon_wheel",
      "name": "每日幸运轮盘",
      "description": "每日抽奖赢取优惠券",
      "config": {
        "min_deposit": 3,
        "daily_limit": 1,
        "currency": "USD"
      },
      "start_time": "2025-01-01T00:00:00Z",
      "end_time": "2025-12-31T23:59:59Z",
      "status": 1
    },
    {
      "id": 3,
      "type": "loss_rebate",
      "name": "亏损返还活动",
      "description": "活动期间净亏损按6%返还",
      "config": {
        "rebate_rate": 0.06,
        "min_loss": 10,
        "currency": "USD",
        "settlement_type": "daily"
      },
      "start_time": "2025-02-01T00:00:00Z",
      "end_time": "2025-02-14T23:59:59Z",
      "status": 1
    }
  ]
}
```

**Go实现：**
```go
// GetActivities 获取活动列表
func GetActivities(ctx *fiber.Ctx) error {
    userID := QueryUserIdFromJwt(ctx)

    // 查询所有进行中的活动
    var activities []dtos.Activity
    if err := db.Where("status = ? AND start_time <= ? AND end_time >= ?",
        1, time.Now(), time.Now()).Find(&activities).Error; err != nil {
        return ctx.Status(500).JSON(fiber.Map{
            "code": 500,
            "message": "获取活动列表失败",
        })
    }

    // 查询用户各活动进度
    var progresses []dtos.UserActivityProgress
    today := time.Now().Format("2006-01-02")
    db.Where("user_id = ? AND progress_date = ?", userID, today).Find(&progresses)

    progressMap := make(map[string]dtos.UserActivityProgress)
    for _, p := range progresses {
        progressMap[p.ActivityType] = p
    }

    // 组装响应
    var result []fiber.Map
    for _, act := range activities {
        item := fiber.Map{
            "id": act.ID,
            "type": act.Type,
            "name": act.Name,
            "config": act.Config,
            "start_time": act.StartTime,
            "end_time": act.EndTime,
            "status": act.Status,
        }

        if progress, ok := progressMap[act.Type]; ok {
            item["user_progress"] = progress
        }

        result = append(result, item)
    }

    return ctx.JSON(fiber.Map{
        "code": 0,
        "message": "success",
        "data": result,
    })
}
```

### 2. 获取活动进度

**请求信息：**
- **Method**: GET
- **Path**: `/activities/:type/progress`
- **Auth**: 需要JWT

**路径参数：**
| 参数 | 类型 | 说明 |
|------|------|------|
| type | string | 活动类型 (daily_recharge / coupon_wheel) |

**响应示例（储值返利活动）：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "activity_type": "daily_recharge",
    "current_day": 2,
    "total_days": 4,
    "today_progress": {
      "date": "2025-02-02",
      "deposit_amount": 0,
      "min_deposit": 3,
      "reward_rate": 0.75,
      "status": "pending",
      "claimed": false
    },
    "history": [
      {
        "date": "2025-02-01",
        "deposit_amount": 50,
        "reward_amount": 25,
        "status": "completed",
        "claimed": true
      }
    ]
  }
}
```

### 3. 领取活动奖励

**请求信息：**
- **Method**: POST
- **Path**: `/activities/:type/claim`
- **Auth**: 需要JWT

**请求体：**
```json
{
  "date": "2025-02-01",
  "day_number": 1
}
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "reward_amount": 25.00,
    "new_balance": 125.00,
    "transaction_id": "TXN123456789"
  }
}
```

**错误码：**
| 错误码 | 说明 |
|--------|------|
| 4001 | 未达到领取条件 |
| 4002 | 奖励已领取 |
| 4003 | 活动已过期 |
| 4004 | 不在活动期间 |

**Go实现：**
```go
// ClaimActivityRewardRequest 领取奖励请求
type ClaimActivityRewardRequest struct {
    Date      string `json:"date"`
    DayNumber int    `json:"day_number"`
}

// ClaimActivityReward 领取活动奖励
func ClaimActivityReward(ctx *fiber.Ctx) error {
    userID := QueryUserIdFromJwt(ctx)
    activityType := ctx.Params("type")

    var req ClaimActivityRewardRequest
    if err := ctx.BodyParser(&req); err != nil {
        return ctx.Status(400).JSON(fiber.Map{
            "code": 400,
            "message": "请求参数错误",
        })
    }

    // 根据活动类型处理
    switch activityType {
    case "daily_recharge":
        return claimDailyRechargeReward(ctx, uint64(userID), req)
    default:
        return ctx.Status(400).JSON(fiber.Map{
            "code": 400,
            "message": "未知活动类型",
        })
    }
}
```

### 4. 轮盘抽奖

**请求信息：**
- **Method**: POST
- **Path**: `/wheel/spin`
- **Auth**: 需要JWT

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "coupon_code": "COUPON_20250202_123456",
    "coupon_value": 10.00,
    "min_deposit": 3.00,
    "valid_until": "2025-02-03T00:00:00Z",
    "wheel_position": 3,
    "result_label": "$10 Coupon"
  }
}
```

**错误码：**
| 错误码 | 说明 |
|--------|------|
| 4005 | 今日已抽奖 |
| 4006 | 轮盘活动未开启 |

**Go实现：**
```go
// SpinWheelResponse 抽奖响应
type SpinWheelResponse struct {
    CouponCode    string    `json:"coupon_code"`
    CouponValue   float64   `json:"coupon_value"`
    MinDeposit    float64   `json:"min_deposit"`
    ValidUntil    time.Time `json:"valid_until"`
    WheelPosition int       `json:"wheel_position"`
    ResultLabel   string    `json:"result_label"`
}

// SpinWheel 轮盘抽奖
func SpinWheel(ctx *fiber.Ctx) error {
    userID := QueryUserIdFromJwt(ctx)

    // 检查今日是否已抽奖
    today := time.Now().Format("2006-01-02")
    var existing dtos.UserActivityProgress
    if err := db.Where("user_id = ? AND activity_type = ? AND progress_date = ?",
        userID, "coupon_wheel", today).First(&existing).Error; err == nil {
        return ctx.Status(400).JSON(fiber.Map{
            "code": 4005,
            "message": "今日已抽奖",
        })
    }

    // 获取轮盘配置
    wheelConfig := getWheelConfig()

    // 抽奖逻辑
    result := drawPrize(wheelConfig)

    // 生成优惠券码
    couponCode := generateCouponCode(userID)

    // 创建优惠券记录
    coupon := dtos.UserCoupon{
        UserID:       uint64(userID),
        CouponCode:   couponCode,
        ActivityType: "coupon_wheel",
        CouponValue:  result.Value,
        MinDeposit:   3.00,
        Status:       0, // 未激活
        ValidStart:   time.Now(),
        ValidEnd:     time.Now().Add(24 * time.Hour),
    }

    if err := db.Create(&coupon).Error; err != nil {
        return ctx.Status(500).JSON(fiber.Map{
            "code": 500,
            "message": "创建优惠券失败",
        })
    }

    // 记录活动进度
    progress := dtos.UserActivityProgress{
        UserID:       uint64(userID),
        ActivityType: "coupon_wheel",
        ProgressData: datatypes.JSON(mustJSON(map[string]interface{}{
            "coupon_code": couponCode,
            "coupon_value": result.Value,
            "spun_at": time.Now(),
        })),
        Status:       2, // 已完成
        ProgressDate: time.Now(),
    }
    db.Create(&progress)

    return ctx.JSON(fiber.Map{
        "code": 0,
        "message": "success",
        "data": SpinWheelResponse{
            CouponCode:    couponCode,
            CouponValue:   result.Value,
            MinDeposit:    3.00,
            ValidUntil:    coupon.ValidEnd,
            WheelPosition: result.Position,
            ResultLabel:   result.Label,
        },
    })
}
```

### 5. 获取轮盘状态

**请求信息：**
- **Method**: GET
- **Path**: `/wheel/status`
- **Auth**: 需要JWT

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "can_spin": false,
    "last_spin_at": "2025-02-02T10:30:00Z",
    "next_spin_at": "2025-02-03T00:00:00Z",
    "today_coupon": {
      "code": "COUPON_20250202_123456",
      "value": 10.00,
      "status": "pending",
      "valid_until": "2025-02-03T00:00:00Z"
    }
  }
}
```

### 6. 获取用户优惠券列表

**请求信息：**
- **Method**: GET
- **Path**: `/coupons`
- **Auth**: 需要JWT

**查询参数：**
| 参数 | 类型 | 说明 |
|------|------|------|
| status | int | 过滤状态 (0:未激活 1:已激活 2:已使用 3:已过期) |
| page | int | 页码，默认1 |
| page_size | int | 每页数量，默认10 |

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 5,
    "page": 1,
    "page_size": 10,
    "list": [
      {
        "id": 1,
        "coupon_code": "COUPON_20250202_123456",
        "coupon_value": 10.00,
        "min_deposit": 3.00,
        "status": 1,
        "status_text": "已激活",
        "valid_start": "2025-02-02T10:30:00Z",
        "valid_end": "2025-02-03T10:30:00Z",
        "activated_at": "2025-02-02T14:00:00Z",
        "used_at": null
      }
    ]
  }
}
```

### 7. 激活优惠券

**请求信息：**
- **Method**: POST
- **Path**: `/coupons/:code/activate`
- **Auth**: 需要JWT

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "coupon_code": "COUPON_20250202_123456",
    "status": "activated",
    "activated_at": "2025-02-02T14:00:00Z"
  }
}
```

**激活条件检查：**
- 用户当日充值金额 >= 优惠券min_deposit
- 优惠券状态为未激活
- 优惠券未过期

### 8. 使用优惠券

**请求信息：**
- **Method**: POST
- **Path**: `/coupons/:code/use`
- **Auth**: 需要JWT

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "reward_amount": 10.00,
    "new_balance": 110.00,
    "transaction_id": "TXN123456790"
  }
}
```

### 9. 获取输返活动统计

**请求信息：**
- **Method**: GET
- **Path**: `/activities/loss_rebate/stats`
- **Auth**: 需要JWT

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "activity_id": 3,
    "activity_name": "亏损返还活动",
    "start_time": "2025-02-01T00:00:00Z",
    "end_time": "2025-02-14T23:59:59Z",
    "rebate_rate": 0.06,
    "min_loss_threshold": 10.00,
    "today_stats": {
      "date": "2025-02-02",
      "bet_amount": 500.00,
      "win_amount": 300.00,
      "net_loss": 200.00,
      "rebate_amount": 12.00,
      "claimed": false
    },
    "total_stats": {
      "total_bet": 1500.00,
      "total_win": 1200.00,
      "total_loss": 300.00,
      "total_rebate": 18.00,
      "claimed_rebate": 6.00,
      "pending_rebate": 12.00
    },
    "history": [
      {
        "date": "2025-02-01",
        "bet_amount": 1000.00,
        "win_amount": 900.00,
        "net_loss": 100.00,
        "rebate_amount": 6.00,
        "claimed": true,
        "claimed_at": "2025-02-02T00:00:00Z"
      }
    ]
  }
}
```

**Go实现：**
```go
// GetLossRebateStats 获取输返活动统计
func GetLossRebateStats(ctx *fiber.Ctx) error {
    userID := QueryUserIdFromJwt(ctx)

    // 获取当前输返活动
    var activity dtos.Activity
    if err := db.Where("type = ? AND status = ? AND start_time <= ? AND end_time >= ?",
        "loss_rebate", 1, time.Now(), time.Now()).First(&activity).Error; err != nil {
        return ctx.Status(404).JSON(fiber.Map{
            "code": 404,
            "message": "当前没有进行中的输返活动",
        })
    }

    // 解析活动配置
    var config struct {
        RebateRate       float64 `json:"rebate_rate"`
        MinLoss          float64 `json:"min_loss"`
        SettlementType   string  `json:"settlement_type"`
    }
    json.Unmarshal(activity.Config, &config)

    // 获取今日统计
    today := time.Now().Format("2006-01-02")
    var todayProgress dtos.UserActivityProgress
    var todayStats map[string]interface{}

    if err := db.Where("user_id = ? AND activity_type = ? AND progress_date = ?",
        userID, "loss_rebate", today).First(&todayProgress).Error; err == nil {
        json.Unmarshal(todayProgress.ProgressData, &todayStats)
    }

    // 获取累计统计
    var totalStats struct {
        TotalBet      float64
        TotalWin      float64
        TotalLoss     float64
        TotalRebate   float64
        ClaimedRebate float64
    }
    db.Model(&dtos.UserActivityProgress{}).
        Select(`
            COALESCE(SUM(JSON_EXTRACT(progress_data, '$.bet_amount')), 0) as total_bet,
            COALESCE(SUM(JSON_EXTRACT(progress_data, '$.win_amount')), 0) as total_win,
            COALESCE(SUM(JSON_EXTRACT(progress_data, '$.net_loss')), 0) as total_loss,
            COALESCE(SUM(JSON_EXTRACT(progress_data, '$.rebate_amount')), 0) as total_rebate,
            COALESCE(SUM(CASE WHEN status = 3 THEN JSON_EXTRACT(progress_data, '$.rebate_amount') ELSE 0 END), 0) as claimed_rebate
        `).
        Where("user_id = ? AND activity_type = ? AND progress_date >= ? AND progress_date <= ?",
            userID, "loss_rebate", activity.StartTime.Format("2006-01-02"), activity.EndTime.Format("2006-01-02")).
        Scan(&totalStats)

    return ctx.JSON(fiber.Map{
        "code": 0,
        "message": "success",
        "data": fiber.Map{
            "activity_id":        activity.ID,
            "activity_name":      activity.Name,
            "start_time":         activity.StartTime,
            "end_time":           activity.EndTime,
            "rebate_rate":        config.RebateRate,
            "min_loss_threshold": config.MinLoss,
            "today_stats":        todayStats,
            "total_stats": fiber.Map{
                "total_bet":       totalStats.TotalBet,
                "total_win":       totalStats.TotalWin,
                "total_loss":      totalStats.TotalLoss,
                "total_rebate":    totalStats.TotalRebate,
                "claimed_rebate":  totalStats.ClaimedRebate,
                "pending_rebate":  totalStats.TotalRebate - totalStats.ClaimedRebate,
            },
        },
    })
}
```

## 储值返利业务逻辑

### 充值监听处理

```go
// HandleDeposit 处理用户充值
func HandleDeposit(userID uint64, depositAmount float64) error {
    // 检查储值返利活动
    today := time.Now().Format("2006-01-02")

    var progress dtos.UserActivityProgress
    err := db.Where("user_id = ? AND activity_type = ? AND progress_date = ?",
        userID, "daily_recharge", today).First(&progress).Error

    if err == gorm.ErrRecordNotFound {
        // 创建新的进度记录
        dayNumber := getUserCurrentDayNumber(userID)
        rewardRate := getRewardRateByDay(dayNumber)

        progress = dtos.UserActivityProgress{
            UserID:       userID,
            ActivityType: "daily_recharge",
            DayNumber:    dayNumber,
            ProgressData: datatypes.JSON(mustJSON(map[string]interface{}{
                "deposit_amount": depositAmount,
                "min_deposit": 3.00,
                "reward_rate": rewardRate,
            })),
            Status:       0, // 未开始
            ProgressDate: time.Now(),
        }

        if depositAmount >= 3.00 {
            progress.Status = 2 // 已完成
        }

        return db.Create(&progress).Error
    }

    // 更新现有记录
    var progressData map[string]interface{}
    json.Unmarshal(progress.ProgressData, &progressData)

    currentDeposit := progressData["deposit_amount"].(float64)
    newDeposit := currentDeposit + depositAmount
    progressData["deposit_amount"] = newDeposit

    if newDeposit >= 3.00 && progress.Status < 2 {
        progress.Status = 2 // 标记为已完成
    }

    progress.ProgressData = datatypes.JSON(mustJSON(progressData))
    return db.Save(&progress).Error
}
```

### 奖励计算

```go
// CalculateDailyReward 计算每日返利金额
func CalculateDailyReward(dayNumber int, depositAmount float64) float64 {
    rates := map[int]float64{
        1: 0.50,
        2: 0.75,
        3: 1.00,
        4: 1.50,
    }

    rate, ok := rates[dayNumber]
    if !ok {
        return 0
    }

    return depositAmount * rate
}
```

## 轮盘抽奖概率配置

```go
// WheelReward 轮盘奖项
type WheelReward struct {
    Value       float64 `json:"value"`
    Probability float64 `json:"probability"`
    Label       string  `json:"label"`
    Position    int     `json:"position"`
}

// getWheelConfig 获取轮盘配置
func getWheelConfig() []WheelReward {
    return []WheelReward{
        {Value: 1, Probability: 0.30, Label: "$1 Coupon", Position: 0},
        {Value: 2, Probability: 0.25, Label: "$2 Coupon", Position: 1},
        {Value: 5, Probability: 0.20, Label: "$5 Coupon", Position: 2},
        {Value: 10, Probability: 0.15, Label: "$10 Coupon", Position: 3},
        {Value: 50, Probability: 0.08, Label: "$50 Coupon", Position: 4},
        {Value: 100, Probability: 0.02, Label: "$100 Coupon", Position: 5},
    }
}

// drawPrize 抽奖
func drawPrize(config []WheelReward) WheelReward {
    rand.Seed(time.Now().UnixNano())
    r := rand.Float64()

    cumulative := 0.0
    for _, reward := range config {
        cumulative += reward.Probability
        if r <= cumulative {
            return reward
        }
    }

    return config[0] // 默认返回第一个
}
```

## 输返活动业务逻辑

### 下注/派彩监听处理

```go
// HandleGameBet 处理游戏下注和派彩
func HandleGameBet(userID uint64, betAmount, winAmount float64) error {
    // 检查是否有进行中的输返活动
    var activity dtos.Activity
    err := db.Where("type = ? AND status = ? AND start_time <= ? AND end_time >= ?",
        "loss_rebate", 1, time.Now(), time.Now()).First(&activity).Error

    if err != nil {
        return nil // 没有进行中的输返活动
    }

    // 解析活动配置
    var config struct {
        RebateRate     float64 `json:"rebate_rate"`
        MinLoss        float64 `json:"min_loss"`
        SettlementType string  `json:"settlement_type"`
    }
    json.Unmarshal(activity.Config, &config)

    today := time.Now().Format("2006-01-02")

    // 查找或创建今日进度记录
    var progress dtos.UserActivityProgress
    err = db.Where("user_id = ? AND activity_type = ? AND progress_date = ?",
        userID, "loss_rebate", today).First(&progress).Error

    var progressData map[string]interface{}

    if err == gorm.ErrRecordNotFound {
        // 创建新记录
        netLoss := math.Max(0, betAmount-winAmount)
        rebateAmount := 0.0
        if netLoss >= config.MinLoss {
            rebateAmount = netLoss * config.RebateRate
        }

        progress = dtos.UserActivityProgress{
            UserID:       userID,
            ActivityType: "loss_rebate",
            ProgressData: datatypes.JSON(mustJSON(map[string]interface{}{
                "activity_id":      activity.ID,
                "bet_amount":       betAmount,
                "win_amount":       winAmount,
                "net_loss":         netLoss,
                "rebate_rate":      config.RebateRate,
                "rebate_amount":    rebateAmount,
                "min_loss_threshold": config.MinLoss,
                "claimed":          false,
                "calculated_at":    time.Now(),
            })),
            Status:       2, // 已完成（待领取）
            ProgressDate: time.Now(),
        }

        return db.Create(&progress).Error
    }

    // 更新现有记录
    json.Unmarshal(progress.ProgressData, &progressData)

    currentBet := progressData["bet_amount"].(float64)
    currentWin := progressData["win_amount"].(float64)

    newBet := currentBet + betAmount
    newWin := currentWin + winAmount
    newNetLoss := math.Max(0, newBet-newWin)

    newRebateAmount := 0.0
    if newNetLoss >= config.MinLoss {
        newRebateAmount = newNetLoss * config.RebateRate
    }

    progressData["bet_amount"] = newBet
    progressData["win_amount"] = newWin
    progressData["net_loss"] = newNetLoss
    progressData["rebate_amount"] = newRebateAmount
    progressData["calculated_at"] = time.Now()

    progress.ProgressData = datatypes.JSON(mustJSON(progressData))
    return db.Save(&progress).Error
}
```

### 输返金额计算

```go
// CalculateLossRebate 计算输返金额
func CalculateLossRebate(betAmount, winAmount, rebateRate, minLossThreshold float64) float64 {
    netLoss := betAmount - winAmount
    if netLoss <= 0 {
        return 0
    }
    if netLoss < minLossThreshold {
        return 0
    }
    return netLoss * rebateRate
}
```

### 定时结算任务（如果配置为日结）

```go
// DailyLossRebateSettlement 每日输返结算
func DailyLossRebateSettlement() {
    yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

    // 查找所有有输返待领取的用户
    var progresses []dtos.UserActivityProgress
    db.Where("activity_type = ? AND progress_date = ? AND status = ? AND JSON_EXTRACT(progress_data, '$.rebate_amount') > 0",
        "loss_rebate", yesterday, 2).Find(&progresses)

    for _, progress := range progresses {
        var data map[string]interface{}
        json.Unmarshal(progress.ProgressData, &data)

        rebateAmount := data["rebate_amount"].(float64)
        if rebateAmount <= 0 {
            continue
        }

        // 发放奖励到用户余额
        // ... 发放逻辑 ...

        // 更新状态为已领取
        progress.Status = 3
        data["claimed"] = true
        data["claimed_at"] = time.Now()
        progress.ProgressData = datatypes.JSON(mustJSON(data))
        db.Save(&progress)
    }
}
```

## 文件结构

```
be/
├── controllers/
│   ├── router.go              # 路由定义（新增活动路由）
│   ├── activity_handler.go    # 活动相关handler（新增）
│   └── coupon_handler.go      # Coupon相关handler（新增）
├── models/
│   └── dtos/
│       ├── models.gen.go      # 现有模型
│       └── activity_models.go # 活动相关模型（新增）
├── services/
│   ├── activity_service.go    # 活动业务逻辑（新增）
│   └── coupon_service.go      # Coupon业务逻辑（新增）
└── utils/
    └── activity_utils.go      # 活动工具函数（新增）
```

## 错误码定义

| 错误码 | 英文标识 | 中文说明 |
|--------|----------|----------|
| 4001 | ERR_REWARD_CONDITION_NOT_MET | 未达到领取条件 |
| 4002 | ERR_REWARD_ALREADY_CLAIMED | 奖励已领取 |
| 4003 | ERR_ACTIVITY_EXPIRED | 活动已过期 |
| 4004 | ERR_ACTIVITY_NOT_STARTED | 活动未开始 |
| 4005 | ERR_WHEEL_ALREADY_SPUN | 今日已抽奖 |
| 4006 | ERR_WHEEL_NOT_AVAILABLE | 轮盘活动未开启 |
| 4007 | ERR_COUPON_NOT_FOUND | 优惠券不存在 |
| 4008 | ERR_COUPON_ALREADY_USED | 优惠券已使用 |
| 4009 | ERR_COUPON_EXPIRED | 优惠券已过期 |
| 4010 | ERR_COUPON_NOT_ACTIVATED | 优惠券未激活 |
| 4011 | ERR_COUPON_ACTIVATE_CONDITION | 未达到激活条件 |
| 4012 | ERR_LOSS_REBATE_NO_ACTIVITY | 当前无输返活动 |
| 4013 | ERR_LOSS_REBATE_NO_LOSS | 无亏损不需返还 |
| 4014 | ERR_LOSS_REBATE_BELOW_THRESHOLD | 亏损未达到返还门槛 |

## 性能考虑

1. **缓存策略**: 活动配置可缓存5分钟，减少数据库查询
2. **并发控制**: 抽奖和领取奖励使用数据库事务和乐观锁
3. **限流**: 抽奖API添加限流，防止恶意请求
