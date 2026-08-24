-- 迁移脚本: migration_0203.sql
-- 说明: 2025-02-03 活动功能数据库变更
-- 创建用户活动进度表、用户优惠券表、活动详细配置表

-- ============================================
-- 0. 清理旧表（如需要重新执行迁移）
-- 警告：执行此脚本将清空活动相关表数据！
-- ============================================
DROP TABLE IF EXISTS `activity_config`;
DROP TABLE IF EXISTS `user_coupons`;
DROP TABLE IF EXISTS `user_activity_progress`;
DROP TABLE IF EXISTS `activities`;

-- ============================================
-- 1. 创建活动基础表（activities）
-- ============================================
CREATE TABLE `activities` (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `type` varchar(50) NOT NULL COMMENT '活动类型 (recharge_rebate, coupon_wheel, loss_rebate)',
  `name` varchar(100) NOT NULL COMMENT '活动名称',
  `description` varchar(255) DEFAULT NULL COMMENT '活动描述',
  `config` json NULL COMMENT '活动配置JSON',
  `start_time` datetime NOT NULL COMMENT '活动开始时间',
  `end_time` datetime NOT NULL COMMENT '活动结束时间',
  `status` tinyint(1) DEFAULT 1 COMMENT '1:启用 0:禁用',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_type` (`type`),
  KEY `idx_status` (`status`),
  KEY `idx_start_time` (`start_time`),
  KEY `idx_end_time` (`end_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='运营活动配置';

-- ============================================
-- 2. 创建用户活动进度表
-- ============================================
CREATE TABLE IF NOT EXISTS `user_activity_progress` (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` bigint UNSIGNED NOT NULL COMMENT '用户ID',
  `activity_type` varchar(50) NOT NULL COMMENT '活动类型 (recharge_rebate, coupon_wheel, loss_rebate)',
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

-- ============================================
-- 3. 创建用户优惠券表
-- ============================================
CREATE TABLE IF NOT EXISTS `user_coupons` (
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

-- ============================================
-- 4. 创建活动详细配置表
-- ============================================
CREATE TABLE IF NOT EXISTS `activity_config` (
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

-- ============================================
-- 5. 插入初始活动配置数据
-- ============================================

-- 5.1 储值返利活动配置（基于充值次数，非连续日期）
-- 1st: 50%, 2nd: 75%, 3rd: 100%, 4th: 150%
INSERT INTO `activities` (`type`, `name`, `config`, `start_time`, `end_time`, `status`)
SELECT 'recharge_rebate', 'Welcome Bonus', '{"min_deposit": 3, "currency": "USD", "repeatable": false}', '2025-02-01 00:00:00', '2025-02-28 23:59:59', 1
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1 FROM `activities`
    WHERE `type` = 'recharge_rebate'
    AND `start_time` = '2025-02-01 00:00:00'
);

-- 获取刚插入的储值返利活动ID并插入详细配置
SET @recharge_activity_id = LAST_INSERT_ID();

INSERT INTO `activity_config` (`activity_id`, `config_key`, `config_value`, `description`)
SELECT @recharge_activity_id, 'tier_rewards', '[
  {"day": 1, "rate": 0.50, "label": "1st", "description": "50% on the first deposit"},
  {"day": 2, "rate": 0.75, "label": "2nd", "description": "75% on the second deposit"},
  {"day": 3, "rate": 1.00, "label": "3rd", "description": "100% + 75 FS on the third deposit"},
  {"day": 4, "rate": 1.50, "label": "4th", "description": "Up to 150% + 125 FS on the fourth deposit"}
]', '充值返利档次配置（1st-4th deposit）'
FROM DUAL
WHERE @recharge_activity_id > 0
AND NOT EXISTS (
    SELECT 1 FROM `activity_config`
    WHERE `activity_id` = @recharge_activity_id
    AND `config_key` = 'tier_rewards'
);

-- 5.2 轮盘活动配置
INSERT INTO `activities` (`type`, `name`, `config`, `start_time`, `end_time`, `status`)
SELECT 'coupon_wheel', 'Daily Lucky Wheel', '{"min_deposit": 3, "daily_limit": 1, "currency": "USD"}', '2025-01-01 00:00:00', '2025-12-31 23:59:59', 1
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1 FROM `activities`
    WHERE `type` = 'coupon_wheel'
);

SET @wheel_activity_id = LAST_INSERT_ID();

INSERT INTO `activity_config` (`activity_id`, `config_key`, `config_value`, `description`)
SELECT @wheel_activity_id, 'wheel_rewards', '[
  {"value": 1, "probability": 0.30, "label": "$1 Coupon"},
  {"value": 2, "probability": 0.25, "label": "$2 Coupon"},
  {"value": 5, "probability": 0.20, "label": "$5 Coupon"},
  {"value": 10, "probability": 0.15, "label": "$10 Coupon"},
  {"value": 50, "probability": 0.08, "label": "$50 Coupon"},
  {"value": 100, "probability": 0.02, "label": "$100 Coupon"}
]', '轮盘奖项配置'
FROM DUAL
WHERE @wheel_activity_id > 0
AND NOT EXISTS (
    SELECT 1 FROM `activity_config`
    WHERE `activity_id` = @wheel_activity_id
    AND `config_key` = 'wheel_rewards'
);

-- 5.3 输返活动配置（14天输返活动）
INSERT INTO `activities` (`type`, `name`, `config`, `start_time`, `end_time`, `status`)
SELECT 'loss_rebate', 'Loss Rebate', '{"rebate_rate": 0.06, "min_loss": 10, "currency": "USD", "settlement_type": "daily"}', '2025-02-01 00:00:00', '2025-02-14 23:59:59', 1
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1 FROM `activities`
    WHERE `type` = 'loss_rebate'
    AND `start_time` = '2025-02-01 00:00:00'
);

SET @loss_activity_id = LAST_INSERT_ID();

INSERT INTO `activity_config` (`activity_id`, `config_key`, `config_value`, `description`)
SELECT @loss_activity_id, 'rebate_rules', '{
  "rebate_rate": 0.06,
  "min_loss_threshold": 10.00,
  "max_rebate_amount": 1000.00,
  "settlement_cycle": "daily",
  "settlement_time": "00:00:00",
  "description": "活动期间内，每日净亏损按6%返还，最低返还门槛$10"
}', '输返活动规则配置'
FROM DUAL
WHERE @loss_activity_id > 0
AND NOT EXISTS (
    SELECT 1 FROM `activity_config`
    WHERE `activity_id` = @loss_activity_id
    AND `config_key` = 'rebate_rules'
);

-- ============================================
-- 迁移完成
-- ============================================
-- 执行命令: mysql -u root -p tols < be/migrations/migration_0203.sql
