-- ============================================
-- Migration: migration_0217.sql
-- Description: add user last login tracking fields
-- ============================================

SET @add_last_login_at = (
  SELECT IF(
    EXISTS(
      SELECT 1
      FROM INFORMATION_SCHEMA.COLUMNS
      WHERE TABLE_SCHEMA = DATABASE()
        AND TABLE_NAME = 'users'
        AND COLUMN_NAME = 'last_login_at'
    ),
    'SELECT 1',
    'ALTER TABLE users ADD COLUMN last_login_at DATETIME NULL AFTER total_withdraw'
  )
);
PREPARE stmt FROM @add_last_login_at;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @add_last_login_ip = (
  SELECT IF(
    EXISTS(
      SELECT 1
      FROM INFORMATION_SCHEMA.COLUMNS
      WHERE TABLE_SCHEMA = DATABASE()
        AND TABLE_NAME = 'users'
        AND COLUMN_NAME = 'last_login_ip'
    ),
    'SELECT 1',
    'ALTER TABLE users ADD COLUMN last_login_ip VARCHAR(64) NOT NULL DEFAULT '''' AFTER last_login_at'
  )
);
PREPARE stmt FROM @add_last_login_ip;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
