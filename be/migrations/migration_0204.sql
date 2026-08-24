-- ============================================
-- 迁移脚本: migration_0204.sql
-- 说明: 2025-02-04 支付功能数据库变更
-- 创建支付订单表、扩展用户余额表
-- ============================================

-- ============================================
-- 1. 创建支付订单表 (payment_orders)
-- ============================================
DROP TABLE IF EXISTS `payment_orders`;

CREATE TABLE `payment_orders` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `order_id` VARCHAR(64) NOT NULL COMMENT '商户订单号',
    `plat_order_id` VARCHAR(64) DEFAULT NULL COMMENT '平台订单号',
    `amount` DECIMAL(15,2) NOT NULL COMMENT '金额（印尼盾）',
    `cost` DECIMAL(15,2) DEFAULT 0.00 COMMENT '手续费',
    `type` VARCHAR(20) NOT NULL COMMENT '支付类型: bank/ewallet/qris',
    `dst_code` VARCHAR(50) NOT NULL COMMENT '支付方式编码',
    `status` TINYINT NOT NULL DEFAULT 0 COMMENT '0:待支付 1:成功 2:失败 3:处理中',
    `ref_code` INT DEFAULT NULL COMMENT '业务状态码',
    `ref_msg` VARCHAR(255) DEFAULT NULL COMMENT '业务描述',
    `pay_url` VARCHAR(500) DEFAULT NULL COMMENT '支付链接',
    `product_info` VARCHAR(255) DEFAULT NULL COMMENT '产品信息',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `paid_at` DATETIME DEFAULT NULL COMMENT '支付完成时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_order_id` (`order_id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_status` (`status`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付订单表';

-- ============================================
-- 2. 用户表扩展字段 (users.balance 已存在)
-- ============================================
-- 检查字段是否存在，不存在则添加

-- 添加累计充值字段
SET @exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'users'
    AND COLUMN_NAME = 'total_deposit');
SET @sql := IF(@exists = 0,
    'ALTER TABLE `users` ADD COLUMN `total_deposit` DECIMAL(15,2) DEFAULT 0.00 COMMENT "累计充值"',
    'SELECT "total_deposit字段已存在"');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 添加累计提现字段
SET @exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'users'
    AND COLUMN_NAME = 'total_withdraw');
SET @sql := IF(@exists = 0,
    'ALTER TABLE `users` ADD COLUMN `total_withdraw` DECIMAL(15,2) DEFAULT 0.00 COMMENT "累计提现"',
    'SELECT "total_withdraw字段已存在"');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ============================================
-- 3. 余额变动记录使用现有的 transactions 表
-- ============================================
-- 支付回调成功后，需要写入 transactions 表:
-- type = 1 (充值)
-- amount = 支付金额 (正数, decimal(15,2))
-- reference_id = payment_orders.order_id
-- remark = 支付方式，如 "DANA充值"、"BCA银行转账"
-- ============================================
-- 检查 transactions 表是否存在，不存在则创建（兼容新部署）
SET @exists := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
    WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'transactions');

SET @sql := IF(@exists = 0, '
CREATE TABLE `transactions` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) UNSIGNED NOT NULL,
  `type` tinyint(4) NOT NULL COMMENT "1:充值 2:提现 3:游戏下注 4:游戏派彩 5:分销返点 6:活动赠送",
  `amount` decimal(15,2) NOT NULL COMMENT "变动金额 (正负)",
  `before_balance` decimal(15,2) NOT NULL COMMENT "变动前余额",
  `after_balance` decimal(15,2) NOT NULL COMMENT "变动后余额",
  `reference_id` varchar(100) DEFAULT "" COMMENT "关联业务ID (订单号/游戏局号)",
  `remark` varchar(255) DEFAULT "",
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_created` (`user_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT="资金流水表"
', 'SELECT "transactions表已存在，跳过创建"');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ============================================
-- 4. 创建代付订单表 (withdraw_orders) - 预留
-- ============================================
DROP TABLE IF EXISTS `withdraw_orders`;

CREATE TABLE `withdraw_orders` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `order_id` VARCHAR(64) NOT NULL COMMENT '商户订单号',
    `plat_order_id` VARCHAR(64) DEFAULT NULL COMMENT '平台订单号',
    `amount` DECIMAL(15,2) NOT NULL COMMENT '金额（印尼盾）',
    `cost` DECIMAL(15,2) DEFAULT 0.00 COMMENT '手续费',
    `type` VARCHAR(20) NOT NULL COMMENT '类型: bankcard/ewallet',
    `dst_code` VARCHAR(50) NOT NULL COMMENT '代付编码',
    `account` VARCHAR(50) NOT NULL COMMENT '银行卡号/钱包账号',
    `account_name` VARCHAR(100) NOT NULL COMMENT '账户姓名',
    `phone` VARCHAR(20) NOT NULL COMMENT '手机号',
    `email` VARCHAR(100) NOT NULL COMMENT '邮箱',
    `address` VARCHAR(255) DEFAULT NULL COMMENT '地址',
    `status` TINYINT NOT NULL DEFAULT 0 COMMENT '0:待处理 1:成功 2:失败 3:处理中',
    `ref_code` INT DEFAULT NULL COMMENT '业务状态码',
    `ref_msg` VARCHAR(255) DEFAULT NULL COMMENT '业务描述',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `completed_at` DATETIME DEFAULT NULL COMMENT '完成时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_order_id` (`order_id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_status` (`status`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='代付订单表(提款)';

-- ============================================
-- 5. 初始化测试数据（可选）
-- ============================================
-- 注意：生产环境请删除以下测试数据

-- ============================================
-- 迁移完成
-- ============================================
-- 执行命令: mysql -u root -p tols < be/migrations/migration_0204.sql
