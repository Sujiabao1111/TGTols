# 数据库设计文档 - 活动功能

## 概述

本文档描述活动功能相关的数据库表结构设计。

## 现有表结构

### activities 表（已有）

存储活动基础配置。

```sql
CREATE TABLE `activities` (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `type` varchar(50) NOT NULL COMMENT '活动类型 (sign_in, recharge_bonus)',
  `name` varchar(100) NOT NULL,
  `config` json NULL COMMENT 'JSON存储规则配置',
  `start_time` datetime NULL,
  `end_time` datetime NULL,
  `status` tinyint(1) NULL DEFAULT 1 COMMENT '1:启用 0:禁用',
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_type` (`type`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='运营活动配置';
```

## 新增表结构

### 1. user_activity_progress 表

存储用户参与活动的进度记录。

```sql
CREATE TABLE `user_activity_progress` (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` bigint UNSIGNED NOT NULL COMMENT '用户ID',
  `activity_type` varchar(50) NOT NULL COMMENT '活动类型 (daily_recharge, coupon_wheel, loss_rebate)',
  `day_number` int DEFAULT 1 COMMENT '第几天（用于多日活动）',
  `progress_data` json NULL COMMENT '进度JSON数据',
  `status` tinyint(1) DEFAULT 0 COMMENT '0:未开始 1:进行中 2:已完成 3:已领取',
  `progress_date` date NOT NULL COMMENT '进度日期',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_activity_date` (`user_id`, `activity_type`, `progress_date`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_activity_type` (`activity_type`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户活动进度记录';
```

**progress_data JSON 结构示例：**

储值返利活动（可重复活动）：
```json
{
  "activity_id": 1,
  "day_number": 2,
  "deposit_amount": 50.00,
  "reward_amount": 37.50,
  "reward_rate": 0.75,
  "min_deposit": 3.00,
  "claimed": false,
  "activity_start": "2025-02-01T00:00:00Z"
}
```

轮盘活动：
```json
{
  "coupon_code": "COUPON_20250202_123456",
  "coupon_value": 10.00,
  "min_deposit_to_activate": 3.00,
  "spun_at": "2025-02-02T10:30:00Z",
  "activated": false,
  "activated_at": null
}
```

输返活动：
```json
{
  "activity_id": 3,
  "bet_amount": 1000.00,
  "win_amount": 800.00,
  "net_loss": 200.00,
  "rebate_rate": 0.06,
  "rebate_amount": 12.00,
  "min_loss_threshold": 10.00,
  "claimed": false,
  "calculated_at": "2025-02-02T23:59:59Z"
}
```

### 2. user_coupons 表

存储用户获得的coupon详情。

```sql
CREATE TABLE `user_coupons` (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` bigint UNSIGNED NOT NULL COMMENT '用户ID',
  `coupon_code` varchar(100) NOT NULL COMMENT '优惠券码',
  `activity_type` varchar(50) NOT NULL COMMENT '活动类型',
  `coupon_value` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '优惠券面值',
  `min_deposit` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '最低充值金额',
  `status` tinyint(1) DEFAULT 0 COMMENT '0:未激活 1:已激活 2:已使用 3:已过期',
  `valid_start` datetime NOT NULL COMMENT '有效期开始',
  `valid_end` datetime NOT NULL COMMENT '有效期结束',
  `activated_at` datetime NULL COMMENT '激活时间',
  `used_at` datetime NULL COMMENT '使用时间',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_coupon_code` (`coupon_code`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_valid_end` (`valid_end`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户优惠券表';
```

### 3. activity_config 表

存储活动详细配置（扩展activities表）。

```sql
CREATE TABLE `activity_config` (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `activity_id` int UNSIGNED NOT NULL COMMENT '关联activities表',
  `config_key` varchar(100) NOT NULL COMMENT '配置键名',
  `config_value` json NULL COMMENT '配置值（JSON格式）',
  `description` varchar(255) DEFAULT NULL COMMENT '配置说明',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_activity_key` (`activity_id`, `config_key`),
  KEY `idx_activity_id` (`activity_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='活动详细配置表';
```

## 活动配置数据示例

### 储值返利活动配置

```sql
-- 基础活动配置（可重复进行的活动）
INSERT INTO `activities` (`type`, `name`, `config`, `start_time`, `end_time`, `status`) VALUES
('recharge_rebate', '储值返利活动', '{"min_deposit": 3, "currency": "USD", "repeatable": true, "cycle_days": 4}', '2025-02-01 00:00:00', '2025-02-28 23:59:59', 1);

-- 详细配置
INSERT INTO `activity_config` (`activity_id`, `config_key`, `config_value`, `description`) VALUES
(1, 'day_rewards', '[
  {"day": 1, "rate": 0.50, "description": "第1日充值返利50%"},
  {"day": 2, "rate": 0.75, "description": "第2日充值返利75%"},
  {"day": 3, "rate": 1.00, "description": "第3日充值返利100%"},
  {"day": 4, "rate": 1.50, "description": "第4日充值返利150%"}
]', '每日返利比例配置');
```

### 轮盘活动配置

```sql
-- 基础活动配置
INSERT INTO `activities` (`type`, `name`, `config`, `start_time`, `end_time`, `status`) VALUES
('coupon_wheel', '每日幸运轮盘', '{"min_deposit": 3, "daily_limit": 1, "currency": "USD"}', '2025-01-01 00:00:00', '2025-12-31 23:59:59', 1);

-- 详细配置
INSERT INTO `activity_config` (`activity_id`, `config_key`, `config_value`, `description`) VALUES
(2, 'wheel_rewards', '[
  {"value": 1, "probability": 0.30, "label": "$1 Coupon"},
  {"value": 2, "probability": 0.25, "label": "$2 Coupon"},
  {"value": 5, "probability": 0.20, "label": "$5 Coupon"},
  {"value": 10, "probability": 0.15, "label": "$10 Coupon"},
  {"value": 50, "probability": 0.08, "label": "$50 Coupon"},
  {"value": 100, "probability": 0.02, "label": "$100 Coupon"}
]', '轮盘奖项配置');
```

### 输返活动配置

```sql
-- 基础活动配置（14天输返活动）
INSERT INTO `activities` (`type`, `name`, `config`, `start_time`, `end_time`, `status`) VALUES
('loss_rebate', '亏损返还活动', '{"rebate_rate": 0.06, "min_loss": 10, "currency": "USD", "settlement_type": "daily"}', '2025-02-01 00:00:00', '2025-02-14 23:59:59', 1);

-- 详细配置
INSERT INTO `activity_config` (`activity_id`, `config_key`, `config_value`, `description`) VALUES
(3, 'rebate_rules', '{
  "rebate_rate": 0.06,
  "min_loss_threshold": 10.00,
  "max_rebate_amount": 1000.00,
  "settlement_cycle": "daily",
  "settlement_time": "00:00:00",
  "description": "活动期间内，每日净亏损按6%返还，最低返还门槛$10"
}', '输返活动规则配置');
```

## GORM模型定义（Go）

### UserActivityProgress 模型

```go
package dtos

import (
	"time"
	"gorm.io/datatypes"
)

// UserActivityProgress 用户活动进度
type UserActivityProgress struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       uint64         `gorm:"not null;index:idx_user_activity_date" json:"user_id"`
	ActivityType string         `gorm:"size:50;not null;index:idx_user_activity_date" json:"activity_type"`
	DayNumber    int            `gorm:"default:1" json:"day_number"`
	ProgressData datatypes.JSON `gorm:"type:json" json:"progress_data"`
	Status       int            `gorm:"default:0;comment:0:未开始 1:进行中 2:已完成 3:已领取" json:"status"`
	ProgressDate time.Time      `gorm:"type:date;not null;index:idx_user_activity_date" json:"progress_date"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

func (UserActivityProgress) TableName() string {
	return "user_activity_progress"
}
```

### UserCoupon 模型

```go
// UserCoupon 用户优惠券
type UserCoupon struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       uint64    `gorm:"not null;index:idx_user_id" json:"user_id"`
	CouponCode   string    `gorm:"size:100;not null;uniqueIndex:idx_coupon_code" json:"coupon_code"`
	ActivityType string    `gorm:"size:50;not null" json:"activity_type"`
	CouponValue  float64   `gorm:"type:decimal(15,2);default:0.00" json:"coupon_value"`
	MinDeposit   float64   `gorm:"type:decimal(15,2);default:0.00" json:"min_deposit"`
	Status       int       `gorm:"default:0;comment:0:未激活 1:已激活 2:已使用 3:已过期" json:"status"`
	ValidStart   time.Time `json:"valid_start"`
	ValidEnd     time.Time `json:"valid_end"`
	ActivatedAt  *time.Time `json:"activated_at,omitempty"`
	UsedAt       *time.Time `json:"used_at,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (UserCoupon) TableName() string {
	return "user_coupons"
}
```

### ActivityConfig 模型

```go
// ActivityConfig 活动详细配置
type ActivityConfig struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	ActivityID  uint           `gorm:"not null;uniqueIndex:idx_activity_key" json:"activity_id"`
	ConfigKey   string         `gorm:"size:100;not null;uniqueIndex:idx_activity_key" json:"config_key"`
	ConfigValue datatypes.JSON `gorm:"type:json" json:"config_value"`
	Description string         `gorm:"size:255" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

func (ActivityConfig) TableName() string {
	return "activity_config"
}
```

## 索引说明

| 表名 | 索引名 | 字段 | 说明 |
|------|--------|------|------|
| user_activity_progress | idx_user_activity_date | user_id + activity_type + progress_date | 唯一索引，确保用户每日每活动一条记录 |
| user_activity_progress | idx_user_id | user_id | 查询用户所有活动进度 |
| user_activity_progress | idx_activity_type | activity_type | 按活动类型查询 |
| user_coupons | idx_coupon_code | coupon_code | 唯一索引，优惠券码唯一 |
| user_coupons | idx_user_id | user_id | 查询用户所有优惠券 |
| user_coupons | idx_valid_end | valid_end | 清理过期优惠券 |
| activity_config | idx_activity_key | activity_id + config_key | 唯一索引，活动配置键唯一 |

## 数据迁移脚本

```sql
-- 迁移脚本: add_activity_tables.sql
-- 创建用户活动进度表
CREATE TABLE IF NOT EXISTS `user_activity_progress` (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` bigint UNSIGNED NOT NULL COMMENT '用户ID',
  `activity_type` varchar(50) NOT NULL COMMENT '活动类型',
  `day_number` int DEFAULT 1 COMMENT '第几天',
  `progress_data` json NULL COMMENT '进度JSON数据',
  `status` tinyint(1) DEFAULT 0 COMMENT '状态',
  `progress_date` date NOT NULL COMMENT '进度日期',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_activity_date` (`user_id`, `activity_type`, `progress_date`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_activity_type` (`activity_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户活动进度记录';

-- 创建用户优惠券表
CREATE TABLE IF NOT EXISTS `user_coupons` (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` bigint UNSIGNED NOT NULL,
  `coupon_code` varchar(100) NOT NULL,
  `activity_type` varchar(50) NOT NULL,
  `coupon_value` decimal(15,2) NOT NULL DEFAULT 0.00,
  `min_deposit` decimal(15,2) NOT NULL DEFAULT 0.00,
  `status` tinyint(1) DEFAULT 0,
  `valid_start` datetime NOT NULL,
  `valid_end` datetime NOT NULL,
  `activated_at` datetime NULL,
  `used_at` datetime NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_coupon_code` (`coupon_code`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户优惠券表';

-- 创建活动详细配置表
CREATE TABLE IF NOT EXISTS `activity_config` (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `activity_id` int UNSIGNED NOT NULL,
  `config_key` varchar(100) NOT NULL,
  `config_value` json NULL,
  `description` varchar(255) DEFAULT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_activity_key` (`activity_id`, `config_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='活动详细配置表';
```

## 注意事项

1. **时区处理**: 所有datetime字段统一使用UTC存储，前端根据用户时区显示
2. **JSON字段**: 使用MySQL 5.7+的JSON类型，便于查询和索引
3. **软删除**: 本设计不使用软删除，活动记录为历史数据，保留原始状态
4. **数据归档**: 建议定期归档超过90天的user_activity_progress记录
