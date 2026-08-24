-- ============================================
-- 迁移脚本: migration_0212.sql
-- 说明: 为统一游戏交易统计表增加金额单位标记，避免 U / IDR / PHP 混用
-- ============================================

SET @amount_unit_exists := (
  SELECT COUNT(*)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'user_game_transaction_stats'
    AND COLUMN_NAME = 'amount_unit'
);

SET @add_amount_unit_sql := IF(
  @amount_unit_exists = 0,
  'ALTER TABLE `user_game_transaction_stats` ADD COLUMN `amount_unit` VARCHAR(8) NOT NULL DEFAULT "" COMMENT "金额单位: U / legacy-empty" AFTER `login_id`',
  'SELECT "amount_unit already exists"'
);

PREPARE stmt FROM @add_amount_unit_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
