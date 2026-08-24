-- 迁移脚本: migration_0209.sql
-- 说明: 新增后台奖励提现打码锁表，桌面奖励/发放奖励需完成 3 倍打码量后方可提现

CREATE TABLE IF NOT EXISTS `user_wager_locks` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `source_type` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '来源类型: desktop_reward/manual_reward',
  `reference_id` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '关联流水参考ID',
  `reward_amount_u` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '奖励金额(U)',
  `wager_multiplier` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '打码倍数',
  `wager_required` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '所需打码量',
  `granted_at` DATETIME NOT NULL COMMENT '奖励发放时间',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_wager_locks_reference_id` (`reference_id`),
  KEY `idx_user_wager_locks_user_id` (`user_id`),
  KEY `idx_user_wager_locks_source_type` (`source_type`),
  KEY `idx_user_wager_locks_granted_at` (`granted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户提现打码锁表';

-- 执行命令: mysql -u root -p tols < be/migrations/migration_0209.sql
