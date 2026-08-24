-- ============================================
-- 迁移脚本: migration_0208.sql
-- 说明: 新增玩家第三方日汇总表，用于落库 TransactionSummaryByPlayer 报表
-- ============================================

CREATE TABLE IF NOT EXISTS `user_game_transaction_summaries` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '站内用户ID',
  `stat_date` DATE NOT NULL COMMENT '统计日期',
  `login_id` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '第三方登录账号',
  `count` INT NOT NULL DEFAULT 0 COMMENT '注单数',
  `turnover` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '打码量/有效投注',
  `bet` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '下注流水',
  `win` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '派彩',
  `winlose` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '输赢',
  `jp_share` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT 'JP分摊',
  `jp_win` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT 'JP中奖',
  `synced_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '最近同步时间',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_stat_date` (`user_id`, `stat_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='玩家第三方日汇总表';

-- 执行命令: mysql -u root -p tols < be/migrations/migration_0208.sql
