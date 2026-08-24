-- ============================================
-- 迁移脚本: migration_0205.sql
-- 说明: 2026-05-20 新增 Add Desktop / World Cup 活动领取记录表
-- ============================================

CREATE TABLE IF NOT EXISTS `add_desktop_reward_claims` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `claimed_at` DATETIME NOT NULL COMMENT '领取时间',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_add_desktop_user` (`user_id`),
  KEY `idx_add_desktop_claimed_at` (`claimed_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Add Desktop 活动领取记录';

CREATE TABLE IF NOT EXISTS `world_cup_reward_claims` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `claimed_at` DATETIME NOT NULL COMMENT '领取时间',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_world_cup_user` (`user_id`),
  KEY `idx_world_cup_claimed_at` (`claimed_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='World Cup 活动领取记录';

-- 执行命令: mysql -u root -p tols < be/migrations/migration_0205.sql
