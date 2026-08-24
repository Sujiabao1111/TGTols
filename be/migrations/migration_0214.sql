-- ============================================
-- Migration: migration_0214.sql
-- Description: add daily deposit and game metrics columns to stats_retention
-- ============================================

SET @deposit_users_exists := (
  SELECT COUNT(*)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'stats_retention'
    AND COLUMN_NAME = 'deposit_users'
);

SET @add_deposit_users_sql := IF(
  @deposit_users_exists = 0,
  'ALTER TABLE `stats_retention`
     ADD COLUMN `deposit_users` INT NOT NULL DEFAULT 0 COMMENT ''daily deposit users'' AFTER `new_users`',
  'SELECT "deposit_users already exists"'
);

PREPARE stmt FROM @add_deposit_users_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @deposit_amount_exists := (
  SELECT COUNT(*)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'stats_retention'
    AND COLUMN_NAME = 'deposit_amount'
);

SET @add_deposit_amount_sql := IF(
  @deposit_amount_exists = 0,
  'ALTER TABLE `stats_retention`
     ADD COLUMN `deposit_amount` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT ''daily deposit amount'' AFTER `deposit_users`',
  'SELECT "deposit_amount already exists"'
);

PREPARE stmt FROM @add_deposit_amount_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @game_winlose_users_exists := (
  SELECT COUNT(*)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'stats_retention'
    AND COLUMN_NAME = 'game_winlose_users'
);

SET @add_game_winlose_users_sql := IF(
  @game_winlose_users_exists = 0,
  'ALTER TABLE `stats_retention`
     ADD COLUMN `game_winlose_users` INT NOT NULL DEFAULT 0 COMMENT ''daily players with non-zero winlose'' AFTER `deposit_amount`',
  'SELECT "game_winlose_users already exists"'
);

PREPARE stmt FROM @add_game_winlose_users_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
