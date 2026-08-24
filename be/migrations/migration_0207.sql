-- ============================================
-- 迁移脚本: migration_0207.sql
-- 说明: 新增玩家第三方游戏逐笔流水表，用于落库 TransactionDetail 报表
-- ============================================

CREATE TABLE IF NOT EXISTS `user_game_transaction_details` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '站内用户ID',
  `stat_date` DATE NOT NULL COMMENT '按北京时间归档的统计日期',
  `agent_id` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '第三方代理号',
  `provider_user_id` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '第三方用户ID',
  `login_id` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '第三方登录账号',
  `game_provider_code` INT NOT NULL COMMENT '第三方厂商编号',
  `game_code` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '游戏编码',
  `external_id` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '第三方流水ID',
  `round_id` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '局号',
  `type` INT NOT NULL DEFAULT 0 COMMENT '第三方流水类型',
  `start_date` DATETIME NULL COMMENT '开始时间(UTC)',
  `end_date` DATETIME NULL COMMENT '结束时间(UTC)',
  `start_balance` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '期初余额',
  `end_balance` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '期末余额',
  `deposit` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '存款额',
  `turnover` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '有效投注/打码量',
  `bet` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '总投注',
  `win` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '总派彩',
  `winlose` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '输赢',
  `jp_share` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT 'JP分摊',
  `jp_win` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT 'JP中奖',
  `remark` TEXT NULL COMMENT '备注',
  `record_hash` VARCHAR(64) NOT NULL COMMENT '幂等去重哈希',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_record_hash` (`record_hash`),
  KEY `idx_user_stat_date_start` (`user_id`, `stat_date`, `start_date`),
  KEY `idx_login_stat_date` (`login_id`, `stat_date`),
  KEY `idx_provider_external` (`game_provider_code`, `external_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='玩家第三方游戏逐笔流水表';

-- 执行命令: mysql -u root -p tols < be/migrations/migration_0207.sql
