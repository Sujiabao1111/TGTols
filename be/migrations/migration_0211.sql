-- ============================================
-- 迁移脚本: migration_0211.sql
-- 说明: 新增玩家统一游戏交易统计表，按 daily / weekly / total 落库
-- ============================================

CREATE TABLE IF NOT EXISTS `user_game_transaction_stats` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '站内用户ID',
  `period_type` VARCHAR(16) NOT NULL COMMENT '统计类型: daily/weekly/total',
  `period_key` VARCHAR(32) NOT NULL COMMENT '统计键: 日=YYYY-MM-DD, 周=周起始日, 总=all',
  `period_start` DATETIME NOT NULL COMMENT '统计开始时间(UTC)',
  `period_end` DATETIME NOT NULL COMMENT '统计结束时间(UTC)',
  `login_id` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '第三方登录账号',
  `count` INT NOT NULL DEFAULT 0 COMMENT '注单数',
  `turnover` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '打码量/有效投注',
  `bet` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '下注量',
  `win` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '派彩',
  `winlose` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '输赢',
  `jp_share` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT 'JP分摊',
  `jp_win` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT 'JP中奖',
  `synced_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '最近同步时间',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_period_key` (`user_id`, `period_type`, `period_key`),
  KEY `idx_user_game_transaction_stats_user` (`user_id`),
  KEY `idx_user_game_transaction_stats_period_type` (`period_type`),
  KEY `idx_user_game_transaction_stats_period_start` (`period_start`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='玩家统一游戏交易统计表';

-- 执行命令: mysql -u root -p tols < be/migrations/migration_0211.sql
 
 