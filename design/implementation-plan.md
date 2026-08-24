# 实施计划文档 - 活动功能

## 项目概述

本实施计划描述游戏平台活动功能的开发、测试和部署流程。

## 功能范围

### 已实现（现有代码）
- [x] 基础活动表结构
- [x] 活动页面框架
- [x] JWT认证系统
- [x] 用户系统
- [x] 充值系统

### 待开发功能

#### 阶段一：数据库和后端基础
- [ ] 新增活动相关数据表
- [ ] 活动模型定义
- [ ] 活动配置API
- [ ] 用户活动进度API

#### 阶段二：储值返利活动（可重复活动）
- [ ] 充值监听和返利计算
- [ ] 返利领取API
- [ ] 进度查询API（支持多活动实例并行）
- [ ] 活动周期管理（开始时间、结束时间、4天周期）

#### 阶段三：轮盘抽奖活动
- [ ] 抽奖API
- [ ] 优惠券系统
- [ ] 优惠券激活/使用API

#### 阶段四：前端组件
- [ ] 类型定义
- [ ] API服务
- [ ] 活动卡片组件
- [ ] 轮盘弹窗组件
- [ ] **UI设计图实现（design/photo/）**
  - [ ] 1-page.png 活动页初始状态（盒子三种状态）
  - [ ] 1-activity-1-dialog.png 活动1详情弹窗
  - [ ] 2-activity-2-input-code-dialog.png 输入code弹窗
  - [ ] 2-2/2-3 验证成功/失败弹窗
  - [ ] spin-done.png 转盘结果弹窗
  - [ ] 3-activity-1-dialog.png 活动3说明弹窗
- [ ] 页面集成

#### 阶段五：管理后台
- [ ] 活动配置管理
- [ ] 用户活动记录查询
- [ ] 优惠券管理

#### 阶段六：测试与部署
- [ ] 单元测试
- [ ] 集成测试
- [ ] 部署上线

## 详细任务分解

### 阶段一：数据库和后端基础（预计2天）

#### 1.1 数据库迁移

**本次更新迁移文件：**
```sql
-- 文件: be/migrations/migration_0203.sql
-- 执行: mysql -u root -p tols < be/migrations/migration_0203.sql
-- 说明: 2025-02-03 活动功能数据库变更
```

**迁移文件清单：**
- [x] 创建 `migration_0203.sql` 迁移脚本
  - [x] 创建 `user_activity_progress` 表
  - [x] 创建 `user_coupons` 表
  - [x] 创建 `activity_config` 表
  - [x] 插入初始活动配置数据（储值返利、轮盘、输返活动）
- [ ] 审核迁移脚本
- [ ] 执行数据库迁移
- [ ] 验证表结构

**历史迁移文件（参考）：**
```sql
-- 文件: be/migrations/001_add_activity_tables.sql (模板参考)
```

#### 1.2 模型定义
文件位置：
- `be/models/dtos/activity_models.go` (新增)

任务清单：
- [ ] UserActivityProgress 结构体
- [ ] UserCoupon 结构体
- [ ] ActivityConfig 结构体
- [ ] JSON序列化方法
- [ ] 表名方法

#### 1.3 基础API
文件位置：
- `be/controllers/activity_handler.go` (新增)
- `be/services/activity_service.go` (新增)

任务清单：
- [ ] 获取活动列表 API
- [ ] 获取活动配置 Service
- [ ] 路由注册

### 阶段二：储值返利活动（预计2天）

#### 2.1 充值监听
文件位置：
- `be/services/deposit_service.go` (修改/新增)

任务清单：
- [ ] 充值事件监听
- [ ] 活动进度更新逻辑
- [ ] 返利金额计算

#### 2.2 返利API
文件位置：
- `be/controllers/activity_handler.go`

任务清单：
- [ ] 获取返利进度 API
- [ ] 领取返利奖励 API
- [ ] 奖励发放事务处理

### 阶段三：输返活动（预计2天）

#### 3.1 下注/派彩监听
文件位置：
- `be/services/game_service.go` (修改)

任务清单：
- [ ] 下注事件监听
- [ ] 派彩事件监听
- [ ] 净亏损计算
- [ ] 输返金额实时计算

#### 3.2 输返API
文件位置：
- `be/controllers/activity_handler.go`
- `be/services/activity_service.go`

任务清单：
- [ ] 获取输返统计API
- [ ] 领取输返奖励API
- [ ] 活动周期管理（开始时间、持续时间）
- [ ] 定时结算任务（如配置为日结）

#### 3.3 输返配置
```sql
-- 插入输返活动配置
INSERT INTO `activities` (`type`, `name`, `config`, `start_time`, `end_time`, `status`) VALUES
('loss_rebate', '亏损返还活动', '{"rebate_rate": 0.06, "min_loss": 10, "currency": "USD", "settlement_type": "daily"}', '2025-02-01 00:00:00', '2025-02-14 23:59:59', 1);
```

### 阶段四：轮盘抽奖活动（预计2天）

#### 4.1 抽奖逻辑
文件位置：
- `be/controllers/coupon_handler.go` (新增)
- `be/services/coupon_service.go` (新增)

任务清单：
- [ ] 抽奖概率算法
- [ ] 优惠券生成
- [ ] 抽奖限制检查

#### 4.2 优惠券系统
任务清单：
- [ ] 优惠券激活逻辑
- [ ] 优惠券使用逻辑
- [ ] 过期检查任务

### 阶段五：前端组件（预计3天）

#### 5.1 基础设置
文件位置：
- `fe/types/activity.ts` (新增)
- `fe/services/activityApi.ts` (新增)
- `fe/hooks/useActivity.ts` (新增)

任务清单：
- [ ] TypeScript类型定义
- [ ] API服务封装
- [ ] 自定义Hook

#### 5.2 组件开发
文件位置：
- `fe/components/activity/DailyRechargeCard.tsx`
- `fe/components/activity/CouponWheelCard.tsx`
- `fe/components/activity/LossRebateCard.tsx` (新增)
- `fe/components/activity/WheelModal.tsx`
- `fe/components/activity/CouponList.tsx`
- `fe/components/activity/RebateProgressBox.tsx` (新增，1-page盒子状态)
- `fe/components/activity/ActivityDetailModal.tsx` (新增，活动详情弹窗)
- `fe/components/activity/CouponInputModal.tsx` (新增，输入code弹窗)
- `fe/components/activity/CouponValidationResult.tsx` (新增，验证结果弹窗)
- `fe/components/ui/Modal.tsx` (新增，通用弹窗组件)

任务清单：
- [ ] 储值返利卡片UI
- [ ] 轮盘活动卡片UI
- [ ] 输返活动卡片UI
- [ ] 轮盘弹窗组件
- [ ] 优惠券列表组件
- [ ] **UI设计实现（根据design/photo/）**
  - [ ] RebateProgressBox组件（盒子三种状态）
  - [ ] ActivityDetailModal组件（活动1详情，含三张图片）
  - [ ] CouponInputModal组件（输入code）
  - [ ] CouponValidationResult组件（成功/失败提示）
  - [ ] WheelModal扩展（spin-done结果展示）
  - [ ] InfoModal组件（活动3说明）
  - [ ] 图片资源导入public目录

#### 5.3 页面集成
文件位置：
- `fe/components/ActivityPage.tsx` (修改)

任务清单：
- [ ] 整合新组件
- [ ] 状态管理
- [ ] 错误处理

### 阶段六：管理后台（预计2天）

#### 5.1 活动管理
文件位置：
- `bg/server/model/example/activities.go` (新增)
- `bg/server/api/v1/example/activity.go` (新增)

任务清单：
- [ ] 活动CRUD API
- [ ] 活动配置管理
- [ ] 前端页面

#### 5.2 用户活动记录
任务清单：
- [ ] 用户进度查询
- [ ] 优惠券记录查询
- [ ] 数据统计报表

### 阶段六：测试与部署（预计2天）

#### 6.1 测试
任务清单：
- [ ] 单元测试（后端）
- [ ] API接口测试
- [ ] 前端组件测试
- [ ] 集成测试

#### 6.2 部署
任务清单：
- [ ] 数据库迁移
- [ ] 后端部署
- [ ] 前端部署
- [ ] 监控配置

## 文件创建清单

### 后端文件 (be/)

```
be/
├── models/dtos/
│   └── activity_models.go          # 活动相关模型
├── controllers/
│   ├── activity_handler.go         # 活动API处理器
│   └── coupon_handler.go           # 优惠券API处理器
├── services/
│   ├── activity_service.go         # 活动业务逻辑
│   └── coupon_service.go           # 优惠券业务逻辑
├── middleware/
│   └── activity_middleware.go      # 活动相关中间件（如需）
└── migrations/
    └── 001_add_activity_tables.sql # 数据库迁移脚本
```

### 前端文件 (fe/)

```
fe/
├── types/
│   └── activity.ts                 # 活动类型定义
├── services/
│   └── activityApi.ts              # 活动API服务
├── hooks/
│   └── useActivity.ts              # 活动相关Hook
└── components/activity/
    ├── DailyRechargeCard.tsx       # 储值返利卡片
    ├── CouponWheelCard.tsx         # 轮盘活动卡片
    ├── LossRebateCard.tsx          # 输返活动卡片
    ├── WheelModal.tsx              # 轮盘弹窗
    ├── CouponList.tsx              # 优惠券列表
    └── RewardClaimModal.tsx        # 奖励领取弹窗
```

### 管理后台文件 (bg/)

```
bg/server/
├── model/example/
│   ├── activities.go               # 活动模型
│   └── request/
│       └── activities.go           # 活动请求结构
├── api/v1/example/
│   └── activity.go                 # 活动API
└── router/example/
    └── activity.go                 # 活动路由
```

## 依赖关系

```
数据库迁移
    ↓
模型定义
    ↓
后端API开发 ──┬── 储值返利API
             ├── 输返活动API
             └── 轮盘抽奖API
                  ↓
前端类型定义
    ↓
前端API服务 ──┬── 活动API
              └── 优惠券API
                   ↓
前端组件开发 ──┬── 储值返利卡片
              ├── 输返卡片
              ├── 轮盘卡片
              ├── 轮盘弹窗
              └── 优惠券列表
                   ↓
页面集成
    ↓
管理后台开发
    ↓
测试
    ↓
部署
```

## 风险与应对措施

| 风险 | 可能性 | 影响 | 应对措施 |
|------|--------|------|----------|
| 抽奖概率算法问题 | 中 | 高 | 充分测试，使用随机数种子 |
| 并发抽奖冲突 | 中 | 高 | 数据库唯一索引+事务 |
| 充值监听遗漏 | 低 | 高 | 添加重试机制和日志 |
| 下注/派彩统计遗漏 | 低 | 高 | 异步补偿机制，定时对账 |
| 输返金额计算误差 | 低 | 高 | 使用Decimal类型，避免浮点误差 |
| 时区问题 | 中 | 中 | 统一使用UTC，前端转换 |
| 性能问题 | 低 | 中 | 添加缓存，优化查询 |

## 数据库回滚方案

```sql
-- 回滚脚本: rollback_activity_tables.sql
DROP TABLE IF EXISTS `user_activity_progress`;
DROP TABLE IF EXISTS `user_coupons`;
DROP TABLE IF EXISTS `activity_config`;
```

## 配置说明

### 活动配置（数据库）

```sql
-- 储值返利活动配置（可重复进行的活动，每次活动独立计算4天进度）
INSERT INTO `activities` (`type`, `name`, `config`, `start_time`, `end_time`, `status`) VALUES
('recharge_rebate', '储值返利活动', '{"min_deposit": 3, "currency": "USD", "repeatable": true, "cycle_days": 4}', '2025-02-01 00:00:00', '2025-02-28 23:59:59', 1);

INSERT INTO `activity_config` (`activity_id`, `config_key`, `config_value`, `description`) VALUES
(1, 'day_rewards', '[
  {"day": 1, "rate": 0.50, "label": "+50%", "description": "第1日充值返利50%"},
  {"day": 2, "rate": 0.75, "label": "+75%", "description": "第2日充值返利75%"},
  {"day": 3, "rate": 1.00, "label": "+100%", "description": "第3日充值返利100%"},
  {"day": 4, "rate": 1.50, "label": "+150%", "description": "第4日充值返利150%"}
]', '每日返利比例配置');

-- 轮盘活动配置
INSERT INTO `activities` (`type`, `name`, `config`, `start_time`, `end_time`, `status`) VALUES
('coupon_wheel', '每日幸运轮盘', '{"min_deposit": 3, "daily_limit": 1, "currency": "USD"}', '2025-01-01 00:00:00', '2025-12-31 23:59:59', 1);

INSERT INTO `activity_config` (`activity_id`, `config_key`, `config_value`, `description`) VALUES
(2, 'wheel_rewards', '[
  {"value": 1, "probability": 0.30, "label": "$1 Coupon"},
  {"value": 2, "probability": 0.25, "label": "$2 Coupon"},
  {"value": 5, "probability": 0.20, "label": "$5 Coupon"},
  {"value": 10, "probability": 0.15, "label": "$10 Coupon"},
  {"value": 50, "probability": 0.08, "label": "$50 Coupon"},
  {"value": 100, "probability": 0.02, "label": "$100 Coupon"}
]', '轮盘奖项配置');

-- 输返活动配置（14天活动）
INSERT INTO `activities` (`type`, `name`, `config`, `start_time`, `end_time`, `status`) VALUES
('loss_rebate', '亏损返还活动', '{"rebate_rate": 0.06, "min_loss": 10, "currency": "USD", "settlement_type": "daily"}', '2025-02-01 00:00:00', '2025-02-14 23:59:59', 1);

INSERT INTO `activity_config` (`activity_id`, `config_key`, `config_value`, `description`) VALUES
(3, 'rebate_rules', '{
  "rebate_rate": 0.06,
  "min_loss_threshold": 10.00,
  "max_rebate_amount": 1000.00,
  "settlement_cycle": "daily",
  "description": "活动期间内，每日净亏损按6%返还，最低返还门槛$10"
}', '输返活动规则配置');
```

## UI设计实现计划

根据设计图 `design/photo/` 目录下的示意图，制定以下详细实现计划。

### 设计图清单

| 设计图 | 对应功能 | 说明 |
|--------|----------|------|
| `1-page.png` | 活动展示页初始状态 | 上方活动卡片 + 下方储值返利进度 |
| `1-activity-1-dialog.png` | 活动1详情弹窗 | More details弹窗，含三张说明图片 |
| `2-activity-2-input-code-dialog.png` | 活动2输入code弹窗 | 输入coupon code的弹出框 |
| `2-2-failed-input.png` | 验证失败弹窗 | 服务端验证失败后的提示 |
| `2-3-successful-input.png` | 验证成功弹窗 | 验证成功，含Deposite引导按钮 |
| `spin-done.png` | 转盘成功弹窗 | 获得coupon code的弹出框 |
| `3-activity-1-dialog.png` | 活动3说明弹窗 | 右上角按钮点击后的说明对话框 |

### 图片资源清单

**位于 `design/resource/` 目录：**

| 图片名 | 用途 | 所在设计图 |
|--------|------|------------|
| `box_available.png` | 可领奖但未领奖状态 | 1-page.png |
| `box_opened.png` | 已领奖状态 | 1-page.png |
| `box_unavailable.png` | 当前不可领奖状态 | 1-page.png |
| `gameconsole.png` | 活动1详情图片1 | 1-activity-1-dialog.png |
| `depositepig.png` | 活动1详情图片2 | 1-activity-1-dialog.png |
| `creditcard.png` | 活动1详情图片3 | 1-activity-1-dialog.png |
| `activity-2-desktop.webp` | 活动2 PC端背景图 | CouponWheelCard |
| `activity-2-mobile.webp` | 活动2 移动端背景图 | CouponWheelCard |
| `activity-3-desktop.webp` | 活动3 PC端背景图 | LossRebateCard |
| `activity-3-mobile.webp` | 活动3 移动端背景图 | LossRebateCard |

**图片使用说明：**
- 活动卡片背景图使用响应式设计，PC端和移动端分别加载对应尺寸图片
- 盒子状态图片用于储值返利活动进度展示
- 活动详情弹窗图片需要导入public目录供前端使用

### 各设计图实现任务

#### 1. 1-page.png - 活动展示页初始状态

**实现组件：** `ActivityPage.tsx`, `ActivityCard.tsx`

**实现要点：**
- [ ] 顶部活动卡片网格布局（响应式3列/2列/1列）
- [ ] 储值返利活动状态展示区域（支持多活动实例，按当前活动周期显示）
- [ ] 4个阶段盒子状态显示（每日进度独立计算）：
  - [ ] 导入 `box_available.png`（可领未领）
  - [ ] 导入 `box_opened.png`（已领取）
  - [ ] 导入 `box_unavailable.png`（未解锁）
- [ ] 盒子点击领取交互
- [ ] 活动进度指示器

**文件位置：**
- `fe/components/ActivityPage.tsx`
- `fe/components/activity/RebateProgressBox.tsx` (新增)

---

#### 2. 1-activity-1-dialog.png - 活动1详情弹窗

**实现组件：** `ActivityDetailModal.tsx`

**实现要点：**
- [ ] 弹窗组件基础结构（标题、内容区、关闭按钮）
- [ ] 三张说明图片横向排列：
  - [ ] 导入 `gameconsole.png`
  - [ ] 导入 `depositepig.png`
  - [ ] 导入 `creditcard.png`
- [ ] 活动规则文字说明区域
- [ ] "More details" 按钮点击触发

**文件位置：**
- `fe/components/activity/ActivityDetailModal.tsx` (新增)

**Props定义：**
```typescript
interface ActivityDetailModalProps {
  isOpen: boolean;
  onClose: () => void;
  activityType: 'recharge_rebate' | 'coupon_wheel' | 'loss_rebate';
}
```

---

#### 3. 2-activity-2-input-code-dialog.png - 活动2输入code弹窗

**实现组件：** `CouponInputModal.tsx`

**实现要点：**
- [ ] 输入框组件（大写字母+数字）
- [ ] 提交按钮
- [ ] 输入验证（前端基础验证）
- [ ] 背景遮罩和点击关闭

**文件位置：**
- `fe/components/activity/CouponInputModal.tsx` (新增)

**API集成：**
- 调用 `activityApi.validateCouponCode(code)`

---

#### 4. 2-2-failed-input.png - 验证失败弹窗

**实现组件：** `CouponValidationResult.tsx` 或通用 `AlertModal.tsx`

**实现要点：**
- [ ] 错误图标（红色X或警告图标）
- [ ] 错误信息文字
- [ ] "Try again" 按钮
- [ ] 关闭按钮

**状态管理：**
- 显示条件：API返回验证失败状态

---

#### 5. 2-3-successful-input.png - 验证成功弹窗

**实现组件：** `CouponValidationResult.tsx`

**实现要点：**
- [ ] 成功图标（绿色对勾）
- [ ] 成功信息文字
- [ ] "Deposite" 充值引导按钮（点击进入充值页面）
- [ ] 关闭按钮

**交互逻辑：**
- Deposite按钮跳转：`router.push('/wallet')`

---

#### 6. spin-done.png - 转盘成功弹窗

**实现组件：** `WheelResultModal.tsx` (已有WheelModal扩展)

**实现要点：**
- [ ] 转盘停止动画
- [ ] 中奖信息展示（coupon金额）
- [ ] Coupon code显示（可复制）
- [ ] "立即充值激活" 引导按钮
- [ ] 关闭按钮

**文件位置：**
- 复用/扩展 `fe/components/activity/WheelModal.tsx`

---

#### 7. 3-activity-1-dialog.png - 活动3说明弹窗

**实现组件：** `ActivityInfoDialog.tsx` 或通用 `InfoModal.tsx`

**实现要点：**
- [ ] 简洁的说明弹窗
- [ ] 活动规则文字
- [ ] 关闭按钮

**触发位置：**
- 活动3卡片右上角信息图标按钮

---

### 通用弹窗组件抽象

考虑到多个弹窗有相似结构，建议提取通用组件：

**文件位置：** `fe/components/ui/Modal.tsx`

```typescript
interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title?: string;
  children: React.ReactNode;
  footer?: React.ReactNode;
  size?: 'sm' | 'md' | 'lg';
}
```

### 图片资源导入清单

**需要在public目录下放置的图片：**

```
public/
├── images/
│   └── activity/
│       ├── box_available.png          # 盒子可领状态
│       ├── box_opened.png             # 盒子已开状态
│       ├── box_unavailable.png        # 盒子未解锁状态
│       ├── gameconsole.png            # 活动1详情图1
│       ├── depositepig.png            # 活动1详情图2
│       ├── creditcard.png             # 活动1详情图3
│       ├── activity-2-desktop.webp    # 活动2 PC背景
│       ├── activity-2-mobile.webp     # 活动2 移动端背景
│       ├── activity-3-desktop.webp    # 活动3 PC背景
│       └── activity-3-mobile.webp     # 活动3 移动端背景
```

### 状态流转图

```
用户操作                      弹窗状态
   │                            │
   ├─ 点击More details ───────► 显示ActivityDetailModal
   │                            │
   ├─ 点击输入code ───────────► 显示CouponInputModal
   │                            │
   │   ├─ 验证失败 ───────────► 显示失败提示
   │   │                        │
   │   └─ 验证成功 ───────────► 显示成功提示 + Deposite按钮
   │                            │
   ├─ 点击转盘 ───────────────► 显示WheelModal
   │                            │
   │   └─ 转盘结束 ───────────► 显示spin-done结果
   │                            │
   └─ 点击活动3信息图标 ──────► 显示InfoModal
```

## 开发与验收分工说明

> **注意**：以下工作由程序员和测试人员执行，不在AI实施范围内：

### 编译与构建
- 后端代码编译（Go编译）
- 前端代码构建（Next.js构建）
- 依赖包安装与版本管理
- 环境配置（开发/测试/生产环境）

### 测试执行
- 单元测试编写与执行
- 集成测试执行
- 端到端测试
- 性能测试
- 安全测试

### 部署与运维
- 服务器部署
- 数据库迁移执行
- 监控配置
- 日志收集

## 验收标准

### 功能验收

- [ ] 用户可以查看储值返利活动（可重复进行的活动）
- [ ] 用户可以查看当前活动周期内的4天进度
- [ ] 用户充值后返利进度更新（每日独立计算）
- [ ] 用户可以领取返利奖励
- [ ] 活动结束后用户可参与新的活动周期
- [ ] 用户可以进行每日轮盘抽奖
- [ ] 抽奖结果显示正确
- [ ] 优惠券有效期24小时
- [ ] 充值达标后优惠券可激活
- [ ] 激活后的优惠券可使用
- [ ] 用户可以查看输返活动
- [ ] 用户下注后输返统计实时更新
- [ ] 输返金额计算正确（亏损×6%）
- [ ] 达到门槛后可以领取输返奖励
- [ ] 活动时间范围限制有效

### 性能验收

- [ ] 活动列表加载 < 500ms
- [ ] 抽奖响应 < 1s
- [ ] 页面渲染无卡顿
- [ ] 并发请求正常处理

### 安全验收

- [ ] 抽奖次数限制有效
- [ ] 防重复领取
- [ ] 奖励计算准确
- [ ] 无越权访问

### UI设计验收（根据design/photo/设计图）

- [ ] 1-page.png: 活动页初始状态实现
  - [ ] 盒子状态图片正确显示（available/opened/unavailable）
  - [ ] 储值返利4个阶段状态正确
- [ ] 1-activity-1-dialog.png: 活动1详情弹窗
  - [ ] 三张说明图片正确显示（gameconsole/depositepig/creditcard）
  - [ ] 弹窗布局与设计图一致
- [ ] 2-activity-2-input-code-dialog.png: 输入code弹窗
  - [ ] 输入框样式正确
  - [ ] 提交按钮正常
- [ ] 2-2-failed-input.png: 验证失败提示
  - [ ] 错误提示样式正确
  - [ ] Try again按钮功能正常
- [ ] 2-3-successful-input.png: 验证成功提示
  - [ ] Deposite按钮跳转充值页
- [ ] spin-done.png: 转盘结果弹窗
  - [ ] Coupon code显示正确
  - [ ] 充值引导按钮功能正常
- [ ] 3-activity-1-dialog.png: 活动3说明弹窗
  - [ ] 信息按钮触发弹窗
  - [ ] 说明文字显示正确

## 后续优化建议

1. **数据分析**
   - 活动参与率统计
   - 用户留存分析
   - ROI计算

2. **功能扩展**
   - 更多活动类型
   - 分享裂变活动
   - VIP专属活动

3. **体验优化**
   - 动画效果增强
   - 推送通知
   - 活动提醒
